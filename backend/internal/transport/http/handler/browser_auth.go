package handler

import (
	"time"

	"server_nesting_optimizer/internal/config"
	"server_nesting_optimizer/internal/transport/http/dto"
	httperror "server_nesting_optimizer/internal/transport/http/errors"
	"server_nesting_optimizer/internal/transport/http/mapper"
	"server_nesting_optimizer/internal/transport/http/middleware"
	userusecase "server_nesting_optimizer/internal/usecase/user"
	"server_nesting_optimizer/pkg/apperror"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

const (
	browserCookieName = "nesting_refresh"
	browserCookiePath = "/api/v1/auth/browser"
)

type BrowserAuthHandler struct {
	loginUseCase   *userusecase.LoginUseCase
	refreshUseCase *userusecase.RefreshUseCase
	logoutUseCase  *userusecase.LogoutUseCase

	allowedOrigin string
	cookieSecure  bool
	cookieMaxAge  int
}

func NewBrowserAuthHandler(
	login *userusecase.LoginUseCase,
	refresh *userusecase.RefreshUseCase,
	logout *userusecase.LogoutUseCase,
	cfg config.Config,
) *BrowserAuthHandler {
	return &BrowserAuthHandler{
		loginUseCase:   login,
		refreshUseCase: refresh,
		logoutUseCase:  logout,
		allowedOrigin:  cfg.HTTP.BrowserOrigin,
		cookieSecure:   cfg.App.Env != "local",
		cookieMaxAge:   cfg.JWT.RefreshTTLMin * 60,
	}
}

func (h *BrowserAuthHandler) checkOrigin(c fiber.Ctx) error {
	if c.Get("Origin") != h.allowedOrigin {
		return apperror.Forbidden("invalid request origin")
	}

	return nil
}

func (h *BrowserAuthHandler) setRefreshCookie(
	c fiber.Ctx,
	token string,
) {
	c.Cookie(&fiber.Cookie{
		Name:     browserCookieName,
		Value:    token,
		Path:     browserCookiePath,
		MaxAge:   h.cookieMaxAge,
		HTTPOnly: true,
		Secure:   h.cookieSecure,
		SameSite: "Strict",
	})
}

func (h *BrowserAuthHandler) clearRefreshCookie(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     browserCookieName,
		Value:    "",
		Path:     browserCookiePath,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HTTPOnly: true,
		Secure:   h.cookieSecure,
		SameSite: "Strict",
	})
}

func (h *BrowserAuthHandler) Login(c fiber.Ctx) error {
	if err := h.checkOrigin(c); err != nil {
		return err
	}

	var req dto.LoginRequest

	if err := parseBody(c, &req); err != nil {
		return err
	}

	result, err := h.loginUseCase.Execute(
		c.Context(),
		mapper.ToLoginInput(req),
	)
	if err != nil {
		return httperror.Handle(c, err)
	}

	h.setRefreshCookie(c, result.RefreshToken)
	c.Set("Cache-Control", "no-store")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"access_token": result.AccessToken,
	})
}

func (h *BrowserAuthHandler) Refresh(c fiber.Ctx) error {
	if err := h.checkOrigin(c); err != nil {
		return err
	}

	token := c.Cookies(browserCookieName)

	result, err := h.refreshUseCase.Execute(
		c.Context(),
		userusecase.RefreshInput{
			RefreshToken: token,
		},
	)
	if err != nil {
		return httperror.Handle(c, err)
	}

	h.setRefreshCookie(c, result.RefreshToken)
	c.Set("Cache-Control", "no-store")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"access_token": result.AccessToken,
	})
}

func (h *BrowserAuthHandler) Logout(c fiber.Ctx) error {
	if err := h.checkOrigin(c); err != nil {
		return err
	}

	sessionID, ok := c.Locals(
		middleware.SessionIDLocalKey,
	).(uuid.UUID)

	if !ok {
		return httperror.Handle(
			c,
			apperror.Unauthorized("invalid session"),
		)
	}

	if err := h.logoutUseCase.Execute(
		c.Context(),
		userusecase.LogoutInput{
			SessionID: sessionID,
		},
	); err != nil {
		return httperror.Handle(c, err)
	}

	h.clearRefreshCookie(c)
	c.Set("Cache-Control", "no-store")

	return c.SendStatus(fiber.StatusNoContent)
}
