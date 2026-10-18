package nesting

import (
	"testing"

	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry/simplefeatures"

	"github.com/stretchr/testify/require"
)

func TestPlacementFeasibilityWithBoundsMatchesExact(t *testing.T) {
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

	surface := rect(0, 0, 100, 100)
	occupied := []domaingeometry.Polygon{
		rect(20, 20, 40, 40),
	}

	tests := []struct {
		name      string
		candidate domaingeometry.Polygon
		want      bool
	}{
		{
			name:      "separated",
			candidate: rect(60, 60, 70, 70),
			want:      true,
		},
		{
			name:      "touching edge",
			candidate: rect(40, 20, 50, 40),
			want:      true,
		},
		{
			name:      "touching corner",
			candidate: rect(40, 40, 50, 50),
			want:      true,
		},
		{
			name:      "small positive overlap",
			candidate: rect(39.999999, 20, 50, 40),
			want:      false,
		},
		{
			name:      "partial overlap",
			candidate: rect(35, 35, 55, 55),
			want:      false,
		},
		{
			name:      "fully contained",
			candidate: rect(25, 25, 35, 35),
			want:      false,
		},
		{
			name:      "outside surface",
			candidate: rect(95, 95, 105, 105),
			want:      false,
		},
	}

	occupiedBounds := make(
		[]domaingeometry.Bounds,
		0,
		len(occupied),
	)

	for _, polygon := range occupied {
		bounds, err := engine.Bounds(polygon)
		require.NoError(t, err)

		occupiedBounds = append(
			occupiedBounds,
			bounds,
		)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidateBounds, err := engine.Bounds(
				tt.candidate,
			)
			require.NoError(t, err)

			exact, err := isPlacementFeasible(
				engine,
				tt.candidate,
				surface,
				occupied,
			)
			require.NoError(t, err)

			optimized, err := isPlacementFeasibleWithBounds(
				engine,
				tt.candidate,
				candidateBounds,
				surface,
				occupied,
				occupiedBounds,
			)
			require.NoError(t, err)

			require.Equal(t, tt.want, exact)
			require.Equal(t, exact, optimized)
		})
	}
}

func TestPlacementFeasibilityWithBoundsRejectsMismatchedCounts(
	t *testing.T,
) {
	engine := simplefeatures.NewEngine()

	surface := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 0, Y: 0},
				{X: 100, Y: 0},
				{X: 100, Y: 100},
				{X: 0, Y: 100},
			},
		},
	}

	bounds, err := engine.Bounds(surface)
	require.NoError(t, err)

	_, err = isPlacementFeasibleWithBounds(
		engine,
		surface,
		bounds,
		surface,
		[]domaingeometry.Polygon{surface},
		nil,
	)

	require.ErrorContains(
		t,
		err,
		"occupied polygons and bounds count mismatch",
	)
}
