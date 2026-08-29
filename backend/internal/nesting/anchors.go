package nesting

import (
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
)

func collectAnchors(
	surface domaingeometry.Polygon,
	occupied []domaingeometry.Polygon,
) []domaingeometry.Point {
	var anchors []domaingeometry.Point

	surfaceVertices := ringVertices(surface.Exterior)
	anchors = append(anchors, surfaceVertices...)

	for _, hole := range surface.Holes {
		anchors = append(anchors, ringVertices(hole)...)
	}

	for _, polygon := range occupied {
		anchors = append(anchors, ringVertices(polygon.Exterior)...)
	}

	return anchors
}
