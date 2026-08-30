package nfp

import domaingeometry "server_nesting_optimizer/internal/domain/geometry"

func reflectAtOrigin(
	polygon domaingeometry.Polygon,
) domaingeometry.Polygon {
	newPolygon := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: make([]domaingeometry.Point, 0, len(polygon.Exterior.Points)),
		},
	}

	for _, exteriorPoint := range polygon.Exterior.Points {
		newPolygon.Exterior.Points = append(
			newPolygon.Exterior.Points,
			domaingeometry.Point{
				X: -exteriorPoint.X,
				Y: -exteriorPoint.Y,
			},
		)
	}

	if len(polygon.Holes) != 0 {
		newPolygon.Holes = make([]domaingeometry.Ring, len(polygon.Holes))
	}

	for i, hole := range polygon.Holes {
		newPolygon.Holes[i].Points = make([]domaingeometry.Point, len(hole.Points))
		for j, holePoint := range hole.Points {
			newPolygon.Holes[i].Points[j] = domaingeometry.Point{
				X: -holePoint.X,
				Y: -holePoint.Y,
			}
		}
	}

	return newPolygon
}
