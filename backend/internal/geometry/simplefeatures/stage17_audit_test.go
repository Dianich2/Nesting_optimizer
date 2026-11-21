package simplefeatures

import (
	"encoding/json"
	"errors"
	"github.com/peterstace/simplefeatures/geom"
	"math"
	"os"
	dg "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
	"testing"
)

func stage17Densify(p dg.Polygon, n int) dg.Polygon {
	points := p.Exterior.Points
	p.Exterior.Points = nil
	for i, a := range points {
		b := points[(i+1)%len(points)]
		for j := 0; j < n; j++ {
			f := float64(j) / float64(n)
			p.Exterior.Points = append(p.Exterior.Points, dg.Point{X: a.X + (b.X-a.X)*f, Y: a.Y + (b.Y-a.Y)*f})
		}
	}
	return p
}
func TestStage17CoverageHoleCounterexample(t *testing.T) {
	data, err := os.ReadFile("testdata/stage17_coverage_counterexample.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct{ Inner, Outer dg.Polygon }
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	first := dg.Polygon{Exterior: fixture.Outer.Holes[0]}
	second, surface := fixture.Inner, fixture.Outer
	legacy, err := geom.CoveredBy(toSimpleFeaturesPolygon(second).AsGeometry(), toSimpleFeaturesPolygon(surface).AsGeometry())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("MINIMAL_COVERAGE legacy_covered=%v vertices=9", legacy)
	engine := NewEngine()
	covered, err := engine.CoveredBy(second, surface)
	if err != nil {
		t.Fatal(err)
	}
	// Independent original exact proof: first and second have two proper crossings.
	// Thus second enters the hole and cannot be covered by the usable surface.
	exactOverlap := strictScanOracle(first, second)
	if !exactOverlap {
		t.Fatal("exact proof missing")
	}
	t.Logf("COVERAGE_AUDIT exact_hole_overlap=%v covered=%v", exactOverlap, covered)
	if covered {
		t.Fatal("strict coverage admits a polygon entering the hole")
	}
}
func TestStage17RealComplexity(t *testing.T) {
	a := stage17Densify(strictPoly(0, 0, 10, 0, 10, 2, 2, 2, 2, 8, 10, 8, 10, 10, 0, 10), 128)
	b := stage17Densify(strictRect(4, 4, 6, 6), 256)
	e := NewEngine()
	if err := e.ValidatePolygon(a); err != nil {
		t.Fatal(err)
	}
	if err := e.ValidatePolygon(b); err != nil {
		t.Fatal(err)
	}
	_, err := e.InteriorsIntersect(a, b)
	if !errors.Is(err, geometry.ErrComplexity) {
		t.Fatalf("expected real work-limit error: %v", err)
	}
}

func TestStage17ExactBudgetKeepsComplexityCause(t *testing.T) {
	p := strictPoly(0, 0, 1, 1, 2, math.Nextafter(2, 3))
	if err := NewEngine().ValidatePolygon(p); err != nil {
		t.Fatal(err)
	}
	state := strictState{exact: strictExactLimit}
	state.edges(p)
	if !errors.Is(state.err, geometry.ErrComplexity) {
		t.Fatalf("exact-budget cause replaced: %v", state.err)
	}
}
