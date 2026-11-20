package nfp

import (
	"context"
	"errors"
	"fmt"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"

	"github.com/peterstace/simplefeatures/geom"
)

func unionPolygons(
	polygons []domaingeometry.Polygon,
) (domaingeometry.MultiPolygon, error) {
	return unionPolygonsContext(context.Background(), polygons)
}

func unionPolygonsContext(ctx context.Context, polygons []domaingeometry.Polygon) (domaingeometry.MultiPolygon, error) {
	if err := ctx.Err(); err != nil {
		return domaingeometry.MultiPolygon{}, err
	}
	if len(polygons) == 0 {
		return domaingeometry.MultiPolygon{}, fmt.Errorf(
			"union polygons: %w",
			errors.New("polygons must not be empty"),
		)
	}

	geometries := make([]geom.Geometry, 0, len(polygons))
	for _, polygon := range polygons {
		if err := ctx.Err(); err != nil {
			return domaingeometry.MultiPolygon{}, err
		}
		sfPolygon := toSimpleFeaturesPolygon(polygon)
		geometries = append(geometries, sfPolygon.AsGeometry())
	}

	// Bound each indivisible third-party union and check cancellation between batches.
	for len(geometries) > 1 {
		next := make([]geom.Geometry, 0, (len(geometries)+63)/64)
		for start := 0; start < len(geometries); start += 64 {
			if err := ctx.Err(); err != nil {
				return domaingeometry.MultiPolygon{}, err
			}
			part, err := geom.UnionMany(geometries[start:min(start+64, len(geometries))])
			if ctx.Err() != nil {
				return domaingeometry.MultiPolygon{}, ctx.Err()
			}
			if err != nil {
				return domaingeometry.MultiPolygon{}, fmt.Errorf("union polygons: %w", err)
			}
			next = append(next, part)
		}
		geometries = next
	}
	union := geometries[0]

	if union.IsPolygon() {
		return domaingeometry.MultiPolygon{
			Polygons: []domaingeometry.Polygon{
				fromSimpleFeaturesPolygon(
					union.MustAsPolygon(),
				),
			},
		}, nil
	}

	if union.IsMultiPolygon() {
		return fromSimpleFeaturesMultiPolygon(
			union.MustAsMultiPolygon(),
		), nil
	}

	return domaingeometry.MultiPolygon{}, fmt.Errorf(
		"union polygons: unexpected geometry type: %s",
		union.Type(),
	)
}
