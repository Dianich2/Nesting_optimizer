package user

import (
	"context"
	"errors"
	domainuser "server_nesting_optimizer/internal/domain/user"
	"server_nesting_optimizer/pkg/apperror"
)

type DeleteCurrentUserUseCase struct {
	repo   UserRepository
	hasher PasswordHasher
	uow    UnitOfWork
}

func NewDeleteCurrentUserUseCase(
	repo UserRepository,
	hasher PasswordHasher,
	unitOfWork UnitOfWork,
) *DeleteCurrentUserUseCase {
	return &DeleteCurrentUserUseCase{
		repo:   repo,
		hasher: hasher,
		uow:    unitOfWork,
	}
}

func (uc *DeleteCurrentUserUseCase) Execute(
	ctx context.Context,
	input DeleteCurrentUserInput,
	userID int64,
) error {
	details := input.Validate()
	if len(details) > 0 {
		return apperror.Validation(
			"validation failed",
			details...,
		)
	}

	user, err := uc.repo.GetByID(
		ctx,
		userID,
	)
	if err != nil {
		if errors.Is(err, domainuser.ErrNotFound) {
			return apperror.NotFound(
				"user not found",
			)
		}

		return apperror.Internal(
			"failed to get user",
			err,
		)
	}

	if err := uc.hasher.Compare(
		user.PasswordHash,
		input.Password,
	); err != nil {
		return apperror.Unauthorized(
			"invalid current password",
		)
	}

	err = uc.uow.WithinTransaction(
		ctx,
		func(repositories TransactionRepositories) error {
			if err := repositories.Users.SoftDelete(
				ctx,
				userID,
				user.PasswordHash,
			); err != nil {
				return err
			}

			if err := repositories.Sessions.DeleteByUserID(
				ctx,
				userID,
			); err != nil {
				return err
			}

			return nil
		},
	)

	if err != nil {
		switch {
		case errors.Is(err, domainuser.ErrUserChanged):
			return apperror.Conflict(
				"user was changed, please try again",
			)

		default:
			return apperror.Internal(
				"failed to delete user",
				err,
			)
		}
	}

	return nil
}
