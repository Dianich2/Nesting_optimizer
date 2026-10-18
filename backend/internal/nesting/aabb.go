package nesting

import domaingeometry "server_nesting_optimizer/internal/domain/geometry"

func aabbInteriorsMayOverlap(
	first domaingeometry.Bounds,
	second domaingeometry.Bounds,
) bool {
	return first.MinX < second.MaxX &&
		second.MinX < first.MaxX &&
		first.MinY < second.MaxY &&
		second.MinY < first.MaxY
}
