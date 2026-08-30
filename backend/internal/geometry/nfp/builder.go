package nfp

import (
	"fmt"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
)

var _ geometry.NFPBuilder = (*Builder)(nil)

type Builder struct {
	engine geometry.Engine
}

func NewBuilder(
	engine geometry.Engine,
) *Builder {
	return &Builder{
		engine: engine,
	}
}

func (b *Builder) BuildExternal(
	stationary domaingeometry.Polygon,
	moving domaingeometry.Polygon,
) (domaingeometry.MultiPolygon, error) {
	if err := b.engine.ValidatePolygon(stationary); err != nil {
		return domaingeometry.MultiPolygon{}, fmt.Errorf(
			"build external validate stationary: %w",
			err,
		)
	}

	if err := b.engine.ValidatePolygon(moving); err != nil {
		return domaingeometry.MultiPolygon{}, fmt.Errorf(
			"build external validate moving: %w",
			err,
		)
	}

	centeredMoving, err := b.engine.CenterAtOrigin(moving)
	if err != nil {
		return domaingeometry.MultiPolygon{}, fmt.Errorf(
			"build external center at origin moving: %w",
			err,
		)
	}

	reflectedMoving := reflectAtOrigin(centeredMoving)

	stationaryTriangles, err := triangulate(stationary)
	if err != nil {
		return domaingeometry.MultiPolygon{}, fmt.Errorf(
			"build external triangulate stationary: %w",
			err,
		)
	}

	movingTriangles, err := triangulate(reflectedMoving)
	if err != nil {
		return domaingeometry.MultiPolygon{}, fmt.Errorf(
			"build external triangulate moving: %w",
			err,
		)
	}

	pieces := make([]domaingeometry.Polygon, 0, len(stationaryTriangles)*len(movingTriangles))

	for _, stationaryTriangle := range stationaryTriangles {
		for _, movingTriangle := range movingTriangles {
			piece, err := convexMinkowskiSum(
				stationaryTriangle,
				movingTriangle,
			)
			if err != nil {
				return domaingeometry.MultiPolygon{}, fmt.Errorf(
					"build external convex Minkowski sum: %w",
					err,
				)
			}

			pieces = append(pieces, piece)
		}
	}

	result, err := unionPolygons(pieces)
	if err != nil {
		return domaingeometry.MultiPolygon{}, fmt.Errorf(
			"build external union polygons: %w",
			err,
		)
	}

	return result, nil
}
