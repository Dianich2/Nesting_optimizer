package handler

import (
	"errors"
	"server_nesting_optimizer/internal/transport/http/dto"
	httperror "server_nesting_optimizer/internal/transport/http/errors"
	"server_nesting_optimizer/internal/transport/http/mapper"
	"server_nesting_optimizer/internal/transport/http/middleware"
	projectusecase "server_nesting_optimizer/internal/usecase/project"
	"server_nesting_optimizer/pkg/apperror"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type ProjectHandler struct {
	createProjectUseCase  *projectusecase.CreateProjectUseCase
	getProjectByIDUseCase *projectusecase.GetProjectByIDUseCase
	listProjectsUseCase   *projectusecase.ListProjectsUseCase
	updateProjectUseCase  *projectusecase.UpdateProjectUseCase
	deleteProjectUseCase  *projectusecase.DeleteProjectUseCase
}

func NewProjectHandler(
	createProjectUseCase *projectusecase.CreateProjectUseCase,
	getProjectByIDUseCase *projectusecase.GetProjectByIDUseCase,
	listProjectsUseCase *projectusecase.ListProjectsUseCase,
	updateProjectUseCase *projectusecase.UpdateProjectUseCase,
	deleteProjectUseCase *projectusecase.DeleteProjectUseCase,
) *ProjectHandler {
	return &ProjectHandler{
		createProjectUseCase:  createProjectUseCase,
		getProjectByIDUseCase: getProjectByIDUseCase,
		listProjectsUseCase:   listProjectsUseCase,
		updateProjectUseCase:  updateProjectUseCase,
		deleteProjectUseCase:  deleteProjectUseCase,
	}
}

// godoc: CreateProject godoc
// @Summary Create Project
// @Description Create Project
// @Accept json
// @Produce json
// @Param project body dto.CreateProjectRequest true "Project"
// @Success 201 {object} dto.CreateProjectResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource Projects
// @Security BearerAuth
// @Router /api/v1/projects [post]
func (h *ProjectHandler) Create(c fiber.Ctx) error {
	var projectReq dto.CreateProjectRequest
	err := c.Bind().Body(&projectReq)
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

	inputCreateProjectUseCase := mapper.ToCreateProjectInput(projectReq, u)

	outputCreateProjectUseCase, err := h.createProjectUseCase.Execute(
		c.Context(),
		inputCreateProjectUseCase,
	)
	if err != nil {
		return httperror.Handle(
			c,
			err,
		)
	}

	createProjectResp := mapper.ToCreateProjectResponse(outputCreateProjectUseCase)

	return c.Status(fiber.StatusCreated).JSON(createProjectResp)
}

// godoc: GetProjectByID godoc
// @Summary Get Project By ID
// @Description Get Project By ID
// @Produce json
// @Param id path int true "Project ID"
// @Success 200 {object} dto.GetProjectByIDResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource Projects
// @Security BearerAuth
// @Router /api/v1/projects/{id} [get]
func (h *ProjectHandler) GetByID(c fiber.Ctx) error {
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

	inputGetProjectByIDUseCase := mapper.ToGetProjectByIDInput(id, u)

	outputGetProjectByIDUseCase, err := h.getProjectByIDUseCase.Execute(
		c.Context(),
		inputGetProjectByIDUseCase,
	)
	if err != nil {
		return httperror.Handle(
			c,
			err,
		)
	}

	getProjectByIDResp := mapper.ToGetProjectByIDResponse(outputGetProjectByIDUseCase)

	return c.Status(fiber.StatusOK).JSON(getProjectByIDResp)
}

// godoc: ListProjects godoc
// @Summary List Projects
// @Description List Projects
// @Produce json
// @Param page query int false "Page" default(1)
// @Param page_size query int false "Page Size" default(20)
// @Success 200 {object} dto.ListProjectsResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource Projects
// @Security BearerAuth
// @Router /api/v1/projects [get]
func (h *ProjectHandler) ListProjects(c fiber.Ctx) error {
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

	inputListProjectsUseCase := mapper.ToListProjectsInput(u, page, pageSize)

	outputListProjectsUseCase, err := h.listProjectsUseCase.Execute(
		c.Context(),
		inputListProjectsUseCase,
	)
	if err != nil {
		return httperror.Handle(
			c,
			err,
		)
	}

	listProjectsResp := mapper.ToListProjectsResponse(outputListProjectsUseCase)

	return c.Status(fiber.StatusOK).JSON(listProjectsResp)
}

// godoc: UpdateProject godoc
// @Summary Update Project
// @Description Update Project
// @Accept json
// @Produce json
// @Param id path int true "Project ID"
// @Param project body dto.UpdateProjectRequest true "Project"
// @Success 200 {object} dto.UpdateProjectResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource Projects
// @Security BearerAuth
// @Router /api/v1/projects/{id} [patch]
func (h *ProjectHandler) Update(c fiber.Ctx) error {
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

	var projectReq dto.UpdateProjectRequest
	err = c.Bind().Body(&projectReq)
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

	inputUpdateProjectUseCase := mapper.ToUpdateProjectInput(projectReq, id, u)

	outputUpdateProjectUseCase, err := h.updateProjectUseCase.Execute(
		c.Context(),
		inputUpdateProjectUseCase,
	)
	if err != nil {
		return httperror.Handle(
			c,
			err,
		)
	}

	updateProjectResp := mapper.ToUpdateProjectResponse(outputUpdateProjectUseCase)

	return c.Status(fiber.StatusOK).JSON(updateProjectResp)
}

// godoc: DeleteProject godoc
// @Summary Delete Project
// @Description Delete Project
// @Param id path int true "Project ID"
// @Success 204 "No Content"
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource Projects
// @Security BearerAuth
// @Router /api/v1/projects/{id} [delete]
func (h *ProjectHandler) DeleteProject(c fiber.Ctx) error {
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

	inputDeleteProjectUseCase := mapper.ToDeleteProjectInput(id, u)

	err = h.deleteProjectUseCase.Execute(
		c.Context(),
		inputDeleteProjectUseCase,
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
