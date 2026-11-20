package nesting

import (
	"context"
	"fmt"
	"math"
	dg "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
)

// chooseBootstrap is a bounded fallback, used only when the legacy search found
// no feasible candidate. Transform(p,0,0,r) fixes the coordinate convention:
// candidate XY translates those rotated vertices, even when their bbox center
// differs from the original rotation center. Seeds are not assumed feasible.
func chooseBootstrap(ctx context.Context, engine geometry.Engine, pattern, surface dg.Polygon, occupied []dg.Polygon, rotations []float64, remaining int) (CandidatePlacement, bool, error) {
	if err := ctx.Err(); err != nil {
		return CandidatePlacement{}, false, err
	}
	remaining = min(remaining, MaxCandidates)
	const seedsPerRotation = 5
	if remaining < 0 || len(rotations) > remaining/seedsPerRotation {
		return CandidatePlacement{}, false, geometry.ErrComplexity
	}
	bounds, err := engine.Bounds(surface)
	if err != nil {
		return CandidatePlacement{}, false, err
	}
	buffer := newCandidateBuffer(remaining, true)
	for _, rotation := range rotations {
		if err = ctx.Err(); err != nil {
			return CandidatePlacement{}, false, err
		}
		rotation = geometry.NormalizeDegrees(rotation)
		rotated, e := engine.Transform(pattern, 0, 0, rotation)
		if e != nil {
			return CandidatePlacement{}, false, e
		}
		rb, e := engine.Bounds(rotated)
		if e != nil {
			return CandidatePlacement{}, false, e
		}
		left, right := bounds.MinX-rb.MinX, bounds.MaxX-rb.MaxX
		bottom, top := bounds.MinY-rb.MinY, bounds.MaxY-rb.MaxY
		positions := [][2]float64{{left, bottom}, {right, bottom}, {left, top}, {right, top}, {bounds.MinX/2 + bounds.MaxX/2 - rb.MinX/2 - rb.MaxX/2, bounds.MinY/2 + bounds.MaxY/2 - rb.MinY/2 - rb.MaxY/2}}
		for _, xy := range positions {
			if math.IsNaN(xy[0]) || math.IsInf(xy[0], 0) || math.IsNaN(xy[1]) || math.IsInf(xy[1], 0) {
				return CandidatePlacement{}, false, geometry.ErrInvalidTransform
			}
			c := Candidate{X: xy[0], Y: xy[1], Rotation: rotation}
			if e := buffer.add(ctx, c); e != nil {
				return CandidatePlacement{}, false, e
			}
		}
	}
	result, found, err := chooseBestCandidate(ctx, engine, pattern, buffer.items, surface, occupied)
	if err != nil {
		return CandidatePlacement{}, false, fmt.Errorf("bootstrap candidate: %w", err)
	}
	if err = ctx.Err(); err != nil {
		return CandidatePlacement{}, false, err
	}
	return result, found, nil
}
