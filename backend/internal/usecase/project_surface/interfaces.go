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
}
