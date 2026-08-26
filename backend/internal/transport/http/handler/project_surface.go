package handler

import (
	"server_nesting_optimizer/internal/transport/http/dto"
	httperror "server_nesting_optimizer/internal/transport/http/errors"
	"server_nesting_optimizer/internal/transport/http/mapper"
	projectsurfaceusecase "server_nesting_optimizer/internal/usecase/project_surface"

	"github.com/gofiber/fiber/v3"
)

type ProjectSurfaceHandler struct {
	createProjectSurfaceUseCase  *projectsurfaceusecase.CreateProjectSurfaceUseCase
	getProjectSurfaceByIDUseCase *projectsurfaceusecase.GetProjectSurfaceByIDUseCase
	listProjectSurfacesUseCase   *projectsurfaceusecase.ListProjectSurfacesUseCase
	updateProjectSurfaceUseCase  *projectsurfaceusecase.UpdateProjectSurfaceUseCase
	deleteProjectSurfaceUseCase  *projectsurfaceusecase.DeleteProjectSurfaceUseCase
}

func NewProjectSurfaceHandler(
	createProjectSurfaceUseCase *projectsurfaceusecase.CreateProjectSurfaceUseCase,
	getProjectSurfaceByIDUseCase *projectsurfaceusecase.GetProjectSurfaceByIDUseCase,
	listProjectSurfacesUseCase *projectsurfaceusecase.ListProjectSurfacesUseCase,
	updateProjectSurfaceUseCase *projectsurfaceusecase.UpdateProjectSurfaceUseCase,
	deleteProjectSurfaceUseCase *projectsurfaceusecase.DeleteProjectSurfaceUseCase,
) *ProjectSurfaceHandler {
	return &ProjectSurfaceHandler{
		createProjectSurfaceUseCase:  createProjectSurfaceUseCase,
		getProjectSurfaceByIDUseCase: getProjectSurfaceByIDUseCase,
		listProjectSurfacesUseCase:   listProjectSurfacesUseCase,
		updateProjectSurfaceUseCase:  updateProjectSurfaceUseCase,
		deleteProjectSurfaceUseCase:  deleteProjectSurfaceUseCase,
	}
}

// godoc: CreateProjectSurface godoc
// @Summary Create ProjectSurface
// @Description Create ProjectSurface
// @Accept json
// @Produce json
// @Param project_id path int true "Project ID"
// @Param projectSurface body dto.CreateProjectSurfaceRequest true "ProjectSurface"
// @Success 201 {object} dto.CreateProjectSurfaceResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource ProjectSurfaces
// @Security BearerAuth
// @Router /api/v1/projects/{project_id}/surfaces [post]
func (h *ProjectSurfaceHandler) Create(c fiber.Ctx) error {
	projectID, err := getIDFromPath(c, "project_id")
	if err != nil {
		return err
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var projectSurfaceReq dto.CreateProjectSurfaceRequest
	err = parseBody(c, &projectSurfaceReq)
	if err != nil {
		return err
	}

	inputCreateProjectSurfaceUseCase := mapper.ToCreateProjectSurfaceInput(
		projectSurfaceReq,
		userID,
		projectID,
	)

	outputCreateProjectSurfaceUseCase, err := h.createProjectSurfaceUseCase.Execute(
		c.Context(),
		inputCreateProjectSurfaceUseCase,
	)
	if err != nil {
		return httperror.Handle(
			c,
			err,
		)
	}

	createProjectSurfaceResp := mapper.ToCreateProjectSurfaceResponse(outputCreateProjectSurfaceUseCase)

	return c.Status(fiber.StatusCreated).JSON(createProjectSurfaceResp)
}

// godoc: GetProjectSurfaceByID godoc
// @Summary Get ProjectSurface by ID
// @Description Get ProjectSurface by ID
// @Produce json
// @Param project_id path int true "Project ID"
// @Param id path int true "ProjectSurface ID"
// @Success 200 {object} dto.GetProjectSurfaceByIDResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource ProjectSurfaces
// @Security BearerAuth
// @Router /api/v1/projects/{project_id}/surfaces/{id} [get]
func (h *ProjectSurfaceHandler) GetByID(c fiber.Ctx) error {
	id, err := getIDFromPath(c, "id")
	if err != nil {
		return err
	}

	projectID, err := getIDFromPath(c, "project_id")
	if err != nil {
		return err
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	inputGetProjectSurfaceByIDUseCase := mapper.ToGetProjectSurfaceByIDInput(
		id,
		userID,
		projectID,
	)

	outputGetProjectSurfaceByIDUseCase, err := h.getProjectSurfaceByIDUseCase.Execute(
		c.Context(),
		inputGetProjectSurfaceByIDUseCase,
	)
	if err != nil {
		return httperror.Handle(
			c,
			err,
		)
	}

	getProjectSurfaceByIDResp := mapper.ToGetProjectSurfaceByIDResponse(outputGetProjectSurfaceByIDUseCase)

	return c.Status(fiber.StatusOK).JSON(getProjectSurfaceByIDResp)
}

// godoc: ListProjectSurfaces godoc
// @Summary List Project Surfaces
// @Description List Project Surfaces
// @Produce json
// @Param page query int false "Page" default(1)
// @Param page_size query int false "Page Size" default(20)
// @Param project_id path int true "Project ID"
// @Success 200 {object} dto.ListProjectSurfacesResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource ProjectSurfaces
// @Security BearerAuth
// @Router /api/v1/projects/{project_id}/surfaces [get]
func (h *ProjectSurfaceHandler) ListProjectSurfaces(c fiber.Ctx) error {
	projectID, err := getIDFromPath(c, "project_id")
	if err != nil {
		return err
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	page, pageSize, err := parsePagination(c)
	if err != nil {
		return err
	}

	inputListProjectSurfacesUseCase := mapper.ToListProjectSurfacesInput(
		userID,
		projectID,
		page,
		pageSize,
	)

	outputListProjectSurfacesUseCase, err := h.listProjectSurfacesUseCase.Execute(
		c.Context(),
		inputListProjectSurfacesUseCase,
	)
	if err != nil {
		return httperror.Handle(
			c,
			err,
		)
	}

	listProjectSurfacesResp := mapper.ToListProjectSurfacesResponse(outputListProjectSurfacesUseCase)

	return c.Status(fiber.StatusOK).JSON(listProjectSurfacesResp)
}

// godoc: UpdateProjectSurface godoc
// @Summary Update Project Surface
// @Description Update Project Surface
// @Accept json
// @Produce json
// @Param id path int true "Project Surface ID"
// @Param project_id path int true "Project ID"
// @Param surface body dto.UpdateProjectSurfaceRequest true "Project Surface"
// @Success 200 {object} dto.UpdateProjectSurfaceResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource ProjectSurfaces
// @Security BearerAuth
// @Router /api/v1/projects/{project_id}/surfaces/{id} [patch]
func (h *ProjectSurfaceHandler) Update(c fiber.Ctx) error {
	id, err := getIDFromPath(c, "id")
	if err != nil {
		return err
	}

	projectID, err := getIDFromPath(c, "project_id")
	if err != nil {
		return err
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var projectSurfaceReq dto.UpdateProjectSurfaceRequest
	err = parseBody(c, &projectSurfaceReq)
	if err != nil {
		return err
	}

	inputUpdateProjectSurfaceUseCase := mapper.ToUpdateProjectSurfaceInput(
		projectSurfaceReq,
		userID,
		projectID,
		id,
	)

	outputUpdateProjectSurfaceUseCase, err := h.updateProjectSurfaceUseCase.Execute(
		c.Context(),
		inputUpdateProjectSurfaceUseCase,
	)
	if err != nil {
		return httperror.Handle(
			c,
			err,
		)
	}

	updateProjectSurfaceResp := mapper.ToUpdateProjectSurfaceResponse(outputUpdateProjectSurfaceUseCase)

	return c.Status(fiber.StatusOK).JSON(updateProjectSurfaceResp)
}

// godoc: DeleteProjectSurface godoc
// @Summary Delete Project Surface
// @Description Delete Project Surface
// @Param project_id path int true "Project ID"
// @Param id path int true "Project Surface ID"
// @Success 204 "No Content"
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource ProjectSurfaces
// @Security BearerAuth
// @Router /api/v1/projects/{project_id}/surfaces/{id} [delete]
func (h *ProjectSurfaceHandler) Delete(c fiber.Ctx) error {
	id, err := getIDFromPath(c, "id")
	if err != nil {
		return err
	}

	projectID, err := getIDFromPath(c, "project_id")
	if err != nil {
		return err
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	inputDeleteProjectSurfaceUseCase := mapper.ToDeleteProjectSurfaceInput(
		userID,
		projectID,
		id,
	)

	err = h.deleteProjectSurfaceUseCase.Execute(
		c.Context(),
		inputDeleteProjectSurfaceUseCase,
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
