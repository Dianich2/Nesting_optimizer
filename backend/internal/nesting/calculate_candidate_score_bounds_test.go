package nesting

import (
	"testing"

	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry/simplefeatures"

	"github.com/stretchr/testify/require"
)

func TestCalculateCandidateScoreWithBoundsMatchesOriginal(t *testing.T) {
	engine := simplefeatures.NewEngine()

	rect := func(
		minX, minY, maxX, maxY float64,
	) domaingeometry.Polygon {
		return domaingeometry.Polygon{
			Exterior: domaingeometry.Ring{
				Points: []domaingeometry.Point{
					{X: minX, Y: minY},
					{X: maxX, Y: minY},
					{X: maxX, Y: maxY},
					{X: minX, Y: maxY},
				},
			},
		}
	}

	surface := rect(100, 200, 300, 400)

	occupied := []domaingeometry.Polygon{
		rect(110, 210, 140, 240),
		rect(160, 220, 190, 270),
	}

	candidate := rect(145, 280, 175, 310)

	original, err := calculateCandidateScore(
		engine,
		surface,
		occupied,
		candidate,
	)
	require.NoError(t, err)

	surfaceBounds, err := engine.Bounds(surface)
	require.NoError(t, err)

	candidateBounds, err := engine.Bounds(candidate)
	require.NoError(t, err)

	occupiedBounds := make(
		[]domaingeometry.Bounds,
		0,
		len(occupied),
	)

	for _, polygon := range occupied {
		bounds, err := engine.Bounds(polygon)
		require.NoError(t, err)

		occupiedBounds = append(occupiedBounds, bounds)
	}

	optimized := calculateCandidateScoreWithBounds(
		surfaceBounds,
		occupiedBounds,
		candidateBounds,
	)

	require.Equal(t, original, optimized)

	require.Equal(t, CandidateScore{
		UsedWidth:  90,
		UsedHeight: 110,
		UsedArea:   9900,
	}, optimized)
}

func TestCalculateCandidateScoreWithExtentsMatchesBounds(t *testing.T) {
	tests := []struct {
		name      string
		surface   domaingeometry.Bounds
		occupied  []domaingeometry.Bounds
		candidate domaingeometry.Bounds
	}{
		{
			name: "empty occupied",
			surface: domaingeometry.Bounds{
				MinX: 100, MinY: 200,
				MaxX: 300, MaxY: 400,
			},
			candidate: domaingeometry.Bounds{
				MinX: 110, MinY: 210,
				MaxX: 140, MaxY: 240,
			},
		},
		{
			name: "occupied extends right and up",
			surface: domaingeometry.Bounds{
				MinX: 100, MinY: 200,
				MaxX: 300, MaxY: 400,
			},
			occupied: []domaingeometry.Bounds{
				{MinX: 110, MinY: 210, MaxX: 190, MaxY: 250},
				{MinX: 115, MinY: 220, MaxX: 145, MaxY: 310},
			},
			candidate: domaingeometry.Bounds{
				MinX: 120, MinY: 230,
				MaxX: 160, MaxY: 280,
			},
		},
		{
			name: "negative coordinates",
			surface: domaingeometry.Bounds{
				MinX: -200, MinY: -200,
				MaxX: -10, MaxY: -10,
			},
			occupied: []domaingeometry.Bounds{
				{MinX: -170, MinY: -170, MaxX: -120, MaxY: -100},
				{MinX: -130, MinY: -150, MaxX: -70, MaxY: -80},
			},
			candidate: domaingeometry.Bounds{
				MinX: -110, MinY: -110,
				MaxX: -50, MaxY: -50,
			},
		},
		{
			name: "candidate determines both maxima",
			surface: domaingeometry.Bounds{
				MinX: 0, MinY: 0,
				MaxX: 100, MaxY: 100,
			},
			occupied: []domaingeometry.Bounds{
				{MinX: 5, MinY: 5, MaxX: 20, MaxY: 20},
			},
			candidate: domaingeometry.Bounds{
				MinX: 30, MinY: 30,
				MaxX: 60, MaxY: 70,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := calculateCandidateScoreWithBounds(
				tt.surface,
				tt.occupied,
				tt.candidate,
			)

			extents := calculateOccupiedScoreExtents(
				tt.occupied,
			)

			optimized := calculateCandidateScoreWithExtents(
				tt.surface,
				extents,
				tt.candidate,
			)

			require.Equal(t, original, optimized)
		})
	}
}
