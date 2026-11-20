package nfp

import (
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
)

// These are computation budgets, independent of library resource storage limits.
// Triangle pairs are limited before allocating the Cartesian product. Each union
// batch has at most 64 parts, so cancellation can run between library calls.
const maxMinkowskiPieces = 65536
const maxTriangulationVertices = 4096
const maxTriangulationHoles = 64

var ErrComplexity = geometry.ErrComplexity

func checkComplexity(polygons ...domaingeometry.Polygon) error {
	for _, polygon := range polygons {
		vertices := openRingVertices(polygon.Exterior)
		if len(polygon.Holes) > maxTriangulationHoles {
			return ErrComplexity
		}
		for _, hole := range polygon.Holes {
			vertices += openRingVertices(hole)
		}
		if vertices > maxTriangulationVertices {
			return ErrComplexity
		}
	}
	return nil
}

func openRingVertices(ring domaingeometry.Ring) int {
	count := len(ring.Points)
	if count > 1 && ring.Points[0] == ring.Points[count-1] {
		count--
	}
	return count
}
