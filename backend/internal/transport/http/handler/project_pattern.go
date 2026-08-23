package handler

import (
	"errors"
	"server_nesting_optimizer/internal/transport/http/dto"
	httperror "server_nesting_optimizer/internal/transport/http/errors"
	"server_nesting_optimizer/internal/transport/http/mapper"
	"server_nesting_optimizer/internal/transport/http/middleware"
	projectpatternusecase "server_nesting_optimizer/internal/usecase/project_pattern"
	"server_nesting_optimizer/pkg/apperror"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type ProjectPatternHandler struct {
	createProjectPatternUseCase  *projectpatternusecase.CreateProjectPatternUseCase
	getProjectPatternByIDUseCase *projectpatternusecase.GetProjectPatternByIDUseCase
	listProjectPatternsUseCase   *projectpatternusecase.ListProjectPatternsUseCase
	updateProjectPatternUseCase  *projectpatternusecase.UpdateProjectPatternUseCase
	deleteProjectPatternUseCase  *projectpatternusecase.DeleteProjectPatternUseCase
}

func NewProjectPatternHandler(
	createProjectPatternUseCase *projectpatternusecase.CreateProjectPatternUseCase,
	getProjectPatternByIDUseCase *projectpatternusecase.GetProjectPatternByIDUseCase,
	listProjectPatternsUseCase *projectpatternusecase.ListProjectPatternsUseCase,
	updateProjectPatternUseCase *projectpatternusecase.UpdateProjectPatternUseCase,
	deleteProjectPatternUseCase *projectpatternusecase.DeleteProjectPatternUseCase,
) *ProjectPatternHandler {
	return &ProjectPatternHandler{
		createProjectPatternUseCase:  createProjectPatternUseCase,
		getProjectPatternByIDUseCase: getProjectPatternByIDUseCase,
		listProjectPatternsUseCase:   listProjectPatternsUseCase,
		updateProjectPatternUseCase:  updateProjectPatternUseCase,
		deleteProjectPatternUseCase:  deleteProjectPatternUseCase,
	}
}

// godoc: CreateProjectPattern godoc
// @Summary Create ProjectPattern
// @Description Create ProjectPattern
// @Accept json
// @Produce json
// @Param project_id path int true "Project ID"
// @Param projectPattern body dto.CreateProjectPatternRequest true "ProjectPattern"
// @Success 201 {object} dto.CreateProjectPatternResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource ProjectPatterns
// @Security BearerAuth
// @Router /api/v1/projects/{project_id}/patterns [post]
func (h *ProjectPatternHandler) Create(c fiber.Ctx) error {
	var projectPatternReq dto.CreateProjectPatternRequest
	err := c.Bind().Body(&projectPatternReq)
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

	inputCreateProjectPatternUseCase := mapper.ToCreateProjectPatternInput(projectPatternReq, u, projectID)

	outputCreateProjectPatternUseCase, err := h.createProjectPatternUseCase.Execute(
		c.Context(),
		inputCreateProjectPatternUseCase,
	)
	if err != nil {
		return httperror.Handle(
			c,
			err,
		)
	}

	createProjectPatternResp := mapper.ToCreateProjectPatternResponse(outputCreateProjectPatternUseCase)

	return c.Status(fiber.StatusCreated).JSON(createProjectPatternResp)
}

// godoc: GetProjectPatternByID godoc
// @Summary Get ProjectPattern by ID
// @Description Get ProjectPattern by ID
// @Produce json
// @Param project_id path int true "Project ID"
// @Param id path int true "ProjectPattern ID"
// @Success 200 {object} dto.GetProjectPatternByIDResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource ProjectPatterns
// @Security BearerAuth
// @Router /api/v1/projects/{project_id}/patterns/{id} [get]
func (h *ProjectPatternHandler) GetByID(c fiber.Ctx) error {
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

	inputGetProjectPatternByIDUseCase := mapper.ToGetProjectPatternByIDInput(id, u, projectID)

	outputGetProjectPatternByIDUseCase, err := h.getProjectPatternByIDUseCase.Execute(
		c.Context(),
		inputGetProjectPatternByIDUseCase,
	)
	if err != nil {
		return httperror.Handle(
			c,
			err,
		)
	}

	getProjectPatternByIDResp := mapper.ToGetProjectPatternByIDResponse(outputGetProjectPatternByIDUseCase)

	return c.Status(fiber.StatusOK).JSON(getProjectPatternByIDResp)
}

// godoc: ListProjectPatterns godoc
// @Summary List Project Patterns
// @Description List Project Patterns
// @Produce json
// @Param page query int false "Page" default(1)
// @Param page_size query int false "Page Size" default(20)
// @Param project_id path int true "Project ID"
// @Success 200 {object} dto.ListProjectPatternsResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource ProjectPatterns
// @Security BearerAuth
// @Router /api/v1/projects/{project_id}/patterns [get]
func (h *ProjectPatternHandler) ListProjectPatterns(c fiber.Ctx) error {
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

	inputListProjectPatternsUseCase := mapper.ToListProjectPatternsInput(u, projectID, page, pageSize)

	outputListProjectPatternsUseCase, err := h.listProjectPatternsUseCase.Execute(
		c.Context(),
		inputListProjectPatternsUseCase,
	)
	if err != nil {
		return httperror.Handle(
			c,
			err,
		)
	}

	listProjectPatternsResp := mapper.ToListProjectPatternsResponse(outputListProjectPatternsUseCase)

	return c.Status(fiber.StatusOK).JSON(listProjectPatternsResp)
}

// godoc: UpdateProjectPattern godoc
// @Summary Update Project Pattern
// @Description Update Project Pattern
// @Accept json
// @Produce json
// @Param id path int true "Project Pattern ID"
// @Param project_id path int true "Project ID"
// @Param projectPattern body dto.UpdateProjectPatternRequest true "Project Pattern"
// @Success 200 {object} dto.UpdateProjectPatternResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource ProjectPatterns
// @Security BearerAuth
// @Router /api/v1/projects/{project_id}/patterns/{id} [patch]
func (h *ProjectPatternHandler) Update(c fiber.Ctx) error {
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
				"invalid project id",
				apperror.NewFieldError(
					"project_id",
					apperror.FieldCodeInvalid,
					"project_id must be a valid positive integer",
				),
			),
		)
	}

	var projectPatternReq dto.UpdateProjectPatternRequest
	err = c.Bind().Body(&projectPatternReq)
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

	inputUpdateProjectPatternUseCase := mapper.ToUpdateProjectPatternInput(projectPatternReq, u, projectID, id)

	outputUpdateProjectPatternUseCase, err := h.updateProjectPatternUseCase.Execute(
		c.Context(),
		inputUpdateProjectPatternUseCase,
	)
	if err != nil {
		return httperror.Handle(
			c,
			err,
		)
	}

	updateProjectPatternResp := mapper.ToUpdateProjectPatternResponse(outputUpdateProjectPatternUseCase)

	return c.Status(fiber.StatusOK).JSON(updateProjectPatternResp)
}

// godoc: DeleteProjectPattern godoc
// @Summary Delete Project Pattern
// @Description Delete Project Pattern
// @Param project_id path int true "Project ID"
// @Param id path int true "Project Pattern ID"
// @Success 204 "No Content"
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource ProjectPatterns
// @Security BearerAuth
// @Router /api/v1/projects/{project_id}/patterns/{id} [delete]
func (h *ProjectPatternHandler) Delete(c fiber.Ctx) error {
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
				"invalid project id",
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

	inputDeleteProjectPatternUseCase := mapper.ToDeleteProjectPatternInput(u, projectID, id)

	err = h.deleteProjectPatternUseCase.Execute(
		c.Context(),
		inputDeleteProjectPatternUseCase,
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
