package mapper

import (
	"server_nesting_optimizer/internal/transport/http/dto"
	projectsurfaceusecase "server_nesting_optimizer/internal/usecase/project_surface"
)

func ToCreateProjectSurfaceInput(
	req dto.CreateProjectSurfaceRequest,
	userID int64,
	projectID int64,
) projectsurfaceusecase.CreateProjectSurfaceInput {
	return projectsurfaceusecase.CreateProjectSurfaceInput{
		UserID:          userID,
		ProjectID:       projectID,
		SourceSurfaceID: req.SourceSurfaceID,
		Scale:           req.Scale,
	}
}

func ToCreateProjectSurfaceResponse(
	resp projectsurfaceusecase.CreateProjectSurfaceOutput,
) dto.CreateProjectSurfaceResponse {
	return dto.CreateProjectSurfaceResponse{
		ID:              resp.ID,
		ProjectID:       resp.ProjectID,
		SourceSurfaceID: resp.SourceSurfaceID,
		Name:            resp.Name,
		Geometry:        toGeometryPolygon(resp.Geometry),
		CreatedAt:       resp.CreatedAt,
		UpdatedAt:       resp.UpdatedAt,
	}
}

func ToGetProjectSurfaceByIDInput(
	projectSurfaceID int64,
	userID int64,
	projectID int64,
) projectsurfaceusecase.GetProjectSurfaceByIDInput {
	return projectsurfaceusecase.GetProjectSurfaceByIDInput{
		UserID:           userID,
		ProjectID:        projectID,
		ProjectSurfaceID: projectSurfaceID,
	}
}

func ToGetProjectSurfaceByIDResponse(
	resp projectsurfaceusecase.GetProjectSurfaceByIDOutput,
) dto.GetProjectSurfaceByIDResponse {
	return dto.GetProjectSurfaceByIDResponse{
		ID:              resp.ID,
		ProjectID:       resp.ProjectID,
		SourceSurfaceID: resp.SourceSurfaceID,
		Name:            resp.Name,
		Geometry:        toGeometryPolygon(resp.Geometry),
		CreatedAt:       resp.CreatedAt,
		UpdatedAt:       resp.UpdatedAt,
	}
}

func ToListProjectSurfacesInput(
	userID int64,
	projectID int64,
	page int,
	pageSize int,
) projectsurfaceusecase.ListProjectSurfacesInput {
	return projectsurfaceusecase.ListProjectSurfacesInput{
		UserID:    userID,
		ProjectID: projectID,
		Page:      page,
		PageSize:  pageSize,
	}
}

func ToListProjectSurfacesResponse(
	resp projectsurfaceusecase.ListProjectSurfacesOutput,
) dto.ListProjectSurfacesResponse {
	listProjectSurfacesResponse := dto.ListProjectSurfacesResponse{
		Items: make([]dto.ListProjectSurfacesItemResponse, 0),
	}

	for _, item := range resp.Items {
		curItem := dto.ListProjectSurfacesItemResponse{
			ID:              item.ID,
			ProjectID:       item.ProjectID,
			SourceSurfaceID: item.SourceSurfaceID,
			Name:            item.Name,
			Geometry:        toGeometryPolygon(item.Geometry),
			CreatedAt:       item.CreatedAt,
			UpdatedAt:       item.UpdatedAt,
		}

		listProjectSurfacesResponse.Items = append(
			listProjectSurfacesResponse.Items,
			curItem,
		)
	}

	listProjectSurfacesResponse.Page = resp.Page
	listProjectSurfacesResponse.PageSize = resp.PageSize
	listProjectSurfacesResponse.Total = resp.Total
	listProjectSurfacesResponse.TotalPages = resp.TotalPages

	return listProjectSurfacesResponse
}
