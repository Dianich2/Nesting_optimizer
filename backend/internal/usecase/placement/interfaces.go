package placement

import (
	"context"
	domainplacement "server_nesting_optimizer/internal/domain/placement"
	domainprojectpattern "server_nesting_optimizer/internal/domain/project_pattern"
	domainprojectsurface "server_nesting_optimizer/internal/domain/project_surface"
)

type PlacementRepository interface {
	Create(
		ctx context.Context,
		placement domainplacement.Placement,
		projectID int64,
		userID int64,
	) (domainplacement.Placement, error)

	ListForCollisionCheck(
		ctx context.Context,
		projectSurfaceID int64,
		projectID int64,
		userID int64,
	) ([]CollisionPlacement, error)

	GetByIDWithPatternGeometry(
		ctx context.Context,
		userID int64,
		projectID int64,
		placementID int64,
	) (PlacementWithPatternGeometry, error)
}

type ProjectSurfaceRepository interface {
	GetByID(
		ctx context.Context,
		userID int64,
		projectID int64,
		projectSurfaceID int64,
	) (domainprojectsurface.ProjectSurface, error)
}

type ProjectPatternRepository interface {
	GetByID(
		ctx context.Context,
		userID int64,
		projectID int64,
		projectPatternID int64,
	) (domainprojectpattern.ProjectPattern, error)
}
