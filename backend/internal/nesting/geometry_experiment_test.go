package nesting

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"os"
	"runtime"
	dg "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/internal/geometry/nfp"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

type geometryTrace struct {
	delegate geometry.NFPBuilder
	calls    atomic.Uint64
	maxNS    atomic.Int64
}

func (g *geometryTrace) BuildExternal(a, b dg.Polygon) (dg.MultiPolygon, error) {
	return g.BuildExternalContext(context.Background(), a, b)
}
func (g *geometryTrace) BuildExternalContext(ctx context.Context, a, b dg.Polygon) (dg.MultiPolygon, error) {
	g.calls.Add(1)
	start := time.Now()
	result, err := geometry.BuildExternal(ctx, g.delegate, a, b)
	elapsed := time.Since(start).Nanoseconds()
	for old := g.maxNS.Load(); elapsed > old; old = g.maxNS.Load() {
		if g.maxNS.CompareAndSwap(old, elapsed) {
			break
		}
	}
	return result, err
}
func experimentOptimizer(t *testing.T, algorithm Algorithm) (Optimizer, *geometryTrace, *geometryTrace) {
	t.Helper()
	engine := simplefeatures.NewEngine()
	miss := &geometryTrace{delegate: nfp.NewBuilder(engine)}
	calls := &geometryTrace{delegate: nfp.NewCachedBuilder(miss)}
	profile := comparisonBenchmarkProfile()
	genetic, err := NewGeneticOptimizer(engine, calls, profile.geneticConfig)
	if err != nil {
		t.Fatal(err)
	}
	genetic.rng = newBenchmarkRandom(20261120)
	annealing, err := NewSimulatedAnnealingOptimizer(engine, calls, profile.annealingConfig)
	if err != nil {
		t.Fatal(err)
	}
	annealing.rng = newBenchmarkRandom(20261120)
	hybrid, err := NewHybridOptimizer(engine, genetic, annealing, DefaultCompactionConfig())
	if err != nil {
		t.Fatal(err)
	}
	options := map[Algorithm]Optimizer{BaselineAlgorithm: NewBaselineOptimizer(engine), NFPGreedyAlgorithm: NewNFPGreedyOptimizer(engine, calls), GeneticAlgorithm: genetic, SimulatedAnnealingAlgorithm: annealing, HybridAlgorithm: hybrid}
	return options[algorithm], calls, miss
}
func experimentPolygon(vertices int) dg.Polygon {
	p := dg.Polygon{}
	for i := 0; i < vertices; i++ {
		r := 5.0
		if i%2 == 1 {
			r = 3.0
		}
		angle := 2 * math.Pi * float64(i) / float64(vertices)
		p.Exterior.Points = append(p.Exterior.Points, dg.Point{X: 6 + r*math.Cos(angle), Y: 6 + r*math.Sin(angle)})
	}
	return p
}
func validateExperiment(t *testing.T, p Problem, result Result) {
	t.Helper()
	engine := simplefeatures.NewEngine()
	patterns := map[int64]dg.Polygon{}
	for _, item := range p.Patterns {
		patterns[item.PatternID] = item.Geometry
	}
	occupied := append([]dg.Polygon(nil), p.Obstacles...)
	for _, placement := range result.Placements {
		transformed, err := engine.Transform(patterns[placement.PatternID], placement.X, placement.Y, placement.Rotation)
		if err != nil {
			t.Fatal(err)
		}
		if err = engine.ValidatePolygon(transformed); err != nil {
			t.Fatal(err)
		}
		inside, err := engine.CoveredBy(transformed, p.Surface)
		if err != nil || !inside {
			t.Fatalf("invalid placement: outside surface %v", err)
		}
		for _, other := range occupied {
			overlap, err := engine.InteriorsIntersect(transformed, other)
			if err != nil || overlap {
				t.Fatalf("invalid placement overlap %v", err)
			}
		}
		occupied = append(occupied, transformed)
	}
	if len(result.Placements) != result.Metrics.PlacedCount {
		t.Fatal("metric mismatch")
	}
}
func TestGeometryReliabilityExperiment(t *testing.T) {
	if os.Getenv("NESTING_GEOMETRY_EXPERIMENT") != "1" {
		t.Skip("explicit bounded geometry experiment")
	}
	algorithms := []Algorithm{BaselineAlgorithm, NFPGreedyAlgorithm, GeneticAlgorithm, SimulatedAnnealingAlgorithm, HybridAlgorithm}
	names := []string{"rectangle", "concave", "one_hole", "two_holes", "occupied", "close_points", "radial_without_seed"}
	for _, name := range names {
		for _, algorithm := range algorithms {
			surface := rectangle(100, 100)
			pattern := rectangle(10, 10)
			obstacles := []dg.Polygon{}
			if name == "concave" {
				pattern = dg.Polygon{Exterior: dg.Ring{Points: []dg.Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 4}, {X: 4, Y: 4}, {X: 4, Y: 10}, {X: 0, Y: 10}}}}
			}
			if name == "radial_without_seed" {
				pattern = experimentPolygon(12)
			}
			if name == "one_hole" || name == "two_holes" {
				surface.Holes = append(surface.Holes, dg.Ring{Points: []dg.Point{{X: 40, Y: 40}, {X: 60, Y: 40}, {X: 60, Y: 60}, {X: 40, Y: 60}}})
			}
			if name == "two_holes" {
				surface.Holes = append(surface.Holes, dg.Ring{Points: []dg.Point{{X: 70, Y: 70}, {X: 80, Y: 70}, {X: 80, Y: 80}, {X: 70, Y: 80}}})
			}
			if name == "occupied" {
				obstacles = append(obstacles, rectangle(12, 12))
			}
			if name == "close_points" {
				pattern.Exterior.Points = append([]dg.Point{{X: 0, Y: 0}, {X: 0.00001, Y: 0}}, pattern.Exterior.Points[1:]...)
				var err error
				surface, err = simplefeatures.NewEngine().Transform(surface, 100000, 100000, 0)
				if err != nil {
					t.Fatal(err)
				}
			}
			problem := Problem{Surface: surface, Patterns: []PatternItem{{PatternID: 1, Geometry: pattern, Quantity: 6}}, AllowedRotations: []float64{0, 90}, Obstacles: obstacles}
			runGeometryExperiment(t, name, algorithm, problem, 10*time.Second, false)
		}
	}
	for _, count := range []int{10, 50, 100, 500} {
		for _, algorithm := range []Algorithm{BaselineAlgorithm, NFPGreedyAlgorithm} {
			seed, err := simplefeatures.NewEngine().Transform(rectangle(10, 10), 200, 200, 0)
			if err != nil {
				t.Fatal(err)
			}
			problem := Problem{Surface: rectangle(1000, 1000), Patterns: []PatternItem{{PatternID: 1, Geometry: experimentPolygon(12), Quantity: count}}, AllowedRotations: []float64{0, 90}, Obstacles: []dg.Polygon{seed}}
			budget := 40 * time.Second
			if count == 500 {
				budget = 3 * time.Second
			}
			runGeometryExperiment(t, "quantity", algorithm, problem, budget, count >= 100)
		}
	}
	for _, vertices := range []int{32, 64, 128} {
		seed, err := simplefeatures.NewEngine().Transform(rectangle(10, 10), 100, 100, 0)
		if err != nil {
			t.Fatal(err)
		}
		problem := Problem{Surface: rectangle(500, 500), Patterns: []PatternItem{{PatternID: 1, Geometry: experimentPolygon(vertices), Quantity: 2}}, AllowedRotations: []float64{0, 90}, Obstacles: []dg.Polygon{seed}}
		runGeometryExperiment(t, "vertices", NFPGreedyAlgorithm, problem, 5*time.Second, true)
	}
	// 36 is accepted, 37 remains a server validation rejection (existing API tests).
	angles := make([]float64, 36)
	for i := range angles {
		angles[i] = float64(i) * 10
	}
	runGeometryExperiment(t, "36_angles", NFPGreedyAlgorithm, Problem{Surface: rectangle(100, 100), Patterns: []PatternItem{{PatternID: 1, Geometry: rectangle(10, 10), Quantity: 2}}, AllowedRotations: angles}, 5*time.Second, true)
}
func runGeometryExperiment(t *testing.T, name string, algorithm Algorithm, problem Problem, budget time.Duration, allowTimeout bool) {
	t.Helper()
	opt, calls, miss := experimentOptimizer(t, algorithm)
	ctx, cancel := context.WithTimeout(t.Context(), budget)
	defer cancel()
	var memBefore, memAfter runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)
	var cpuBefore, cpuAfter syscall.Rusage
	_ = syscall.Getrusage(syscall.RUSAGE_SELF, &cpuBefore)
	var peak atomic.Int64
	peak.Store(int64(runtime.NumGoroutine()))
	stop := make(chan struct{})
	joined := make(chan struct{})
	go func() {
		defer close(joined)
		ticker := time.NewTicker(2 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				n := int64(runtime.NumGoroutine())
				if n > peak.Load() {
					peak.Store(n)
				}
			}
		}
	}()
	started := time.Now()
	result, err := opt.Optimize(ctx, problem)
	elapsed := time.Since(started)
	close(stop)
	<-joined
	_ = syscall.Getrusage(syscall.RUSAGE_SELF, &cpuAfter)
	runtime.ReadMemStats(&memAfter)
	state := "completed"
	if err != nil {
		if !allowTimeout || !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("%s/%s: %v", name, algorithm, err)
		}
		state = "timeout"
		if len(result.Placements) != 0 {
			t.Fatal("timeout exposed partial placement result")
		}
	} else {
		validateExperiment(t, problem, result)
		if name != "radial_without_seed" && result.Metrics.PlacedCount == 0 {
			t.Fatalf("fixture %s did not exercise actual placements", name)
		}
	}
	cpu := float64(cpuAfter.Utime.Sec-cpuBefore.Utime.Sec+cpuAfter.Stime.Sec-cpuBefore.Stime.Sec)*1000 + float64(cpuAfter.Utime.Usec-cpuBefore.Utime.Usec+cpuAfter.Stime.Usec-cpuBefore.Stime.Usec)/1000
	row := map[string]any{"fixture": name, "algorithm": algorithm, "quantity": problem.Patterns[0].Quantity, "vertices": len(problem.Patterns[0].Geometry.Exterior.Points), "holes": len(problem.Surface.Holes), "angles": len(problem.AllowedRotations), "wall_ms": float64(elapsed) / float64(time.Millisecond), "cpu_ms": cpu, "bytes": memAfter.TotalAlloc - memBefore.TotalAlloc, "allocs": memAfter.Mallocs - memBefore.Mallocs, "peak_goroutines_sampled": peak.Load(), "maxrss_kib_process_cumulative": cpuAfter.Maxrss, "nfp_calls": calls.calls.Load(), "cache_misses": miss.calls.Load(), "cache_hits": calls.calls.Load() - miss.calls.Load(), "max_nfp_call_ms": float64(miss.maxNS.Load()) / float64(time.Millisecond), "placed": result.Metrics.PlacedCount, "unplaced": problem.Patterns[0].Quantity - result.Metrics.PlacedCount, "utilization": result.Metrics.Utilization, "status": state, "valid": err == nil}
	if err != nil {
		row["cache_hits"] = nil
	}
	data, _ := json.Marshal(row)
	t.Log("GEOMETRY_ROW " + string(data))
}
