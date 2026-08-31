package nesting

import (
	"context"
	"fmt"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
)

func placeOnePatternNFP(
	ctx context.Context,
	engine geometry.Engine,
	nfpBuilder geometry.NFPBuilder,
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

	candidates, err := generateNFPCandidatesForRotations(
		engine,
		nfpBuilder,
		pattern,
		surface,
		occupied,
		rotations,
	)
	if err != nil {
		return CandidatePlacement{}, false, fmt.Errorf(
			"place one pattern NFP: %w",
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
			"place one pattern NFP: %w",
			err,
		)
	}

	if !isFound {
		return CandidatePlacement{}, false, nil
	}

	return bestCandidate, true, nil
}
