package projectpattern

import (
	"context"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	domainprojectpattern "server_nesting_optimizer/internal/domain/project_pattern"
)

type ProjectPatternRepository interface {
	Create(
		ctx context.Context,
		projectPattern domainprojectpattern.ProjectPattern,
		userID int64,
	) (domainprojectpattern.ProjectPattern, error)

	GetByID(
		ctx context.Context,
		userID int64,
		projectID int64,
		projectPatternID int64,
	) (domainprojectpattern.ProjectPattern, error)

	ListByProjectID(
		ctx context.Context,
		userID int64,
		projectID int64,
		limit int,
		offset int,
	) (ProjectPatternListResult, error)

	Update(
		ctx context.Context,
		projectPatternID int64,
		projectID int64,
		userID int64,
		name *string,
		geometry *domaingeometry.Polygon,
	) (domainprojectpattern.ProjectPattern, error)

	SoftDelete(
		ctx context.Context,
		projectPatternID int64,
		projectID int64,
		userID int64,
	) error
}
