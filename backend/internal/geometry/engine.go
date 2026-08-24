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

	Transform(
		polygon domaingeometry.Polygon,
		x float64,
		y float64,
		rotation float64,
	) (domaingeometry.Polygon, error)

	CoveredBy(
		inner domaingeometry.Polygon,
		outer domaingeometry.Polygon,
	) (bool, error)

	InteriorsIntersect(
		first domaingeometry.Polygon,
		second domaingeometry.Polygon,
	) (bool, error)
}
