package simplefeatures

import (
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestRectangleCoverageMatchesTopology(t *testing.T) {
	outer := domaingeometry.Polygon{Exterior: domaingeometry.Ring{Points: []domaingeometry.Point{
		{X: -10, Y: -20}, {X: 100, Y: -20}, {X: 100, Y: 100}, {X: -10, Y: 100},
	}}}
	// A concave detail with an interior hole exercises more than rectangle bounds.
	inner := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{Points: []domaingeometry.Point{
			{X: 0, Y: 0}, {X: 60, Y: 0}, {X: 60, Y: 20},
			{X: 30, Y: 20}, {X: 30, Y: 60}, {X: 0, Y: 60},
		}},
		Holes: []domaingeometry.Ring{{Points: []domaingeometry.Point{
			{X: 5, Y: 5}, {X: 5, Y: 10}, {X: 10, Y: 10}, {X: 10, Y: 5},
		}}},
	}
	for _, closed := range []bool{false, true} {
		surface := outer
		if closed {
			surface.Exterior.Points = append(append([]domaingeometry.Point{}, outer.Exterior.Points...), outer.Exterior.Points[0])
		}
		for _, offset := range []float64{-11, -10, 0, 40, 40.000001, 120} {
			detail := inner
			detail.Exterior.Points = append([]domaingeometry.Point{}, inner.Exterior.Points...)
			detail.Holes = []domaingeometry.Ring{{Points: append([]domaingeometry.Point{}, inner.Holes[0].Points...)}}
			for i := range detail.Exterior.Points {
				detail.Exterior.Points[i].X += offset
			}
			for i := range detail.Holes[0].Points {
				detail.Holes[0].Points[i].X += offset
			}
			want, err := geom.CoveredBy(toSimpleFeaturesPolygon(detail).AsGeometry(), toSimpleFeaturesPolygon(surface).AsGeometry())
			if err != nil {
				t.Fatal(err)
			}
			got, err := NewEngine().CoveredBy(detail, surface)
			if err != nil || got != want {
				t.Fatalf("closed=%v offset=%v coverage=%v want=%v err=%v", closed, offset, got, want, err)
			}
		}
	}
}
