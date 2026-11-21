package simplefeatures

import (
	"encoding/json"
	"fmt"
	"github.com/peterstace/simplefeatures/geom"
	"math"
	"math/big"
	"math/rand"
	"os"
	dg "server_nesting_optimizer/internal/domain/geometry"
	"slices"
	"testing"
)

// Independent rational slab inclusion oracle. All input vertex and boundary-
// crossing Y values partition intervals; no production predicate is called.
func stage17CoverageOracle(a, b dg.Polygon) bool {
	type edge struct{ x, y, u, v *big.Rat }
	edges := func(p dg.Polygon) []edge {
		out := []edge{}
		rings := append([]dg.Ring{p.Exterior}, p.Holes...)
		for _, r := range rings {
			points := r.Points
			for i, a := range points {
				b := points[(i+1)%len(points)]
				out = append(out, edge{rat(a.X), rat(a.Y), rat(b.X), rat(b.Y)})
			}
		}
		return out
	}
	ea, eb := edges(a), edges(b)
	ys := []*big.Rat{}
	for _, e := range append(slices.Clone(ea), eb...) {
		ys = append(ys, e.y)
	}
	sub := func(a, b *big.Rat) *big.Rat { return new(big.Rat).Sub(a, b) }
	mul := func(a, b *big.Rat) *big.Rat { return new(big.Rat).Mul(a, b) }
	cross := func(x, y, u, v *big.Rat) *big.Rat { return sub(mul(x, v), mul(y, u)) }
	for _, e := range ea {
		for _, f := range eb {
			dx, dy, ux, uy := sub(e.u, e.x), sub(e.v, e.y), sub(f.u, f.x), sub(f.v, f.y)
			det := cross(dx, dy, ux, uy)
			if det.Sign() == 0 {
				continue
			}
			t := new(big.Rat).Quo(cross(sub(f.x, e.x), sub(f.y, e.y), ux, uy), det)
			q := new(big.Rat).Quo(cross(sub(f.x, e.x), sub(f.y, e.y), dx, dy), det)
			if t.Sign() > 0 && t.Cmp(big.NewRat(1, 1)) < 0 && q.Sign() > 0 && q.Cmp(big.NewRat(1, 1)) < 0 {
				ys = append(ys, new(big.Rat).Add(e.y, mul(t, dy)))
			}
		}
	}
	slices.SortFunc(ys, func(a, b *big.Rat) int { return a.Cmp(b) })
	xs := func(edges []edge, y *big.Rat) []*big.Rat {
		out := []*big.Rat{}
		for _, e := range edges {
			if (e.y.Cmp(y) > 0) == (e.v.Cmp(y) > 0) {
				continue
			}
			x := new(big.Rat).Add(e.x, new(big.Rat).Quo(mul(sub(y, e.y), sub(e.u, e.x)), sub(e.v, e.y)))
			out = append(out, x)
		}
		slices.SortFunc(out, func(a, b *big.Rat) int { return a.Cmp(b) })
		return out
	}
	for i := 1; i < len(ys); i++ {
		if ys[i].Cmp(ys[i-1]) == 0 {
			continue
		}
		y := new(big.Rat).Quo(new(big.Rat).Add(ys[i], ys[i-1]), big.NewRat(2, 1))
		x, z := xs(ea, y), xs(eb, y)
		for j := 0; j+1 < len(x); j += 2 {
			included := false
			for k := 0; k+1 < len(z); k += 2 {
				if z[k].Cmp(x[j]) <= 0 && z[k+1].Cmp(x[j+1]) >= 0 {
					included = true
					break
				}
			}
			if !included {
				return false
			}
		}
	}
	return true
}

type stage17CoverageRow struct {
	Name         string
	Inner, Outer dg.Polygon
	Want         bool
}

func TestStage17CoverageMatrix(t *testing.T) {
	tri := strictPoly(0, 0, 10, 0, 0, 10)
	concave := strictPoly(0, 0, 10, 0, 10, 2, 2, 2, 2, 8, 10, 8, 10, 10, 0, 10)
	holes := strictRect(0, 0, 20, 20)
	holes.Holes = []dg.Ring{strictRect(2, 2, 4, 4).Exterior, strictPoly(10, 10, 15, 10, 10, 15).Exterior}
	first, second := strictFixture(t)
	original := strictRect(-10, -10, 20, 20)
	original.Holes = []dg.Ring{first.Exterior}
	rows := []stage17CoverageRow{
		{"original-hole", second, original, false},
		{"triangle-in", strictRect(1, 1, 2, 2), tri, true},
		{"triangle-out", strictRect(8, 8, 9, 9), tri, false},
		{"triangle-edge", strictPoly(1, 1, 5, 5, 1, 5), tri, true},
		{"triangle-ulp-out", strictPoly(1, 1, 5, math.Nextafter(5, 10), 1, 5), tri, false},
		{"triangle-ulp-in", strictPoly(1, 1, 5, math.Nextafter(5, 0), 1, 5), tri, true},
		{"concave-in", strictRect(0, 0, 2, 2), concave, true},
		{"concave-bridge", strictPoly(1, 1, 9, 1, 1, 9), concave, false},
		{"concave-long-edge", strictPoly(1, 1, 9, 1, 9, 9, 1, 9), concave, false},
		{"hole-in", strictRect(2.5, 2.5, 3.5, 3.5), holes, false},
		{"hole-equal", strictRect(2, 2, 4, 4), holes, false},
		{"hole-touch", strictRect(0, 2, 2, 4), holes, true},
		{"hole-ulp-enter", strictRect(0, 2, math.Nextafter(2, 3), 4), holes, false},
		{"hole-ulp-gap", strictRect(0, 2, math.Nextafter(2, 1), 4), holes, true},
		{"hole-enclosed", strictRect(1, 1, 5, 5), holes, false},
		{"identical-holes", holes, holes, true},
		{"collinear", strictPoly(0, 0, 1, 0, 2, 0, 0, 2), tri, true},
	}
	// Powers-of-two scaling preserves source doubles exactly; integer
	// translations below preserve these small dyadic fixtures too.
	transform := func(p dg.Polygon, exp int, delta float64) dg.Polygon {
		ring := func(r dg.Ring) dg.Ring {
			q := dg.Ring{}
			for _, v := range r.Points {
				q.Points = append(q.Points, dg.Point{X: math.Ldexp(v.X, exp) + delta, Y: math.Ldexp(v.Y, exp) + delta})
			}
			return q
		}
		q := dg.Polygon{Exterior: ring(p.Exterior)}
		for _, h := range p.Holes {
			q.Holes = append(q.Holes, ring(h))
		}
		return q
	}
	for _, exp := range []int{-100, 40} {
		for _, base := range append([]stage17CoverageRow(nil), rows[1:6]...) {
			rows = append(rows, stage17CoverageRow{Name: "dyadic-scale", Inner: transform(base.Inner, exp, 0), Outer: transform(base.Outer, exp, 0), Want: base.Want})
		}
	}
	rows = append(rows, stage17CoverageRow{Name: "negative", Inner: transform(strictRect(1, 1, 2, 2), 0, -20), Outer: transform(tri, 0, -20), Want: true})
	data, err := os.ReadFile("testdata/stage17_coverage_counterexample.json")
	if err != nil {
		t.Fatal(err)
	}
	var minimal struct{ Inner, Outer dg.Polygon }
	if err = json.Unmarshal(data, &minimal); err != nil {
		t.Fatal(err)
	}
	rows = append(rows, stage17CoverageRow{Name: "minimal", Inner: minimal.Inner, Outer: minimal.Outer, Want: false})
	engine := NewEngine()
	random := rand.New(rand.NewSource(20261127))
	for i := 0; i < 120; i++ {
		x, y := random.Float64()*12, random.Float64()*12
		inner := strictRect(x, y, x+.1+random.Float64()*2, y+.1+random.Float64()*2)
		rows = append(rows, stage17CoverageRow{Name: "seed20261127", Inner: inner, Outer: holes, Want: stage17CoverageOracle(inner, holes)})
	}
	for _, angle := range []float64{0, 30, 45, 90} {
		p, err := engine.Transform(strictRect(0, 0, 1, 1), 2, 2, angle)
		if err != nil {
			t.Fatal(err)
		}
		rows = append(rows, stage17CoverageRow{Name: "formed-rotation", Inner: p, Outer: tri, Want: stage17CoverageOracle(p, tri)})
	}
	data, err = os.ReadFile("../../nesting/testdata/stage12_property12_counterexample.json")
	if err != nil {
		t.Fatal(err)
	}
	var transformedFixture struct {
		Problem struct {
			Patterns []struct{ Geometry dg.Polygon }
		}
		Result struct {
			Placements []struct{ X, Y, Rotation float64 }
		}
	}
	if err = json.Unmarshal(data, &transformedFixture); err != nil {
		t.Fatal(err)
	}
	pose := transformedFixture.Result.Placements[1]
	actual, err := engine.Transform(transformedFixture.Problem.Patterns[0].Geometry, pose.X, pose.Y, pose.Rotation)
	if err != nil {
		t.Fatal(err)
	}
	rows = append(rows, stage17CoverageRow{Name: "actual-manual-transform", Inner: actual, Outer: minimal.Outer, Want: false})
	for _, row := range rows {
		for _, rev := range []bool{false, true} {
			for _, closed := range []bool{false, true} {
				a, b := strictRingVariant(row.Inner, rev, closed), strictRingVariant(row.Outer, rev, closed)
				want := stage17CoverageOracle(a, b)
				if want != row.Want {
					t.Fatalf("oracle changed %s", row.Name)
				}
				got, err := engine.CoveredBy(a, b)
				if err != nil || got != want {
					t.Fatalf("%s reverse%v closed%v got%v want%v err%v", row.Name, rev, closed, got, want, err)
				}
			}
		}
	}
	if path := os.Getenv("NESTING_COVERAGE_CORPUS"); path != "" {
		data, err := json.Marshal(rows)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("COVERAGE_MATRIX rows=%d variants=%d", len(rows), len(rows)*4)
}

func BenchmarkStage17Coverage(b *testing.B) {
	tri := strictPoly(0, 0, 10, 0, 0, 10)
	holes := strictRect(0, 0, 20, 20)
	holes.Holes = []dg.Ring{strictRect(2, 2, 4, 4).Exterior, strictPoly(10, 10, 15, 10, 10, 15).Exterior}
	rows := []stage17CoverageRow{{Name: "triangle-in", Inner: strictRect(1, 1, 2, 2), Outer: tri}, {Name: "triangle-ulp-out", Inner: strictPoly(1, 1, 5, math.Nextafter(5, 10), 1, 5), Outer: tri}, {Name: "holes-touch", Inner: strictRect(0, 2, 2, 4), Outer: holes}, {Name: "holes-contained", Inner: strictRect(1, 1, 5, 5), Outer: holes}}
	for _, row := range rows {
		for _, old := range []bool{true, false} {
			b.Run(fmt.Sprintf("%s/old%v", row.Name, old), func(b *testing.B) {
				b.ReportAllocs()
				engine := NewEngine()
				for i := 0; i < b.N; i++ {
					if old {
						if err := engine.ValidatePolygon(row.Inner); err != nil {
							b.Fatal(err)
						}
						if err := engine.ValidatePolygon(row.Outer); err != nil {
							b.Fatal(err)
						}
						if _, err := geom.CoveredBy(toSimpleFeaturesPolygon(row.Inner).AsGeometry(), toSimpleFeaturesPolygon(row.Outer).AsGeometry()); err != nil {
							b.Fatal(err)
						}
					} else {
						if _, err := engine.CoveredBy(row.Inner, row.Outer); err != nil {
							b.Fatal(err)
						}
					}
				}
			})
		}
	}
}
