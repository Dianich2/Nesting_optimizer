package nesting

import (
	"math"
	"testing"

	dg "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
	"server_nesting_optimizer/internal/nesting"
)

func rectangle(x, y, width, height float64) dg.Ring {
	return dg.Ring{Points: []dg.Point{{X: x, Y: y}, {X: x + width, Y: y}, {X: x + width, Y: y + height}, {X: x, Y: y + height}}}
}

func TestValidateResult(t *testing.T) {
	engine := simplefeatures.NewEngine()
	for _, tc := range []struct {
		name string
		edit func(*nesting.Problem, *nesting.Result)
	}{
		{"edge touching", nil},
		{"two placements overlap", func(_ *nesting.Problem, r *nesting.Result) { r.Placements[1].X = 14 }},
		{"partly outside", func(_ *nesting.Problem, r *nesting.Result) { r.Placements[0].X = 4 }},
		{"inside hole", func(p *nesting.Problem, _ *nesting.Result) { p.Surface.Holes = []dg.Ring{rectangle(4, 4, 12, 12)} }},
		{"overlaps obstacle", func(p *nesting.Problem, _ *nesting.Result) {
			p.Obstacles = []dg.Polygon{{Exterior: rectangle(8, 8, 2, 2)}}
		}},
		{"unknown pattern", func(_ *nesting.Problem, r *nesting.Result) { r.Placements[0].PatternID = 2 }},
		{"invalid coordinate", func(_ *nesting.Problem, r *nesting.Result) { r.Placements[0].X = math.Inf(1) }},
		{"disallowed rotation", func(_ *nesting.Problem, r *nesting.Result) { r.Placements[0].Rotation = 45 }},
		{"wrong area", func(_ *nesting.Problem, r *nesting.Result) { r.Metrics.PlacedArea = 80 }},
		{"wrong unplaced", func(_ *nesting.Problem, r *nesting.Result) {
			r.Unplaced = []nesting.UnplacedPattern{{PatternID: 1, Quantity: 1}}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			problem := nesting.Problem{Surface: dg.Polygon{Exterior: rectangle(0, 0, 100, 100)}, AllowedRotations: []float64{0, 90},
				Patterns: []nesting.PatternItem{{PatternID: 1, Quantity: 2, Geometry: dg.Polygon{Exterior: rectangle(0, 0, 10, 10)}}}}
			result := nesting.Result{Placements: []nesting.Placement{{PatternID: 1, X: 10, Y: 10}, {PatternID: 1, X: 20, Y: 10}},
				Metrics: nesting.Metrics{RequestedCount: 2, PlacedCount: 2, SurfaceArea: 10000, PlacedArea: 200, Utilization: .02}}
			if tc.edit != nil {
				tc.edit(&problem, &result)
			}
			prepared, err := validateResult(t.Context(), engine, 1, problem, result)
			if tc.edit == nil {
				if err != nil || len(prepared) != 2 {
					t.Fatalf("expected valid result: %v", err)
				}
			} else if err == nil || prepared != nil {
				t.Fatalf("invalid result accepted: %v", prepared)
			}
		})
	}
}

func TestValidateConcaveSurfaceWithTwoHolesAndRotations(t *testing.T) {
	engine := simplefeatures.NewEngine()
	surface := dg.Polygon{
		Exterior: dg.Ring{Points: []dg.Point{{X: 0, Y: 0}, {X: 100, Y: 0}, {X: 100, Y: 40}, {X: 60, Y: 40}, {X: 60, Y: 100}, {X: 0, Y: 100}}},
		Holes:    []dg.Ring{rectangle(20, 20, 10, 10), rectangle(40, 70, 10, 10)},
	}
	pattern := dg.Polygon{Exterior: rectangle(0, 0, 12, 6)}
	problem := nesting.Problem{Surface: surface, AllowedRotations: []float64{0, 90, 180},
		Patterns: []nesting.PatternItem{{PatternID: 1, Quantity: 3, Geometry: pattern}}}
	result := nesting.Result{Placements: []nesting.Placement{
		{PatternID: 1, X: 6, Y: 10, Rotation: 90}, {PatternID: 1, X: 15, Y: 10, Rotation: 180}, {PatternID: 1, X: 58, Y: 85, Rotation: 90},
	}, Metrics: nesting.Metrics{RequestedCount: 3, PlacedCount: 3, SurfaceArea: 7400, PlacedArea: 216, Utilization: 216.0 / 7400}}
	if _, err := validateResult(t.Context(), engine, 1, problem, result); err == nil {
		t.Fatal("placement crossed concave boundary")
	}
	result.Placements[2].X = 55
	if actual, err := validateResult(t.Context(), engine, 1, problem, result); err != nil || len(actual) != 3 {
		t.Fatalf("valid rotations on a concave surface rejected: %v", err)
	}
	for _, center := range []dg.Point{{X: 25, Y: 25}, {X: 45, Y: 75}} {
		result.Placements[2].X, result.Placements[2].Y = center.X, center.Y
		if _, err := validateResult(t.Context(), engine, 1, problem, result); err == nil {
			t.Fatalf("hole at %v accepted", center)
		}
	}
}
