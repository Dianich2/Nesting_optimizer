package handler

import (
	"server_nesting_optimizer/internal/transport/http/dto"
	httperror "server_nesting_optimizer/internal/transport/http/errors"
	"server_nesting_optimizer/internal/transport/http/mapper"
	userusecase "server_nesting_optimizer/internal/usecase/user"
	"server_nesting_optimizer/pkg/apperror"

	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	createUserUseCase *userusecase.CreateUserUseCase
}

func NewUserHandler(
	createUserUseCase *userusecase.CreateUserUseCase,
) *UserHandler {
	return &UserHandler{
		createUserUseCase: createUserUseCase,
	}
}

// godoc: CreateUser godoc
// @Summary Create User
// @Description Create User
// @Accept json
// @Produce json
// @Param user body dto.CreateUserRequest true "User"
// @Success 201 {object} dto.CreateUserResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 409 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource Users
// @Router /api/v1/users [post]
func (h *UserHandler) Create(c fiber.Ctx) error {
	var userReq dto.CreateUserRequest
	err := c.Bind().Body(&userReq)
	if err != nil {
		return httperror.Handle(c, apperror.Validation("invalid request body"))
	}

	inputUserUseCase := mapper.ToCreateUserInput(userReq)

	outputUserUseCase, err := h.createUserUseCase.Execute(c, inputUserUseCase)
	if err != nil {
		return httperror.Handle(c, err)
	}

	userResp := mapper.ToCreateUserResponse(outputUserUseCase)

	return c.Status(fiber.StatusCreated).JSON(userResp)
}
