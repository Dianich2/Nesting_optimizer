package nesting

import (
	"fmt"

	"server_nesting_optimizer/internal/geometry"
)

func buildResultFromDecoded(
	engine geometry.Engine,
	problem Problem,
	decoded DecodedChromosome,
) (Result, error) {
	surfaceArea, err := engine.Area(
		problem.Surface,
	)
	if err != nil {
		return Result{}, fmt.Errorf(
			"build result from decoded: %w",
			err,
		)
	}

	placedCountByPattern := make(
		map[int64]int,
	)

	for _, placement := range decoded.Placements {
		placedCountByPattern[placement.PatternID]++
	}

	unplaced := make(
		[]UnplacedPattern,
		0,
	)

	requestedCount := 0
	placedArea := 0.0

	for _, patternItem := range problem.Patterns {
		requestedCount += patternItem.Quantity

		placedCount := placedCountByPattern[patternItem.PatternID]

		remaining := patternItem.Quantity - placedCount

		if remaining > 0 {
			unplaced = append(
				unplaced,
				UnplacedPattern{
					PatternID: patternItem.PatternID,
					Quantity:  remaining,
				},
			)
		}

		patternArea, err := engine.Area(
			patternItem.Geometry,
		)
		if err != nil {
			return Result{}, fmt.Errorf(
				"build result from decoded: %w",
				err,
			)
		}

		placedArea += patternArea * float64(
			placedCount,
		)
	}

	result := Result{
		Placements: decoded.Placements,
		Unplaced:   unplaced,
		Metrics: Metrics{
			RequestedCount: requestedCount,
			PlacedCount:    len(decoded.Placements),
			SurfaceArea:    surfaceArea,
			PlacedArea:     placedArea,
		},
	}

	if surfaceArea > 0 {
		result.Metrics.Utilization =
			placedArea / surfaceArea
	}

	return result, nil
}
