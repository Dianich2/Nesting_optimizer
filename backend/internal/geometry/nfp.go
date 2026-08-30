package geometry

import domaingeometry "server_nesting_optimizer/internal/domain/geometry"

type NFPBuilder interface {
	BuildExternal(
		stationary domaingeometry.Polygon,
		moving domaingeometry.Polygon,
	) (domaingeometry.MultiPolygon, error)
}
