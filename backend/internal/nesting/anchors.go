package nesting

import (
	"context"
	dg "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
)

func anchorRings(ctx context.Context, surface dg.Polygon, occupied []dg.Polygon, visit func(dg.Ring) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := visit(surface.Exterior); err != nil {
		return err
	}
	for _, hole := range surface.Holes {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := visit(hole); err != nil {
			return err
		}
	}
	for _, p := range occupied {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := visit(p.Exterior); err != nil {
			return err
		}
	}
	return ctx.Err()
}
func countAnchors(ctx context.Context, surface dg.Polygon, occupied []dg.Polygon) (int, error) {
	total := 0
	err := anchorRings(ctx, surface, occupied, func(r dg.Ring) error {
		n := len(ringVertices(r))
		if n > MaxCandidates-total {
			return geometry.ErrComplexity
		}
		total += n
		return nil
	})
	return total, err
}
func collectAnchors(surface dg.Polygon, occupied []dg.Polygon) ([]dg.Point, error) {
	return collectAnchorsContext(context.Background(), surface, occupied)
}
func collectAnchorsContext(ctx context.Context, surface dg.Polygon, occupied []dg.Polygon) ([]dg.Point, error) {
	total, err := countAnchors(ctx, surface, occupied)
	if err != nil {
		return nil, err
	}
	var result []dg.Point
	if total > 0 {
		result = make([]dg.Point, 0, total)
	}
	err = anchorRings(ctx, surface, occupied, func(r dg.Ring) error {
		for _, p := range ringVertices(r) {
			if err := ctx.Err(); err != nil {
				return err
			}
			result = append(result, p)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
