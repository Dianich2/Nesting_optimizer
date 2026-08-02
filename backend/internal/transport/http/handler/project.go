package handler

import (
	"errors"
	"server_nesting_optimizer/internal/transport/http/dto"
	httperror "server_nesting_optimizer/internal/transport/http/errors"
	"server_nesting_optimizer/internal/transport/http/mapper"
	"server_nesting_optimizer/internal/transport/http/middleware"
	projectusecase "server_nesting_optimizer/internal/usecase/project"
	"server_nesting_optimizer/pkg/apperror"

	"github.com/gofiber/fiber/v3"
)

type ProjectHandler struct {
	createProjectUseCase *projectusecase.CreateProjectUseCase
}

func NewProjectHandler(
	createProjectUseCase *projectusecase.CreateProjectUseCase,
) *ProjectHandler {
	return &ProjectHandler{
		createProjectUseCase: createProjectUseCase,
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
