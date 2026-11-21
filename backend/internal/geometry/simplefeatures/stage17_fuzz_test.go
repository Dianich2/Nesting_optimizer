package simplefeatures

import (
	"math"
	dg "server_nesting_optimizer/internal/domain/geometry"
	"testing"
)

// Bounded differential fuzzing: at most nine vertices, dyadic scales, optional
// one-ULP perturbations. Expectations come from separate exact rational slabs.
func FuzzStrictGeometryAgainstRational(f *testing.F) {
	for _, seed := range [][]byte{{0, 0, 16, 0, 0, 16, 10, 0}, {240, 0, 16, 0, 0, 16, 10, 1}, {0, 0, 32, 0, 0, 32, 10, 0}, {16, 16, 32, 16, 16, 32, 10, 7}, {240, 240, 240, 16, 16, 240, 10, 0}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) != 8 {
			t.Skip()
		}
		scale := math.Ldexp(1, int(data[6]%21)-10)
		inner := dg.Polygon{}
		for i := 0; i < 3; i++ {
			x, y := float64(int8(data[2*i]))/4, float64(int8(data[2*i+1]))/4
			if data[7]&(1<<i) != 0 {
				x = math.Nextafter(x, math.Inf(1))
			}
			inner.Exterior.Points = append(inner.Exterior.Points, dg.Point{X: x * scale, Y: y * scale})
		}
		ring := func(points ...dg.Point) dg.Ring {
			for i := range points {
				points[i].X *= scale
				points[i].Y *= scale
			}
			return dg.Ring{Points: points}
		}
		hole := dg.Polygon{Exterior: ring(dg.Point{X: 0, Y: 0}, dg.Point{X: 8, Y: 0}, dg.Point{X: 0, Y: 8})}
		outer := dg.Polygon{Exterior: ring(dg.Point{X: -64, Y: -64}, dg.Point{X: 128, Y: -64}, dg.Point{X: -64, Y: 128}), Holes: []dg.Ring{hole.Exterior}}
		engine := NewEngine()
		if engine.ValidatePolygon(inner) != nil {
			t.Skip()
		}
		covered, e := engine.CoveredBy(inner, outer)
		if e != nil {
			t.Fatal(e)
		}
		if want := stage17CoverageOracle(inner, outer); covered != want {
			t.Fatalf("coverage=%v rational=%v", covered, want)
		}
		overlap, e := engine.InteriorsIntersect(inner, hole)
		if e != nil {
			t.Fatal(e)
		}
		if want := strictScanOracle(inner, hole); overlap != want {
			t.Fatalf("overlap=%v rational=%v", overlap, want)
		}
	})
}
