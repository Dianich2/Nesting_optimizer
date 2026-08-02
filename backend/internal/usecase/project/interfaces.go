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
}
