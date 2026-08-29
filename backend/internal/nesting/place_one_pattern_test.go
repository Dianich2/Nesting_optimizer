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

func TestPlaceOnePattern(t *testing.T) {
	tests := []struct {
		name      string
		pattern   domaingeometry.Polygon
		surface   domaingeometry.Polygon
		rotations []float64
		candidate Candidate
		occupied  []domaingeometry.Polygon
		bounds    domaingeometry.Bounds
		found     bool
		wantErr   error
	}{
		{
			name: "empty surface / one pattern",
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
			rotations: []float64{0},
			candidate: Candidate{
				X:        10,
				Y:        10,
				Rotation: 0,
			},
			bounds: domaingeometry.Bounds{
				MinX: 0,
				MinY: 0,
				MaxX: 20,
				MaxY: 20,
			},
			found: true,
		},
		{
			name: "existing occupied",
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
			rotations: []float64{0},
			occupied: []domaingeometry.Polygon{
				{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 0, Y: 0},
							{X: 20, Y: 0},
							{X: 20, Y: 20},
							{X: 0, Y: 20},
						},
					},
				},
			},
			candidate: Candidate{
				X:        30,
				Y:        10,
				Rotation: 0,
			},
			bounds: domaingeometry.Bounds{
				MinX: 20,
				MinY: 0,
				MaxX: 40,
				MaxY: 20,
			},
			found: true,
		},
		{
			name: "rotation needed",
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
			rotations: []float64{0, 90},
			candidate: Candidate{
				X:        10,
				Y:        20,
				Rotation: 90,
			},
			bounds: domaingeometry.Bounds{
				MinX: 0,
				MinY: 0,
				MaxX: 20,
				MaxY: 40,
			},
			found: true,
		},
		{
			name: "no more space",
			surface: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 20, Y: 0},
						{X: 20, Y: 20},
						{X: 0, Y: 20},
					},
				},
			},
			pattern: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 10, Y: 0},
						{X: 10, Y: 10},
						{X: 0, Y: 10},
					},
				},
			},
			occupied: []domaingeometry.Polygon{
				{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 0, Y: 0},
							{X: 20, Y: 0},
							{X: 20, Y: 20},
							{X: 0, Y: 20},
						},
					},
				},
			},
			rotations: []float64{0},
			candidate: Candidate{},
			bounds:    domaingeometry.Bounds{},
			found:     false,
		},
		{
			name: "empty rotations",
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
			rotations: []float64{},
			candidate: Candidate{},
			bounds:    domaingeometry.Bounds{},
			found:     false,
		},
		{
			name: "invalid pattern",
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
			pattern: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 0, Y: 10},
						{X: 0, Y: 20},
						{X: 0, Y: 30},
					},
				},
			},
			rotations: []float64{0},
			candidate: Candidate{},
			bounds:    domaingeometry.Bounds{},
			wantErr:   geometry.ErrInvalidPolygon,
		},
		{
			name: "surface don't start with (0,0)",
			surface: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 100, Y: 200},
						{X: 200, Y: 200},
						{X: 200, Y: 300},
						{X: 100, Y: 300},
					},
				},
			},
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
			rotations: []float64{0},
			candidate: Candidate{
				X:        110,
				Y:        210,
				Rotation: 0,
			},
			bounds: domaingeometry.Bounds{
				MinX: 100,
				MinY: 200,
				MaxX: 120,
				MaxY: 220,
			},
			found: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sfEngine := simplefeatures.NewEngine()

			placement, isFound, err := placeOnePattern(
				context.Background(),
				sfEngine,
				tt.pattern,
				tt.surface,
				tt.occupied,
				tt.rotations,
			)

			if tt.wantErr == nil {
				require.NoError(t, err)

				assert.Equal(
					t,
					tt.found,
					isFound,
				)

				if !tt.found {
					assert.Zero(t, placement)

					return
				}

				assert.InDelta(
					t,
					tt.candidate.X,
					placement.Candidate.X,
					config.Epsilon,
				)

				assert.InDelta(
					t,
					tt.candidate.Y,
					placement.Candidate.Y,
					config.Epsilon,
				)

				assert.InDelta(
					t,
					tt.candidate.Rotation,
					placement.Candidate.Rotation,
					config.Epsilon,
				)

				bounds, err := sfEngine.Bounds(
					placement.Geometry,
				)
				require.NoError(t, err)

				assert.InDelta(
					t,
					tt.bounds.MinX,
					bounds.MinX,
					config.Epsilon,
				)

				assert.InDelta(
					t,
					tt.bounds.MinY,
					bounds.MinY,
					config.Epsilon,
				)

				assert.InDelta(
					t,
					tt.bounds.MaxX,
					bounds.MaxX,
					config.Epsilon,
				)

				assert.InDelta(
					t,
					tt.bounds.MaxY,
					bounds.MaxY,
					config.Epsilon,
				)

				return
			}

			assert.Zero(t, isFound)
			assert.Zero(t, placement)

			assert.ErrorIs(
				t,
				err,
				tt.wantErr,
			)

		})
	}
}

func TestPlaceOnePatternCanceledContext(t *testing.T) {
	sfEngine := simplefeatures.NewEngine()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, found, err := placeOnePattern(
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
		domaingeometry.Polygon{
			Exterior: domaingeometry.Ring{
				Points: []domaingeometry.Point{
					{X: 0, Y: 0},
					{X: 100, Y: 0},
					{X: 100, Y: 100},
					{X: 0, Y: 100},
				},
			},
		},
		nil,
		[]float64{0, 90},
	)

	assert.Zero(t, result)
	assert.False(t, found)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestPlaceOnePatternDoesNotMutateOccupied(t *testing.T) {
	sfEngine := simplefeatures.NewEngine()

	occupied := []domaingeometry.Polygon{
		{
			Exterior: domaingeometry.Ring{
				Points: []domaingeometry.Point{
					{X: 0, Y: 0},
					{X: 10, Y: 0},
					{X: 10, Y: 10},
					{X: 0, Y: 10},
				},
			},
		},
	}

	wantOccupied := []domaingeometry.Polygon{
		{
			Exterior: domaingeometry.Ring{
				Points: []domaingeometry.Point{
					{X: 0, Y: 0},
					{X: 10, Y: 0},
					{X: 10, Y: 10},
					{X: 0, Y: 10},
				},
			},
		},
	}

	result, found, err := placeOnePattern(
		context.Background(),
		sfEngine,
		domaingeometry.Polygon{
			Exterior: domaingeometry.Ring{
				Points: []domaingeometry.Point{
					{X: 20, Y: 20},
					{X: 40, Y: 20},
					{X: 40, Y: 40},
					{X: 20, Y: 40},
				},
			},
		},
		domaingeometry.Polygon{
			Exterior: domaingeometry.Ring{
				Points: []domaingeometry.Point{
					{X: 0, Y: 0},
					{X: 100, Y: 0},
					{X: 100, Y: 100},
					{X: 0, Y: 100},
				},
			},
		},
		occupied,
		[]float64{0, 90},
	)

	require.NoError(t, err)

	assert.NotEmpty(
		t,
		result,
	)

	assert.Equal(
		t,
		found,
		true,
	)

	assert.Equal(
		t,
		wantOccupied,
		occupied,
	)
}
