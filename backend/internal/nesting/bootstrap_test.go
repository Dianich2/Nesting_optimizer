package nesting

import (
	"context"
	"errors"
	"math"
	"reflect"
	dg "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
	"testing"
	"time"
)

func TestRadialBootstrapRegression(t *testing.T) {
	engine := simplefeatures.NewEngine()
	pattern := experimentPolygon(12)
	surface := rectangle(100, 100)
	witness, err := engine.Transform(pattern, 50, 50, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err = engine.ValidatePolygon(witness); err != nil {
		t.Fatal(err)
	}
	if ok, err := isPlacementFeasible(engine, witness, surface, nil); err != nil || !ok {
		t.Fatalf("constructive witness invalid: %v", err)
	}
	for _, algo := range []Algorithm{BaselineAlgorithm, NFPGreedyAlgorithm} {
		for _, rotations := range [][]float64{{0}, {0, 90}} {
			t.Run(string(algo)+fmtRotations(rotations), func(t *testing.T) {
				opt, _, _ := experimentOptimizer(t, algo)
				ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
				defer cancel()
				p := Problem{Surface: surface, Patterns: []PatternItem{{PatternID: 1, Geometry: pattern, Quantity: 1}}, AllowedRotations: rotations}
				result, err := opt.Optimize(ctx, p)
				if err != nil {
					t.Fatal(err)
				}
				if len(result.Placements) != 1 {
					t.Fatalf("placed=%d; valid interior witness exists", len(result.Placements))
				}
				validateExperiment(t, p, result)
			})
		}
	}
}
func fmtRotations(r []float64) string {
	if len(r) == 1 {
		return "/one_angle"
	}
	return "/two_angles"
}

func TestBootstrapGeometryMatrix(t *testing.T) {
	engine := simplefeatures.NewEngine()
	hole := func(x, y, size float64) dg.Ring {
		return dg.Ring{Points: []dg.Point{{X: x, Y: y}, {X: x, Y: y + size}, {X: x + size, Y: y + size}, {X: x + size, Y: y}}}
	}
	for _, algo := range []Algorithm{BaselineAlgorithm, NFPGreedyAlgorithm, GeneticAlgorithm, SimulatedAnnealingAlgorithm, HybridAlgorithm} {
		t.Run(string(algo)+"/radial", func(t *testing.T) {
			opt, _, _ := experimentOptimizer(t, algo)
			p := Problem{Surface: rectangle(100, 100), Patterns: []PatternItem{{PatternID: 1, Geometry: experimentPolygon(12), Quantity: 1}}, AllowedRotations: []float64{90}}
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			defer cancel()
			result, err := opt.Optimize(ctx, p)
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Placements) != 1 || result.Placements[0].Rotation != 90 {
				t.Fatalf("wrong radial placement: %+v", result)
			}
			validateExperiment(t, p, result)
		})
	}
	for _, algo := range []Algorithm{BaselineAlgorithm, NFPGreedyAlgorithm} {
		for _, name := range []string{"rectangle_six", "concave_L", "one_hole", "two_holes", "occupied", "full", "too_large", "diamond", "shifted", "rotated_asymmetric"} {
			t.Run(string(algo)+"/"+name, func(t *testing.T) {
				surface, pattern := rectangle(100, 100), experimentPolygon(12)
				rotations := []float64{0, 90}
				quantity := 1
				var occupied []dg.Polygon
				want := 1
				switch name {
				case "rectangle_six":
					pattern = rectangle(10, 10)
					quantity = 6
					want = 6
				case "concave_L":
					pattern = dg.Polygon{Exterior: dg.Ring{Points: []dg.Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 4}, {X: 4, Y: 4}, {X: 4, Y: 10}, {X: 0, Y: 10}}}}
				case "one_hole":
					surface.Holes = []dg.Ring{hole(40, 40, 20)}
				case "two_holes":
					surface.Holes = []dg.Ring{hole(0, 0, 12), hole(40, 40, 20)}
					surface.Holes[0] = hole(1, 1, 12)
				case "occupied":
					for _, xy := range [][2]float64{{7, 7}, {93, 7}, {7, 93}} {
						p, e := engine.Transform(rectangle(14, 14), xy[0], xy[1], 0)
						if e != nil {
							t.Fatal(e)
						}
						occupied = append(occupied, p)
					}
				case "full":
					occupied = []dg.Polygon{surface}
					want = 0
				case "too_large":
					surface = rectangle(4, 4)
					want = 0
				case "diamond":
					surface = dg.Polygon{Exterior: dg.Ring{Points: []dg.Point{{X: 50, Y: 0}, {X: 100, Y: 50}, {X: 50, Y: 100}, {X: 0, Y: 50}}}}
				case "shifted":
					var e error
					surface, e = engine.Transform(surface, 100000, -100000, 0)
					if e != nil {
						t.Fatal(e)
					}
				case "rotated_asymmetric":
					pattern = dg.Polygon{Exterior: dg.Ring{Points: []dg.Point{{X: 20, Y: 40}, {X: 28, Y: 40}, {X: 23, Y: 46}}}}
					rotations = []float64{33}
				}
				p := Problem{Surface: surface, Patterns: []PatternItem{{PatternID: 1, Geometry: pattern, Quantity: quantity}}, Obstacles: occupied, AllowedRotations: rotations}
				opt, _, _ := experimentOptimizer(t, algo)
				ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
				defer cancel()
				result, err := opt.Optimize(ctx, p)
				if err != nil {
					t.Fatal(err)
				}
				if len(result.Placements) != want {
					t.Fatalf("placed %d want %d", len(result.Placements), want)
				}
				if result.Metrics.RequestedCount != quantity || result.Metrics.PlacedCount != want {
					t.Fatal("inconsistent counts")
				}
				unplaced := 0
				for _, u := range result.Unplaced {
					unplaced += u.Quantity
				}
				if unplaced+want != quantity {
					t.Fatal("unplaced mismatch")
				}
				if math.Abs(result.Metrics.Utilization-result.Metrics.PlacedArea/result.Metrics.SurfaceArea) > 1e-12 {
					t.Fatal("utilization mismatch")
				}
				for _, p := range result.Placements {
					allowed := false
					for _, r := range rotations {
						allowed = allowed || p.Rotation == r
					}
					if !allowed {
						t.Fatal("unauthorized rotation")
					}
				}
				validateExperiment(t, p, result)
			})
		}
	}
}

type bootstrapCancelEngine struct {
	geometry.Engine
	cancel          context.CancelFunc
	duringSelection bool
}

func (e *bootstrapCancelEngine) Transform(p dg.Polygon, x, y, r float64) (dg.Polygon, error) {
	v, err := e.Engine.Transform(p, x, y, r)
	if !e.duringSelection {
		e.cancel()
	}
	return v, err
}
func (e *bootstrapCancelEngine) CoveredBy(a, b dg.Polygon) (bool, error) {
	v, err := e.Engine.CoveredBy(a, b)
	if e.duringSelection {
		e.cancel()
	}
	return v, err
}
func TestBootstrapCancellationAndBudget(t *testing.T) {
	for _, mode := range []string{"before", "generation", "selection"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			engine := &bootstrapCancelEngine{Engine: simplefeatures.NewEngine(), cancel: cancel, duringSelection: mode == "selection"}
			if mode == "before" {
				cancel()
			}
			placement, found, err := chooseBootstrap(ctx, engine, experimentPolygon(12), rectangle(100, 100), nil, []float64{0, 90}, 200000)
			if !errors.Is(err, context.Canceled) || found || len(placement.Geometry.Exterior.Points) != 0 {
				t.Fatalf("partial cancelled fallback: %v", err)
			}
		})
	}
	_, found, err := chooseBootstrap(t.Context(), simplefeatures.NewEngine(), experimentPolygon(12), rectangle(100, 100), nil, []float64{0, 90}, 9)
	if !errors.Is(err, geometry.ErrComplexity) || found {
		t.Fatal("candidate budget bypass")
	}
}
func TestBootstrapPreservesLegacySuccess(t *testing.T) {
	engine := simplefeatures.NewEngine()
	surface := rectangle(100, 100)
	pattern := rectangle(10, 10)
	candidates, err := generateCandidates(engine, pattern, mustAnchors(t, surface), []float64{0, 90})
	if err != nil {
		t.Fatal(err)
	}
	expected, found, err := chooseBestCandidate(t.Context(), engine, pattern, candidates, surface, nil)
	if err != nil || !found {
		t.Fatal("legacy witness missing")
	}
	got, found, err := placeOnePattern(t.Context(), engine, pattern, surface, nil, []float64{0, 90})
	if err != nil || !found || !reflect.DeepEqual(expected, got) {
		t.Fatal("legacy priority changed")
	}
	bootstrap, found, err := chooseBootstrap(t.Context(), engine, experimentPolygon(12), surface, nil, []float64{0}, 5)
	if err != nil || !found {
		t.Fatal(err)
	}
	bounds, err := engine.Bounds(bootstrap.Geometry)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(bounds.MinX) > 1e-12 || math.Abs(bounds.MinY) > 1e-12 {
		t.Fatalf("expected allowed boundary contact, got %+v", bounds)
	}
}

func mustAnchors(t *testing.T, p dg.Polygon) []dg.Point {
	t.Helper()
	v, e := collectAnchors(p, nil)
	if e != nil {
		t.Fatal(e)
	}
	return v
}
