package nesting

import (
	"math"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateProblem(t *testing.T) {
	tests := []struct {
		name    string
		problem Problem
		wantErr error
	}{
		{
			name: "valid problem",
			problem: Problem{
				Surface: domaingeometry.Polygon{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 0, Y: 0},
							{X: 100, Y: 0},
							{X: 100, Y: 100},
							{X: 0, Y: 100},
						},
					},
				},
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Geometry: domaingeometry.Polygon{
							Exterior: domaingeometry.Ring{
								Points: []domaingeometry.Point{
									{X: 0, Y: 0},
									{X: 20, Y: 0},
									{X: 20, Y: 20},
									{X: 0, Y: 20},
								},
							},
						},
						Quantity: 2,
					},
				},
				Obstacles:        nil,
				AllowedRotations: []float64{0, 90, 270},
			},
			wantErr: nil,
		},
		{
			name: "invalid surface",
			problem: Problem{
				Surface: domaingeometry.Polygon{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 0, Y: 0},
							{X: 0, Y: 10},
							{X: 0, Y: 20},
							{X: 0, Y: 30},
						},
					},
				},
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Geometry: domaingeometry.Polygon{
							Exterior: domaingeometry.Ring{
								Points: []domaingeometry.Point{
									{X: 0, Y: 0},
									{X: 20, Y: 0},
									{X: 20, Y: 20},
									{X: 0, Y: 20},
								},
							},
						},
						Quantity: 2,
					},
				},
				Obstacles:        nil,
				AllowedRotations: []float64{0, 90, 270},
			},
			wantErr: geometry.ErrInvalidPolygon,
		},
		{
			name: "invalid pattern",
			problem: Problem{
				Surface: domaingeometry.Polygon{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 0, Y: 0},
							{X: 100, Y: 0},
							{X: 100, Y: 100},
							{X: 0, Y: 100},
						},
					},
				},
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Geometry: domaingeometry.Polygon{
							Exterior: domaingeometry.Ring{
								Points: []domaingeometry.Point{
									{X: 0, Y: 0},
									{X: 0, Y: 10},
									{X: 0, Y: 20},
									{X: 0, Y: 30},
								},
							},
						},
						Quantity: 2,
					},
				},
				Obstacles:        nil,
				AllowedRotations: []float64{0, 90, 270},
			},
			wantErr: geometry.ErrInvalidPolygon,
		},
		{
			name: "zero quantity",
			problem: Problem{
				Surface: domaingeometry.Polygon{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 0, Y: 0},
							{X: 100, Y: 0},
							{X: 100, Y: 100},
							{X: 0, Y: 100},
						},
					},
				},
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Geometry: domaingeometry.Polygon{
							Exterior: domaingeometry.Ring{
								Points: []domaingeometry.Point{
									{X: 0, Y: 0},
									{X: 10, Y: 0},
									{X: 10, Y: 10},
									{X: 0, Y: 10},
								},
							},
						},
						Quantity: 0,
					},
				},
				Obstacles:        nil,
				AllowedRotations: []float64{0, 90, 270},
			},
			wantErr: ErrInvalidQuantity,
		},
		{
			name: "negative quantity",
			problem: Problem{
				Surface: domaingeometry.Polygon{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 0, Y: 0},
							{X: 100, Y: 0},
							{X: 100, Y: 100},
							{X: 0, Y: 100},
						},
					},
				},
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Geometry: domaingeometry.Polygon{
							Exterior: domaingeometry.Ring{
								Points: []domaingeometry.Point{
									{X: 0, Y: 0},
									{X: 10, Y: 0},
									{X: 10, Y: 10},
									{X: 0, Y: 10},
								},
							},
						},
						Quantity: -2,
					},
				},
				Obstacles:        nil,
				AllowedRotations: []float64{0, 90, 270},
			},
			wantErr: ErrInvalidQuantity,
		},
		{
			name: "empty rotations",
			problem: Problem{
				Surface: domaingeometry.Polygon{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 0, Y: 0},
							{X: 100, Y: 0},
							{X: 100, Y: 100},
							{X: 0, Y: 100},
						},
					},
				},
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Geometry: domaingeometry.Polygon{
							Exterior: domaingeometry.Ring{
								Points: []domaingeometry.Point{
									{X: 0, Y: 0},
									{X: 10, Y: 0},
									{X: 10, Y: 10},
									{X: 0, Y: 10},
								},
							},
						},
						Quantity: 2,
					},
				},
				Obstacles:        nil,
				AllowedRotations: []float64{},
			},
			wantErr: ErrEmptyRotations,
		},
		{
			name: "NaN rotation",
			problem: Problem{
				Surface: domaingeometry.Polygon{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 0, Y: 0},
							{X: 100, Y: 0},
							{X: 100, Y: 100},
							{X: 0, Y: 100},
						},
					},
				},
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Geometry: domaingeometry.Polygon{
							Exterior: domaingeometry.Ring{
								Points: []domaingeometry.Point{
									{X: 0, Y: 0},
									{X: 10, Y: 0},
									{X: 10, Y: 10},
									{X: 0, Y: 10},
								},
							},
						},
						Quantity: 2,
					},
				},
				Obstacles:        nil,
				AllowedRotations: []float64{math.NaN()},
			},
			wantErr: ErrInvalidRotation,
		},
		{
			name: "+Inf rotation",
			problem: Problem{
				Surface: domaingeometry.Polygon{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 0, Y: 0},
							{X: 100, Y: 0},
							{X: 100, Y: 100},
							{X: 0, Y: 100},
						},
					},
				},
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Geometry: domaingeometry.Polygon{
							Exterior: domaingeometry.Ring{
								Points: []domaingeometry.Point{
									{X: 0, Y: 0},
									{X: 10, Y: 0},
									{X: 10, Y: 10},
									{X: 0, Y: 10},
								},
							},
						},
						Quantity: 2,
					},
				},
				Obstacles:        nil,
				AllowedRotations: []float64{math.Inf(1)},
			},
			wantErr: ErrInvalidRotation,
		},
		{
			name: "-Inf rotation",
			problem: Problem{
				Surface: domaingeometry.Polygon{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 0, Y: 0},
							{X: 100, Y: 0},
							{X: 100, Y: 100},
							{X: 0, Y: 100},
						},
					},
				},
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Geometry: domaingeometry.Polygon{
							Exterior: domaingeometry.Ring{
								Points: []domaingeometry.Point{
									{X: 0, Y: 0},
									{X: 10, Y: 0},
									{X: 10, Y: 10},
									{X: 0, Y: 10},
								},
							},
						},
						Quantity: 2,
					},
				},
				Obstacles:        nil,
				AllowedRotations: []float64{math.Inf(-1)},
			},
			wantErr: ErrInvalidRotation,
		},
		{
			name: "invalid obstacle",
			problem: Problem{
				Surface: domaingeometry.Polygon{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 0, Y: 0},
							{X: 100, Y: 0},
							{X: 100, Y: 100},
							{X: 0, Y: 100},
						},
					},
				},
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Geometry: domaingeometry.Polygon{
							Exterior: domaingeometry.Ring{
								Points: []domaingeometry.Point{
									{X: 0, Y: 0},
									{X: 10, Y: 0},
									{X: 10, Y: 10},
									{X: 0, Y: 10},
								},
							},
						},
						Quantity: 2,
					},
				},
				Obstacles: []domaingeometry.Polygon{
					{
						Exterior: domaingeometry.Ring{
							Points: []domaingeometry.Point{
								{X: 0, Y: 30},
								{X: 0, Y: 30},
								{X: 0, Y: 40},
								{X: 0, Y: 40},
							},
						},
					},
				},
				AllowedRotations: []float64{0, 90},
			},
			wantErr: geometry.ErrInvalidPolygon,
		},
		{
			name: "no patterns",
			problem: Problem{
				Surface: domaingeometry.Polygon{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 0, Y: 0},
							{X: 100, Y: 0},
							{X: 100, Y: 100},
							{X: 0, Y: 100},
						},
					},
				},
				Patterns:         []PatternItem{},
				Obstacles:        nil,
				AllowedRotations: []float64{0, 90},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sfEngine := simplefeatures.NewEngine()

			err := validateProblem(
				sfEngine,
				tt.problem,
			)

			if tt.wantErr == nil {
				require.NoError(t, err)
				return
			}

			assert.ErrorIs(
				t,
				err,
				tt.wantErr,
			)

		})
	}
}
