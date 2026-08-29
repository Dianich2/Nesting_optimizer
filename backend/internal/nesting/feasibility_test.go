package nesting

import (
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsPlacementFeasible(t *testing.T) {
	tests := []struct {
		name         string
		pattern      domaingeometry.Polygon
		candidate    Candidate
		surface      domaingeometry.Polygon
		occupied     []domaingeometry.Polygon
		wantFeasible bool
		wantErr      error
	}{
		{
			name: "basic candidate",
			pattern: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 20, Y: 0},
						{X: 20, Y: 20},
						{X: 0, Y: 20},
					},
				},
			},
			surface: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 100, Y: 0},
						{X: 100, Y: 100},
						{X: 0, Y: 100},
					},
				},
			},
			candidate: Candidate{
				X:        50,
				Y:        50,
				Rotation: 0,
			},
			occupied:     nil,
			wantFeasible: true,
			wantErr:      nil,
		},
		{
			name: "touch boundary",
			pattern: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 20, Y: 0},
						{X: 20, Y: 20},
						{X: 0, Y: 20},
					},
				},
			},
			surface: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 100, Y: 0},
						{X: 100, Y: 100},
						{X: 0, Y: 100},
					},
				},
			},
			candidate: Candidate{
				X:        10,
				Y:        50,
				Rotation: 0,
			},
			occupied:     nil,
			wantFeasible: true,
			wantErr:      nil,
		},
		{
			name: "partially outside surface",
			pattern: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 20, Y: 0},
						{X: 20, Y: 20},
						{X: 0, Y: 20},
					},
				},
			},
			surface: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 100, Y: 0},
						{X: 100, Y: 100},
						{X: 0, Y: 100},
					},
				},
			},
			candidate: Candidate{
				X:        5,
				Y:        50,
				Rotation: 0,
			},
			occupied:     nil,
			wantFeasible: false,
			wantErr:      nil,
		},
		{
			name: "completely on the outside",
			pattern: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 20, Y: 0},
						{X: 20, Y: 20},
						{X: 0, Y: 20},
					},
				},
			},
			surface: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 100, Y: 0},
						{X: 100, Y: 100},
						{X: 0, Y: 100},
					},
				},
			},
			candidate: Candidate{
				X:        150,
				Y:        50,
				Rotation: 0,
			},
			occupied:     nil,
			wantFeasible: false,
			wantErr:      nil,
		},
		{
			name: "intersect occupied",
			pattern: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 20, Y: 0},
						{X: 20, Y: 20},
						{X: 0, Y: 20},
					},
				},
			},
			surface: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 100, Y: 0},
						{X: 100, Y: 100},
						{X: 0, Y: 100},
					},
				},
			},
			candidate: Candidate{
				X:        50,
				Y:        50,
				Rotation: 0,
			},
			occupied: []domaingeometry.Polygon{
				{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 50, Y: 40},
							{X: 70, Y: 40},
							{X: 70, Y: 60},
							{X: 50, Y: 60},
						},
					},
				},
			},
			wantFeasible: false,
			wantErr:      nil,
		},
		{
			name: "touch occupied",
			pattern: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 20, Y: 0},
						{X: 20, Y: 20},
						{X: 0, Y: 20},
					},
				},
			},
			surface: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 100, Y: 0},
						{X: 100, Y: 100},
						{X: 0, Y: 100},
					},
				},
			},
			candidate: Candidate{
				X:        50,
				Y:        50,
				Rotation: 0,
			},
			occupied: []domaingeometry.Polygon{
				{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 60, Y: 40},
							{X: 80, Y: 40},
							{X: 80, Y: 60},
							{X: 60, Y: 60},
						},
					},
				},
			},
			wantFeasible: true,
			wantErr:      nil,
		},
		{
			name: "multiple occupied non-intersect",
			pattern: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 20, Y: 0},
						{X: 20, Y: 20},
						{X: 0, Y: 20},
					},
				},
			},
			surface: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 100, Y: 0},
						{X: 100, Y: 100},
						{X: 0, Y: 100},
					},
				},
			},
			candidate: Candidate{
				X:        50,
				Y:        50,
				Rotation: 0,
			},
			occupied: []domaingeometry.Polygon{
				{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 60, Y: 40},
							{X: 80, Y: 40},
							{X: 80, Y: 60},
							{X: 60, Y: 60},
						},
					},
				},
				{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 10, Y: 80},
							{X: 15, Y: 80},
							{X: 15, Y: 90},
							{X: 10, Y: 90},
						},
					},
				},
				{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 5, Y: 5},
							{X: 8, Y: 5},
							{X: 8, Y: 8},
							{X: 5, Y: 8},
						},
					},
				},
			},
			wantFeasible: true,
			wantErr:      nil,
		},
		{
			name: "multiple occupied intersect",
			pattern: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 20, Y: 0},
						{X: 20, Y: 20},
						{X: 0, Y: 20},
					},
				},
			},
			surface: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 100, Y: 0},
						{X: 100, Y: 100},
						{X: 0, Y: 100},
					},
				},
			},
			candidate: Candidate{
				X:        50,
				Y:        50,
				Rotation: 0,
			},
			occupied: []domaingeometry.Polygon{
				{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 60, Y: 40},
							{X: 80, Y: 40},
							{X: 80, Y: 60},
							{X: 60, Y: 60},
						},
					},
				},
				{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 10, Y: 80},
							{X: 15, Y: 80},
							{X: 15, Y: 90},
							{X: 10, Y: 90},
						},
					},
				},
				{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 45, Y: 45},
							{X: 48, Y: 45},
							{X: 48, Y: 48},
							{X: 45, Y: 48},
						},
					},
				},
			},
			wantFeasible: false,
			wantErr:      nil,
		},
		{
			name: "rotation 90",
			pattern: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 40, Y: 0},
						{X: 40, Y: 20},
						{X: 0, Y: 20},
					},
				},
			},
			surface: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 30, Y: 0},
						{X: 30, Y: 50},
						{X: 0, Y: 50},
					},
				},
			},
			candidate: Candidate{
				X:        15,
				Y:        25,
				Rotation: 90,
			},
			occupied:     nil,
			wantFeasible: true,
			wantErr:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sfEngine := simplefeatures.NewEngine()
			transformedPattern, err := sfEngine.Transform(
				tt.pattern,
				tt.candidate.X,
				tt.candidate.Y,
				tt.candidate.Rotation,
			)
			require.NoError(t, err)

			isFeasible, err := isPlacementFeasible(
				sfEngine,
				transformedPattern,
				tt.surface,
				tt.occupied,
			)

			if tt.wantErr == nil {
				require.NoError(t, err)

				assert.Equal(
					t,
					tt.wantFeasible,
					isFeasible,
				)

				return
			}

			assert.Zero(t, isFeasible)

			assert.ErrorIs(
				t,
				err,
				tt.wantErr,
			)
		})
	}
}
