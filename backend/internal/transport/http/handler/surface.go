package handler

import (
	"errors"
	"server_nesting_optimizer/internal/transport/http/dto"
	httperror "server_nesting_optimizer/internal/transport/http/errors"
	"server_nesting_optimizer/internal/transport/http/mapper"
	"server_nesting_optimizer/internal/transport/http/middleware"
	surfaceusecase "server_nesting_optimizer/internal/usecase/surface"
	"server_nesting_optimizer/pkg/apperror"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type SurfaceHandler struct {
	createSurfaceUseCase  *surfaceusecase.CreateSurfaceUseCase
	getSurfaceByIDUseCase *surfaceusecase.GetSurfaceByIDUseCase
	listSurfacesUseCase   *surfaceusecase.ListSurfacesUseCase
	updateSurfaceUseCase  *surfaceusecase.UpdateSurfaceUseCase
}

func NewSurfaceHandler(
	createSurfaceUseCase *surfaceusecase.CreateSurfaceUseCase,
	getSurfaceByIDUseCase *surfaceusecase.GetSurfaceByIDUseCase,
	listSurfacesUseCase *surfaceusecase.ListSurfacesUseCase,
	updateSurfaceUseCase *surfaceusecase.UpdateSurfaceUseCase,
) *SurfaceHandler {
	return &SurfaceHandler{
		createSurfaceUseCase:  createSurfaceUseCase,
		getSurfaceByIDUseCase: getSurfaceByIDUseCase,
		listSurfacesUseCase:   listSurfacesUseCase,
		updateSurfaceUseCase:  updateSurfaceUseCase,
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

// godoc: GetSurfaceByID godoc
// @Summary Get Surface by ID
// @Description Get Surface by ID
// @Produce json
// @Param id path int true "Surface ID"
// @Success 200 {object} dto.GetSurfaceResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource Surfaces
// @Security BearerAuth
// @Router /api/v1/surfaces/{id} [get]
func (h *SurfaceHandler) GetByID(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return httperror.Handle(
			c,
			apperror.Validation(
				"invalid id",
				apperror.NewFieldError(
					"id",
					apperror.FieldCodeInvalid,
					"id must be a valid positive integer",
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

	inputGetSurfaceUseCase := mapper.ToGetSurfaceInput(id, u)

	outputGetSurfaceUseCase, err := h.getSurfaceByIDUseCase.Execute(
		c.Context(),
		inputGetSurfaceUseCase,
	)
	if err != nil {
		return httperror.Handle(
			c,
			err,
		)
	}

	getSurfaceResp := mapper.ToGetSurfaceResponse(outputGetSurfaceUseCase)

	return c.Status(fiber.StatusOK).JSON(getSurfaceResp)
}

// godoc: ListSurfaces godoc
// @Summary List Surfaces
// @Description List Surfaces
// @Produce json
// @Param page query int false "Page" default(1)
// @Param page_size query int false "Page Size" default(20)
// @Success 200 {object} dto.ListSurfacesResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource Surfaces
// @Security BearerAuth
// @Router /api/v1/surfaces [get]
func (h *SurfaceHandler) ListSurfaces(c fiber.Ctx) error {
	pageSizeQuery, pageQuery := c.Query("page_size"), c.Query("page")

	if pageSizeQuery == "" {
		pageSizeQuery = "20"
	}

	if pageQuery == "" {
		pageQuery = "1"
	}

	pageSize, err := strconv.Atoi(pageSizeQuery)
	if err != nil {
		return httperror.Handle(
			c,
			apperror.Validation(
				"invalid page size",
				apperror.NewFieldError(
					"page_size",
					apperror.FieldCodeInvalid,
					"page size must be a valid positive integer",
				),
			),
		)
	}

	page, err := strconv.Atoi(pageQuery)
	if err != nil {
		return httperror.Handle(
			c,
			apperror.Validation(
				"invalid page",
				apperror.NewFieldError(
					"page",
					apperror.FieldCodeInvalid,
					"page must be a valid positive integer",
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

	inputListSurfacesUseCase := mapper.ToListSurfacesInput(page, pageSize, u)

	outputListSurfacesUseCase, err := h.listSurfacesUseCase.Execute(
		c.Context(),
		inputListSurfacesUseCase,
	)
	if err != nil {
		return httperror.Handle(
			c,
			err,
		)
	}

	listSurfacesResp := mapper.ToListSurfacesResponse(outputListSurfacesUseCase)

	return c.Status(fiber.StatusOK).JSON(listSurfacesResp)
}

// godoc: UpdateSurface godoc
// @Summary Update Surface
// @Description Update Surface
// @Accept json
// @Produce json
// @Param id path int true "Surface ID"
// @Param surface body dto.UpdateSurfaceRequest true "Surface"
// @Success 200 {object} dto.UpdateSurfaceResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource Surfaces
// @Security BearerAuth
// @Router /api/v1/surfaces/{id} [patch]
func (h *SurfaceHandler) Update(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return httperror.Handle(
			c,
			apperror.Validation(
				"invalid id",
				apperror.NewFieldError(
					"id",
					apperror.FieldCodeInvalid,
					"id must be a valid positive integer",
				),
			),
		)
	}

	var surfaceReq dto.UpdateSurfaceRequest
	err = c.Bind().Body(&surfaceReq)
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

	inputUpdateSurfaceUseCase := mapper.ToUpdateSurfaceInput(surfaceReq, u, id)

	outputUpdateSurfaceUseCase, err := h.updateSurfaceUseCase.Execute(
		c.Context(),
		inputUpdateSurfaceUseCase,
	)
	if err != nil {
		return httperror.Handle(
			c,
			err,
		)
	}

	updateSurfaceResp := mapper.ToUpdateSurfaceResponse(outputUpdateSurfaceUseCase)

	return c.Status(fiber.StatusOK).JSON(updateSurfaceResp)
}
