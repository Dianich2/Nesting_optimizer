package handler

import (
	"server_nesting_optimizer/internal/transport/http/dto"
	httperror "server_nesting_optimizer/internal/transport/http/errors"
	"server_nesting_optimizer/internal/transport/http/mapper"
	nestingusecase "server_nesting_optimizer/internal/usecase/nesting"

	"github.com/gofiber/fiber/v3"
)

type NestingHandler struct {
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
// @Description Run automatic nesting for a project surface
// @Accept json
// @Produce json
// @Param project_id path int true "Project ID"
// @Param project_surface_id path int true "Project Surface ID"
// @Param nestingParams body dto.RunNestingRequest true "Nesting Params"
// @Success 200 {object} dto.RunNestingResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource Nesting
// @Security BearerAuth
// @Router /api/v1/projects/{project_id}/surfaces/{project_surface_id}/nesting [post]
func (h *NestingHandler) Run(c fiber.Ctx) error {
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

	outputRunNestingUseCase, err := h.runNestingUseCase.Execute(
		c.Context(),
		inputRunNestingUseCase,
	)
	if err != nil {
		return httperror.Handle(
			c,
			err,
		)
	}

	runNestingResp := mapper.ToRunNestingResponse(outputRunNestingUseCase)

	return c.Status(fiber.StatusOK).JSON(runNestingResp)
}
