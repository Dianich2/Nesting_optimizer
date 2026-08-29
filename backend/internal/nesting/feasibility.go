package nesting

import (
	"fmt"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
)

func isPlacementFeasible(
	engine geometry.Engine,
	transformedPattern domaingeometry.Polygon,
	surface domaingeometry.Polygon,
	occupied []domaingeometry.Polygon,
) (bool, error) {
	covered, err := engine.CoveredBy(
		transformedPattern,
		surface,
	)
	if err != nil {
		return false, fmt.Errorf(
			"check is placement feasible: %w",
			err,
		)
	}

	if !covered {
		return false, nil
	}

	for _, polygon := range occupied {
		intersects, err := engine.InteriorsIntersect(
			transformedPattern,
			polygon,
		)
		if err != nil {
			return false, fmt.Errorf(
				"check is placement feasible: %w",
				err,
			)
		}

		if intersects {
			return false, nil
		}
	}

	return true, nil
}
