package mapper

import (
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/transport/http/dto"
)

func toDomainRing(
	ring dto.GeometryRing,
) domaingeometry.Ring {
	points := make(
		[]domaingeometry.Point,
		0,
		len(ring.Points),
	)

	for _, point := range ring.Points {
		points = append(
			points,
			domaingeometry.Point{
				X: point.X,
				Y: point.Y,
			},
		)
	}

	return domaingeometry.Ring{
		Points: points,
	}
}

func toGeometryRing(
	ring domaingeometry.Ring,
) dto.GeometryRing {
	points := make(
		[]dto.GeometryPoint,
		0,
		len(ring.Points),
	)

	for _, point := range ring.Points {
		points = append(
			points,
			dto.GeometryPoint{
				X: point.X,
				Y: point.Y,
			},
		)
	}

	return dto.GeometryRing{
		Points: points,
	}
}

func toDomainPolygon(
	polygon dto.PolygonGeometry,
) domaingeometry.Polygon {
	holes := make(
		[]domaingeometry.Ring,
		0,
		len(polygon.Holes),
	)

	for _, hole := range polygon.Holes {
		holes = append(
			holes,
			toDomainRing(hole),
		)
	}

	return domaingeometry.Polygon{
		Exterior: toDomainRing(polygon.Exterior),
		Holes:    holes,
	}
}

func toGeometryPolygon(
	polygon domaingeometry.Polygon,
) dto.PolygonGeometry {
	holes := make(
		[]dto.GeometryRing,
		0,
		len(polygon.Holes),
	)

	for _, hole := range polygon.Holes {
		holes = append(
			holes,
			toGeometryRing(hole),
		)
	}

	return dto.PolygonGeometry{
		Exterior: toGeometryRing(polygon.Exterior),
		Holes:    holes,
	}
}
