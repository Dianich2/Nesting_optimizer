package user

import (
	"context"
	"server_nesting_optimizer/pkg/apperror"

	"github.com/google/uuid"
)

type LogoutInput struct {
	SessionID uuid.UUID
}

type LogoutUseCase struct {
	sessionRepo SessionRepository
}

func NewLogoutUseCase(
	sessionRepo SessionRepository,
) *LogoutUseCase {
	return &LogoutUseCase{
		sessionRepo: sessionRepo,
	}
}

func (uc *LogoutUseCase) Execute(
	ctx context.Context,
	input LogoutInput,
) error {
	if err := uc.sessionRepo.DeleteBySessionID(
		ctx,
		input.SessionID,
	); err != nil {
		return apperror.Internal(
			"failed to delete session",
			err,
		)
	}

	return nil
}
