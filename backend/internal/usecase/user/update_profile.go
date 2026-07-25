package user

import (
	"context"
	"errors"
	domainuser "server_nesting_optimizer/internal/domain/user"
	"server_nesting_optimizer/pkg/apperror"
)

type UpdateProfileUseCase struct {
	repo UserRepository
}

func NewUpdateProfileUseCase(
	repo UserRepository,
) *UpdateProfileUseCase {
	return &UpdateProfileUseCase{
		repo: repo,
	}
}

func (uc *UpdateProfileUseCase) Execute(
	ctx context.Context,
	input UpdateProfileInput,
	id int64,
) (UpdateProfileOutput, error) {
	input = normalizeUpdateProfileInput(input)
	details := input.Validate()
	if len(details) > 0 {
		return UpdateProfileOutput{}, apperror.Validation(
			"validation failed",
			details...,
		)
	}

	updatedProfile, err := uc.repo.UpdateProfile(
		ctx,
		input.FirstName,
		input.LastName,
		id,
	)

	if err != nil {
		switch {
		case errors.Is(err, domainuser.ErrNotFound):
			return UpdateProfileOutput{}, apperror.NotFound(
				"user profile not found",
			)

		default:
			return UpdateProfileOutput{}, apperror.Internal(
				"failed to update user profile",
				err,
			)
		}
	}

	return toUpdateProfileOutput(updatedProfile), nil
}
