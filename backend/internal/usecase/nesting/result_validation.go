package nesting

import (
	"context"
	"fmt"
	"math"
	"slices"

	dg "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/domain/placement"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/internal/nesting"
)

func validateResult(ctx context.Context, engine geometry.Engine, surfaceID int64, problem nesting.Problem, result nesting.Result) ([]preparedPlacement, error) {
	patterns := make(map[int64]nesting.PatternItem, len(problem.Patterns))
	requested := 0
	for _, item := range problem.Patterns {
		patterns[item.PatternID] = item
		requested += item.Quantity
	}
	remaining := make(map[int64]int, len(patterns))
	for id, item := range patterns {
		remaining[id] = item.Quantity
	}
	occupied := append([]dg.Polygon(nil), problem.Obstacles...)
	prepared := make([]preparedPlacement, 0, len(result.Placements))
	placedArea := 0.0
	for _, item := range result.Placements {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if remaining[item.PatternID] <= 0 || !slices.Contains(problem.AllowedRotations, item.Rotation) {
			return nil, fmt.Errorf("unexpected pattern, quantity or rotation for pattern %d", item.PatternID)
		}
		polygon, err := engine.Transform(patterns[item.PatternID].Geometry, item.X, item.Y, item.Rotation)
		if err != nil {
			return nil, fmt.Errorf("transform result: %w", err)
		}
		if err = engine.ValidatePolygon(polygon); err != nil {
			return nil, err
		}
		covered, err := engine.CoveredBy(polygon, problem.Surface)
		if err != nil {
			return nil, err
		}
		if !covered {
			return nil, fmt.Errorf("pattern %d is outside the usable surface", item.PatternID)
		}
		for _, other := range occupied {
			overlap, err := engine.InteriorsIntersect(polygon, other)
			if err != nil {
				return nil, err
			}
			if overlap {
				return nil, fmt.Errorf("pattern %d overlaps another placement", item.PatternID)
			}
		}
		area, err := engine.Area(polygon)
		if err != nil {
			return nil, err
		}
		placedArea += area
		remaining[item.PatternID]--
		occupied = append(occupied, polygon)
		prepared = append(prepared, preparedPlacement{Placement: placement.Placement{
			ProjectSurfaceID: surfaceID, ProjectPatternID: item.PatternID,
			X: item.X, Y: item.Y, Rotation: item.Rotation,
		}, Geometry: polygon})
	}
	for _, item := range result.Unplaced {
		if item.Quantity <= 0 || remaining[item.PatternID] != item.Quantity {
			return nil, fmt.Errorf("invalid unplaced quantity for pattern %d", item.PatternID)
		}
		delete(remaining, item.PatternID)
	}
	for _, count := range remaining {
		if count != 0 {
			return nil, fmt.Errorf("missing unplaced patterns")
		}
	}
	surfaceArea, err := engine.Area(problem.Surface)
	if err != nil {
		return nil, err
	}
	metrics := result.Metrics
	close := func(actual, expected float64) bool {
		return math.Abs(actual-expected) <= 1e-8*math.Max(math.Abs(expected), 1e-12)
	}
	if surfaceArea <= 0 || metrics.RequestedCount != requested || metrics.PlacedCount != len(prepared) ||
		!close(metrics.SurfaceArea, surfaceArea) || !close(metrics.PlacedArea, placedArea) ||
		!close(metrics.Utilization, placedArea/surfaceArea) || metrics.Utilization < 0 || metrics.Utilization > 1 {
		return nil, fmt.Errorf("optimizer metrics do not match validated placements")
	}
	return prepared, nil
}
