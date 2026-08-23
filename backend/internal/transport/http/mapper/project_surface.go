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
