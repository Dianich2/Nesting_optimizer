package nfp

import (
	"context"
	"fmt"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
)

var _ geometry.NFPBuilder = (*Builder)(nil)
var _ geometry.ContextNFPBuilder = (*Builder)(nil)

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
	return b.BuildExternalContext(context.Background(), stationary, moving)
}

func (b *Builder) BuildExternalContext(ctx context.Context, stationary, moving domaingeometry.Polygon) (domaingeometry.MultiPolygon, error) {
	if err := ctx.Err(); err != nil {
		return domaingeometry.MultiPolygon{}, err
	}
	if err := checkComplexity(stationary, moving); err != nil {
		return domaingeometry.MultiPolygon{}, err
	}
	if err := b.engine.ValidatePolygon(stationary); err != nil {
		return domaingeometry.MultiPolygon{}, fmt.Errorf(
			"build external validate stationary: %w",
			err,
		)
	}

	if err := ctx.Err(); err != nil {
		return domaingeometry.MultiPolygon{}, err
	}
	if err := b.engine.ValidatePolygon(moving); err != nil {
		return domaingeometry.MultiPolygon{}, fmt.Errorf(
			"build external validate moving: %w",
			err,
		)
	}

	if err := ctx.Err(); err != nil {
		return domaingeometry.MultiPolygon{}, err
	}
	centeredMoving, err := b.engine.CenterAtOrigin(moving)
	if err != nil {
		return domaingeometry.MultiPolygon{}, fmt.Errorf(
			"build external center at origin moving: %w",
			err,
		)
	}

	reflectedMoving := reflectAtOrigin(centeredMoving)

	stationaryTriangles, err := triangulateContext(ctx, stationary)
	if err != nil {
		return domaingeometry.MultiPolygon{}, fmt.Errorf(
			"build external triangulate stationary: %w",
			err,
		)
	}

	movingTriangles, err := triangulateContext(ctx, reflectedMoving)
	if err != nil {
		return domaingeometry.MultiPolygon{}, fmt.Errorf(
			"build external triangulate moving: %w",
			err,
		)
	}

	if len(movingTriangles) != 0 && len(stationaryTriangles) > maxMinkowskiPieces/len(movingTriangles) {
		return domaingeometry.MultiPolygon{}, ErrComplexity
	}
	pieces := make([]domaingeometry.Polygon, 0, len(stationaryTriangles)*len(movingTriangles))

	for _, stationaryTriangle := range stationaryTriangles {
		if err := ctx.Err(); err != nil {
			return domaingeometry.MultiPolygon{}, err
		}
		for _, movingTriangle := range movingTriangles {
			if len(pieces)%64 == 0 {
				if err := ctx.Err(); err != nil {
					return domaingeometry.MultiPolygon{}, err
				}
			}
			piece, err := convexMinkowskiSumContext(ctx,
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

	result, err := unionPolygonsContext(ctx, pieces)
	if err != nil {
		return domaingeometry.MultiPolygon{}, fmt.Errorf(
			"build external union polygons: %w",
			err,
		)
	}

	return result, nil
}
