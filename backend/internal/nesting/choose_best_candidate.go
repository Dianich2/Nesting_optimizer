package nesting

import (
	"context"
	"fmt"
	"math"
	"server_nesting_optimizer/internal/config"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
)

func floatsEqual(a, b float64) bool {

	diff := math.Abs(a - b)

	scale := math.Max(
		1,
		math.Max(math.Abs(a), math.Abs(b)),
	)

	return diff <= config.Epsilon*scale
}

func isScoreBetter(
	candidate CandidateScore,
	best CandidateScore,
) bool {
	if !floatsEqual(candidate.UsedArea, best.UsedArea) {
		return candidate.UsedArea < best.UsedArea
	}

	if !floatsEqual(candidate.UsedHeight, best.UsedHeight) {
		return candidate.UsedHeight < best.UsedHeight
	}

	return false
}

func chooseBestCandidate(
	ctx context.Context,
	engine geometry.Engine,
	pattern domaingeometry.Polygon,
	candidates []Candidate,
	surface domaingeometry.Polygon,
	occupied []domaingeometry.Polygon,
) (CandidatePlacement, bool, error) {
	var bestCandidate Candidate
	var bestScore CandidateScore
	var bestCandidateGeometry domaingeometry.Polygon
	found := false

	if err := ctx.Err(); err != nil {
		return CandidatePlacement{}, false, fmt.Errorf(
			"choose best candidate: %w",
			err,
		)
	}

	var surfaceBounds domaingeometry.Bounds

	if len(candidates) > 0 {
		var err error

		surfaceBounds, err = engine.Bounds(surface)
		if err != nil {
			return CandidatePlacement{}, false, fmt.Errorf(
				"choose best candidate: surface bounds: %w",
				err,
			)
		}
	}

	occupiedBounds := make(
		[]domaingeometry.Bounds,
		0,
		len(occupied),
	)

	if len(candidates) > 0 {
		for _, polygon := range occupied {
			bounds, err := engine.Bounds(polygon)
			if err != nil {
				return CandidatePlacement{}, false, fmt.Errorf(
					"choose best candidate: occupied bounds: %w",
					err,
				)
			}

			occupiedBounds = append(occupiedBounds, bounds)
		}
	}

	occupiedExtents := calculateOccupiedScoreExtents(
		occupiedBounds,
	)

	recordCandidateStatistics(ctx, len(candidates))
	for _, candidate := range candidates {
		if err := ctx.Err(); err != nil {
			return CandidatePlacement{}, false, fmt.Errorf(
				"choose best candidate: %w",
				err,
			)
		}

		transformedPattern, err := engine.Transform(
			pattern,
			candidate.X,
			candidate.Y,
			candidate.Rotation,
		)
		if err != nil {
			return CandidatePlacement{}, false, fmt.Errorf(
				"choose best candidate: %w",
				err,
			)
		}

		candidateBounds, err := engine.Bounds(transformedPattern)
		if err != nil {
			return CandidatePlacement{}, false, fmt.Errorf(
				"choose best candidate: candidate bounds: %w",
				err,
			)
		}

		feasible, err := isPlacementFeasibleWithBounds(
			engine,
			transformedPattern,
			candidateBounds,
			surface,
			occupied,
			occupiedBounds,
		)
		if err != nil {
			return CandidatePlacement{}, false, fmt.Errorf(
				"choose best candidate: %w",
				err,
			)
		}

		if !feasible {
			continue
		}

		score := calculateCandidateScoreWithExtents(
			surfaceBounds,
			occupiedExtents,
			candidateBounds,
		)

		if !found {
			bestCandidate = candidate
			bestScore = score
			bestCandidateGeometry = transformedPattern
			found = true
			continue
		}

		if isScoreBetter(score, bestScore) {
			bestCandidate = candidate
			bestScore = score
			bestCandidateGeometry = transformedPattern
		}
	}

	if !found {
		return CandidatePlacement{}, false, nil
	}

	return CandidatePlacement{
		Candidate: bestCandidate,
		Geometry:  bestCandidateGeometry,
	}, true, nil
}
