package nesting

import (
	"fmt"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
)

func generateCandidates(
	engine geometry.Engine,
	pattern domaingeometry.Polygon,
	anchors []domaingeometry.Point,
	rotations []float64,
) ([]Candidate, error) {
	var candidates []Candidate

	for _, rotation := range rotations {
		normalizedRotation := geometry.NormalizeDegrees(rotation)

		rotatedPattern, err := engine.Transform(
			pattern,
			0,
			0,
			normalizedRotation,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"generate candidates: %w",
				err,
			)
		}

		rotationCandidates := candidatesFromAnchors(
			anchors,
			rotatedPattern,
			normalizedRotation,
		)

		candidates = append(candidates, rotationCandidates...)
	}

	return candidates, nil
}
