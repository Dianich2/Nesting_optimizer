package nesting

import (
	"context"
	"fmt"

	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
)

type compactionAxis int

const (
	compactionAxisX compactionAxis = iota
	compactionAxisY
)

func compactPlacements(
	ctx context.Context,
	engine geometry.Engine,
	problem Problem,
	placements []Placement,
	config CompactionConfig,
) ([]Placement, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf(
			"compact placements: %w",
			err,
		)
	}

	if err := validateCompactionConfig(config); err != nil {
		return nil, fmt.Errorf(
			"compact placements: %w",
			err,
		)
	}

	if len(placements) == 0 {
		return []Placement{}, nil
	}

	surfaceBounds, err := engine.Bounds(
		problem.Surface,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"compact placements: %w",
			err,
		)
	}

	surfaceWidth :=
		surfaceBounds.MaxX - surfaceBounds.MinX

	surfaceHeight :=
		surfaceBounds.MaxY - surfaceBounds.MinY

	if surfaceWidth <= 0 || surfaceHeight <= 0 {
		return nil, fmt.Errorf(
			"compact placements: invalid surface bounds",
		)
	}

	initialStepX :=
		surfaceWidth * config.InitialStepFactor

	initialStepY :=
		surfaceHeight * config.InitialStepFactor

	minStepX :=
		surfaceWidth * config.MinStepFactor

	minStepY :=
		surfaceHeight * config.MinStepFactor

	patternLookup := buildPatternLookup(
		problem.Patterns,
	)

	compacted := append(
		[]Placement(nil),
		placements...,
	)

	for pass := 0; pass < config.MaxPasses; pass++ {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf(
				"compact placements: %w",
				err,
			)
		}

		passChanged := false

		for i := range compacted {
			if err := ctx.Err(); err != nil {
				return nil, fmt.Errorf(
					"compact placements: %w",
					err,
				)
			}

			patternItem, ok := patternLookup[compacted[i].PatternID]
			if !ok {
				return nil, fmt.Errorf(
					"compact placements: pattern item %d not found",
					compacted[i].PatternID,
				)
			}

			occupied, err := buildCompactionOccupied(
				engine,
				problem,
				compacted,
				patternLookup,
				i,
			)
			if err != nil {
				return nil, fmt.Errorf(
					"compact placements: %w",
					err,
				)
			}

			leftCompacted, changed, err :=
				compactPlacementAlongAxis(
					ctx,
					engine,
					problem.Surface,
					patternItem.Geometry,
					occupied,
					compacted[i],
					compactionAxisX,
					initialStepX,
					minStepX,
				)
			if err != nil {
				return nil, fmt.Errorf(
					"compact placements: %w",
					err,
				)
			}

			compacted[i] = leftCompacted

			if changed {
				passChanged = true
			}

			downCompacted, changed, err :=
				compactPlacementAlongAxis(
					ctx,
					engine,
					problem.Surface,
					patternItem.Geometry,
					occupied,
					compacted[i],
					compactionAxisY,
					initialStepY,
					minStepY,
				)
			if err != nil {
				return nil, fmt.Errorf(
					"compact placements: %w",
					err,
				)
			}

			compacted[i] = downCompacted

			if changed {
				passChanged = true
			}
		}

		if !passChanged {
			break
		}
	}

	return compacted, nil
}

func compactPlacementAlongAxis(
	ctx context.Context,
	engine geometry.Engine,
	surface domaingeometry.Polygon,
	pattern domaingeometry.Polygon,
	occupied []domaingeometry.Polygon,
	placement Placement,
	axis compactionAxis,
	initialStep float64,
	minStep float64,
) (Placement, bool, error) {
	current := placement
	changed := false
	step := initialStep

	for {
		if err := ctx.Err(); err != nil {
			return Placement{}, false, fmt.Errorf(
				"compact placement along axis: %w",
				err,
			)
		}

		candidate := current

		switch axis {
		case compactionAxisX:
			candidate.X -= step

		case compactionAxisY:
			candidate.Y -= step

		default:
			return Placement{}, false, fmt.Errorf(
				"compact placement along axis: unknown axis",
			)
		}

		transformed, err := engine.Transform(
			pattern,
			candidate.X,
			candidate.Y,
			candidate.Rotation,
		)
		if err != nil {
			return Placement{}, false, fmt.Errorf(
				"compact placement along axis: %w",
				err,
			)
		}

		feasible, err := isPlacementFeasible(
			engine,
			transformed,
			surface,
			occupied,
		)
		if err != nil {
			return Placement{}, false, fmt.Errorf(
				"compact placement along axis: %w",
				err,
			)
		}

		if feasible {
			current = candidate
			changed = true
			continue
		}

		if step <= minStep {
			break
		}

		step /= 2

		if step < minStep {
			step = minStep
		}
	}

	return current, changed, nil
}

func buildCompactionOccupied(
	engine geometry.Engine,
	problem Problem,
	placements []Placement,
	patternLookup map[int64]PatternItem,
	excludedIndex int,
) ([]domaingeometry.Polygon, error) {
	occupied := append(
		[]domaingeometry.Polygon(nil),
		problem.Obstacles...,
	)

	for i, placement := range placements {
		if i == excludedIndex {
			continue
		}

		patternItem, ok := patternLookup[placement.PatternID]
		if !ok {
			return nil, fmt.Errorf(
				"build compaction occupied: pattern item %d not found",
				placement.PatternID,
			)
		}

		transformed, err := engine.Transform(
			patternItem.Geometry,
			placement.X,
			placement.Y,
			placement.Rotation,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"build compaction occupied: %w",
				err,
			)
		}

		occupied = append(
			occupied,
			transformed,
		)
	}

	return occupied, nil
}
