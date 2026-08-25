package handler

import (
	"errors"
	"server_nesting_optimizer/internal/transport/http/dto"
	httperror "server_nesting_optimizer/internal/transport/http/errors"
	"server_nesting_optimizer/internal/transport/http/mapper"
	"server_nesting_optimizer/internal/transport/http/middleware"
	placementusecase "server_nesting_optimizer/internal/usecase/placement"
	"server_nesting_optimizer/pkg/apperror"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type PlacementHandler struct {
	createPlacementUseCase  *placementusecase.CreatePlacementUseCase
	getPlacementByIDUseCase *placementusecase.GetPlacementByIDUseCase
	listPlacementsUseCase   *placementusecase.ListPlacementsUseCase
	updatePlacementUseCase  *placementusecase.UpdatePlacementUseCase
	deletePlacementUseCase  *placementusecase.DeletePlacementUseCase
}

func NewPlacementHandler(
	createPlacementUseCase *placementusecase.CreatePlacementUseCase,
	getPlacementByIDUseCase *placementusecase.GetPlacementByIDUseCase,
	listPlacementsUseCase *placementusecase.ListPlacementsUseCase,
	updatePlacementUseCase *placementusecase.UpdatePlacementUseCase,
	deletePlacementUseCase *placementusecase.DeletePlacementUseCase,
) *PlacementHandler {
	return &PlacementHandler{
		createPlacementUseCase:  createPlacementUseCase,
		getPlacementByIDUseCase: getPlacementByIDUseCase,
		listPlacementsUseCase:   listPlacementsUseCase,
		updatePlacementUseCase:  updatePlacementUseCase,
		deletePlacementUseCase:  deletePlacementUseCase,
	}
}

// godoc: CreatePlacement godoc
// @Summary Create Placement
// @Description Create Placement
// @Accept json
// @Produce json
// @Param project_id path int true "Project ID"
// @Param project_surface_id path int true "Project Surface ID"
// @Param placement body dto.CreatePlacementRequest true "Placement"
// @Success 201 {object} dto.CreatePlacementResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 409 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource Placements
// @Security BearerAuth
// @Router /api/v1/projects/{project_id}/surfaces/{project_surface_id}/placements [post]
func (h *PlacementHandler) Create(c fiber.Ctx) error {
	var placementReq dto.CreatePlacementRequest
	err := c.Bind().Body(&placementReq)
	if err != nil {
		return httperror.Handle(
			c,
			apperror.Validation("invalid request body"),
		)
	}

	projectID, err := strconv.ParseInt(c.Params("project_id"), 10, 64)
	if err != nil {
		return httperror.Handle(
			c,
			apperror.Validation(
				"invalid id",
				apperror.NewFieldError(
					"project_id",
					apperror.FieldCodeInvalid,
					"project id must be a valid positive integer",
				),
			),
		)
	}

	projectSurfaceID, err := strconv.ParseInt(c.Params("project_surface_id"), 10, 64)
	if err != nil {
		return httperror.Handle(
			c,
			apperror.Validation(
				"invalid id",
				apperror.NewFieldError(
					"project_surface_id",
					apperror.FieldCodeInvalid,
					"project surface id must be a valid positive integer",
				),
			),
		)
	}

	userID := c.Locals(middleware.UserIDLocalKey)
	u, ok := userID.(int64)
	if !ok {
		return httperror.Handle(
			c,
			apperror.Internal("failed to get user id from request context", errors.New("user_id local is missing or invalid")),
		)
	}

	inputCreatePlacementUseCase := mapper.ToCreatePlacementInput(
		placementReq,
		u,
		projectID,
		projectSurfaceID,
	)

	outputCreatePlacementUseCase, err := h.createPlacementUseCase.Execute(
		c.Context(),
		inputCreatePlacementUseCase,
	)
	if err != nil {
		return httperror.Handle(
			c,
			err,
		)
	}

	createPlacementResp := mapper.ToCreatePlacementResponse(outputCreatePlacementUseCase)

	return c.Status(fiber.StatusCreated).JSON(createPlacementResp)
}

// godoc: GetPlacementByID godoc
// @Summary Get Placement by ID
// @Description Get Placement
// @Produce json
// @Param project_id path int true "Project ID"
// @Param placement_id path int true "Placement ID"
// @Success 200 {object} dto.GetPlacementByIDResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource Placements
// @Security BearerAuth
// @Router /api/v1/projects/{project_id}/placements/{placement_id} [get]
func (h *PlacementHandler) GetByID(c fiber.Ctx) error {
	projectID, err := strconv.ParseInt(c.Params("project_id"), 10, 64)
	if err != nil {
		return httperror.Handle(
			c,
			apperror.Validation(
				"invalid id",
				apperror.NewFieldError(
					"project_id",
					apperror.FieldCodeInvalid,
					"project id must be a valid positive integer",
				),
			),
		)
	}

	placementID, err := strconv.ParseInt(c.Params("placement_id"), 10, 64)
	if err != nil {
		return httperror.Handle(
			c,
			apperror.Validation(
				"invalid id",
				apperror.NewFieldError(
					"placement_id",
					apperror.FieldCodeInvalid,
					"placement id must be a valid positive integer",
				),
			),
		)
	}

	userID := c.Locals(middleware.UserIDLocalKey)
	u, ok := userID.(int64)
	if !ok {
		return httperror.Handle(
			c,
			apperror.Internal("failed to get user id from request context", errors.New("user_id local is missing or invalid")),
		)
	}

	inputGetPlacementByIDUseCase := mapper.ToGetPlacementByIDInput(
		u,
		projectID,
		placementID,
	)

	outputGetPlacementByIDUseCase, err := h.getPlacementByIDUseCase.Execute(
		c.Context(),
		inputGetPlacementByIDUseCase,
	)
	if err != nil {
		return httperror.Handle(
			c,
			err,
		)
	}

	getPlacementByIDResp := mapper.ToGetPlacementByIDResponse(outputGetPlacementByIDUseCase)

	return c.Status(fiber.StatusOK).JSON(getPlacementByIDResp)
}

// godoc: ListPlacements godoc
// @Summary List Placements
// @Description List Placements
// @Produce json
// @Param project_id path int true "Project ID"
// @Param project_surface_id path int true "Project Surface ID"
// @Success 200 {object} dto.ListPlacementsResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource Placements
// @Security BearerAuth
// @Router /api/v1/projects/{project_id}/surfaces/{project_surface_id}/placements [get]
func (h *PlacementHandler) ListPlacements(c fiber.Ctx) error {
	projectID, err := strconv.ParseInt(c.Params("project_id"), 10, 64)
	if err != nil {
		return httperror.Handle(
			c,
			apperror.Validation(
				"invalid id",
				apperror.NewFieldError(
					"project_id",
					apperror.FieldCodeInvalid,
					"project id must be a valid positive integer",
				),
			),
		)
	}

	projectSurfaceID, err := strconv.ParseInt(c.Params("project_surface_id"), 10, 64)
	if err != nil {
		return httperror.Handle(
			c,
			apperror.Validation(
				"invalid id",
				apperror.NewFieldError(
					"project_surface_id",
					apperror.FieldCodeInvalid,
					"project surface id must be a valid positive integer",
				),
			),
		)
	}

	userID := c.Locals(middleware.UserIDLocalKey)
	u, ok := userID.(int64)
	if !ok {
		return httperror.Handle(
			c,
			apperror.Internal("failed to get user id from request context", errors.New("user_id local is missing or invalid")),
		)
	}

	inputListPlacementsUseCase := mapper.ToListPlacementsInput(
		u,
		projectID,
		projectSurfaceID,
	)

	outputListPlacementsUseCase, err := h.listPlacementsUseCase.Execute(
		c.Context(),
		inputListPlacementsUseCase,
	)
	if err != nil {
		return httperror.Handle(
			c,
			err,
		)
	}

	listPlacementsResp := mapper.ToListPlacementsResponse(outputListPlacementsUseCase)

	return c.Status(fiber.StatusOK).JSON(listPlacementsResp)
}

// godoc: UpdatePlacement godoc
// @Summary Update Placement
// @Description Update Placement
// @Accept json
// @Produce json
// @Param project_id path int true "Project ID"
// @Param placement_id path int true "Placement ID"
// @Param placement body dto.UpdatePlacementRequest true "Placement"
// @Success 200 {object} dto.UpdatePlacementResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 409 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource Placements
// @Security BearerAuth
// @Router /api/v1/projects/{project_id}/placements/{placement_id} [patch]
func (h *PlacementHandler) Update(c fiber.Ctx) error {
	var placementReq dto.UpdatePlacementRequest
	err := c.Bind().Body(&placementReq)
	if err != nil {
		return httperror.Handle(
			c,
			apperror.Validation("invalid request body"),
		)
	}

	projectID, err := strconv.ParseInt(c.Params("project_id"), 10, 64)
	if err != nil {
		return httperror.Handle(
			c,
			apperror.Validation(
				"invalid id",
				apperror.NewFieldError(
					"project_id",
					apperror.FieldCodeInvalid,
					"project id must be a valid positive integer",
				),
			),
		)
	}

	placementID, err := strconv.ParseInt(c.Params("placement_id"), 10, 64)
	if err != nil {
		return httperror.Handle(
			c,
			apperror.Validation(
				"invalid id",
				apperror.NewFieldError(
					"placement_id",
					apperror.FieldCodeInvalid,
					"placement id must be a valid positive integer",
				),
			),
		)
	}

	userID := c.Locals(middleware.UserIDLocalKey)
	u, ok := userID.(int64)
	if !ok {
		return httperror.Handle(
			c,
			apperror.Internal("failed to get user id from request context", errors.New("user_id local is missing or invalid")),
		)
	}

	inputUpdatePlacementUseCase := mapper.ToUpdatePlacementInput(
		placementReq,
		u,
		projectID,
		placementID,
	)

	outputUpdatePlacementUseCase, err := h.updatePlacementUseCase.Execute(
		c.Context(),
		inputUpdatePlacementUseCase,
	)
	if err != nil {
		return httperror.Handle(
			c,
			err,
		)
	}

	updatePlacementResp := mapper.ToUpdatePlacementResponse(outputUpdatePlacementUseCase)

	return c.Status(fiber.StatusOK).JSON(updatePlacementResp)
}

// godoc: DeletePlacement godoc
// @Summary Delete Placement
// @Description Delete Placement
// @Param project_id path int true "Project ID"
// @Param placement_id path int true "Placement ID"
// @Success 204 "No Content"
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource Placements
// @Security BearerAuth
// @Router /api/v1/projects/{project_id}/placements/{placement_id} [delete]
func (h *PlacementHandler) Delete(c fiber.Ctx) error {
	placementID, err := strconv.ParseInt(c.Params("placement_id"), 10, 64)
	if err != nil {
		return httperror.Handle(
			c,
			apperror.Validation(
				"invalid id",
				apperror.NewFieldError(
					"placement_id",
					apperror.FieldCodeInvalid,
					"placement id must be a valid positive integer",
				),
			),
		)
	}

	projectID, err := strconv.ParseInt(c.Params("project_id"), 10, 64)
	if err != nil {
		return httperror.Handle(
			c,
			apperror.Validation(
				"invalid project id",
				apperror.NewFieldError(
					"project_id",
					apperror.FieldCodeInvalid,
					"project_id must be a valid positive integer",
				),
			),
		)
	}

	userID := c.Locals(middleware.UserIDLocalKey)
	u, ok := userID.(int64)
	if !ok {
		return httperror.Handle(
			c,
			apperror.Internal("failed to get user id from request context", errors.New("user_id local is missing or invalid")),
		)
	}

	inputDeletePlacementUseCase := mapper.ToDeletePlacementInput(u, projectID, placementID)

	err = h.deletePlacementUseCase.Execute(
		c.Context(),
		inputDeletePlacementUseCase,
	)
	if err != nil {
		return httperror.Handle(
			c,
			err,
		)
	}

	c.Status(fiber.StatusNoContent)
	return nil
}
