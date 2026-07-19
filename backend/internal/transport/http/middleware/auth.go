package middleware

import (
	"context"
	"errors"
	domainsession "server_nesting_optimizer/internal/domain/session"
	"server_nesting_optimizer/pkg/apperror"
	"server_nesting_optimizer/pkg/jwt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type SessionRepository interface {
	GetBySessionID(
		ctx context.Context,
		sessionID uuid.UUID,
	) (domainsession.Session, error)
}

func AuthRequired(
	jwtManager *jwt.Manager,
	sessionRepo SessionRepository,
) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := strings.TrimSpace(c.Get("Authorization"))
		if authHeader == "" {
			return apperror.Unauthorized(
				"authorization header must not be empty",
			)
		}

		if !strings.HasPrefix(
			authHeader,
			"Bearer ",
		) {
			return apperror.Unauthorized(
				"invalid authorization header",
			)
		}

		token := strings.TrimSpace(
			strings.TrimPrefix(
				authHeader,
				"Bearer ",
			),
		)

		if token == "" {
			return apperror.Unauthorized(
				"token must not be empty",
			)
		}

		claims, err := jwtManager.ParseAccessToken(token)
		if err != nil {
			if errors.Is(err, jwt.ErrExpiredToken) {
				return apperror.Unauthorized(
					"token expired",
				)
			}

			if errors.Is(err, jwt.ErrUnexpectedTokenType) {
				return apperror.Unauthorized(
					"invalid token type",
				)
			}

			return apperror.Unauthorized(
				"invalid token",
			)
		}

		currentSession, err := sessionRepo.GetBySessionID(
			c.Context(),
			claims.SessionID,
		)
		if err != nil {
			if errors.Is(err, domainsession.ErrNotFound) {
				return apperror.Unauthorized(
					"invalid token",
				)
			}

			return apperror.Internal(
				"failed to get session by session id",
				err,
			)
		}

		if currentSession.UserID != claims.UserID {
			return apperror.Unauthorized(
				"invalid token",
			)
		}

		if currentSession.IsExpired(time.Now()) {
			return apperror.Unauthorized(
				"token expired",
			)
		}

		c.Locals(UserIDLocalKey, claims.UserID)
		c.Locals(SessionIDLocalKey, claims.SessionID)

		return c.Next()
	}
}
