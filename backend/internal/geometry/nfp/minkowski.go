package nfp

import (
	"fmt"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
)

func convexMinkowskiSum(
	first domaingeometry.Polygon,
	second domaingeometry.Polygon,
) (domaingeometry.Polygon, error) {
	if len(first.Holes) != 0 || len(second.Holes) != 0 {
		return domaingeometry.Polygon{}, fmt.Errorf(
			"calculate convex Minkowski sum: invalid polygon",
		)
	}

	points := make([]domaingeometry.Point, 0,
		len(first.Exterior.Points)*len(second.Exterior.Points),
	)

	for _, firstPoint := range first.Exterior.Points {
		for _, secondPoint := range second.Exterior.Points {
			points = append(points, domaingeometry.Point{
				X: firstPoint.X + secondPoint.X,
				Y: firstPoint.Y + secondPoint.Y,
			})
		}
	}

	result, err := pointsToConvexHull(points)
	if err != nil {
		return domaingeometry.Polygon{}, fmt.Errorf(
			"calculate convex Minkowski sum: %w",
			err,
		)
	}

	return result, nil
}
