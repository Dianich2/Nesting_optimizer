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

		feasible, err := isPlacementFeasible(
			engine,
			transformedPattern,
			surface,
			occupied,
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

		score, err := calculateCandidateScore(
			engine,
			surface,
			occupied,
			transformedPattern,
		)
		if err != nil {
			return CandidatePlacement{}, false, fmt.Errorf(
				"choose best candidate: %w",
				err,
			)
		}

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
