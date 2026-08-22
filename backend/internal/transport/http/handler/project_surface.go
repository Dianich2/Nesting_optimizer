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
	createProjectSurfaceUseCase *projectsurfaceusecase.CreateProjectSurfaceUseCase
}

func NewProjectSurfaceHandler(
	createProjectSurfaceUseCase *projectsurfaceusecase.CreateProjectSurfaceUseCase,
) *ProjectSurfaceHandler {
	return &ProjectSurfaceHandler{
		createProjectSurfaceUseCase: createProjectSurfaceUseCase,
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
