package nesting

import (
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateCandidates(t *testing.T) {
	tests := []struct {
		name           string
		pattern        domaingeometry.Polygon
		anchors        []domaingeometry.Point
		rotations      []float64
		wantCandidates []Candidate
		wantErr        error
	}{
		{
			name: "multiple rotations square",
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
			anchors: []domaingeometry.Point{
				{
					X: 100,
					Y: 100,
				},
			},
			rotations: []float64{0, 90},
			wantCandidates: []Candidate{
				{
					X:        110,
					Y:        110,
					Rotation: 0,
				},
				{
					X:        90,
					Y:        110,
					Rotation: 0,
				},
				{
					X:        90,
					Y:        90,
					Rotation: 0,
				},
				{
					X:        110,
					Y:        90,
					Rotation: 0,
				},
				{
					X:        90,
					Y:        110,
					Rotation: 90,
				},
				{
					X:        90,
					Y:        90,
					Rotation: 90,
				},
				{
					X:        110,
					Y:        90,
					Rotation: 90,
				},
				{
					X:        110,
					Y:        110,
					Rotation: 90,
				},
			},
		},
		{
			name: "multiple rotations rectangle",
			pattern: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 20, Y: 0},
						{X: 20, Y: 10},
						{X: 0, Y: 10},
					},
				},
			},
			anchors: []domaingeometry.Point{
				{
					X: 100,
					Y: 100,
				},
			},
			rotations: []float64{0, 90},
			wantCandidates: []Candidate{
				{
					X:        110,
					Y:        105,
					Rotation: 0,
				},
				{
					X:        90,
					Y:        105,
					Rotation: 0,
				},
				{
					X:        90,
					Y:        95,
					Rotation: 0,
				},
				{
					X:        110,
					Y:        95,
					Rotation: 0,
				},
				{
					X:        95,
					Y:        110,
					Rotation: 90,
				},
				{
					X:        95,
					Y:        90,
					Rotation: 90,
				},
				{
					X:        105,
					Y:        90,
					Rotation: 90,
				},
				{
					X:        105,
					Y:        110,
					Rotation: 90,
				},
			},
		},
		{
			name: "empty rotations",
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
			anchors: []domaingeometry.Point{
				{
					X: 100,
					Y: 100,
				},
			},
			rotations:      nil,
			wantCandidates: nil,
		},
		{
			name: "invalid polygon",
			pattern: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 0, Y: 20},
						{X: 0, Y: 40},
						{X: 0, Y: 60},
					},
				},
			},
			anchors: []domaingeometry.Point{
				{
					X: 100,
					Y: 100,
				},
			},
			rotations: []float64{90},
			wantErr:   geometry.ErrInvalidPolygon,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sfEngine := simplefeatures.NewEngine()

			candidates, err := generateCandidates(
				sfEngine,
				tt.pattern,
				tt.anchors,
				tt.rotations,
			)

			if tt.wantErr == nil {
				require.NoError(t, err)

				assert.Equal(
					t,
					tt.wantCandidates,
					candidates,
				)

				return
			}

			assert.Zero(t, candidates)

			assert.ErrorIs(
				t,
				err,
				tt.wantErr,
			)
		})
	}
}
