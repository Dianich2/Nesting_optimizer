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

func ToLoginInput(
	req dto.LoginRequest,
) userusecase.LoginInput {
	return userusecase.LoginInput{
		Identifier: req.Identifier,
		Password:   req.Password,
	}
}

func ToLoginResponse(
	resp userusecase.LoginOutput,
) dto.LoginResponse {
	return dto.LoginResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}
}

func ToRefreshInput(
	req dto.RefreshRequest,
) userusecase.RefreshInput {
	return userusecase.RefreshInput{
		RefreshToken: req.RefreshToken,
	}
}

func ToRefreshResponse(
	resp userusecase.RefreshOutput,
) dto.RefreshResponse {
	return dto.RefreshResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}
}

func ToGetCurrentUserInput(
	userID int64,
) userusecase.GetCurrentUserInput {
	return userusecase.GetCurrentUserInput{
		ID: userID,
	}
}

func ToGetCurrentUserResponse(
	resp userusecase.GetCurrentUserOutput,
) dto.GetCurrentUserResponse {
	return dto.GetCurrentUserResponse{
		ID:        resp.ID,
		Login:     resp.Login,
		Email:     resp.Email,
		FirstName: resp.FirstName,
		LastName:  resp.LastName,
		CreatedAt: resp.CreatedAt,
		UpdatedAt: resp.UpdatedAt,
	}
}

func ToUpdateProfileInput(
	req dto.UpdateProfileRequest,
) userusecase.UpdateProfileInput {
	return userusecase.UpdateProfileInput{
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}
}

func ToUpdateProfileResponse(
	resp userusecase.UpdateProfileOutput,
) dto.UpdateProfileResponse {
	return dto.UpdateProfileResponse{
		ID:        resp.ID,
		Login:     resp.Login,
		Email:     resp.Email,
		FirstName: resp.FirstName,
		LastName:  resp.LastName,
		CreatedAt: resp.CreatedAt,
		UpdatedAt: resp.UpdatedAt,
	}
}
