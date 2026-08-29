package nesting

import (
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
)

func ringVertices(
	ring domaingeometry.Ring,
) []domaingeometry.Point {
	points := ring.Points
	l := len(points)

	if l == 0 {
		return nil
	}

	if l > 1 {
		first := points[0]
		last := points[l-1]

		if first.X == last.X && first.Y == last.Y {
			l = l - 1
		}
	}

	res := make([]domaingeometry.Point, l)

	copy(res, points[:l])

	return res
}
