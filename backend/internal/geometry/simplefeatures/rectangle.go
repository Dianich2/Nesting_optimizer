package simplefeatures

import domaingeometry "server_nesting_optimizer/internal/domain/geometry"

// axisAlignedRectangle expects a polygon that has already passed validation.
func axisAlignedRectangle(polygon domaingeometry.Polygon) (domaingeometry.Bounds, bool) {
	if len(polygon.Holes) > 0 {
		return domaingeometry.Bounds{}, false
	}
	points := polygon.Exterior.Points
	if len(points) == 5 && points[0] == points[4] {
		points = points[:4]
	}
	if len(points) != 4 {
		return domaingeometry.Bounds{}, false
	}
	bounds := domaingeometry.Bounds{MinX: points[0].X, MaxX: points[0].X, MinY: points[0].Y, MaxY: points[0].Y}
	for i, point := range points {
		next := points[(i+1)%4]
		if (point.X == next.X) == (point.Y == next.Y) {
			return domaingeometry.Bounds{}, false
		}
		bounds.MinX = min(bounds.MinX, point.X)
		bounds.MaxX = max(bounds.MaxX, point.X)
		bounds.MinY = min(bounds.MinY, point.Y)
		bounds.MaxY = max(bounds.MaxY, point.Y)
	}
	return bounds, true
}
