package mapper

import (
	"server_nesting_optimizer/internal/transport/http/dto"
	projectpatternusecase "server_nesting_optimizer/internal/usecase/project_pattern"
)

func ToCreateProjectPatternInput(
	req dto.CreateProjectPatternRequest,
	userID int64,
	projectID int64,
) projectpatternusecase.CreateProjectPatternInput {
	return projectpatternusecase.CreateProjectPatternInput{
		UserID:          userID,
		ProjectID:       projectID,
		SourcePatternID: req.SourcePatternID,
		Scale:           req.Scale,
	}
}

func ToCreateProjectPatternResponse(
	resp projectpatternusecase.CreateProjectPatternOutput,
) dto.CreateProjectPatternResponse {
	return dto.CreateProjectPatternResponse{
		ID:              resp.ID,
		ProjectID:       resp.ProjectID,
		SourcePatternID: resp.SourcePatternID,
		Name:            resp.Name,
		Geometry:        toGeometryPolygon(resp.Geometry),
		CreatedAt:       resp.CreatedAt,
		UpdatedAt:       resp.UpdatedAt,
	}
}

func ToGetProjectPatternByIDInput(
	projectPatternID int64,
	userID int64,
	projectID int64,
) projectpatternusecase.GetProjectPatternByIDInput {
	return projectpatternusecase.GetProjectPatternByIDInput{
		UserID:           userID,
		ProjectID:        projectID,
		ProjectPatternID: projectPatternID,
	}
}

func ToGetProjectPatternByIDResponse(
	resp projectpatternusecase.GetProjectPatternByIDOutput,
) dto.GetProjectPatternByIDResponse {
	return dto.GetProjectPatternByIDResponse{
		ID:              resp.ID,
		ProjectID:       resp.ProjectID,
		SourcePatternID: resp.SourcePatternID,
		Name:            resp.Name,
		Geometry:        toGeometryPolygon(resp.Geometry),
		CreatedAt:       resp.CreatedAt,
		UpdatedAt:       resp.UpdatedAt,
	}
}

func ToListProjectPatternsInput(
	userID int64,
	projectID int64,
	page int,
	pageSize int,
) projectpatternusecase.ListProjectPatternsInput {
	return projectpatternusecase.ListProjectPatternsInput{
		UserID:    userID,
		ProjectID: projectID,
		Page:      page,
		PageSize:  pageSize,
	}
}

func ToListProjectPatternsResponse(
	resp projectpatternusecase.ListProjectPatternsOutput,
) dto.ListProjectPatternsResponse {
	listProjectPatternsResponse := dto.ListProjectPatternsResponse{
		Items: make([]dto.ListProjectPatternsItemResponse, 0),
	}

	for _, item := range resp.Items {
		curItem := dto.ListProjectPatternsItemResponse{
			ID:              item.ID,
			ProjectID:       item.ProjectID,
			SourcePatternID: item.SourcePatternID,
			Name:            item.Name,
			Geometry:        toGeometryPolygon(item.Geometry),
			CreatedAt:       item.CreatedAt,
			UpdatedAt:       item.UpdatedAt,
		}

		listProjectPatternsResponse.Items = append(
			listProjectPatternsResponse.Items,
			curItem,
		)
	}

	listProjectPatternsResponse.Page = resp.Page
	listProjectPatternsResponse.PageSize = resp.PageSize
	listProjectPatternsResponse.Total = resp.Total
	listProjectPatternsResponse.TotalPages = resp.TotalPages

	return listProjectPatternsResponse
}

func ToUpdateProjectPatternInput(
	req dto.UpdateProjectPatternRequest,
	userID int64,
	projectID int64,
	projectPatternID int64,
) projectpatternusecase.UpdateProjectPatternInput {
	return projectpatternusecase.UpdateProjectPatternInput{
		UserID:           userID,
		ProjectID:        projectID,
		ProjectPatternID: projectPatternID,
		Name:             req.Name,
		Scale:            req.Scale,
	}
}

func ToUpdateProjectPatternResponse(
	resp projectpatternusecase.UpdateProjectPatternOutput,
) dto.UpdateProjectPatternResponse {
	return dto.UpdateProjectPatternResponse{
		ID:              resp.ID,
		ProjectID:       resp.ProjectID,
		SourcePatternID: resp.SourcePatternID,
		Name:            resp.Name,
		Geometry:        toGeometryPolygon(resp.Geometry),
		CreatedAt:       resp.CreatedAt,
		UpdatedAt:       resp.UpdatedAt,
	}
}

func ToDeleteProjectPatternInput(
	userID int64,
	projectID int64,
	projectPatternID int64,
) projectpatternusecase.DeleteProjectPatternInput {
	return projectpatternusecase.DeleteProjectPatternInput{
		UserID:           userID,
		ProjectID:        projectID,
		ProjectPatternID: projectPatternID,
	}
}
