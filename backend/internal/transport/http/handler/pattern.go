package handler

import (
	"server_nesting_optimizer/internal/transport/http/dto"
	httperror "server_nesting_optimizer/internal/transport/http/errors"
	"server_nesting_optimizer/internal/transport/http/mapper"
	patternusecase "server_nesting_optimizer/internal/usecase/pattern"

	"github.com/gofiber/fiber/v3"
)

type PatternHandler struct {
	createPatternUseCase  *patternusecase.CreatePatternUseCase
	getPatternByIDUseCase *patternusecase.GetPatternByIDUseCase
	listPatternsUseCase   *patternusecase.ListPatternsUseCase
	updatePatternUseCase  *patternusecase.UpdatePatternUseCase
	deletePatternUseCase  *patternusecase.DeletePatternUseCase
}

func NewPatternHandler(
	createPatternUseCase *patternusecase.CreatePatternUseCase,
	getPatternByIDUseCase *patternusecase.GetPatternByIDUseCase,
	listPatternsUseCase *patternusecase.ListPatternsUseCase,
	updatePatternUseCase *patternusecase.UpdatePatternUseCase,
	deletePatternUseCase *patternusecase.DeletePatternUseCase,
) *PatternHandler {
	return &PatternHandler{
		createPatternUseCase:  createPatternUseCase,
		getPatternByIDUseCase: getPatternByIDUseCase,
		listPatternsUseCase:   listPatternsUseCase,
		updatePatternUseCase:  updatePatternUseCase,
		deletePatternUseCase:  deletePatternUseCase,
	}
}

// godoc: CreatePattern godoc
// @Summary Create Pattern
// @Description Create Pattern
// @Accept json
// @Produce json
// @Param pattern body dto.CreatePatternRequest true "Pattern"
// @Success 201 {object} dto.CreatePatternResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource Patterns
// @Security BearerAuth
// @Router /api/v1/patterns [post]
func (h *PatternHandler) Create(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var patternReq dto.CreatePatternRequest
	err = parseBody(c, &patternReq)
	if err != nil {
		return err
	}

	inputCreatePatternUseCase := mapper.ToCreatePatternInput(
		patternReq, userID,
	)

	outputCreatePatternUseCase, err := h.createPatternUseCase.Execute(
		c.Context(),
		inputCreatePatternUseCase,
	)
	if err != nil {
		return httperror.Handle(
			c,
			err,
		)
	}

	createPatternResp := mapper.ToCreatePatternResponse(outputCreatePatternUseCase)

	return c.Status(fiber.StatusCreated).JSON(createPatternResp)
}

// godoc: GetPatternByID godoc
// @Summary Get Pattern by ID
// @Description Get Pattern by ID
// @Produce json
// @Param id path int true "Pattern ID"
// @Success 200 {object} dto.GetPatternResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource Patterns
// @Security BearerAuth
// @Router /api/v1/patterns/{id} [get]
func (h *PatternHandler) GetByID(c fiber.Ctx) error {
	id, err := getIDFromPath(c, "id")
	if err != nil {
		return err
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	inputGetPatternUseCase := mapper.ToGetPatternInput(
		id,
		userID,
	)

	outputGetPatternUseCase, err := h.getPatternByIDUseCase.Execute(
		c.Context(),
		inputGetPatternUseCase,
	)
	if err != nil {
		return httperror.Handle(
			c,
			err,
		)
	}

	getPatternResp := mapper.ToGetPatternResponse(outputGetPatternUseCase)

	return c.Status(fiber.StatusOK).JSON(getPatternResp)
}

// godoc: ListPatterns godoc
// @Summary List Patterns
// @Description List Patterns
// @Produce json
// @Param page query int false "Page" default(1)
// @Param page_size query int false "Page Size" default(20)
// @Success 200 {object} dto.ListPatternsResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource Patterns
// @Security BearerAuth
// @Router /api/v1/patterns [get]
func (h *PatternHandler) ListPatterns(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	page, pageSize, err := parsePagination(c)
	if err != nil {
		return err
	}

	inputListPatternsUseCase := mapper.ToListPatternsInput(
		page,
		pageSize,
		userID,
	)

	outputListPatternsUseCase, err := h.listPatternsUseCase.Execute(
		c.Context(),
		inputListPatternsUseCase,
	)
	if err != nil {
		return httperror.Handle(
			c,
			err,
		)
	}

	listPatternsResp := mapper.ToListPatternsResponse(outputListPatternsUseCase)

	return c.Status(fiber.StatusOK).JSON(listPatternsResp)
}

// godoc: UpdatePattern godoc
// @Summary Update Pattern
// @Description Update Pattern
// @Accept json
// @Produce json
// @Param id path int true "Pattern ID"
// @Param pattern body dto.UpdatePatternRequest true "Pattern"
// @Success 200 {object} dto.UpdatePatternResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource Patterns
// @Security BearerAuth
// @Router /api/v1/patterns/{id} [patch]
func (h *PatternHandler) Update(c fiber.Ctx) error {
	id, err := getIDFromPath(c, "id")
	if err != nil {
		return err
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var patternReq dto.UpdatePatternRequest
	err = parseBody(c, &patternReq)
	if err != nil {
		return err
	}

	inputUpdatePatternUseCase := mapper.ToUpdatePatternInput(
		patternReq,
		userID,
		id,
	)

	outputUpdatePatternUseCase, err := h.updatePatternUseCase.Execute(
		c.Context(),
		inputUpdatePatternUseCase,
	)
	if err != nil {
		return httperror.Handle(
			c,
			err,
		)
	}

	updatePatternResp := mapper.ToUpdatePatternResponse(outputUpdatePatternUseCase)

	return c.Status(fiber.StatusOK).JSON(updatePatternResp)
}

// godoc: DeletePattern godoc
// @Summary Delete Pattern
// @Description Delete Pattern
// @Param id path int true "Pattern ID"
// @Success 204 "No Content"
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource Patterns
// @Security BearerAuth
// @Router /api/v1/patterns/{id} [delete]
func (h *PatternHandler) Delete(c fiber.Ctx) error {
	id, err := getIDFromPath(c, "id")
	if err != nil {
		return err
	}

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	inputDeletePatternUseCase := mapper.ToDeletePatternInput(
		id,
		userID,
	)

	err = h.deletePatternUseCase.Execute(
		c.Context(),
		inputDeletePatternUseCase,
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
