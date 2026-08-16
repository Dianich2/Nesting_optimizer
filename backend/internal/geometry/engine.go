package geometry

import (
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
)

type Engine interface {
	ValidatePolygon(
		polygon domaingeometry.Polygon,
	) error

	Area(
		polygon domaingeometry.Polygon,
	) (float64, error)

	Scale(
		polygon domaingeometry.Polygon,
		factor float64,
	) (domaingeometry.Polygon, error)

	Normalize(
		polygon domaingeometry.Polygon,
	) (domaingeometry.Polygon, error)
}
