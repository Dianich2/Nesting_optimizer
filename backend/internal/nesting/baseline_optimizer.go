package nesting

import (
	"context"
	"fmt"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
	"sort"
)

var _ Optimizer = (*BaselineOptimizer)(nil)

type BaselineOptimizer struct {
	engine geometry.Engine
}

func NewBaselineOptimizer(
	engine geometry.Engine,
) *BaselineOptimizer {
	return &BaselineOptimizer{
		engine: engine,
	}
}

type preparedPattern struct {
	item PatternItem
	area float64
}

func (o *BaselineOptimizer) Optimize(
	ctx context.Context,
	problem Problem,
) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, fmt.Errorf(
			"optimize: %w",
			err,
		)
	}

	err := validateProblem(
		o.engine,
		problem,
	)
	if err != nil {
		return Result{}, fmt.Errorf(
			"optimize: %w",
			err,
		)
	}

	surfaceArea, err := o.engine.Area(problem.Surface)
	if err != nil {
		return Result{}, fmt.Errorf(
			"optimize: %w",
			err,
		)
	}

	patterns := make([]preparedPattern, 0, len(problem.Patterns))

	for _, pattern := range problem.Patterns {
		patternArea, err := o.engine.Area(pattern.Geometry)
		if err != nil {
			return Result{}, fmt.Errorf(
				"optimize: %w",
				err,
			)
		}

		patterns = append(patterns, preparedPattern{
			item: pattern,
			area: patternArea,
		})
	}

	sort.SliceStable(patterns, func(i, j int) bool {
		return patterns[i].area > patterns[j].area
	})

	occupied := append(
		[]domaingeometry.Polygon(nil),
		problem.Obstacles...,
	)

	result := Result{
		Placements: make([]Placement, 0),
		Unplaced:   make([]UnplacedPattern, 0),
		Metrics: Metrics{
			SurfaceArea: surfaceArea,
		},
	}

	for _, preparedPattern := range patterns {
		if err := ctx.Err(); err != nil {
			return Result{}, fmt.Errorf(
				"optimize: %w",
				err,
			)
		}

		item := preparedPattern.item

		result.Metrics.RequestedCount += item.Quantity

		remaining := item.Quantity

		for remaining > 0 {
			if err := ctx.Err(); err != nil {
				return Result{}, fmt.Errorf(
					"optimize: %w",
					err,
				)
			}

			placement, found, err := placeOnePattern(
				ctx,
				o.engine,
				item.Geometry,
				problem.Surface,
				occupied,
				problem.AllowedRotations,
			)
			if err != nil {
				return Result{}, fmt.Errorf(
					"optimize: %w",
					err,
				)
			}

			if !found {
				result.Unplaced = append(
					result.Unplaced,
					UnplacedPattern{
						PatternID: item.PatternID,
						Quantity:  remaining,
					},
				)

				break
			}

			result.Placements = append(
				result.Placements,
				Placement{
					PatternID: item.PatternID,
					X:         placement.Candidate.X,
					Y:         placement.Candidate.Y,
					Rotation:  placement.Candidate.Rotation,
				},
			)

			occupied = append(
				occupied,
				placement.Geometry,
			)

			result.Metrics.PlacedCount++
			result.Metrics.PlacedArea += preparedPattern.area

			remaining--
		}
	}

	if surfaceArea > 0 {
		result.Metrics.Utilization = result.Metrics.PlacedArea / surfaceArea
	}

	return result, nil
}
