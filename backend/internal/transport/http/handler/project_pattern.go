package handler

import (
	"server_nesting_optimizer/internal/transport/http/dto"
	httperror "server_nesting_optimizer/internal/transport/http/errors"
	"server_nesting_optimizer/internal/transport/http/mapper"
	projectpatternusecase "server_nesting_optimizer/internal/usecase/project_pattern"

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
	projectID, err := getIDFromPath(c, "project_id")
	if err != nil {
		return err
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var projectPatternReq dto.CreateProjectPatternRequest
	err = parseBody(c, &projectPatternReq)
	if err != nil {
		return err
	}

	inputCreateProjectPatternUseCase := mapper.ToCreateProjectPatternInput(
		projectPatternReq,
		userID,
		projectID,
	)

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

	inputGetProjectPatternByIDUseCase := mapper.ToGetProjectPatternByIDInput(
		id,
		userID,
		projectID,
	)

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

	inputListProjectPatternsUseCase := mapper.ToListProjectPatternsInput(
		userID,
		projectID,
		page,
		pageSize,
	)

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

	var projectPatternReq dto.UpdateProjectPatternRequest
	err = parseBody(c, &projectPatternReq)
	if err != nil {
		return err
	}

	inputUpdateProjectPatternUseCase := mapper.ToUpdateProjectPatternInput(
		projectPatternReq,
		userID,
		projectID,
		id,
	)

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

	inputDeleteProjectPatternUseCase := mapper.ToDeleteProjectPatternInput(
		userID,
		projectID,
		id,
	)

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
