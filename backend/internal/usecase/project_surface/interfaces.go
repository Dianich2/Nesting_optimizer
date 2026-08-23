package projectsurface

import (
	"context"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
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

	Update(
		ctx context.Context,
		projectSurfaceID int64,
		projectID int64,
		userID int64,
		name *string,
		geometry *domaingeometry.Polygon,
	) (domainprojectsurface.ProjectSurface, error)

	SoftDelete(
		ctx context.Context,
		projectSurfaceID int64,
		projectID int64,
		userID int64,
	) error
}
