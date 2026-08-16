package handler

import (
	"errors"
	"server_nesting_optimizer/internal/transport/http/dto"
	httperror "server_nesting_optimizer/internal/transport/http/errors"
	"server_nesting_optimizer/internal/transport/http/mapper"
	"server_nesting_optimizer/internal/transport/http/middleware"
	surfaceusecase "server_nesting_optimizer/internal/usecase/surface"
	"server_nesting_optimizer/pkg/apperror"

	"github.com/gofiber/fiber/v3"
)

type SurfaceHandler struct {
	createSurfaceUseCase *surfaceusecase.CreateSurfaceUseCase
}

func NewSurfaceHandler(
	createSurfaceUseCase *surfaceusecase.CreateSurfaceUseCase,
) *SurfaceHandler {
	return &SurfaceHandler{
		createSurfaceUseCase: createSurfaceUseCase,
	}
}

// godoc: CreateSurface godoc
// @Summary Create Surface
// @Description Create Surface
// @Accept json
// @Produce json
// @Param surface body dto.CreateSurfaceRequest true "Surface"
// @Success 201 {object} dto.CreateSurfaceResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource Surfaces
// @Security BearerAuth
// @Router /api/v1/surfaces [post]
func (h *SurfaceHandler) Create(c fiber.Ctx) error {
	var surfaceReq dto.CreateSurfaceRequest
	err := c.Bind().Body(&surfaceReq)
	if err != nil {
		return httperror.Handle(
			c,
			apperror.Validation("invalid request body"),
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

	inputCreateSurfaceUseCase := mapper.ToCreateSurfaceInput(surfaceReq, u)

	outputCreateSurfaceUseCase, err := h.createSurfaceUseCase.Execute(
		c.Context(),
		inputCreateSurfaceUseCase,
	)
	if err != nil {
		return httperror.Handle(
			c,
			err,
		)
	}

	createSurfaceResp := mapper.ToCreateSurfaceResponse(outputCreateSurfaceUseCase)

	return c.Status(fiber.StatusCreated).JSON(createSurfaceResp)
}
