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

	anchorCount, err := countAnchors(ctx, surface, occupied)
	if err != nil {
		return CandidatePlacement{}, false, err
	}
	if _, err = candidateProduct(MaxCandidates, anchorCount, len(ringVertices(pattern.Exterior)), len(rotations)); err != nil {
		return CandidatePlacement{}, false, err
	}
	anchors, err := collectAnchorsContext(ctx, surface, occupied)
	if err != nil {
		return CandidatePlacement{}, false, err
	}
	candidates, err := generateCandidatesContext(ctx, engine, pattern, anchors, rotations)
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
		return chooseBootstrap(ctx, engine, pattern, surface, occupied, rotations, MaxCandidates-len(candidates))
	}

	return bestCandidate, true, nil
}
