package user

import (
	"context"
	"errors"

	domainuser "server_nesting_optimizer/internal/domain/user"
	"server_nesting_optimizer/pkg/apperror"
)

type LoginInput struct {
	Identifier string
	Password   string
}

type LoginOutput struct {
	AccessToken  string
	RefreshToken string
}

type LoginUseCase struct {
	userRepo       UserRepository
	sessionRepo    SessionRepository
	hasher         PasswordHasher
	tokens         TokenManager
	sessionFactory SessionFactory
}

func NewLoginUseCase(
	userRepo UserRepository,
	sessionRepo SessionRepository,
	hasher PasswordHasher,
	tokens TokenManager,
	sessionFactory SessionFactory,
) *LoginUseCase {
	return &LoginUseCase{
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		hasher:         hasher,
		tokens:         tokens,
		sessionFactory: sessionFactory,
	}
}

func (uc *LoginUseCase) Execute(
	ctx context.Context,
	input LoginInput,
) (LoginOutput, error) {
	input = normalizeLoginInput(input)
	details := validateLoginInput(input)
	if len(details) > 0 {
		return LoginOutput{}, apperror.Validation(
			"validation failed",
			details...,
		)
	}

	user, err := uc.userRepo.GetByIdentifier(
		ctx,
		input.Identifier,
	)
	if err != nil {
		if errors.Is(err, domainuser.ErrNotFound) {
			return LoginOutput{}, apperror.Unauthorized(
				"invalid credentials",
			)
		}
		return LoginOutput{}, apperror.Internal(
			"failed to get user",
			err,
		)
	}

	if err := uc.hasher.Compare(
		user.PasswordHash,
		input.Password,
	); err != nil {
		return LoginOutput{}, apperror.Unauthorized(
			"invalid credentials",
		)
	}

	session := uc.sessionFactory.New(
		user.ID,
	)

	accessToken, err := uc.tokens.GenerateAccessToken(
		user.ID,
		session.SessionID,
	)
	if err != nil {
		return LoginOutput{}, apperror.Internal(
			"failed to generate access token",
			err,
		)
	}

	refreshToken, err := uc.tokens.GenerateRefreshToken(
		user.ID,
		session.SessionID,
	)
	if err != nil {
		return LoginOutput{}, apperror.Internal(
			"failed to generate refresh token",
			err,
		)
	}

	session.SetRefreshToken(refreshToken)

	if _, err := uc.sessionRepo.Upsert(
		ctx,
		session,
	); err != nil {
		return LoginOutput{}, apperror.Internal(
			"failed to save session",
			err,
		)
	}

	return LoginOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
