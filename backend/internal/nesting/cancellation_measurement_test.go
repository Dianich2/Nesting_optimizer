package nesting

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	dg "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/internal/geometry/nfp"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
	"sync/atomic"
	"testing"
	"time"
)

func TestAllOptimizerCancellationTiming(t *testing.T) {
	if os.Getenv("NESTING_GEOMETRY_EXPERIMENT") != "1" {
		t.Skip("explicit cancellation timing experiment")
	}
	for _, algorithm := range []Algorithm{BaselineAlgorithm, NFPGreedyAlgorithm, GeneticAlgorithm, SimulatedAnnealingAlgorithm, HybridAlgorithm} {
		for repeat := 0; repeat < 5; repeat++ {
			opt, _, _ := experimentOptimizer(t, algorithm)
			angles := make([]float64, 36)
			for i := range angles {
				angles[i] = float64(i) * 10
			}
			p := Problem{Surface: rectangle(10000, 10000), Patterns: []PatternItem{{PatternID: 1, Geometry: rectangle(10, 10), Quantity: 500}}, AllowedRotations: angles}
			for i := 0; i < 20; i++ {
				obstacle, err := simplefeatures.NewEngine().Transform(rectangle(10, 10), 1000+float64(i)*20, 1000, 0)
				if err != nil {
					t.Fatal(err)
				}
				p.Obstacles = append(p.Obstacles, obstacle)
			}
			ctx, cancel := context.WithCancel(t.Context())
			var requested atomic.Int64
			timer := time.AfterFunc(20*time.Millisecond, func() { requested.Store(time.Now().UnixNano()); cancel() })
			start := time.Now()
			result, err := opt.Optimize(ctx, p)
			finished := time.Now()
			timer.Stop()
			cancel()
			if !errors.Is(err, context.Canceled) || len(result.Placements) != 0 {
				t.Fatalf("%s cancellation result %v", algorithm, err)
			}
			latency := float64(finished.UnixNano()-requested.Load()) / 1e6
			if latency > 3000 {
				t.Fatalf("managed cancellation took %.1f ms", latency)
			}
			row := map[string]any{"algorithm": algorithm, "repeat": repeat, "cancel_to_return_ms": latency, "wall_ms": float64(finished.Sub(start)) / float64(time.Millisecond), "quantity": 500, "angles": 36, "obstacles": len(p.Obstacles), "partial_placements": len(result.Placements)}
			encoded, _ := json.Marshal(row)
			t.Log("OPT_CANCEL_ROW " + string(encoded))
		}
	}
}

func TestProfileOccupiedConcaveGeometry(t *testing.T) {
	if os.Getenv("NESTING_PROFILE_GEOMETRY") != "1" {
		t.Skip("explicit bounded pprof run")
	}
	seed, err := simplefeatures.NewEngine().Transform(rectangle(10, 10), 200, 200, 0)
	if err != nil {
		t.Fatal(err)
	}
	problem := Problem{Surface: rectangle(1000, 1000), Patterns: []PatternItem{{PatternID: 1, Geometry: experimentPolygon(12), Quantity: 50}}, AllowedRotations: []float64{0, 90}, Obstacles: []dg.Polygon{seed}}
	runGeometryExperiment(t, "profile", BaselineAlgorithm, problem, 40*time.Second, false)
}

// Deterministic barrier at an actual rotation boundary, rather than assuming
// a wall-clock timer happened to fire while the angle loop was executing.
type cancelRotationEngine struct {
	geometry.Engine
	calls     int
	cancel    context.CancelFunc
	requested time.Time
}

func (e *cancelRotationEngine) Transform(p dg.Polygon, x, y, r float64) (dg.Polygon, error) {
	result, err := e.Engine.Transform(p, x, y, r)
	e.calls++
	if e.calls == 3 {
		e.requested = time.Now()
		e.cancel()
	}
	return result, err
}
func TestCancellationAtRotationBoundary(t *testing.T) {
	base := simplefeatures.NewEngine()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	engine := &cancelRotationEngine{Engine: base, cancel: cancel}
	angles := make([]float64, 36)
	for i := range angles {
		angles[i] = float64(i) * 10
	}
	candidates, err := generateNFPCandidatesForRotationsContext(ctx, engine, nfp.NewBuilder(base), rectangle(10, 10), rectangle(100, 100), nil, angles)
	if !errors.Is(err, context.Canceled) || len(candidates) != 0 || engine.calls != 3 {
		t.Fatalf("rotation cancellation: calls=%d err=%v", engine.calls, err)
	}
	t.Logf("rotation boundary=3/36 cancel_to_return=%s", time.Since(engine.requested))
}
