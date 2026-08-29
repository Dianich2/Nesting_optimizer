package nesting

import (
	"context"
	"fmt"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
)

func placeOnePattern(
	ctx context.Context,
	engine geometry.Engine,
	pattern domaingeometry.Polygon,
	surface domaingeometry.Polygon,
	occupied []domaingeometry.Polygon,
	rotations []float64,
) (CandidatePlacement, bool, error) {
	if err := ctx.Err(); err != nil {
		return CandidatePlacement{}, false, fmt.Errorf(
			"place one pattern: %w",
			err,
		)
	}

	anchors := collectAnchors(
		surface,
		occupied,
	)

	candidates, err := generateCandidates(
		engine,
		pattern,
		anchors,
		rotations,
	)
	if err != nil {
		return CandidatePlacement{}, false, fmt.Errorf(
			"place one pattern: %w",
			err,
		)
	}

	if err := ctx.Err(); err != nil {
		return CandidatePlacement{}, false, fmt.Errorf(
			"place one pattern: %w",
			err,
		)
	}

	bestCandidate, isFound, err := chooseBestCandidate(
		ctx,
		engine,
		pattern,
		candidates,
		surface,
		occupied,
	)
	if err != nil {
		return CandidatePlacement{}, false, fmt.Errorf(
			"place one pattern: %w",
			err,
		)
	}

	if !isFound {
		return CandidatePlacement{}, false, nil
	}

	return bestCandidate, true, nil
}
