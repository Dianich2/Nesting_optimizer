package project

import (
	"context"
	domainproject "server_nesting_optimizer/internal/domain/project"
)

type ProjectRepository interface {
	Create(
		ctx context.Context,
		project domainproject.Project,
	) (domainproject.Project, error)

	GetByID(
		ctx context.Context,
		projectID int64,
		userID int64,
	) (domainproject.Project, error)

	ListByUserID(
		ctx context.Context,
		userID int64,
		limit int,
		offset int,
	) (ProjectListResult, error)
}
