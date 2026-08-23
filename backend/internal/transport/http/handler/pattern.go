package handler

import (
	"errors"
	"server_nesting_optimizer/internal/transport/http/dto"
	httperror "server_nesting_optimizer/internal/transport/http/errors"
	"server_nesting_optimizer/internal/transport/http/mapper"
	"server_nesting_optimizer/internal/transport/http/middleware"
	patternusecase "server_nesting_optimizer/internal/usecase/pattern"
	"server_nesting_optimizer/pkg/apperror"
	"strconv"

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
	var patternReq dto.CreatePatternRequest
	err := c.Bind().Body(&patternReq)
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

	inputCreatePatternUseCase := mapper.ToCreatePatternInput(patternReq, u)

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

	inputGetPatternUseCase := mapper.ToGetPatternInput(id, u)

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

	inputListPatternsUseCase := mapper.ToListPatternsInput(page, pageSize, u)

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

	var patternReq dto.UpdatePatternRequest
	err = c.Bind().Body(&patternReq)
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

	inputUpdatePatternUseCase := mapper.ToUpdatePatternInput(patternReq, u, id)

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

	inputDeletePatternUseCase := mapper.ToDeletePatternInput(id, u)

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
