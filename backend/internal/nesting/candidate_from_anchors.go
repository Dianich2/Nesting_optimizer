package nesting

import domaingeometry "server_nesting_optimizer/internal/domain/geometry"

func candidatesFromAnchors(
	anchors []domaingeometry.Point,
	rotatedPattern domaingeometry.Polygon,
	rotation float64,
) []Candidate {
	patternVertices := ringVertices(rotatedPattern.Exterior)

	var candidates []Candidate
	for _, anchor := range anchors {
		for _, vertex := range patternVertices {
			candidate := Candidate{
				X:        anchor.X - vertex.X,
				Y:        anchor.Y - vertex.Y,
				Rotation: rotation,
			}

			candidates = append(candidates, candidate)
		}
	}

	return candidates
}
