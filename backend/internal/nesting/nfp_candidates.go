package nesting

import (
	"fmt"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
)

func generateNFPCandidatesForRotations(
	engine geometry.Engine,
	nfpBuilder geometry.NFPBuilder,
	pattern domaingeometry.Polygon,
	surface domaingeometry.Polygon,
	occupied []domaingeometry.Polygon,
	rotations []float64,
) ([]Candidate, error) {
	allCandidates := make([]Candidate, 0)
	seen := make(map[Candidate]struct{})

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
				"generate NFP candidates for rotations: rotate pattern: %w",
				err,
			)
		}

		rotationCandidates, err := generateNFPCandidates(
			nfpBuilder,
			rotatedPattern,
			surface,
			occupied,
			normalizedRotation,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"generate NFP candidates for rotations: %w",
				err,
			)
		}

		for _, rotationCandidate := range rotationCandidates {
			_, ok := seen[rotationCandidate]
			if !ok {
				seen[rotationCandidate] = struct{}{}
				allCandidates = append(allCandidates, rotationCandidate)
			}
		}
	}

	return allCandidates, nil
}

func generateNFPCandidates(
	nfpBuilder geometry.NFPBuilder,
	rotatedPattern domaingeometry.Polygon,
	surface domaingeometry.Polygon,
	occupied []domaingeometry.Polygon,
	rotation float64,
) ([]Candidate, error) {
	candidates := make([]Candidate, 0)
	seen := make(map[Candidate]struct{}, 0)

	var anchors []domaingeometry.Point
	surfaceVertices := ringVertices(surface.Exterior)
	anchors = append(anchors, surfaceVertices...)

	for _, hole := range surface.Holes {
		anchors = append(
			anchors,
			ringVertices(hole)...,
		)
	}

	surfaceCandidates := candidatesFromAnchors(
		anchors,
		rotatedPattern,
		rotation,
	)

	for _, surfaceCandidate := range surfaceCandidates {
		_, ok := seen[surfaceCandidate]
		if !ok {
			seen[surfaceCandidate] = struct{}{}
			candidates = append(candidates, surfaceCandidate)
		}
	}

	for _, occupiedDetail := range occupied {
		nfp, err := nfpBuilder.BuildExternal(
			occupiedDetail,
			rotatedPattern,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"generate NFP candidates: %w",
				err,
			)
		}

		for _, nfpPolygon := range nfp.Polygons {
			for _, exteriorPoint := range nfpPolygon.Exterior.Points {
				candidate := Candidate{
					X:        exteriorPoint.X,
					Y:        exteriorPoint.Y,
					Rotation: rotation,
				}

				_, ok := seen[candidate]
				if !ok {
					seen[candidate] = struct{}{}
					candidates = append(candidates, candidate)
				}
			}

			for _, hole := range nfpPolygon.Holes {
				for _, holePoint := range hole.Points {
					candidate := Candidate{
						X:        holePoint.X,
						Y:        holePoint.Y,
						Rotation: rotation,
					}

					_, ok := seen[candidate]
					if !ok {
						seen[candidate] = struct{}{}
						candidates = append(candidates, candidate)
					}
				}
			}
		}
	}

	return candidates, nil
}
