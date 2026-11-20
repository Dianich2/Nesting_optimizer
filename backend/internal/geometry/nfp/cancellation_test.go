package nfp

import (
	"context"
	"errors"
	"math"
	dg "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
	"testing"
	"time"
)

func complexStar(vertices int) dg.Polygon {
	points := make([]dg.Point, vertices)
	for i := range points {
		radius := 100.0
		if i%2 == 1 {
			radius = 65
		}
		angle := 2 * math.Pi * float64(i) / float64(vertices)
		points[i] = dg.Point{X: radius * math.Cos(angle), Y: radius * math.Sin(angle)}
	}
	return dg.Polygon{Exterior: dg.Ring{Points: points}}
}

func TestCancellationDuringRealNFP(t *testing.T) {
	builder := NewCachedBuilder(NewBuilder(simplefeatures.NewEngine()))
	ctx, cancel := context.WithCancel(t.Context())
	timer := time.AfterFunc(15*time.Millisecond, cancel)
	defer timer.Stop()
	defer cancel()
	start := time.Now()
	result, err := builder.BuildExternalContext(ctx, complexStar(128), complexStar(126))
	if !errors.Is(err, context.Canceled) || len(result.Polygons) != 0 {
		t.Fatalf("real NFP cancellation: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Fatalf("cooperative cancellation took %s", elapsed)
	}
	if builder.count != 0 {
		t.Fatal("partially cancelled geometry entered cache")
	}
	t.Logf("real NFP cancellation completed in %s", time.Since(start))
	_, err = builder.BuildExternalContext(t.Context(), complexStar(4), complexStar(4))
	if err != nil {
		t.Fatalf("builder unusable after cancellation: %v", err)
	}
}

func TestNFPCartesianProductBudget(t *testing.T) {
	_, err := NewBuilder(simplefeatures.NewEngine()).BuildExternalContext(t.Context(), complexStar(300), complexStar(300))
	if !errors.Is(err, geometry.ErrComplexity) {
		t.Fatalf("unbounded triangle Cartesian product: %v", err)
	}
}
