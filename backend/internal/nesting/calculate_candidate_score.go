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

	occupiedBounds := make(
		[]domaingeometry.Bounds,
		0,
		len(occupied),
	)

	for _, polygon := range occupied {
		bounds, err := engine.Bounds(polygon)
		if err != nil {
			return CandidateScore{}, fmt.Errorf(
				"calculate candidate score: %w",
				err,
			)
		}

		occupiedBounds = append(occupiedBounds, bounds)
	}

	return calculateCandidateScoreWithBounds(
		surfaceBounds,
		occupiedBounds,
		candidateBounds,
	), nil
}

func calculateCandidateScoreWithBounds(
	surfaceBounds domaingeometry.Bounds,
	occupiedBounds []domaingeometry.Bounds,
	candidateBounds domaingeometry.Bounds,
) CandidateScore {
	extents := calculateOccupiedScoreExtents(occupiedBounds)

	return calculateCandidateScoreWithExtents(
		surfaceBounds,
		extents,
		candidateBounds,
	)
}

type occupiedScoreExtents struct {
	MaxX        float64
	MaxY        float64
	HasOccupied bool
}

func calculateOccupiedScoreExtents(
	occupiedBounds []domaingeometry.Bounds,
) occupiedScoreExtents {
	if len(occupiedBounds) == 0 {
		return occupiedScoreExtents{}
	}

	extents := occupiedScoreExtents{
		MaxX:        occupiedBounds[0].MaxX,
		MaxY:        occupiedBounds[0].MaxY,
		HasOccupied: true,
	}

	for _, bounds := range occupiedBounds[1:] {
		if bounds.MaxX > extents.MaxX {
			extents.MaxX = bounds.MaxX
		}

		if bounds.MaxY > extents.MaxY {
			extents.MaxY = bounds.MaxY
		}
	}

	return extents
}

func calculateCandidateScoreWithExtents(
	surfaceBounds domaingeometry.Bounds,
	extents occupiedScoreExtents,
	candidateBounds domaingeometry.Bounds,
) CandidateScore {
	maxX := candidateBounds.MaxX
	maxY := candidateBounds.MaxY

	if extents.HasOccupied {
		if extents.MaxX > maxX {
			maxX = extents.MaxX
		}

		if extents.MaxY > maxY {
			maxY = extents.MaxY
		}
	}

	usedWidth := maxX - surfaceBounds.MinX
	usedHeight := maxY - surfaceBounds.MinY

	return CandidateScore{
		UsedArea:   usedWidth * usedHeight,
		UsedHeight: usedHeight,
		UsedWidth:  usedWidth,
	}
}
