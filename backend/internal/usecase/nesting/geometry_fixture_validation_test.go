package nesting

import (
	"context"
	"errors"
	"math"
	dg "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry/nfp"
	sf "server_nesting_optimizer/internal/geometry/simplefeatures"
	opt "server_nesting_optimizer/internal/nesting"
	"testing"
	"time"
)

func TestGeometryFixturesPassServerValidation(t *testing.T) {
	engine := sf.NewEngine()
	square := func(size float64) dg.Polygon {
		return dg.Polygon{Exterior: dg.Ring{Points: []dg.Point{{X: 0, Y: 0}, {X: size, Y: 0}, {X: size, Y: size}, {X: 0, Y: size}}}}
	}
	for _, fixture := range []string{"concave", "holes", "occupied", "large_coordinates", "radial"} {
		for _, algorithm := range []opt.Algorithm{opt.BaselineAlgorithm, opt.NFPGreedyAlgorithm, opt.GeneticAlgorithm, opt.SimulatedAnnealingAlgorithm, opt.HybridAlgorithm} {
			t.Run(fixture+"/"+string(algorithm), func(t *testing.T) {
				surface := square(100)
				pattern := square(10)
				obstacles := []dg.Polygon{}
				if fixture == "radial" {
					pattern.Exterior.Points = nil
					for i := 0; i < 12; i++ {
						r := 5.0
						if i%2 == 1 {
							r = 3
						}
						a := 2 * math.Pi * float64(i) / 12
						pattern.Exterior.Points = append(pattern.Exterior.Points, dg.Point{X: 6 + r*math.Cos(a), Y: 6 + r*math.Sin(a)})
					}
				}
				if fixture == "concave" {
					pattern = dg.Polygon{Exterior: dg.Ring{Points: []dg.Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 4}, {X: 4, Y: 4}, {X: 4, Y: 10}, {X: 0, Y: 10}}}}
				}
				if fixture == "holes" {
					surface.Holes = []dg.Ring{{Points: []dg.Point{{X: 40, Y: 40}, {X: 40, Y: 60}, {X: 60, Y: 60}, {X: 60, Y: 40}}}}
				}
				if fixture == "occupied" {
					obstacles = []dg.Polygon{square(12)}
				}
				if fixture == "large_coordinates" {
					var err error
					surface, err = engine.Transform(surface, 100000, 100000, 0)
					if err != nil {
						t.Fatal(err)
					}
				}
				quantity := 2
				if fixture == "radial" {
					quantity = 1
				}
				problem := opt.Problem{Surface: surface, Patterns: []opt.PatternItem{{PatternID: 1, Geometry: pattern, Quantity: quantity}}, Obstacles: obstacles, AllowedRotations: []float64{0, 90}}
				ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
				defer cancel()
				builder := nfp.NewCachedBuilder(nfp.NewBuilder(engine))
				genetic, err := opt.NewGeneticOptimizer(engine, builder, opt.DefaultGeneticConfig())
				if err != nil {
					t.Fatal(err)
				}
				annealing, err := opt.NewSimulatedAnnealingOptimizer(engine, builder, opt.DefaultAnnealingConfig())
				if err != nil {
					t.Fatal(err)
				}
				hybrid, err := opt.NewHybridOptimizer(engine, genetic, annealing, opt.DefaultCompactionConfig())
				if err != nil {
					t.Fatal(err)
				}
				algorithms := map[opt.Algorithm]opt.Optimizer{opt.BaselineAlgorithm: opt.NewBaselineOptimizer(engine), opt.NFPGreedyAlgorithm: opt.NewNFPGreedyOptimizer(engine, builder), opt.GeneticAlgorithm: genetic, opt.SimulatedAnnealingAlgorithm: annealing, opt.HybridAlgorithm: hybrid}
				result, err := algorithms[algorithm].Optimize(ctx, problem)
				if err != nil {
					t.Fatal(err)
				}
				if len(result.Placements) == 0 {
					t.Fatal("positive fixture did not place anything")
				}
				placements, err := validateResult(ctx, engine, 1, problem, result)
				if err != nil || len(placements) != len(result.Placements) {
					t.Fatalf("server validation %v", err)
				}
				result.Placements[0].X = 1e8
				if _, err = validateResult(ctx, engine, 1, problem, result); err == nil {
					t.Fatal("server accepted placement outside surface")
				}
			})
		}
	}
}

func TestWaitingSlotDeadlineTiming(t *testing.T) {
	limiter := NewComputeLimiter(1, 1)
	release, err := limiter.Acquire(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err = limiter.Acquire(ctx)
	elapsed := time.Since(start)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("waiting deadline %v", err)
	}
	release()
	next, err := limiter.Acquire(t.Context())
	if err != nil {
		t.Fatal("slot leaked")
	}
	next()
	t.Logf("queue deadline=20ms actual return=%s", elapsed)
}
