package nesting

import (
	"context"
	dg "server_nesting_optimizer/internal/domain/geometry"
)

func candidatesFromAnchors(anchors []dg.Point, pattern dg.Polygon, rotation float64) ([]Candidate, error) {
	return candidatesFromAnchorsContext(context.Background(), anchors, pattern, rotation)
}
func candidatesFromAnchorsContext(ctx context.Context, anchors []dg.Point, pattern dg.Polygon, rotation float64) ([]Candidate, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	n, err := candidateProduct(MaxCandidates, len(anchors), len(ringVertices(pattern.Exterior)))
	if err != nil {
		return nil, err
	}
	b := newCandidateBuffer(MaxCandidates, false)
	if n > 0 {
		b.items = make([]Candidate, 0, n)
	}
	if err = appendAnchorCandidates(ctx, b, anchors, pattern, rotation); err != nil {
		return nil, err
	}
	return b.items, nil
}
func appendAnchorCandidates(ctx context.Context, b *candidateBuffer, anchors []dg.Point, pattern dg.Polygon, rotation float64) error {
	vertices := ringVertices(pattern.Exterior)
	if _, err := candidateProduct(b.limit-b.work, len(anchors), len(vertices)); err != nil {
		return err
	}
	for _, a := range anchors {
		if err := ctx.Err(); err != nil {
			return err
		}
		for _, v := range vertices {
			if err := b.add(ctx, Candidate{X: a.X - v.X, Y: a.Y - v.Y, Rotation: rotation}); err != nil {
				return err
			}
		}
	}
	return ctx.Err()
}
