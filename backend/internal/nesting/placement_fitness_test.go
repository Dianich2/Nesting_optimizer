package nesting

import (
	"testing"

	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry/simplefeatures"

	"github.com/stretchr/testify/require"
)

func TestCalculatePlacedGeometriesFitnessMatchesPlacementsFitness(t *testing.T) {
	engine := simplefeatures.NewEngine()

	triangle := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 0, Y: 0},
				{X: 10, Y: 0},
				{X: 4, Y: 3},
			},
		},
	}

	problem := Problem{
		Surface: rectangle(100, 100),
		Patterns: []PatternItem{
			{
				PatternID: 1,
				Quantity:  3,
				Geometry:  triangle,
			},
		},
	}

	tests := []struct {
		name       string
		placements []Placement
	}{
		{
			name:       "no placements",
			placements: nil,
		},
		{
			name: "single placement",
			placements: []Placement{
				{
					PatternID: 1,
					X:         25,
					Y:         25,
					Rotation:  45,
				},
			},
		},
		{
			name: "multiple placements with rotations",
			placements: []Placement{
				{
					PatternID: 1,
					X:         25,
					Y:         25,
					Rotation:  45,
				},
				{
					PatternID: 1,
					X:         50,
					Y:         50,
					Rotation:  90,
				},
				{
					PatternID: 1,
					X:         75,
					Y:         75,
					Rotation:  180,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			placedGeometries := make(
				[]domaingeometry.Polygon,
				0,
				len(tt.placements),
			)

			for _, placement := range tt.placements {
				transformed, err := engine.Transform(
					triangle,
					placement.X,
					placement.Y,
					placement.Rotation,
				)
				require.NoError(t, err)

				placedGeometries = append(
					placedGeometries,
					transformed,
				)
			}

			originalFitness, err := calculatePlacementsFitness(
				engine,
				problem,
				tt.placements,
			)
			require.NoError(t, err)

			optimizedFitness, err := calculatePlacedGeometriesFitness(
				engine,
				problem,
				tt.placements,
				placedGeometries,
			)
			require.NoError(t, err)

			require.Equal(
				t,
				originalFitness,
				optimizedFitness,
			)
		})
	}
}
