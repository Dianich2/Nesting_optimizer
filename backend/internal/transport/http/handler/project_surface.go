package handler

import (
	"errors"
	"server_nesting_optimizer/internal/transport/http/dto"
	httperror "server_nesting_optimizer/internal/transport/http/errors"
	"server_nesting_optimizer/internal/transport/http/mapper"
	"server_nesting_optimizer/internal/transport/http/middleware"
	projectsurfaceusecase "server_nesting_optimizer/internal/usecase/project_surface"
	"server_nesting_optimizer/pkg/apperror"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type ProjectSurfaceHandler struct {
	createProjectSurfaceUseCase  *projectsurfaceusecase.CreateProjectSurfaceUseCase
	getProjectSurfaceByIDUseCase *projectsurfaceusecase.GetProjectSurfaceByIDUseCase
	listProjectSurfacesUseCase   *projectsurfaceusecase.ListProjectSurfacesUseCase
}

func NewProjectSurfaceHandler(
	createProjectSurfaceUseCase *projectsurfaceusecase.CreateProjectSurfaceUseCase,
	getProjectSurfaceByIDUseCase *projectsurfaceusecase.GetProjectSurfaceByIDUseCase,
	listProjectSurfacesUseCase *projectsurfaceusecase.ListProjectSurfacesUseCase,
) *ProjectSurfaceHandler {
	return &ProjectSurfaceHandler{
		createProjectSurfaceUseCase:  createProjectSurfaceUseCase,
		getProjectSurfaceByIDUseCase: getProjectSurfaceByIDUseCase,
		listProjectSurfacesUseCase:   listProjectSurfacesUseCase,
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
	var projectSurfaceReq dto.CreateProjectSurfaceRequest
	err := c.Bind().Body(&projectSurfaceReq)
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

	inputCreateProjectSurfaceUseCase := mapper.ToCreateProjectSurfaceInput(projectSurfaceReq, u, projectID)

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

	projectID, err := strconv.ParseInt(c.Params("project_id"), 10, 64)
	if err != nil {
		return httperror.Handle(
			c,
			apperror.Validation(
				"invalid id",
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

	inputGetProjectSurfaceByIDUseCase := mapper.ToGetProjectSurfaceByIDInput(id, u, projectID)

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
	projectID, err := strconv.ParseInt(c.Params("project_id"), 10, 64)
	if err != nil {
		return httperror.Handle(
			c,
			apperror.Validation(
				"invalid id",
				apperror.NewFieldError(
					"project_id",
					apperror.FieldCodeInvalid,
					"project_id must be a valid positive integer",
				),
			),
		)
	}

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

	inputListProjectSurfacesUseCase := mapper.ToListProjectSurfacesInput(u, projectID, page, pageSize)

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
