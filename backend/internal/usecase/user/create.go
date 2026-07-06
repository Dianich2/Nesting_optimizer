package user

import (
	"context"
	"server_nesting_optimizer/pkg/apperror"
)

type CreateUserUseCase struct {
	repo   UserRepository
	hasher PasswordHasher
}

func NewCreateUserUseCase(
	repo UserRepository,
	hasher PasswordHasher,
) *CreateUserUseCase {
	return &CreateUserUseCase{
		repo:   repo,
		hasher: hasher,
	}
}

func (uc *CreateUserUseCase) Execute(
	ctx context.Context,
	input CreateUserInput,
) (CreateUserOutput, error) {
	details := validateCreateUserInput(input)
	if len(details) > 0 {
		return CreateUserOutput{}, apperror.Validation("validation failed", details...)
	}

	loginExists, err := uc.repo.ExistsByLogin(ctx, input.Login)
	if err != nil {
		return CreateUserOutput{}, apperror.Internal("failed to check login uniqueness", err)
	}
	if loginExists {
		return CreateUserOutput{}, apperror.Conflict("login already exists")
	}

	emailExists, err := uc.repo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return CreateUserOutput{}, apperror.Internal("failed to check email uniqueness", err)
	}
	if emailExists {
		return CreateUserOutput{}, apperror.Conflict("email already exists")
	}

	passwordHash, err := uc.hasher.Hash(input.Password)
	if err != nil {
		return CreateUserOutput{}, apperror.Internal("failed to hash user password", err)
	}

	domainUser := toUser(input, passwordHash)

	createdUser, err := uc.repo.Create(ctx, domainUser)
	if err != nil {
		return CreateUserOutput{}, apperror.Internal("failed to create user", err)
	}

	return toCreateUserOutput(createdUser), nil
}
