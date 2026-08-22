package surface

import (
	"context"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	domainsurface "server_nesting_optimizer/internal/domain/surface"
)

type SurfaceRepository interface {
	Create(
		ctx context.Context,
		surface domainsurface.Surface,
	) (domainsurface.Surface, error)

	GetByID(
		ctx context.Context,
		surfaceID int64,
		userID int64,
	) (domainsurface.Surface, error)

	ListByUserID(
		ctx context.Context,
		userID int64,
		limit int,
		offset int,
	) (SurfaceListResult, error)

	Update(
		ctx context.Context,
		surfaceID int64,
		userID int64,
		name *string,
		geometry *domaingeometry.Polygon,
	) (domainsurface.Surface, error)

	SoftDelete(
		ctx context.Context,
		surfaceID int64,
		userID int64,
	) error
}
