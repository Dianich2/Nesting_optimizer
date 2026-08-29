package nesting

import (
	"context"
	domainplacement "server_nesting_optimizer/internal/domain/placement"
	domainprojectpattern "server_nesting_optimizer/internal/domain/project_pattern"
	domainprojectsurface "server_nesting_optimizer/internal/domain/project_surface"
)

type PlacementRepository interface {
	ListForCollisionCheck(
		ctx context.Context,
		projectSurfaceID int64,
		projectID int64,
		userID int64,
	) ([]domainplacement.CollisionPlacement, error)
}

type PlacementWriter interface {
	Create(
		ctx context.Context,
		input domainplacement.Placement,
		projectID int64,
		userID int64,
	) (domainplacement.Placement, error)

	DeleteByProjectSurface(
		ctx context.Context,
		userID int64,
		projectID int64,
		projectSurfaceID int64,
	) error
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
	GetByIDs(
		ctx context.Context,
		userID int64,
		projectID int64,
		ids []int64,
	) ([]domainprojectpattern.ProjectPattern, error)
}

type TransactionRepositories struct {
	Placements PlacementWriter
}

type UnitOfWork interface {
	WithinTransaction(
		ctx context.Context,
		fn func(repositories TransactionRepositories) error,
	) error
}
