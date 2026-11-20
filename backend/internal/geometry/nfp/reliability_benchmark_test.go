package nfp

import (
	"math"
	dg "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
	"testing"
)

func reliabilityContour(concave bool, holes int) dg.Polygon {
	p := dg.Polygon{}
	for i := range 64 {
		r := 100.0
		if concave && i%2 == 1 {
			r = 65
		}
		a := 2 * math.Pi * float64(i) / 64
		p.Exterior.Points = append(p.Exterior.Points, dg.Point{X: r * math.Cos(a), Y: r * math.Sin(a)})
	}
	for i := range holes {
		x := -20 + float64(i)*30
		p.Holes = append(p.Holes, dg.Ring{Points: []dg.Point{{X: x, Y: 0}, {X: x, Y: 5}, {X: x + 5, Y: 5}, {X: x + 5, Y: 0}}})
	}
	return p
}
func BenchmarkReliabilityNFPStages(b *testing.B) {
	shapes := []struct {
		name    string
		concave bool
		holes   int
	}{{"convex", false, 0}, {"concave", true, 0}, {"one_hole", true, 1}, {"two_holes", true, 2}}
	for _, shape := range shapes {
		polygon := reliabilityContour(shape.concave, shape.holes)
		b.Run(shape.name, func(b *testing.B) {
			b.Run("triangulation", func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					parts, err := triangulate(polygon)
					if err != nil {
						b.Fatal(err)
					}
					b.ReportMetric(float64(len(parts)), "triangles/op")
				}
			})
			triangles, err := triangulate(polygon)
			if err != nil {
				b.Fatal(err)
			}
			pieces, err := reliabilityMinkowskiParts(triangles)
			if err != nil {
				b.Fatal(err)
			}
			b.Run("minkowski", func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					pieces, err = reliabilityMinkowskiParts(triangles)
					if err != nil {
						b.Fatal(err)
					}
					b.ReportMetric(float64(len(pieces)), "pieces/op")
				}
			})
			b.Run("union", func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					result, err := unionPolygons(pieces)
					if err != nil {
						b.Fatal(err)
					}
					b.ReportMetric(float64(len(result.Polygons)), "polygons/op")
				}
			})
			b.Run("cache_hit", func(b *testing.B) {
				builder := NewCachedBuilder(NewBuilder(simplefeatures.NewEngine()))
				b.StopTimer()
				if _, err := builder.BuildExternal(polygon, polygon); err != nil {
					b.Fatal(err)
				}
				b.StartTimer()
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					if _, err := builder.BuildExternal(polygon, polygon); err != nil {
						b.Fatal(err)
					}
				}
				b.ReportMetric(float64(builder.hits), "hits")
			})
		})
	}
}

func reliabilityMinkowskiParts(triangles []dg.Polygon) ([]dg.Polygon, error) {
	pieces := make([]dg.Polygon, 0, len(triangles)*len(triangles))
	for _, a := range triangles {
		for _, c := range triangles {
			piece, err := convexMinkowskiSum(a, c)
			if err != nil {
				return nil, err
			}
			pieces = append(pieces, piece)
		}
	}
	return pieces, nil
}
