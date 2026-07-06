package mapper

import (
	"server_nesting_optimizer/internal/transport/http/dto"
	userusecase "server_nesting_optimizer/internal/usecase/user"
)

func ToCreateUserInput(
	req dto.CreateUserRequest,
) userusecase.CreateUserInput {
	return userusecase.CreateUserInput{
		Login:     req.Login,
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}
}

func ToCreateUserResponse(
	resp userusecase.CreateUserOutput,
) dto.CreateUserResponse {
	return dto.CreateUserResponse{
		ID:        resp.ID,
		Login:     resp.Login,
		Email:     resp.Email,
		FirstName: resp.FirstName,
		LastName:  resp.LastName,
		CreatedAt: resp.CreatedAt,
		UpdatedAt: resp.UpdatedAt,
	}
}
