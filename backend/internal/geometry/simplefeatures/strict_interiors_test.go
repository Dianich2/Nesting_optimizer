package simplefeatures

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"os"
	"slices"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	dg "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
)

func strictPoly(xy ...float64) dg.Polygon {
	p := dg.Polygon{}
	for i := 0; i < len(xy); i += 2 {
		p.Exterior.Points = append(p.Exterior.Points, dg.Point{X: xy[i], Y: xy[i+1]})
	}
	return p
}
func strictRect(x0, y0, x1, y1 float64) dg.Polygon { return strictPoly(x0, y0, x1, y0, x1, y1, x0, y1) }
func strictFixture(t testing.TB) (dg.Polygon, dg.Polygon) {
	t.Helper()
	data, err := os.ReadFile("testdata/stage12_contact_counterexample.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct{ First, Second dg.Polygon }
	if err = json.Unmarshal(data, &f); err != nil {
		t.Fatal(err)
	}
	return f.First, f.Second
}
func strictRingVariant(p dg.Polygon, reverse, closed bool) dg.Polygon {
	rev := func(r dg.Ring) dg.Ring {
		r.Points = slices.Clone(r.Points)
		if reverse {
			slices.Reverse(r.Points)
		}
		if closed {
			r.Points = append(r.Points, r.Points[0])
		}
		return r
	}
	p.Exterior = rev(p.Exterior)
	p.Holes = slices.Clone(p.Holes)
	for i := range p.Holes {
		p.Holes[i] = rev(p.Holes[i])
	}
	return p
}
func TestStrictTopologyMatrix(t *testing.T) {
	a, b := strictFixture(t)
	hole := strictRect(0, 0, 10, 10)
	hole.Holes = []dg.Ring{strictRect(2, 2, 4, 4).Exterior, strictRect(5, 2, 6, 4).Exterior, strictRect(7, 7, 9, 9).Exterior}
	cases := []struct {
		name string
		a, b dg.Polygon
		want bool
	}{
		{"original", a, b, true},
		{"edge", strictRect(0, 0, 1, 1), strictRect(1, 0, 2, 1), false},
		{"vertex", strictRect(0, 0, 1, 1), strictPoly(1, 1, 2, 1, 1, 2), false},
		{"gap", strictRect(0, 0, 1, 1), strictRect(math.Nextafter(1, 2), 0, 2, 1), false},
		{"diagonal-gap", strictPoly(0, 0, 2, 0, 0, 2), strictPoly(1, math.Nextafter(1, 2), 3, 1, 1, 3), false},
		{"diagonal-ulp-overlap", strictPoly(0, 0, 2, 0, 0, 2), strictPoly(1, math.Nextafter(1, 0), 3, 1, 1, 3), true},
		{"diagonal-edge", strictPoly(0, 0, 2, 0, 0, 2), strictPoly(2, 0, 2, 2, 0, 2), false},
		{"diagonal-shared-inside", strictPoly(0, 0, 2, 0, 0, 2), strictPoly(0, 2, 2, 0, 1, 0), true},
		{"ulp-overlap", strictRect(0, 0, 1, 1), strictRect(math.Nextafter(1, 0), 0, 2, 1), true},
		{"normal", strictRect(-2, -2, 2, 2), strictRect(1, 1, 3, 3), true},
		{"identical", strictRect(0, 0, 1, 1), strictRect(0, 0, 1, 1), true},
		{"contained", strictRect(0, 0, 10, 10), strictRect(2, 2, 3, 3), true},
		{"in-hole", hole, strictRect(2.5, 2.5, 3.5, 3.5), false},
		{"equal-hole", hole, strictRect(2, 2, 4, 4), false},
		{"hole-edge", hole, strictRect(2, 2.5, 3, 3.5), false},
		{"hole-cross", hole, strictRect(1, 2.5, 3, 3.5), true},
		{"hole-ulp-cross", hole, strictRect(math.Nextafter(2, 1), 2.5, 3, 3.5), true},
		{"inscribed-vertices", strictRect(0, 0, 4, 4), strictPoly(2, 0, 4, 2, 2, 4, 0, 2), true},
		{"T-junction-cross", strictPoly(0, 0, 4, 0, 4, 4, 2, 2, 0, 4), strictPoly(-1, 2, 2, 2, 5, 2, 5, 3, -1, 3), true},
		{"concave-gap", strictPoly(0, 0, 4, 0, 4, 1, 1, 1, 1, 4, 0, 4), strictRect(2, 2, 3, 3), false},
		{"large", strictRect(1e12, 1e12, 1e12+10, 1e12+10), strictRect(1e12+9, 1e12+9, 1e12+20, 1e12+20), true},
		{"small", strictRect(0, 0, 1e-100, 1e-100), strictRect(5e-101, 5e-101, 2e-100, 2e-100), true},
	}
	engine := NewEngine()
	if path := os.Getenv("NESTING_STRICT_MATRIX"); path != "" {
		type row struct {
			First, Second dg.Polygon
			Want          bool
		}
		rows := []row{}
		for _, c := range cases {
			rows = append(rows, row{c.a, c.b, c.want})
		}
		data, err := json.Marshal(rows)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, c := range cases {
		if got := strictScanOracle(c.a, c.b); got != c.want {
			t.Fatalf("exact slab oracle %s got %v want %v", c.name, got, c.want)
		}
	}
	for _, c := range cases {
		for _, closed := range []bool{false, true} {
			for _, reverse := range []bool{false, true} {
				for _, swap := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/closed%v/reverse%v/swap%v", c.name, closed, reverse, swap), func(t *testing.T) {
						a, b := c.a, c.b
						if reverse || closed {
							a = strictRingVariant(a, reverse, closed)
							b = strictRingVariant(b, reverse, closed)
						}
						if swap {
							a, b = b, a
						}
						got, err := engine.InteriorsIntersect(a, b)
						if err != nil || got != c.want {
							t.Fatalf("got %v err %v want %v", got, err, c.want)
						}
					})
				}
			}
		}
	}
	for _, angle := range []float64{33, 45, 90} {
		t.Run(fmt.Sprintf("rotation%g", angle), func(t *testing.T) {
			a, err := engine.Transform(strictRect(0, 0, 4, 2), 0, 0, angle)
			if err != nil {
				t.Fatal(err)
			}
			b, err := engine.Transform(strictRect(0, 0, 1, 1), 0, 0, angle)
			if err != nil {
				t.Fatal(err)
			}
			got, err := engine.InteriorsIntersect(a, b)
			if err != nil || !got {
				t.Fatalf("%v %v", got, err)
			}
		})
	}
	for _, p := range []dg.Polygon{{}, strictPoly(0, 0, 1, 1, 0, 1, 1, 0), strictPoly(0, 0, math.NaN(), 1, 1, 0), strictPoly(0, 0, math.Inf(1), 1, 1, 0)} {
		if _, err := engine.InteriorsIntersect(p, strictRect(0, 0, 1, 1)); !errors.Is(err, geometry.ErrInvalidPolygon) {
			t.Fatalf("invalid input: %v", err)
		}
	}
}

// Independent exact oracle: partition the plane into open horizontal slabs at
// ALL input vertex ordinates. Edges cannot change order inside a slab unless
// they cross, so also include exact ordinates of all proper edge intersections.
// At each rational sample Y, compare open inside intervals. No production
// orientation, point classifier, intersection coordinate or tolerance is used.
func strictScanOracle(a, b dg.Polygon) bool {
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
			for k := 0; k+1 < len(z); k += 2 {
				if x[j].Cmp(z[k+1]) < 0 && z[k].Cmp(x[j+1]) < 0 {
					return true
				}
			}
		}
	}
	return false
}
func TestStrictAgainstExactSlabOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(20261124))
	engine := NewEngine()
	for i := 0; i < 300; i++ {
		shape := func() dg.Polygon {
			p := dg.Polygon{}
			n := 3 + rng.Intn(8)
			dx, dy := rng.Float64()*5, rng.Float64()*5
			for j := 0; j < n; j++ {
				a := float64(j) * 2 * math.Pi / float64(n)
				r := 1 + rng.Float64()*3
				p.Exterior.Points = append(p.Exterior.Points, dg.Point{X: dx + r*math.Cos(a), Y: dy + r*math.Sin(a)})
			}
			return p
		}
		a, b := shape(), shape()
		got, err := engine.InteriorsIntersect(a, b)
		if err != nil {
			t.Fatalf("case%d: %v", i, err)
		}
		want := strictScanOracle(a, b)
		if got != want {
			t.Fatalf("case%d: got %v want %v a=%+v b=%+v", i, got, want, a, b)
		}
	}
}
func TestStrictOrientationExtremes(t *testing.T) {
	for _, scale := range []float64{1, 1e-200, 1e200} {
		s := strictState{}
		a, b, c := dg.Point{}, dg.Point{X: scale, Y: scale}, dg.Point{X: scale, Y: math.Nextafter(scale, math.Inf(1))}
		if s.orient(a, b, c) != 1 || s.err != nil {
			t.Fatalf("scale=%g state=%+v", scale, s)
		}
	}
	s := strictState{exact: strictExactLimit}
	s.orient(dg.Point{}, dg.Point{X: 1, Y: 1}, dg.Point{X: 2, Y: 2})
	if !errors.Is(s.err, geometry.ErrComplexity) {
		t.Fatalf("budget: %v", s.err)
	}
	s = strictState{work: strictWorkLimit}
	_, err := strictInteriors(strictRect(0, 0, 2, 2), strictRect(1, 0, 3, 2), &s)
	if !errors.Is(err, geometry.ErrComplexity) {
		t.Fatalf("work budget: %v", err)
	}
}
func BenchmarkStrictInteriors(b *testing.B) {
	a, c := strictFixture(b)
	cases := []struct {
		name string
		a, b dg.Polygon
	}{{"far", strictRect(0, 0, 1, 1), strictRect(3, 0, 4, 1)}, {"touch", strictRect(0, 0, 1, 1), strictRect(1, 0, 2, 1)}, {"overlap", strictRect(0, 0, 2, 2), strictRect(1, 1, 3, 3)}, {"regression", a, c}, {"vertex-crossing", strictRect(0, 0, 4, 4), strictPoly(2, 0, 4, 2, 2, 4, 0, 2)}, {"diagonal-touch", strictPoly(0, 0, 2, 0, 0, 2), strictPoly(1, 1, 3, 1, 1, 3)}}
	for _, n := range []int{10, 50} {
		p := dg.Polygon{}
		for i := 0; i < n; i++ {
			angle := 2 * math.Pi * float64(i) / float64(n)
			r := 5.
			if i%2 == 0 {
				r = 3
			}
			p.Exterior.Points = append(p.Exterior.Points, dg.Point{X: r * math.Cos(angle), Y: r * math.Sin(angle)})
		}
		q := strictRect(-1, -1, 1, 1)
		q.Holes = []dg.Ring{strictRect(-.5, -.5, .5, .5).Exterior}
		cases = append(cases, struct {
			name string
			a, b dg.Polygon
		}{fmt.Sprintf("concave%d-hole", n), p, q})
	}
	for _, c := range cases {
		for _, old := range []bool{true, false} {
			b.Run(fmt.Sprintf("%s/old%v", c.name, old), func(b *testing.B) {
				engine := NewEngine()
				exact := 0
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if old {
						if err := engine.ValidatePolygon(c.a); err != nil {
							b.Fatal(err)
						}
						if err := engine.ValidatePolygon(c.b); err != nil {
							b.Fatal(err)
						}
						if _, err := geom.Relate(toSimpleFeaturesPolygon(c.a).AsGeometry(), toSimpleFeaturesPolygon(c.b).AsGeometry()); err != nil {
							b.Fatal(err)
						}
					} else {
						if err := engine.ValidatePolygon(c.a); err != nil {
							b.Fatal(err)
						}
						if err := engine.ValidatePolygon(c.b); err != nil {
							b.Fatal(err)
						}
						state := strictState{}
						if _, err := strictInteriors(c.a, c.b, &state); err != nil {
							b.Fatal(err)
						}
						exact += state.exact
					}
				}
				b.ReportMetric(float64(exact)/float64(b.N), "exact/op")
			})
		}
	}
}

func TestStrictTransitionExact(t *testing.T) {
	path := os.Getenv("NESTING_STRICT_TRANSITION")
	if path == "" {
		t.Skip("explicit bounded transition evidence")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var evidence struct {
		Pairs []struct {
			First, Second dg.Polygon
			Strict        bool
		}
	}
	if err = json.Unmarshal(data, &evidence); err != nil || len(evidence.Pairs) == 0 {
		t.Fatalf("invalid evidence: %v", err)
	}
	for i, p := range evidence.Pairs {
		if got := strictScanOracle(p.First, p.Second); got != p.Strict {
			t.Fatalf("exact transition pair%d got %v want %v", i, got, p.Strict)
		}
	}
	t.Logf("EXACT_TRANSITION pairs=%d", len(evidence.Pairs))
}

func TestStrictFilteredOrientationAgainstRational(t *testing.T) {
	rng := rand.New(rand.NewSource(20261125))
	exact := func(a, b, c dg.Point) int {
		dx1 := new(big.Rat).Sub(rat(a.X), rat(c.X))
		dy1 := new(big.Rat).Sub(rat(a.Y), rat(c.Y))
		dx2 := new(big.Rat).Sub(rat(b.X), rat(c.X))
		dy2 := new(big.Rat).Sub(rat(b.Y), rat(c.Y))
		return new(big.Rat).Sub(new(big.Rat).Mul(dx1, dy2), new(big.Rat).Mul(dy1, dx2)).Sign()
	}
	fallbacks := 0
	for i := 0; i < 3000; i++ {
		point := func() dg.Point {
			return dg.Point{X: math.Ldexp(rng.Float64()*2-1, rng.Intn(2000)-1000), Y: math.Ldexp(rng.Float64()*2-1, rng.Intn(2000)-1000)}
		}
		a, b, c := point(), point(), point()
		if i%3 == 0 {
			c = b
			c.X = math.Nextafter(c.X, math.Inf(1))
		}
		state := strictState{}
		got := state.orient(a, b, c)
		want := exact(a, b, c)
		if got != want || state.err != nil {
			t.Fatalf("orientation case%d got%d want%d state%+v", i, got, want, state)
		}
		fallbacks += state.exact
	}
	t.Logf("FILTER_PROOF triples=3000 exact_fallbacks=%d", fallbacks)
}

func TestStrictMidpointCannotRoundToBoundary(t *testing.T) {
	s := strictState{}
	lo, hi := 1., math.Nextafter(1, 2)
	edges := s.edges(strictRect(lo, 0, hi, 2))
	if got := s.midpointLocation(dg.Point{X: lo, Y: 1}, dg.Point{X: hi, Y: 1}, edges); got != 1 || s.err != nil || s.exact == 0 {
		t.Fatalf("unrepresentable midpoint got%d state%+v", got, s)
	}
}
