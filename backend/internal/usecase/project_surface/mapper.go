package projectsurface

import domainprojectsurface "server_nesting_optimizer/internal/domain/project_surface"

func toCreateProjectSurfaceOutput(
	projectSurface domainprojectsurface.ProjectSurface,
) CreateProjectSurfaceOutput {
	return CreateProjectSurfaceOutput{
		ID:              projectSurface.ID,
		ProjectID:       projectSurface.ProjectID,
		SourceSurfaceID: *projectSurface.SourceSurfaceID,
		Name:            projectSurface.Name,
		Geometry:        projectSurface.Geometry,
		CreatedAt:       projectSurface.CreatedAt,
		UpdatedAt:       projectSurface.UpdatedAt,
	}
}
