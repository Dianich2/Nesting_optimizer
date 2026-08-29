package nesting

import (
	"context"
	"server_nesting_optimizer/internal/config"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChooseBestCandidate(t *testing.T) {
	tests := []struct {
		name          string
		pattern       domaingeometry.Polygon
		candidates    []Candidate
		surface       domaingeometry.Polygon
		occupied      []domaingeometry.Polygon
		wantCandidate Candidate
		wantBounds    domaingeometry.Bounds
		found         bool
		wantErr       error
	}{
		{
			name: "one feasible candidate",
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
			candidates: []Candidate{
				{
					X:        10,
					Y:        10,
					Rotation: 0,
				},
			},
			occupied: nil,
			wantCandidate: Candidate{
				X:        10,
				Y:        10,
				Rotation: 0,
			},
			wantBounds: domaingeometry.Bounds{
				MinX: 0,
				MinY: 0,
				MaxX: 20,
				MaxY: 20,
			},
			found:   true,
			wantErr: nil,
		},
		{
			name: "first infeasible and second feasible candidate",
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
			candidates: []Candidate{
				{
					X:        -20,
					Y:        -20,
					Rotation: 0,
				},
				{
					X:        10,
					Y:        10,
					Rotation: 0,
				},
			},
			occupied: nil,
			wantCandidate: Candidate{
				X:        10,
				Y:        10,
				Rotation: 0,
			},
			wantBounds: domaingeometry.Bounds{
				MinX: 0,
				MinY: 0,
				MaxX: 20,
				MaxY: 20,
			},
			found:   true,
			wantErr: nil,
		},
		{
			name: "multiple feasible candidates (first has bigger usedArea)",
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
			candidates: []Candidate{
				{
					X:        50,
					Y:        50,
					Rotation: 0,
				},
				{
					X:        10,
					Y:        10,
					Rotation: 0,
				},
			},
			occupied: nil,
			wantCandidate: Candidate{
				X:        10,
				Y:        10,
				Rotation: 0,
			},
			wantBounds: domaingeometry.Bounds{
				MinX: 0,
				MinY: 0,
				MaxX: 20,
				MaxY: 20,
			},
			found:   true,
			wantErr: nil,
		},
		{
			name: "multiple feasible candidates (first has smaller usedHeight)",
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
			candidates: []Candidate{
				{
					X:        30,
					Y:        20,
					Rotation: 0,
				},
				{
					X:        20,
					Y:        30,
					Rotation: 0,
				},
			},
			occupied: nil,
			wantCandidate: Candidate{
				X:        30,
				Y:        20,
				Rotation: 0,
			},
			wantBounds: domaingeometry.Bounds{
				MinX: 20,
				MinY: 10,
				MaxX: 40,
				MaxY: 30,
			},
			found:   true,
			wantErr: nil,
		},
		{
			name: "equals score",
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
			candidates: []Candidate{
				{
					X:        30,
					Y:        20,
					Rotation: 0,
				},
				{
					X:        30,
					Y:        20,
					Rotation: 360,
				},
			},
			occupied: nil,
			wantCandidate: Candidate{
				X:        30,
				Y:        20,
				Rotation: 0,
			},
			wantBounds: domaingeometry.Bounds{
				MinX: 20,
				MinY: 10,
				MaxX: 40,
				MaxY: 30,
			},
			found:   true,
			wantErr: nil,
		},
		{
			name: "no feasible candidates",
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
			candidates: []Candidate{
				{
					X:        -50,
					Y:        -50,
					Rotation: 0,
				},
				{
					X:        150,
					Y:        150,
					Rotation: 0,
				},
			},
			occupied:      nil,
			wantCandidate: Candidate{},
			found:         false,
			wantErr:       nil,
		},
		{
			name: "empty candidates",
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
			candidates:    []Candidate{},
			occupied:      nil,
			wantCandidate: Candidate{},
			found:         false,
			wantErr:       nil,
		},
		{
			name: "candidates with rotation",
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
			candidates: []Candidate{
				{
					X:        15,
					Y:        25,
					Rotation: 0,
				},
				{
					X:        15,
					Y:        25,
					Rotation: 90,
				},
			},
			occupied: nil,
			wantCandidate: Candidate{
				X:        15,
				Y:        25,
				Rotation: 90,
			},
			wantBounds: domaingeometry.Bounds{
				MinX: 5,
				MinY: 5,
				MaxX: 25,
				MaxY: 45,
			},
			found:   true,
			wantErr: nil,
		},
		{
			name: "invalid pattern",
			pattern: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 10},
						{X: 0, Y: 20},
						{X: 0, Y: 30},
						{X: 0, Y: 40},
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
			candidates: []Candidate{
				{
					X:        15,
					Y:        25,
					Rotation: 0,
				},
				{
					X:        15,
					Y:        25,
					Rotation: 90,
				},
			},
			occupied:      nil,
			wantCandidate: Candidate{},
			found:         false,
			wantErr:       geometry.ErrInvalidPolygon,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sfEngine := simplefeatures.NewEngine()

			bestCandidate, isFound, err := chooseBestCandidate(
				context.Background(),
				sfEngine,
				tt.pattern,
				tt.candidates,
				tt.surface,
				tt.occupied,
			)

			if tt.wantErr == nil {
				require.NoError(t, err)

				assert.Equal(
					t,
					tt.found,
					isFound,
				)

				if !tt.found {
					assert.Zero(t, bestCandidate)
					return
				}

				assert.Equal(
					t,
					tt.wantCandidate,
					bestCandidate.Candidate,
				)

				bounds, err := sfEngine.Bounds(
					bestCandidate.Geometry,
				)
				require.NoError(t, err)

				assert.InDelta(
					t,
					tt.wantBounds.MinX,
					bounds.MinX,
					config.Epsilon,
				)

				assert.InDelta(
					t,
					tt.wantBounds.MinY,
					bounds.MinY,
					config.Epsilon,
				)

				assert.InDelta(
					t,
					tt.wantBounds.MaxX,
					bounds.MaxX,
					config.Epsilon,
				)

				assert.InDelta(
					t,
					tt.wantBounds.MaxY,
					bounds.MaxY,
					config.Epsilon,
				)

				return
			}

			assert.Zero(t, bestCandidate)
			assert.Zero(t, isFound)

			assert.ErrorIs(
				t,
				err,
				tt.wantErr,
			)
		})
	}
}

func TestChooseBestCandidateCanceledContext(t *testing.T) {
	sfEngine := simplefeatures.NewEngine()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, found, err := chooseBestCandidate(
		ctx,
		sfEngine,
		domaingeometry.Polygon{
			Exterior: domaingeometry.Ring{
				Points: []domaingeometry.Point{
					{X: 0, Y: 0},
					{X: 40, Y: 0},
					{X: 40, Y: 20},
					{X: 0, Y: 20},
				},
			},
		},
		[]Candidate{},
		domaingeometry.Polygon{
			Exterior: domaingeometry.Ring{
				Points: []domaingeometry.Point{
					{X: 0, Y: 0},
					{X: 30, Y: 0},
					{X: 30, Y: 50},
					{X: 0, Y: 50},
				},
			},
		},
		nil,
	)

	assert.Zero(t, result)
	assert.False(t, found)
	assert.ErrorIs(t, err, context.Canceled)
}
