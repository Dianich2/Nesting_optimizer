package handler

import (
	"server_nesting_optimizer/internal/transport/http/dto"
	httperror "server_nesting_optimizer/internal/transport/http/errors"
	"server_nesting_optimizer/internal/transport/http/mapper"
	placementusecase "server_nesting_optimizer/internal/usecase/placement"

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

	var placementReq dto.CreatePlacementRequest
	err = parseBody(c, &placementReq)
	if err != nil {
		return err
	}

	inputCreatePlacementUseCase := mapper.ToCreatePlacementInput(
		placementReq,
		userID,
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
	projectID, err := getIDFromPath(c, "project_id")
	if err != nil {
		return err
	}

	placementID, err := getIDFromPath(c, "placement_id")
	if err != nil {
		return err
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	inputGetPlacementByIDUseCase := mapper.ToGetPlacementByIDInput(
		userID,
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

	inputListPlacementsUseCase := mapper.ToListPlacementsInput(
		userID,
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
	projectID, err := getIDFromPath(c, "project_id")
	if err != nil {
		return err
	}

	placementID, err := getIDFromPath(c, "placement_id")
	if err != nil {
		return err
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var placementReq dto.UpdatePlacementRequest
	err = parseBody(c, &placementReq)
	if err != nil {
		return err
	}

	inputUpdatePlacementUseCase := mapper.ToUpdatePlacementInput(
		placementReq,
		userID,
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
	projectID, err := getIDFromPath(c, "project_id")
	if err != nil {
		return err
	}

	placementID, err := getIDFromPath(c, "placement_id")
	if err != nil {
		return err
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	inputDeletePlacementUseCase := mapper.ToDeletePlacementInput(
		userID,
		projectID,
		placementID,
	)

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
