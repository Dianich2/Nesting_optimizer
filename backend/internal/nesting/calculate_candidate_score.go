package nesting

import (
	"fmt"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
)

func calculateCandidateScore(
	engine geometry.Engine,
	surface domaingeometry.Polygon,
	occupied []domaingeometry.Polygon,
	candidate domaingeometry.Polygon,
) (CandidateScore, error) {
	surfaceBounds, err := engine.Bounds(surface)
	if err != nil {
		return CandidateScore{}, fmt.Errorf(
			"calculate candidate score: %w",
			err,
		)
	}

	candidateBounds, err := engine.Bounds(candidate)
	if err != nil {
		return CandidateScore{}, fmt.Errorf(
			"calculate candidate score: %w",
			err,
		)
	}

	maxX := candidateBounds.MaxX
	maxY := candidateBounds.MaxY

	for _, occupiedCandidate := range occupied {
		bounds, err := engine.Bounds(occupiedCandidate)
		if err != nil {
			return CandidateScore{}, fmt.Errorf(
				"calculate candidate score: %w",
				err,
			)
		}

		if bounds.MaxX > maxX {
			maxX = bounds.MaxX
		}

		if bounds.MaxY > maxY {
			maxY = bounds.MaxY
		}
	}

	usedWidth := maxX - surfaceBounds.MinX
	usedHeight := maxY - surfaceBounds.MinY
	usedArea := usedWidth * usedHeight

	return CandidateScore{
		UsedArea:   usedArea,
		UsedHeight: usedHeight,
		UsedWidth:  usedWidth,
	}, nil
}
