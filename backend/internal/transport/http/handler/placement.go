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
	createPlacementUseCase *placementusecase.CreatePlacementUseCase
}

func NewPlacementHandler(
	createPlacementUseCase *placementusecase.CreatePlacementUseCase,
) *PlacementHandler {
	return &PlacementHandler{
		createPlacementUseCase: createPlacementUseCase,
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
