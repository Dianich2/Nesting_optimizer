package handler

import (
	"server_nesting_optimizer/internal/transport/http/dto"
	httperror "server_nesting_optimizer/internal/transport/http/errors"
	"server_nesting_optimizer/internal/transport/http/mapper"
	nestingusecase "server_nesting_optimizer/internal/usecase/nesting"

	"github.com/gofiber/fiber/v3"
	"server_nesting_optimizer/pkg/apperror"
)

type NestingHandler struct {
	executions        nestingExecutions
	runNestingUseCase *nestingusecase.RunNestingUseCase
}

func NewNestingHandler(
	runNestingUseCase *nestingusecase.RunNestingUseCase,
) *NestingHandler {
	return &NestingHandler{
		runNestingUseCase: runNestingUseCase,
	}
}

// godoc: RunNesting godoc
// @Summary Run automatic nesting
// @Description Run automatic nesting, or save the exact server-owned preview by including preview_id and its original parameters. Writes are atomic and reject stale source snapshots.
// @Accept json
// @Produce json
// @Param project_id path int true "Project ID"
// @Param project_surface_id path int true "Project Surface ID"
// @Param Idempotency-Key header string false "UUID for safe retry"
// @Failure 409 {object} httperror.ErrorResponse
// @Failure 408 {object} httperror.ErrorResponse
// @Param nestingParams body dto.RunNestingRequest true "Nesting Params"
// @Success 200 {object} dto.RunNestingResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Failure 503 {object} httperror.ErrorResponse
// @Resource Nesting
// @Security BearerAuth
// @Router /api/v1/projects/{project_id}/surfaces/{project_surface_id}/nesting [post]
func (h *NestingHandler) Run(c fiber.Ctx) error {
	return h.run(c, false)
}

// @Summary Calculate a nesting preview without writing placements
// @Description Returns a server-owned preview_id valid for 15 minutes. Save it using POST /nesting with the same parameters and preview_id. Retention is bounded per user.
// @Accept json
// @Produce json
// @Param project_id path int true "Project ID"
// @Param project_surface_id path int true "Project surface ID"
// @Param Idempotency-Key header string false "Execution UUID for cancellation"
// @Param nestingParams body dto.RunNestingRequest true "Nesting Params"
// @Success 200 {object} dto.RunNestingResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 408 {object} httperror.ErrorResponse
// @Failure 409 {object} httperror.ErrorResponse
// @Failure 503 {object} httperror.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/projects/{project_id}/surfaces/{project_surface_id}/nesting/preview [post]
func (h *NestingHandler) Preview(c fiber.Ctx) error {
	return h.run(c, true)
}

func (h *NestingHandler) run(c fiber.Ctx, preview bool) error {
	projectID, err := getIDFromPath(c, "project_id")
	if err != nil {
		return err
	}

	projectSurfaceID, err := getIDFromPath(c, "project_surface_id")
	if err != nil {
		return err
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var req dto.RunNestingRequest
	err = parseBody(c, &req)
	if err != nil {
		return err
	}

	inputRunNestingUseCase := mapper.ToRunNestingInput(
		req,
		userID,
		projectID,
		projectSurfaceID,
	)

	key, err := parseExecutionKey(c.Get("Idempotency-Key"))
	if err != nil {
		return err
	}
	executionKey := nestingExecution{userID: userID, projectID: projectID, surfaceID: projectSurfaceID, key: key, preview: preview}
	ctx, finish := h.executions.begin(c.Context(), executionKey)
	defer finish()
	var outputRunNestingUseCase nestingusecase.RunNestingOutput
	if preview && key != "" {
		outputRunNestingUseCase, err = h.runNestingUseCase.PreviewWithKey(ctx, inputRunNestingUseCase, key)
	} else if preview {
		outputRunNestingUseCase, err = h.runNestingUseCase.Preview(ctx, inputRunNestingUseCase)
	} else {
		outputRunNestingUseCase, err = h.runNestingUseCase.ExecuteWithKey(ctx, inputRunNestingUseCase, key)
	}
	if err != nil {
		if ctx.Err() != nil {
			return nestingResponseError(c, ctx, err)
		}
		return httperror.Handle(c, err)
	}

	runNestingResp := mapper.ToRunNestingResponse(outputRunNestingUseCase)

	return c.Status(fiber.StatusOK).JSON(runNestingResp)
}

// @Summary Read nesting operation status
// @Produce json
// @Param project_id path int true "Project ID"
// @Param project_surface_id path int true "Project surface ID"
// @Param kind query string false "Operation type: save (default) or preview"
// @Param operation_key path string true "Idempotency UUID"
// @Success 200 {object} dto.NestingOperationResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/projects/{project_id}/surfaces/{project_surface_id}/nesting/operations/{operation_key} [get]
func (h *NestingHandler) GetOperation(c fiber.Ctx) error {
	projectID, err := getIDFromPath(c, "project_id")
	if err != nil {
		return err
	}
	surfaceID, err := getIDFromPath(c, "project_surface_id")
	if err != nil {
		return err
	}
	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	read := h.runNestingUseCase.GetOperation
	if c.Query("kind") == "preview" {
		read = h.runNestingUseCase.GetPreviewOperation
	} else if c.Query("kind") != "" && c.Query("kind") != "save" {
		return apperror.Validation("invalid operation kind")
	}
	status, err := read(c.Context(), userID, projectID, surfaceID, c.Params("operation_key"))
	if err != nil {
		return err
	}
	response := dto.NestingOperationResponse{Status: status.Status}
	if status.Failure != nil {
		response.Failure = &dto.NestingOperationFailure{Code: string(status.Failure.Code), Message: status.Failure.Message}
	}
	if status.Result != nil {
		result := mapper.ToRunNestingResponse(*status.Result)
		response.Result = &result
	}
	return c.JSON(response)
}

// @Summary List saved nesting runs
// @Produce json
// @Param project_id path int true "Project ID"
// @Param project_surface_id path int true "Project surface ID"
// @Param page query int false "Page"
// @Param page_size query int false "Page size"
// @Success 200 {object} dto.NestingHistoryResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/projects/{project_id}/surfaces/{project_surface_id}/nesting/history [get]
func (h *NestingHandler) History(c fiber.Ctx) error {
	projectID, err := getIDFromPath(c, "project_id")
	if err != nil {
		return err
	}
	surfaceID, err := getIDFromPath(c, "project_surface_id")
	if err != nil {
		return err
	}
	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	page, size, err := parsePagination(c)
	if err != nil {
		return err
	}
	runs, hasNext, err := h.runNestingUseCase.ListHistory(c.Context(), userID, projectID, surfaceID, page, size)
	if err != nil {
		return httperror.Handle(c, err)
	}
	response := dto.NestingHistoryResponse{Items: make([]dto.NestingHistoryItem, 0, len(runs)), HasNext: hasNext}
	for _, run := range runs {
		response.Items = append(response.Items, dto.NestingHistoryItem{
			Parameters: run.Parameters,
			ID:         run.ID, Algorithm: string(run.Algorithm), KeepExisting: run.KeepExisting,
			RequestedCount: run.RequestedCount, PlacedCount: run.PlacedCount, SurfaceArea: run.SurfaceArea,
			PlacedArea: run.PlacedArea, Utilization: run.Utilization, DurationMS: run.Duration.Milliseconds(), CreatedAt: run.CreatedAt,
		})
	}
	return c.JSON(response)
}
