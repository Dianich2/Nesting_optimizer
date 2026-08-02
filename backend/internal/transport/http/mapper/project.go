package mapper

import (
	"server_nesting_optimizer/internal/transport/http/dto"
	projectusecase "server_nesting_optimizer/internal/usecase/project"
)

func ToCreateProjectInput(
	req dto.CreateProjectRequest,
	userID int64,
) projectusecase.CreateProjectInput {
	return projectusecase.CreateProjectInput{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
	}
}

func ToCreateProjectResponse(
	resp projectusecase.CreateProjectOutput,
) dto.CreateProjectResponse {
	return dto.CreateProjectResponse{
		ID:          resp.ID,
		UserID:      resp.UserID,
		Name:        resp.Name,
		Description: resp.Description,
		CreatedAt:   resp.CreatedAt,
		UpdatedAt:   resp.UpdatedAt,
	}
}
