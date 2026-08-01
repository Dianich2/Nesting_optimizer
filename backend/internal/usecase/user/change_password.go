package user

import (
	"context"
	"errors"
	"fmt"
	domainuser "server_nesting_optimizer/internal/domain/user"
	"server_nesting_optimizer/pkg/apperror"
)

type ChangePasswordUseCase struct {
	repo   UserRepository
	hasher PasswordHasher
	uow    UnitOfWork
}

func NewChangePasswordUseCase(
	repo UserRepository,
	hasher PasswordHasher,
	unitOfWork UnitOfWork,
) *ChangePasswordUseCase {
	return &ChangePasswordUseCase{
		repo:   repo,
		hasher: hasher,
		uow:    unitOfWork,
	}
}

func (uc *ChangePasswordUseCase) Execute(
	ctx context.Context,
	input ChangePasswordInput,
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
		return fmt.Errorf("get user by id: %w", err)
	}

	if err := uc.hasher.Compare(
		user.PasswordHash,
		input.OldPassword,
	); err != nil {
		return apperror.Unauthorized(
			"invalid current password",
		)
	}

	if input.OldPassword == input.NewPassword {
		return nil
	}

	newPasswordHash, err := uc.hasher.Hash(
		input.NewPassword,
	)
	if err != nil {
		return apperror.Internal(
			"failed to hash user new password",
			err,
		)
	}

	err = uc.uow.WithinTransaction(
		ctx,
		func(repositories TransactionRepositories) error {
			if err := repositories.Users.ChangePassword(
				ctx,
				userID,
				user.PasswordHash,
				newPasswordHash,
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
		case errors.Is(err, domainuser.ErrPasswordChanged):
			return apperror.Conflict(
				"user password was changed, please try again",
			)

		default:
			return apperror.Internal(
				"failed to change password",
				err,
			)
		}
	}

	return nil
}
