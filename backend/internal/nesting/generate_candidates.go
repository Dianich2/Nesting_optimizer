package nesting

import (
	"context"
	"fmt"
	dg "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
)

func generateCandidates(engine geometry.Engine, pattern dg.Polygon, anchors []dg.Point, rotations []float64) ([]Candidate, error) {
	return generateCandidatesContext(context.Background(), engine, pattern, anchors, rotations)
}
func generateCandidatesContext(ctx context.Context, engine geometry.Engine, pattern dg.Polygon, anchors []dg.Point, rotations []float64) ([]Candidate, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	n, err := candidateProduct(MaxCandidates, len(anchors), len(ringVertices(pattern.Exterior)), len(rotations))
	if err != nil {
		return nil, err
	}
	b := newCandidateBuffer(MaxCandidates, false)
	if n > 0 {
		b.items = make([]Candidate, 0, n)
	}
	for _, rotation := range rotations {
		if err = ctx.Err(); err != nil {
			return nil, err
		}
		rotation = geometry.NormalizeDegrees(rotation)
		rotated, e := engine.Transform(pattern, 0, 0, rotation)
		if e != nil {
			return nil, fmt.Errorf("generate candidates: %w", e)
		}
		if e = appendAnchorCandidates(ctx, b, anchors, rotated, rotation); e != nil {
			return nil, e
		}
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	return b.items, nil
}
