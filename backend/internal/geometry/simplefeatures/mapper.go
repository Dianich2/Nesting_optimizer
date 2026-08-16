package simplefeatures

import (
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"

	"github.com/peterstace/simplefeatures/geom"
)

func toSimpleFeaturesRing(
	ring domaingeometry.Ring,
) geom.LineString {
	coords := make([]float64, 0, 2*(len(ring.Points)+1))

	for _, point := range ring.Points {
		coords = append(
			coords,
			point.X,
			point.Y,
		)
	}

	if len(ring.Points) > 0 {
		coords = append(
			coords,
			coords[0],
			coords[1],
		)
	}

	return geom.NewLineStringXY(coords...)
}

func toSimpleFeaturesPolygon(
	polygon domaingeometry.Polygon,
) geom.Polygon {
	rings := make([]geom.LineString, 0, len(polygon.Holes)+1)

	rings = append(
		rings,
		toSimpleFeaturesRing(polygon.Exterior),
	)

	for _, hole := range polygon.Holes {
		rings = append(
			rings,
			toSimpleFeaturesRing(hole),
		)
	}

	return geom.NewPolygon(rings)
}

func fromSimpleFeaturesRing(
	ring geom.LineString,
) domaingeometry.Ring {
	coords := ring.Coordinates()

	coordsLen := coords.Length()
	if coordsLen != 0 && coords.GetXY(0) == coords.GetXY(coordsLen-1) {
		coordsLen--
	}

	points := make([]domaingeometry.Point, 0, coordsLen)

	for i := 0; i < coordsLen; i++ {
		coord := coords.GetXY(i)
		points = append(
			points,
			domaingeometry.Point{
				X: coord.X,
				Y: coord.Y,
			},
		)
	}

	return domaingeometry.Ring{
		Points: points,
	}
}

func fromSimpleFeaturesPolygon(
	polygon geom.Polygon,
) domaingeometry.Polygon {
	rings := polygon.DumpRings()

	if len(rings) == 0 {
		return domaingeometry.Polygon{}
	}

	domainRings := make([]domaingeometry.Ring, 0, len(rings))
	for _, ring := range rings {
		domainRings = append(
			domainRings,
			fromSimpleFeaturesRing(ring),
		)
	}

	var holes []domaingeometry.Ring
	if len(rings) > 1 {
		holes = domainRings[1:]
	}

	return domaingeometry.Polygon{
		Exterior: domainRings[0],
		Holes:    holes,
	}
}
