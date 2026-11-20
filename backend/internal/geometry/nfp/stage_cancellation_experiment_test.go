package nfp

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	dg "server_nesting_optimizer/internal/domain/geometry"
	"sync/atomic"
	"testing"
	"time"
)

// Gap includes native work and scheduling between REAL production Err checks;
// it is not advertised as exact time spent in one third-party function.
type measuredContext struct {
	context.Context
	last, maxGap atomic.Int64
}

func (m *measuredContext) Err() error {
	now := time.Now().UnixNano()
	last := m.last.Swap(now)
	if last != 0 {
		gap := now - last
		for old := m.maxGap.Load(); gap > old; old = m.maxGap.Load() {
			if m.maxGap.CompareAndSwap(old, gap) {
				break
			}
		}
	}
	return m.Context.Err()
}
func TestNFPStageCancellationExperiment(t *testing.T) {
	if os.Getenv("NESTING_GEOMETRY_EXPERIMENT") != "1" {
		t.Skip("explicit bounded stage experiment")
	}
	polygon := complexStar(128)
	triangles, err := triangulate(polygon)
	if err != nil {
		t.Fatal(err)
	}
	pieces, err := reliabilityMinkowskiParts(triangles[:32])
	if err != nil {
		t.Fatal(err)
	}
	stages := map[string]func(context.Context) error{
		"earcut": func(ctx context.Context) error { _, e := triangulateContext(ctx, complexStar(4096)); return e },
		"Minkowski_series": func(ctx context.Context) error {
			for _, a := range triangles {
				for _, b := range triangles {
					if _, e := convexMinkowskiSumContext(ctx, a, b); e != nil {
						return e
					}
				}
			}
			return nil
		},
		"Union_batches": func(ctx context.Context) error { _, e := unionPolygonsContext(ctx, pieces); return e },
	}
	for _, name := range []string{"earcut", "Minkowski_series", "Union_batches"} {
		for repeat := 0; repeat < 5; repeat++ {
			full := &measuredContext{Context: t.Context()}
			fullStart := time.Now()
			if err := stages[name](full); err != nil {
				t.Fatal(err)
			}
			fullWall := time.Since(fullStart)
			base, cancel := context.WithCancel(t.Context())
			ctx := &measuredContext{Context: base}
			var requested atomic.Int64
			timer := time.AfterFunc(time.Millisecond, func() { requested.Store(time.Now().UnixNano()); cancel() })
			began := time.Now()
			err := stages[name](ctx)
			finished := time.Now()
			timer.Stop()
			cancel()
			if err != nil && !errors.Is(err, context.Canceled) {
				t.Fatal(err)
			}
			latency := float64(0)
			if requested.Load() != 0 {
				latency = float64(finished.UnixNano()-requested.Load()) / 1e6
			}
			if finished.Sub(began) > 5*time.Second {
				t.Fatal("stage exceeded cancellation experiment budget")
			}
			row := map[string]any{"uncancelled_wall_ms": float64(fullWall) / float64(time.Millisecond), "uncancelled_max_gap_ms": float64(full.maxGap.Load()) / 1e6, "stage": name, "repeat": repeat, "wall_ms": float64(finished.Sub(began)) / float64(time.Millisecond), "cancel_to_return_ms": latency, "cancel_requested": requested.Load() != 0, "cancelled": errors.Is(err, context.Canceled), "max_gap_between_ctx_checks_ms": float64(ctx.maxGap.Load()) / 1e6}
			encoded, _ := json.Marshal(row)
			t.Log("CANCEL_ROW " + string(encoded))
		}
	}
}
func TestNFPComplexityBoundaryChecks(t *testing.T) {
	if err := checkComplexity(complexStar(4096)); err != nil {
		t.Fatal(err)
	}
	if !errors.Is(checkComplexity(complexStar(4097)), ErrComplexity) {
		t.Fatal("vertex budget bypass")
	}
	p := complexStar(4)
	for i := 0; i < 64; i++ {
		p.Holes = append(p.Holes, dg.Ring{Points: []dg.Point{{X: float64(i), Y: 0}, {X: float64(i) + 0.1, Y: 0}, {X: float64(i), Y: 0.1}}})
	}
	if err := checkComplexity(p); err != nil {
		t.Fatal(err)
	}
	p.Holes = append(p.Holes, p.Holes[0])
	if !errors.Is(checkComplexity(p), ErrComplexity) {
		t.Fatal("hole budget bypass")
	}
}
