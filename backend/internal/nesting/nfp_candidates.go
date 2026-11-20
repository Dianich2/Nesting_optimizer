package nesting

import (
	"context"
	"fmt"
	dg "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
)

func generateNFPCandidatesForRotations(engine geometry.Engine, builder geometry.NFPBuilder, pattern, surface dg.Polygon, occupied []dg.Polygon, rotations []float64) ([]Candidate, error) {
	return generateNFPCandidatesForRotationsContext(context.Background(), engine, builder, pattern, surface, occupied, rotations)
}
func generateNFPCandidatesForRotationsContext(ctx context.Context, engine geometry.Engine, builder geometry.NFPBuilder, pattern, surface dg.Polygon, occupied []dg.Polygon, rotations []float64) ([]Candidate, error) {
	b, err := generateNFPCandidateBatch(ctx, engine, builder, pattern, surface, occupied, rotations)
	if err != nil {
		return nil, err
	}
	return b.items, nil
}
func generateNFPCandidateBatch(ctx context.Context, engine geometry.Engine, builder geometry.NFPBuilder, pattern, surface dg.Polygon, occupied []dg.Polygon, rotations []float64) (*candidateBuffer, error) {
	anchors, err := countAnchors(ctx, surface, nil)
	if err != nil {
		return nil, err
	}
	if _, err = candidateProduct(MaxCandidates, anchors, len(ringVertices(pattern.Exterior)), len(rotations)); err != nil {
		return nil, err
	}
	b := newCandidateBuffer(MaxCandidates, true)
	b.items = []Candidate{}
	for _, rotation := range rotations {
		if err = ctx.Err(); err != nil {
			return nil, err
		}
		rotation = geometry.NormalizeDegrees(rotation)
		rotated, e := engine.Transform(pattern, 0, 0, rotation)
		if e != nil {
			return nil, fmt.Errorf("generate NFP candidates: %w", e)
		}
		if e = appendNFPCandidates(ctx, b, builder, rotated, surface, occupied, rotation); e != nil {
			return nil, e
		}
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	return b, nil
}
func generateNFPCandidates(builder geometry.NFPBuilder, pattern, surface dg.Polygon, occupied []dg.Polygon, rotation float64) ([]Candidate, error) {
	return generateNFPCandidatesContext(context.Background(), builder, pattern, surface, occupied, rotation)
}
func generateNFPCandidatesContext(ctx context.Context, builder geometry.NFPBuilder, pattern, surface dg.Polygon, occupied []dg.Polygon, rotation float64) ([]Candidate, error) {
	b := newCandidateBuffer(MaxCandidates, true)
	b.items = []Candidate{}
	if err := appendNFPCandidates(ctx, b, builder, pattern, surface, occupied, rotation); err != nil {
		return nil, err
	}
	return b.items, nil
}
func appendNFPCandidates(ctx context.Context, b *candidateBuffer, builder geometry.NFPBuilder, pattern, surface dg.Polygon, occupied []dg.Polygon, rotation float64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	vertices := ringVertices(pattern.Exterior)
	if len(vertices) == 0 {
		return geometry.ErrInvalidPolygon
	}
	minX, maxX, minY, maxY := vertices[0].X, vertices[0].X, vertices[0].Y, vertices[0].Y
	for _, p := range vertices[1:] {
		if err := ctx.Err(); err != nil {
			return err
		}
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	centerX, centerY := minX/2+maxX/2, minY/2+maxY/2
	anchors, err := collectAnchorsContext(ctx, surface, nil)
	if err != nil {
		return err
	}
	if err = appendAnchorCandidates(ctx, b, anchors, pattern, rotation); err != nil {
		return err
	}
	for _, stationary := range occupied {
		if err = ctx.Err(); err != nil {
			return err
		}
		nfp, err := geometry.BuildExternal(ctx, builder, stationary, pattern)
		if err != nil {
			return fmt.Errorf("generate NFP candidates: %w", err)
		}
		appendRing := func(r dg.Ring) error {
			for _, p := range r.Points {
				if e := b.add(ctx, Candidate{X: p.X - centerX, Y: p.Y - centerY, Rotation: rotation}); e != nil {
					return e
				}
			}
			return ctx.Err()
		}
		for _, p := range nfp.Polygons {
			if err = ctx.Err(); err != nil {
				return err
			}
			if err = appendRing(p.Exterior); err != nil {
				return err
			}
			for _, h := range p.Holes {
				if err = ctx.Err(); err != nil {
					return err
				}
				if err = appendRing(h); err != nil {
					return err
				}
			}
		}
	}
	return ctx.Err()
}
