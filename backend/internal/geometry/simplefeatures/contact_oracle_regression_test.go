package simplefeatures

import (
	"encoding/json"
	"github.com/peterstace/simplefeatures/geom"
	"math/big"
	"os"
	dg "server_nesting_optimizer/internal/domain/geometry"
	"testing"
)

// This small deterministic regression intentionally remains RED until an agreed precision
// policy/robust predicate fix exists. See stage12_geometry_invariants_report.md.
func TestStrictContactOracleRegression(t *testing.T) {

	data, e := os.ReadFile("testdata/stage12_contact_counterexample.json")
	if e != nil {
		t.Fatal(e)
	}
	var fixture struct{ First, Second dg.Polygon }
	if e = json.Unmarshal(data, &fixture); e != nil {
		t.Fatal(e)
	}
	engine := NewEngine()
	if e = engine.ValidatePolygon(fixture.First); e != nil {
		t.Fatal(e)
	}
	if e = engine.ValidatePolygon(fixture.Second); e != nil {
		t.Fatal(e)
	}
	orient := func(a, b, c dg.Point) int {
		rat := func(x float64) *big.Rat { return new(big.Rat).SetFloat64(x) }
		dx1 := new(big.Rat).Sub(rat(b.X), rat(a.X))
		dy1 := new(big.Rat).Sub(rat(b.Y), rat(a.Y))
		dx2 := new(big.Rat).Sub(rat(c.X), rat(a.X))
		dy2 := new(big.Rat).Sub(rat(c.Y), rat(a.Y))
		return new(big.Rat).Sub(new(big.Rat).Mul(dx1, dy2), new(big.Rat).Mul(dy1, dx2)).Sign()
	}
	crossings := 0
	a, b := fixture.First.Exterior.Points, fixture.Second.Exterior.Points
	for i, p := range a {
		q := a[(i+1)%len(a)]
		for j, x := range b {
			y := b[(j+1)%len(b)]
			if orient(p, q, x)*orient(p, q, y) < 0 && orient(x, y, p)*orient(x, y, q) < 0 {
				crossings++
			}
		}
	}
	if crossings != 2 {
		t.Fatalf("counterexample exact proof changed: %d crossings", crossings)
	}
	relation, e := geom.Relate(toSimpleFeaturesPolygon(fixture.First).AsGeometry(), toSimpleFeaturesPolygon(fixture.Second).AsGeometry())
	if e != nil {
		t.Fatal(e)
	}
	overlap, e := engine.InteriorsIntersect(fixture.First, fixture.Second)
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("STRICT_CONTACT exact_rational_proper_crossings=%d go_relation=%s go_interiors=%v", crossings, relation, overlap)
	if !overlap {
		t.Fatal("strict mathematical overlap missed: PostGIS relation212101212 and exact rational crossings prove interiors overlap; no tolerance waiver applied")
	}
}
