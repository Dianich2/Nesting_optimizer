package surface

import (
	"context"
	domainsurface "server_nesting_optimizer/internal/domain/surface"
)

type SurfaceRepository interface {
	Create(
		ctx context.Context,
		surface domainsurface.Surface,
	) (domainsurface.Surface, error)
}
