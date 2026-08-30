package nfp

import (
	"errors"
	"fmt"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"

	"github.com/peterstace/simplefeatures/geom"
)

func unionPolygons(
	polygons []domaingeometry.Polygon,
) (domaingeometry.MultiPolygon, error) {
	if len(polygons) == 0 {
		return domaingeometry.MultiPolygon{}, fmt.Errorf(
			"union polygons: %w",
			errors.New("polygons must not be empty"),
		)
	}

	geometries := make([]geom.Geometry, 0, len(polygons))
	for _, polygon := range polygons {
		sfPolygon := toSimpleFeaturesPolygon(polygon)
		geometries = append(geometries, sfPolygon.AsGeometry())
	}

	union, err := geom.UnionMany(geometries)
	if err != nil {
		return domaingeometry.MultiPolygon{}, fmt.Errorf(
			"union polygons: %w",
			err,
		)
	}

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
