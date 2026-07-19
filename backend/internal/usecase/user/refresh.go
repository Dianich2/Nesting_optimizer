package user

import (
	"context"
	"errors"
	domainsession "server_nesting_optimizer/internal/domain/session"
	"server_nesting_optimizer/pkg/apperror"
	"server_nesting_optimizer/pkg/jwt"
	"strings"
	"time"
)

type RefreshInput struct {
	RefreshToken string
}

type RefreshOutput struct {
	AccessToken  string
	RefreshToken string
}

type RefreshUseCase struct {
	sessionRepo    SessionRepository
	sessionFactory SessionFactory
	tokens         TokenManager
}

func NewRefreshUseCase(
	sessionRepo SessionRepository,
	sessionFactory SessionFactory,
	tokens TokenManager,
) *RefreshUseCase {
	return &RefreshUseCase{
		sessionRepo:    sessionRepo,
		sessionFactory: sessionFactory,
		tokens:         tokens,
	}
}

func (uc *RefreshUseCase) Execute(
	ctx context.Context,
	input RefreshInput,
) (RefreshOutput, error) {
	refreshToken := strings.TrimSpace(input.RefreshToken)

	if refreshToken == "" {
		return RefreshOutput{}, apperror.Validation(
			"refresh token must not be empty",
			apperror.NewFieldError(
				"refresh_token",
				apperror.FieldCodeRequired,
				"empty refresh token",
			),
		)
	}

	claims, err := uc.tokens.ParseRefreshToken(refreshToken)
	if err != nil {
		if errors.Is(err, jwt.ErrExpiredToken) {
			return RefreshOutput{}, apperror.Unauthorized(
				"token expired",
			)
		}

		if errors.Is(err, jwt.ErrUnexpectedTokenType) {
			return RefreshOutput{}, apperror.Unauthorized(
				"invalid token type",
			)
		}

		return RefreshOutput{}, apperror.Unauthorized(
			"invalid token",
		)
	}

	currentSession, err := uc.sessionRepo.GetBySessionID(
		ctx,
		claims.SessionID,
	)
	if err != nil {
		if errors.Is(err, domainsession.ErrNotFound) {
			return RefreshOutput{}, apperror.Unauthorized(
				"invalid token",
			)
		}

		return RefreshOutput{}, apperror.Internal(
			"failed to get session by session id",
			err,
		)
	}

	if currentSession.UserID != claims.UserID {
		return RefreshOutput{}, apperror.Unauthorized(
			"invalid token",
		)
	}

	now := time.Now()

	if currentSession.IsExpired(now) {
		return RefreshOutput{}, apperror.Unauthorized(
			"token expired",
		)
	}

	if !currentSession.MatchesRefreshToken(refreshToken) {
		return RefreshOutput{}, apperror.Unauthorized(
			"invalid token",
		)
	}

	newSession := uc.sessionFactory.New(
		currentSession.UserID,
	)

	newAccessToken, err := uc.tokens.GenerateAccessToken(
		newSession.UserID,
		newSession.SessionID,
	)
	if err != nil {
		return RefreshOutput{}, apperror.Internal(
			"failed to generate new access token",
			err,
		)
	}

	newRefreshToken, err := uc.tokens.GenerateRefreshToken(
		newSession.UserID,
		newSession.SessionID,
	)
	if err != nil {
		return RefreshOutput{}, apperror.Internal(
			"failed to generate new refresh token",
			err,
		)
	}

	newSession.SetRefreshToken(newRefreshToken)

	if _, err := uc.sessionRepo.Rotate(
		ctx,
		currentSession.SessionID,
		currentSession.RefreshTokenHash,
		newSession,
	); err != nil {
		if errors.Is(err, domainsession.ErrSessionChanged) {
			return RefreshOutput{}, apperror.Unauthorized(
				"invalid token",
			)
		}

		return RefreshOutput{}, apperror.Internal(
			"failed to save session",
			err,
		)
	}

	return RefreshOutput{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}
