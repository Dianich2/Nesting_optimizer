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

func ToGetProjectByIDInput(
	projectID int64,
	userID int64,
) projectusecase.GetProjectByIDInput {
	return projectusecase.GetProjectByIDInput{
		UserID:    userID,
		ProjectID: projectID,
	}
}

func ToGetProjectByIDResponse(
	resp projectusecase.GetProjectByIDOutput,
) dto.GetProjectByIDResponse {
	return dto.GetProjectByIDResponse{
		ID:          resp.ID,
		UserID:      resp.UserID,
		Name:        resp.Name,
		Description: resp.Description,
		CreatedAt:   resp.CreatedAt,
		UpdatedAt:   resp.UpdatedAt,
	}
}

func ToListProjectsInput(
	userID int64,
	page int,
	pageSize int,
) projectusecase.ListProjectsInput {
	return projectusecase.ListProjectsInput{
		UserID:   userID,
		Page:     page,
		PageSize: pageSize,
	}
}

func ToListProjectsResponse(
	resp projectusecase.ListProjectsOutput,
) dto.ListProjectsResponse {
	listProjectsResponse := dto.ListProjectsResponse{
		Items: make([]dto.ListProjectsItemResponse, 0),
	}

	for _, item := range resp.Items {
		curItem := dto.ListProjectsItemResponse{
			ID:          item.ID,
			UserID:      item.UserID,
			Name:        item.Name,
			Description: item.Description,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		}

		listProjectsResponse.Items = append(
			listProjectsResponse.Items,
			curItem,
		)
	}

	listProjectsResponse.Page = resp.Page
	listProjectsResponse.PageSize = resp.PageSize
	listProjectsResponse.Total = resp.Total
	listProjectsResponse.TotalPages = resp.TotalPages

	return listProjectsResponse
}

func ToUpdateProjectInput(
	req dto.UpdateProjectRequest,
	projectID int64,
	userID int64,
) projectusecase.UpdateProjectInput {
	return projectusecase.UpdateProjectInput{
		UserID:      userID,
		ProjectID:   projectID,
		Name:        req.Name,
		Description: req.Description,
	}
}

func ToUpdateProjectResponse(
	resp projectusecase.UpdateProjectOutput,
) dto.UpdateProjectResponse {
	return dto.UpdateProjectResponse{
		ID:          resp.ID,
		UserID:      resp.UserID,
		Name:        resp.Name,
		Description: resp.Description,
		CreatedAt:   resp.CreatedAt,
		UpdatedAt:   resp.UpdatedAt,
	}
}

func ToDeleteProjectInput(
	projectID int64,
	userID int64,
) projectusecase.DeleteProjectInput {
	return projectusecase.DeleteProjectInput{
		ProjectID: projectID,
		UserID:    userID,
	}
}
