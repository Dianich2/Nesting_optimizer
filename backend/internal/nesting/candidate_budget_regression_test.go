package nesting

import (
	"context"
	"errors"
	"os"
	"runtime"
	dg "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
	"testing"
	"time"
)

type budgetNFPFixture struct{ result dg.MultiPolygon }

func (b budgetNFPFixture) BuildExternal(a, c dg.Polygon) (dg.MultiPolygon, error) {
	return b.result, nil
}

func TestCandidateBudgetRegression(t *testing.T) {
	t.Run("baseline_raw_product", func(t *testing.T) {
		anchors := make([]dg.Point, 50001)
		result, err := generateCandidates(simplefeatures.NewEngine(), rectangle(10, 10), anchors, []float64{0})
		if !errors.Is(err, geometry.ErrComplexity) || len(result) != 0 {
			t.Fatalf("expected early complexity, got count=%d err=%v", len(result), err)
		}
	})
	t.Run("NFP_inner_polygon", func(t *testing.T) {
		p := dg.Polygon{Exterior: dg.Ring{Points: make([]dg.Point, 200001)}}
		for i := range p.Exterior.Points {
			p.Exterior.Points[i] = dg.Point{X: float64(i) + 1000, Y: 1000}
		}
		result, err := generateNFPCandidatesContext(context.Background(), budgetNFPFixture{dg.MultiPolygon{Polygons: []dg.Polygon{p}}}, rectangle(10, 10), rectangle(100, 100), []dg.Polygon{rectangle(1, 1)}, 0)
		if !errors.Is(err, geometry.ErrComplexity) || len(result) != 0 {
			t.Fatalf("expected inner append guard, got count=%d err=%v", len(result), err)
		}
	})
}

type generationCancelContext struct {
	context.Context
	cancel     context.CancelFunc
	checks, at int
}

func (c *generationCancelContext) Err() error {
	c.checks++
	if c.checks == c.at {
		c.cancel()
	}
	return c.Context.Err()
}

type generationTransformCounter struct {
	geometry.Engine
	calls int
}

func (e *generationTransformCounter) Transform(p dg.Polygon, x, y, r float64) (dg.Polygon, error) {
	e.calls++
	return e.Engine.Transform(p, x, y, r)
}

func TestCandidateProductOverflowAndEdges(t *testing.T) {
	limit := MaxCandidates
	maxInt := int(^uint(0) >> 1)
	for _, tc := range []struct {
		factors []int
		want    int
		fail    bool
	}{{[]int{0, maxInt, maxInt}, 0, false}, {[]int{1, 1}, 1, false}, {[]int{limit - 1, 1}, limit - 1, false}, {[]int{limit, 1}, limit, false}, {[]int{limit + 1, 1}, 0, true}, {[]int{maxInt, maxInt, 36}, 0, true}, {[]int{-1, 0}, 0, true}, {[]int{4096, 4096, 36}, 0, true}} {
		got, err := candidateProduct(limit, tc.factors...)
		if got != tc.want || errors.Is(err, geometry.ErrComplexity) != tc.fail {
			t.Fatalf("product %v = %d,%v", tc.factors, got, err)
		}
	}
}
func TestCandidateBoundaryAllocationAndDuplicates(t *testing.T) {
	pattern := dg.Polygon{Exterior: dg.Ring{Points: []dg.Point{{X: 1, Y: 2}}}}
	for _, n := range []int{0, 1, MaxCandidates - 1, MaxCandidates, MaxCandidates + 1} {
		anchors := make([]dg.Point, n)
		result, err := candidatesFromAnchorsContext(t.Context(), anchors, pattern, 0)
		if n > MaxCandidates {
			if !errors.Is(err, geometry.ErrComplexity) || result != nil {
				t.Fatal("limit+1 accepted")
			}
			continue
		}
		if err != nil || len(result) != n || cap(result) > MaxCandidates {
			t.Fatalf("boundary n=%d count=%d cap=%d err=%v", n, len(result), cap(result), err)
		}
	}
	b := newCandidateBuffer(2, true)
	if err := b.add(t.Context(), Candidate{}); err != nil {
		t.Fatal(err)
	}
	if err := b.add(t.Context(), Candidate{}); err != nil {
		t.Fatal(err)
	}
	if len(b.items) != 1 || b.work != 2 {
		t.Fatal("duplicate work lost")
	}
	if !errors.Is(b.add(t.Context(), Candidate{}), geometry.ErrComplexity) {
		t.Fatal("duplicate work bypassed raw limit")
	}
}
func TestCancellationInsideCandidateGeneration(t *testing.T) {
	base, cancel := context.WithCancel(t.Context())
	defer cancel()
	ctx := &generationCancelContext{Context: base, cancel: cancel, at: 100}
	engine := &generationTransformCounter{Engine: simplefeatures.NewEngine()}
	result, err := generateCandidatesContext(ctx, engine, experimentPolygon(1000), make([]dg.Point, 100), []float64{0})
	if !errors.Is(err, context.Canceled) || result != nil || engine.calls != 1 || ctx.checks != 100 {
		t.Fatalf("inner cancel: checks=%d transforms=%d result=%d err=%v", ctx.checks, engine.calls, len(result), err)
	}
	// Surface anchors, exterior and holes share the same cancellation-aware sink.
	base, cancel = context.WithCancel(t.Context())
	defer cancel()
	ctx = &generationCancelContext{Context: base, cancel: cancel, at: 100}
	polygon := rectangle(1, 1)
	polygon.Holes = []dg.Ring{{Points: make([]dg.Point, 2000)}}
	result, err = generateNFPCandidatesContext(ctx, budgetNFPFixture{dg.MultiPolygon{Polygons: []dg.Polygon{polygon}}}, rectangle(1, 1), rectangle(100, 100), []dg.Polygon{rectangle(1, 1)}, 0)
	if !errors.Is(err, context.Canceled) || result != nil || ctx.checks != 100 {
		t.Fatalf("hole cancel: %d %v", len(result), err)
	}
	expired, stop := context.WithDeadline(t.Context(), time.Now().Add(-time.Second))
	defer stop()
	result, err = generateCandidatesContext(expired, engine, rectangle(1, 1), nil, []float64{0})
	if !errors.Is(err, context.DeadlineExceeded) || result != nil {
		t.Fatal("deadline did not propagate")
	}
}
func TestCandidateBudgetWatchdog(t *testing.T) {
	if os.Getenv("NESTING_CANDIDATE_EXPERIMENT") != "1" {
		t.Skip("explicit hard-capped allocation experiment")
	}
	pattern := experimentPolygon(4096)
	anchors := make([]dg.Point, 4096)
	angles := make([]float64, 36)
	for i := range angles {
		angles[i] = float64(i) * 10
	}
	engine := &generationTransformCounter{Engine: simplefeatures.NewEngine()}
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	start := time.Now()
	result, err := generateCandidatesContext(t.Context(), engine, pattern, anchors, angles)
	runtime.ReadMemStats(&after)
	if !errors.Is(err, geometry.ErrComplexity) || result != nil || engine.calls != 0 {
		t.Fatal("large product reached transformation/allocation")
	}
	if after.TotalAlloc-before.TotalAlloc > 1<<20 {
		t.Fatal("early refusal allocated over1MiB")
	}
	t.Logf("BUDGET_EARLY elapsed=%s allocated=%d transform_calls=%d theoretical_candidates=603979776", time.Since(start), after.TotalAlloc-before.TotalAlloc, engine.calls)
}
