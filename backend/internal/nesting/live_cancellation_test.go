package nesting

import (
	"context"
	"errors"
	"fmt"
	dg "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry/nfp"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
	"testing"
	"time"
)

func TestLiveCancellationOfEveryOptimizer(t *testing.T) {
	for _, algorithm := range []Algorithm{BaselineAlgorithm, NFPGreedyAlgorithm, GeneticAlgorithm, SimulatedAnnealingAlgorithm, HybridAlgorithm} {
		for _, deadline := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/deadline=%v", algorithm, deadline), func(t *testing.T) {
				engine := simplefeatures.NewEngine()
				builder := nfp.NewCachedBuilder(nfp.NewBuilder(engine))
				genetic, err := NewGeneticOptimizer(engine, builder, DefaultGeneticConfig())
				if err != nil {
					t.Fatal(err)
				}
				annealing, err := NewSimulatedAnnealingOptimizer(engine, builder, DefaultAnnealingConfig())
				if err != nil {
					t.Fatal(err)
				}
				hybrid, err := NewHybridOptimizer(engine, genetic, annealing, DefaultCompactionConfig())
				if err != nil {
					t.Fatal(err)
				}
				algorithms := map[Algorithm]Optimizer{
					BaselineAlgorithm:  NewBaselineOptimizer(engine),
					NFPGreedyAlgorithm: NewNFPGreedyOptimizer(engine, builder),
					GeneticAlgorithm:   genetic, SimulatedAnnealingAlgorithm: annealing, HybridAlgorithm: hybrid,
				}
				problem := Problem{Surface: rectangle(10000, 10000), Patterns: []PatternItem{{PatternID: 1, Geometry: rectangle(10, 10), Quantity: 500}}, AllowedRotations: []float64{0, 90, 180, 270}}
				for i := range 20 {
					obstacle := dg.Polygon{Exterior: dg.Ring{Points: []dg.Point{{X: 1000 + float64(i)*20, Y: 1000}, {X: 1010 + float64(i)*20, Y: 1000}, {X: 1010 + float64(i)*20, Y: 1010}, {X: 1000 + float64(i)*20, Y: 1010}}}}
					problem.Obstacles = append(problem.Obstacles, obstacle)
				}
				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()
				expected := error(context.Canceled)
				var timer *time.Timer
				if deadline {
					cancel()
					ctx, cancel = context.WithTimeout(t.Context(), 20*time.Millisecond)
					expected = context.DeadlineExceeded
				} else {
					timer = time.AfterFunc(20*time.Millisecond, cancel)
				}
				defer cancel()
				if timer != nil {
					defer timer.Stop()
				}
				start := time.Now()
				result, err := algorithms[algorithm].Optimize(ctx, problem)
				if !errors.Is(err, expected) || len(result.Placements) != 0 {
					t.Fatalf("live optimizer cancellation: %v (%d placements)", err, len(result.Placements))
				}
				if time.Since(start) > 3*time.Second {
					t.Fatal("optimizer failed to react within bounded test window")
				}
				t.Logf("live cancellation took %s", time.Since(start))
			})
		}
	}
}
