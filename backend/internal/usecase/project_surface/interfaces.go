package projectsurface

import (
	"context"
	domainprojectsurface "server_nesting_optimizer/internal/domain/project_surface"
)

type ProjectSurfaceRepository interface {
	Create(
		ctx context.Context,
		projectSurface domainprojectsurface.ProjectSurface,
		userID int64,
	) (domainprojectsurface.ProjectSurface, error)

	GetByID(
		ctx context.Context,
		userID int64,
		projectID int64,
		projectSurfaceID int64,
	) (domainprojectsurface.ProjectSurface, error)

	ListByProjectID(
		ctx context.Context,
		userID int64,
		projectID int64,
		limit int,
		offset int,
	) (ProjectSurfaceListResult, error)
}
