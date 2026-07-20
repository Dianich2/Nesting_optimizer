package user

import domainuser "server_nesting_optimizer/internal/domain/user"

func toCreateUserOutput(
	user domainuser.User,
) CreateUserOutput {
	return CreateUserOutput{
		ID:        user.ID,
		Login:     user.Login,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func toUser(
	input CreateUserInput,
	passwordHash string,
) domainuser.User {
	return domainuser.User{
		Login:        input.Login,
		Email:        input.Email,
		PasswordHash: passwordHash,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
	}
}

func toGetUserByIDOutput(
	user domainuser.User,
) GetCurrentUserOutput {
	return GetCurrentUserOutput{
		ID:        user.ID,
		Login:     user.Login,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
