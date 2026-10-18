package nesting

import (
	"fmt"

	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
)

type placementGeometryProvider func(
	placement Placement,
	patternItem PatternItem,
	index int,
) (domaingeometry.Polygon, error)

func calculatePlacementsFitness(
	engine geometry.Engine,
	problem Problem,
	placements []Placement,
) (Fitness, error) {
	return calculateFitnessWithGeometry(
		engine,
		problem,
		placements,
		func(
			placement Placement,
			patternItem PatternItem,
			_ int,
		) (domaingeometry.Polygon, error) {
			return engine.Transform(
				patternItem.Geometry,
				placement.X,
				placement.Y,
				placement.Rotation,
			)
		},
		"calculate placements fitness",
	)
}

func calculatePlacedGeometriesFitness(
	engine geometry.Engine,
	problem Problem,
	placements []Placement,
	placedGeometries []domaingeometry.Polygon,
) (Fitness, error) {
	if len(placements) != len(placedGeometries) {
		return Fitness{}, fmt.Errorf(
			"calculate placed geometries fitness: placements and geometries count mismatch",
		)
	}

	return calculateFitnessWithGeometry(
		engine,
		problem,
		placements,
		func(
			_ Placement,
			_ PatternItem,
			index int,
		) (domaingeometry.Polygon, error) {
			return placedGeometries[index], nil
		},
		"calculate placed geometries fitness",
	)
}

func calculateFitnessWithGeometry(
	engine geometry.Engine,
	problem Problem,
	placements []Placement,
	geometryProvider placementGeometryProvider,
	errorContext string,
) (Fitness, error) {
	requestedCount := 0

	for _, patternItem := range problem.Patterns {
		requestedCount += patternItem.Quantity
	}

	if len(placements) > requestedCount {
		return Fitness{}, fmt.Errorf(
			"%s: placed count exceeds requested count",
			errorContext,
		)
	}

	fitness := Fitness{
		UnplacedCount: requestedCount - len(placements),
	}

	if len(placements) == 0 {
		return fitness, nil
	}

	surfaceBounds, err := engine.Bounds(problem.Surface)
	if err != nil {
		return Fitness{}, fmt.Errorf(
			"%s: %w",
			errorContext,
			err,
		)
	}

	patternLookup := buildPatternLookup(problem.Patterns)

	maxX := surfaceBounds.MinX
	maxY := surfaceBounds.MinY

	for index, placement := range placements {
		patternItem, ok := patternLookup[placement.PatternID]
		if !ok {
			return Fitness{}, fmt.Errorf(
				"%s: pattern item %d not found",
				errorContext,
				placement.PatternID,
			)
		}

		area, err := engine.Area(patternItem.Geometry)
		if err != nil {
			return Fitness{}, fmt.Errorf(
				"%s: %w",
				errorContext,
				err,
			)
		}

		fitness.PlacedArea += area

		transformed, err := geometryProvider(
			placement,
			patternItem,
			index,
		)
		if err != nil {
			return Fitness{}, fmt.Errorf(
				"%s: %w",
				errorContext,
				err,
			)
		}

		bounds, err := engine.Bounds(transformed)
		if err != nil {
			return Fitness{}, fmt.Errorf(
				"%s: %w",
				errorContext,
				err,
			)
		}

		if bounds.MaxX > maxX {
			maxX = bounds.MaxX
		}

		if bounds.MaxY > maxY {
			maxY = bounds.MaxY
		}
	}

	usedWidth := maxX - surfaceBounds.MinX
	usedHeight := maxY - surfaceBounds.MinY

	fitness.UsedArea = usedWidth * usedHeight
	fitness.UsedHeight = usedHeight

	return fitness, nil
}
