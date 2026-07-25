package user

import (
	"context"
	"errors"
	domainuser "server_nesting_optimizer/internal/domain/user"
	"server_nesting_optimizer/pkg/apperror"
)

type GetCurrentUserUseCase struct {
	repo UserRepository
}

func NewGetCurrentUserUseCase(
	repo UserRepository,
) *GetCurrentUserUseCase {
	return &GetCurrentUserUseCase{
		repo: repo,
	}
}

func (uc *GetCurrentUserUseCase) Execute(
	ctx context.Context,
	input GetCurrentUserInput,
) (GetCurrentUserOutput, error) {
	user, err := uc.repo.GetByID(
		ctx,
		input.ID,
	)
	if err != nil {
		switch {
		case errors.Is(err, domainuser.ErrNotFound):
			return GetCurrentUserOutput{}, apperror.NotFound(
				"user not found",
			)
		default:
			return GetCurrentUserOutput{}, apperror.Internal(
				"failed to get user",
				err,
			)
		}
	}

	return toGetCurrentUserOutput(user), nil
}
