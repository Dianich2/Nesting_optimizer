package nesting

import (
	"context"
	"fmt"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
	"sort"
)

type NFPGreedyOptimizer struct {
	engine     geometry.Engine
	nfpBuilder geometry.NFPBuilder
}

var _ Optimizer = (*NFPGreedyOptimizer)(nil)

func NewNFPGreedyOptimizer(
	engine geometry.Engine,
	nfpBuilder geometry.NFPBuilder,
) *NFPGreedyOptimizer {
	return &NFPGreedyOptimizer{
		engine:     engine,
		nfpBuilder: nfpBuilder,
	}
}

func (o *NFPGreedyOptimizer) Optimize(
	ctx context.Context,
	problem Problem,
) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, fmt.Errorf(
			"nfp greedy optimizer optimize: %w",
			err,
		)
	}

	err := validateProblem(
		o.engine,
		problem,
	)
	if err != nil {
		return Result{}, fmt.Errorf(
			"nfp greedy optimizer optimize: %w",
			err,
		)
	}

	surfaceArea, err := o.engine.Area(problem.Surface)
	if err != nil {
		return Result{}, fmt.Errorf(
			"nfp greedy optimizer optimize: %w",
			err,
		)
	}

	preparedPatterns := make([]preparedPattern, 0, len(problem.Patterns))

	for _, patternItem := range problem.Patterns {
		area, err := o.engine.Area(patternItem.Geometry)
		if err != nil {
			return Result{}, fmt.Errorf(
				"nfp greedy optimizer optimize: %w",
				err,
			)
		}

		preparedPatterns = append(preparedPatterns, preparedPattern{
			item: patternItem,
			area: area,
		})
	}

	sort.SliceStable(preparedPatterns, func(i, j int) bool {
		return preparedPatterns[i].area > preparedPatterns[j].area
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

	for _, prepreparedPattern := range preparedPatterns {
		if err := ctx.Err(); err != nil {
			return Result{}, fmt.Errorf(
				"nfp greedy optimizer optimize: %w",
				err,
			)
		}

		item := prepreparedPattern.item
		result.Metrics.RequestedCount += item.Quantity

		remaining := item.Quantity

		for remaining > 0 {
			if err := ctx.Err(); err != nil {
				return Result{}, fmt.Errorf(
					"nfp greedy optimizer optimize: %w",
					err,
				)
			}

			placement, found, err := placeOnePatternNFP(
				ctx,
				o.engine,
				o.nfpBuilder,
				item.Geometry,
				problem.Surface,
				occupied,
				problem.AllowedRotations,
			)
			if err != nil {
				return Result{}, fmt.Errorf(
					"nfp greedy optimizer optimize: %w",
					err,
				)
			}

			if !found {
				result.Unplaced = append(result.Unplaced, UnplacedPattern{
					PatternID: item.PatternID,
					Quantity:  remaining,
				})

				break
			}

			result.Placements = append(result.Placements, Placement{
				PatternID: item.PatternID,
				X:         placement.Candidate.X,
				Y:         placement.Candidate.Y,
				Rotation:  placement.Candidate.Rotation,
			})

			occupied = append(occupied, placement.Geometry)
			result.Metrics.PlacedCount++
			result.Metrics.PlacedArea += prepreparedPattern.area

			remaining--
		}
	}

	if surfaceArea > 0 {
		result.Metrics.Utilization = result.Metrics.PlacedArea / result.Metrics.SurfaceArea
	}

	return result, nil
}
