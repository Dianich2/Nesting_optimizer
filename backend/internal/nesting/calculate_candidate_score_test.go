package nesting

import (
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateCandidateScore(t *testing.T) {
	tests := []struct {
		name               string
		candidate          domaingeometry.Polygon
		surface            domaingeometry.Polygon
		occupied           []domaingeometry.Polygon
		wantCandidateScore CandidateScore
		wantErr            error
	}{
		{
			name: "candidate only",
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
			candidate: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 10, Y: 20},
						{X: 30, Y: 20},
						{X: 30, Y: 40},
						{X: 10, Y: 40},
					},
				},
			},
			occupied: nil,
			wantCandidateScore: CandidateScore{
				UsedArea:   1200,
				UsedHeight: 40,
				UsedWidth:  30,
			},
			wantErr: nil,
		},
		{
			name: "occupied expands the cut area only to the right",
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
			candidate: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 10, Y: 10},
						{X: 30, Y: 10},
						{X: 30, Y: 30},
						{X: 10, Y: 30},
					},
				},
			},
			occupied: []domaingeometry.Polygon{
				{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 40, Y: 5},
							{X: 70, Y: 5},
							{X: 70, Y: 20},
							{X: 40, Y: 20},
						},
					},
				},
			},
			wantCandidateScore: CandidateScore{
				UsedArea:   2100,
				UsedHeight: 30,
				UsedWidth:  70,
			},
			wantErr: nil,
		},
		{
			name: "occupied expands the cut area only to the up",
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
			candidate: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 10, Y: 10},
						{X: 30, Y: 10},
						{X: 30, Y: 30},
						{X: 10, Y: 30},
					},
				},
			},
			occupied: []domaingeometry.Polygon{
				{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 5, Y: 40},
							{X: 20, Y: 40},
							{X: 20, Y: 80},
							{X: 5, Y: 80},
						},
					},
				},
			},
			wantCandidateScore: CandidateScore{
				UsedArea:   2400,
				UsedHeight: 80,
				UsedWidth:  30,
			},
			wantErr: nil,
		},
		{
			name: "multiple occupied",
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
			candidate: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 10, Y: 20},
						{X: 30, Y: 20},
						{X: 30, Y: 30},
						{X: 10, Y: 30},
					},
				},
			},
			occupied: []domaingeometry.Polygon{
				{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 5, Y: 10},
							{X: 80, Y: 10},
							{X: 80, Y: 20},
							{X: 5, Y: 20},
						},
					},
				},
				{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 35, Y: 60},
							{X: 40, Y: 60},
							{X: 40, Y: 90},
							{X: 35, Y: 90},
						},
					},
				},
			},
			wantCandidateScore: CandidateScore{
				UsedArea:   7200,
				UsedHeight: 90,
				UsedWidth:  80,
			},
			wantErr: nil,
		},
		{
			name: "surface not in (0,0)",
			surface: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 100, Y: 200},
						{X: 300, Y: 200},
						{X: 300, Y: 400},
						{X: 100, Y: 400},
					},
				},
			},
			candidate: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 120, Y: 230},
						{X: 150, Y: 230},
						{X: 150, Y: 270},
						{X: 120, Y: 270},
					},
				},
			},
			occupied: nil,
			wantCandidateScore: CandidateScore{
				UsedArea:   3500,
				UsedHeight: 70,
				UsedWidth:  50,
			},
			wantErr: nil,
		},
		{
			name: "invalid polygon",
			surface: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 100, Y: 200},
						{X: 300, Y: 200},
						{X: 300, Y: 400},
						{X: 100, Y: 400},
					},
				},
			},
			candidate: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 120, Y: 230},
						{X: 120, Y: 240},
						{X: 120, Y: 250},
						{X: 120, Y: 260},
					},
				},
			},
			occupied: nil,
			wantErr:  geometry.ErrInvalidPolygon,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sfEngine := simplefeatures.NewEngine()

			candidateScore, err := calculateCandidateScore(
				sfEngine,
				tt.surface,
				tt.occupied,
				tt.candidate,
			)

			if tt.wantErr == nil {
				require.NoError(t, err)

				assert.Equal(
					t,
					tt.wantCandidateScore,
					candidateScore,
				)

				return
			}

			assert.Zero(t, candidateScore)

			assert.ErrorIs(
				t,
				err,
				tt.wantErr,
			)
		})
	}
}
