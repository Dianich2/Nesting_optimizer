package handler

import (
	"errors"
	"server_nesting_optimizer/internal/transport/http/dto"
	httperror "server_nesting_optimizer/internal/transport/http/errors"
	"server_nesting_optimizer/internal/transport/http/mapper"
	"server_nesting_optimizer/internal/transport/http/middleware"
	userusecase "server_nesting_optimizer/internal/usecase/user"
	"server_nesting_optimizer/pkg/apperror"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type UserHandler struct {
	createUserUseCase     *userusecase.CreateUserUseCase
	loginUseCase          *userusecase.LoginUseCase
	refreshUseCase        *userusecase.RefreshUseCase
	logoutUseCase         *userusecase.LogoutUseCase
	getCurrentUserUseCase *userusecase.GetCurrentUserUseCase
	updateProfileUseCase  *userusecase.UpdateProfileUseCase
}

func NewUserHandler(
	createUserUseCase *userusecase.CreateUserUseCase,
	loginUseCase *userusecase.LoginUseCase,
	refreshUseCase *userusecase.RefreshUseCase,
	logoutUseCase *userusecase.LogoutUseCase,
	getCurrentUserUseCase *userusecase.GetCurrentUserUseCase,
	updateProfileUseCase *userusecase.UpdateProfileUseCase,
) *UserHandler {
	return &UserHandler{
		createUserUseCase:     createUserUseCase,
		loginUseCase:          loginUseCase,
		logoutUseCase:         logoutUseCase,
		refreshUseCase:        refreshUseCase,
		getCurrentUserUseCase: getCurrentUserUseCase,
		updateProfileUseCase:  updateProfileUseCase,
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
// @Failure 409 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource Users
// @Router /api/v1/users [post]
func (h *UserHandler) Create(c fiber.Ctx) error {
	var userReq dto.CreateUserRequest
	err := c.Bind().Body(&userReq)
	if err != nil {
		return httperror.Handle(
			c,
			apperror.Validation("invalid request body"),
		)
	}

	inputUserUseCase := mapper.ToCreateUserInput(userReq)

	outputUserUseCase, err := h.createUserUseCase.Execute(
		c.Context(),
		inputUserUseCase,
	)
	if err != nil {
		return httperror.Handle(
			c,
			err,
		)
	}

	userResp := mapper.ToCreateUserResponse(outputUserUseCase)

	return c.Status(fiber.StatusCreated).JSON(userResp)
}

// godoc: Login godoc
// @Summary Login
// @Description Login
// @Accept json
// @Produce json
// @Param credentials body dto.LoginRequest true "Credentials"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource Users
// @Router /api/v1/auth/login [post]
func (h *UserHandler) Login(c fiber.Ctx) error {
	var loginReq dto.LoginRequest
	err := c.Bind().Body(&loginReq)
	if err != nil {
		return httperror.Handle(
			c,
			apperror.Validation("invalid request body"),
		)
	}

	inputLoginUseCase := mapper.ToLoginInput(loginReq)

	outputLoginUseCase, err := h.loginUseCase.Execute(
		c.Context(),
		inputLoginUseCase,
	)
	if err != nil {
		return httperror.Handle(c, err)
	}

	loginResp := mapper.ToLoginResponse(outputLoginUseCase)

	return c.Status(fiber.StatusOK).JSON(loginResp)
}

// godoc: Refresh godoc
// @Summary Refresh
// @Description Refresh
// @Accept json
// @Produce json
// @Param credentials body dto.RefreshRequest true "Refresh Token"
// @Success 200 {object} dto.RefreshResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource Users
// @Router /api/v1/auth/refresh [post]
func (h *UserHandler) Refresh(c fiber.Ctx) error {
	var refreshReq dto.RefreshRequest
	err := c.Bind().Body(&refreshReq)
	if err != nil {
		return httperror.Handle(
			c,
			apperror.Validation("invalid request body"),
		)
	}

	inputRefreshUseCase := mapper.ToRefreshInput(refreshReq)

	outputRefreshUseCase, err := h.refreshUseCase.Execute(
		c.Context(),
		inputRefreshUseCase,
	)
	if err != nil {
		return httperror.Handle(c, err)
	}

	refreshResp := mapper.ToRefreshResponse(outputRefreshUseCase)

	return c.Status(fiber.StatusOK).JSON(refreshResp)
}

// godoc: Logout godoc
// @Summary Logout
// @Description Logout
// @Success 204 "No Content"
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource Users
// @Security BearerAuth
// @Router /api/v1/auth/logout [post]
func (h *UserHandler) Logout(c fiber.Ctx) error {
	sessionID := c.Locals(middleware.SessionIDLocalKey)
	s, ok := sessionID.(uuid.UUID)
	if !ok {
		return httperror.Handle(
			c,
			apperror.Internal("failed to get session id from request context", errors.New("session_id local is missing or invalid")),
		)
	}

	inputLogoutUseCase := userusecase.LogoutInput{
		SessionID: s,
	}

	if err := h.logoutUseCase.Execute(
		c.Context(),
		inputLogoutUseCase,
	); err != nil {
		return httperror.Handle(
			c,
			err,
		)
	}

	c.Status(fiber.StatusNoContent)
	return nil
}

// godoc: GetCurrentUser godoc
// @Summary GetCurrentUser
// @Description GetCurrentUser
// @Produce json
// @Success 200 {object} dto.GetCurrentUserResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource Users
// @Security BearerAuth
// @Router /api/v1/users/me [get]
func (h *UserHandler) GetCurrentUser(c fiber.Ctx) error {
	userID := c.Locals(middleware.UserIDLocalKey)
	u, ok := userID.(int64)
	if !ok {
		return httperror.Handle(
			c,
			apperror.Internal("failed to get user id from request context", errors.New("user_id local is missing or invalid")),
		)
	}

	inputGetCurrentUserUseCase := mapper.ToGetCurrentUserInput(u)

	outputGetCurrentUserUseCase, err := h.getCurrentUserUseCase.Execute(
		c.Context(),
		inputGetCurrentUserUseCase,
	)
	if err != nil {
		return httperror.Handle(c, err)
	}

	getCurrentUserResp := mapper.ToGetCurrentUserResponse(outputGetCurrentUserUseCase)

	return c.Status(fiber.StatusOK).JSON(getCurrentUserResp)
}

// godoc: UpdateProfile godoc
// @Summary UpdateProfile
// @Description UpdateProfile
// @Accept json
// @Produce json
// @Param updateProfile body dto.UpdateProfileRequest true "Update Profile"
// @Success 200 {object} dto.UpdateProfileResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Resource Users
// @Security BearerAuth
// @Router /api/v1/users/me [patch]
func (h *UserHandler) UpdateProfile(c fiber.Ctx) error {
	userID := c.Locals(middleware.UserIDLocalKey)
	u, ok := userID.(int64)
	if !ok {
		return httperror.Handle(
			c,
			apperror.Internal("failed to get user id from request context", errors.New("user_id local is missing or invalid")),
		)
	}

	var updateProfileReq dto.UpdateProfileRequest
	err := c.Bind().Body(&updateProfileReq)
	if err != nil {
		return httperror.Handle(
			c,
			apperror.Validation("invalid request body"),
		)
	}

	inputUpdateProfileUseCase := mapper.ToUpdateProfileInput(updateProfileReq)

	outputUpdateProfileUseCase, err := h.updateProfileUseCase.Execute(
		c.Context(),
		inputUpdateProfileUseCase,
		u,
	)
	if err != nil {
		return httperror.Handle(c, err)
	}

	updateProfileResp := mapper.ToUpdateProfileResponse(outputUpdateProfileUseCase)

	return c.Status(fiber.StatusOK).JSON(updateProfileResp)
}
