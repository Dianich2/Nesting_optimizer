package handler

import (
	"server_nesting_optimizer/internal/transport/http/dto"
	httperror "server_nesting_optimizer/internal/transport/http/errors"
	"server_nesting_optimizer/internal/transport/http/mapper"
	surfaceusecase "server_nesting_optimizer/internal/usecase/surface"

	"github.com/gofiber/fiber/v3"
)

type SurfaceHandler struct {
	createSurfaceUseCase  *surfaceusecase.CreateSurfaceUseCase
	getSurfaceByIDUseCase *surfaceusecase.GetSurfaceByIDUseCase
	listSurfacesUseCase   *surfaceusecase.ListSurfacesUseCase
	updateSurfaceUseCase  *surfaceusecase.UpdateSurfaceUseCase
	deleteSurfaceUseCase  *surfaceusecase.DeleteSurfaceUseCase
}

func NewSurfaceHandler(
	createSurfaceUseCase *surfaceusecase.CreateSurfaceUseCase,
	getSurfaceByIDUseCase *surfaceusecase.GetSurfaceByIDUseCase,
	listSurfacesUseCase *surfaceusecase.ListSurfacesUseCase,
	updateSurfaceUseCase *surfaceusecase.UpdateSurfaceUseCase,
	deleteSurfaceUseCase *surfaceusecase.DeleteSurfaceUseCase,
) *SurfaceHandler {
	return &SurfaceHandler{
		createSurfaceUseCase:  createSurfaceUseCase,
		getSurfaceByIDUseCase: getSurfaceByIDUseCase,
		listSurfacesUseCase:   listSurfacesUseCase,
		updateSurfaceUseCase:  updateSurfaceUseCase,
		deleteSurfaceUseCase:  deleteSurfaceUseCase,
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
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var surfaceReq dto.CreateSurfaceRequest
	err = parseBody(c, &surfaceReq)
	if err != nil {
		return err
	}

	inputCreateSurfaceUseCase := mapper.ToCreateSurfaceInput(
		surfaceReq,
		userID,
	)

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
	id, err := getIDFromPath(c, "id")
	if err != nil {
		return err
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	inputGetSurfaceUseCase := mapper.ToGetSurfaceInput(
		id,
		userID,
	)

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
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	page, pageSize, err := parsePagination(c)
	if err != nil {
		return err
	}

	inputListSurfacesUseCase := mapper.ToListSurfacesInput(
		page,
		pageSize,
		userID,
	)

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
	id, err := getIDFromPath(c, "id")
	if err != nil {
		return err
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var surfaceReq dto.UpdateSurfaceRequest
	err = parseBody(c, &surfaceReq)
	if err != nil {
		return err
	}

	inputUpdateSurfaceUseCase := mapper.ToUpdateSurfaceInput(
		surfaceReq,
		userID,
		id,
	)

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

// godoc: DeleteSurface godoc
// @Summary Delete Surface
// @Description Delete Surface
// @Param id path int true "Surface ID"
// @Success 204 "No Content"
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource Surfaces
// @Security BearerAuth
// @Router /api/v1/surfaces/{id} [delete]
func (h *SurfaceHandler) Delete(c fiber.Ctx) error {
	id, err := getIDFromPath(c, "id")
	if err != nil {
		return err
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	inputDeleteSurfaceUseCase := mapper.ToDeleteSurfaceInput(
		id,
		userID,
	)

	err = h.deleteSurfaceUseCase.Execute(
		c.Context(),
		inputDeleteSurfaceUseCase,
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
