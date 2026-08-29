package nesting

import (
	"context"
	"server_nesting_optimizer/internal/config"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaselineOptimizerOptimize(t *testing.T) {
	sfEngine := simplefeatures.NewEngine()
	optimizer := NewBaselineOptimizer(sfEngine)

	tests := []struct {
		name       string
		problem    Problem
		wantResult Result
		wantErr    error
	}{
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
				AllowedRotations: []float64{0},
			},
			wantResult: Result{
				Placements: []Placement{},
				Unplaced:   []UnplacedPattern{},
				Metrics: Metrics{
					RequestedCount: 0,
					PlacedCount:    0,
					SurfaceArea:    10000,
					PlacedArea:     0,
					Utilization:    0,
				},
			},
		},
		{
			name: "single pattern single quantity",
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
						Quantity:  1,
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
					},
				},
				AllowedRotations: []float64{0},
			},
			wantResult: Result{
				Placements: []Placement{
					{
						PatternID: 1,
						X:         10,
						Y:         10,
						Rotation:  0,
					},
				},
				Unplaced: []UnplacedPattern{},
				Metrics: Metrics{
					RequestedCount: 1,
					PlacedCount:    1,
					SurfaceArea:    10000,
					PlacedArea:     400,
					Utilization:    0.04,
				},
			},
		},
		{
			name: "single pattern multiple quantity",
			problem: Problem{
				Surface: domaingeometry.Polygon{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 0, Y: 0},
							{X: 60, Y: 0},
							{X: 60, Y: 20},
							{X: 0, Y: 20},
						},
					},
				},
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Quantity:  3,
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
					},
				},
				AllowedRotations: []float64{0},
			},
			wantResult: Result{
				Placements: []Placement{
					{
						PatternID: 1,
						X:         10,
						Y:         10,
						Rotation:  0,
					},
					{
						PatternID: 1,
						X:         30,
						Y:         10,
						Rotation:  0,
					},
					{
						PatternID: 1,
						X:         50,
						Y:         10,
						Rotation:  0,
					},
				},
				Unplaced: []UnplacedPattern{},
				Metrics: Metrics{
					RequestedCount: 3,
					PlacedCount:    3,
					SurfaceArea:    1200,
					PlacedArea:     1200,
					Utilization:    1,
				},
			},
		},
		{
			name: "partially unplaced quantity",
			problem: Problem{
				Surface: domaingeometry.Polygon{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 0, Y: 0},
							{X: 40, Y: 0},
							{X: 40, Y: 20},
							{X: 0, Y: 20},
						},
					},
				},
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Quantity:  4,
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
					},
				},
				AllowedRotations: []float64{0},
			},
			wantResult: Result{
				Placements: []Placement{
					{
						PatternID: 1,
						X:         10,
						Y:         10,
						Rotation:  0,
					},
					{
						PatternID: 1,
						X:         30,
						Y:         10,
						Rotation:  0,
					},
				},
				Unplaced: []UnplacedPattern{
					{
						PatternID: 1,
						Quantity:  2,
					},
				},
				Metrics: Metrics{
					RequestedCount: 4,
					PlacedCount:    2,
					SurfaceArea:    800,
					PlacedArea:     800,
					Utilization:    1,
				},
			},
		},
		{
			name: "unplaced big pattern does not stop next small pattern",
			problem: Problem{
				Surface: domaingeometry.Polygon{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 0, Y: 0},
							{X: 40, Y: 0},
							{X: 40, Y: 30},
							{X: 0, Y: 30},
						},
					},
				},
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Quantity:  2,
						Geometry: domaingeometry.Polygon{
							Exterior: domaingeometry.Ring{
								Points: []domaingeometry.Point{
									{X: 0, Y: 0},
									{X: 30, Y: 0},
									{X: 30, Y: 30},
									{X: 0, Y: 30},
								},
							},
						},
					},
					{
						PatternID: 2,
						Quantity:  1,
						Geometry: domaingeometry.Polygon{
							Exterior: domaingeometry.Ring{
								Points: []domaingeometry.Point{
									{X: 0, Y: 0},
									{X: 10, Y: 0},
									{X: 10, Y: 30},
									{X: 0, Y: 30},
								},
							},
						},
					},
				},
				AllowedRotations: []float64{0},
			},
			wantResult: Result{
				Placements: []Placement{
					{
						PatternID: 1,
						X:         15,
						Y:         15,
						Rotation:  0,
					},
					{
						PatternID: 2,
						X:         35,
						Y:         15,
						Rotation:  0,
					},
				},
				Unplaced: []UnplacedPattern{
					{
						PatternID: 1,
						Quantity:  1,
					},
				},
				Metrics: Metrics{
					RequestedCount: 3,
					PlacedCount:    2,
					SurfaceArea:    1200,
					PlacedArea:     1200,
					Utilization:    1,
				},
			},
		},
		{
			name: "largest pattern goes first",
			problem: Problem{
				Surface: domaingeometry.Polygon{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 0, Y: 0},
							{X: 40, Y: 0},
							{X: 40, Y: 20},
							{X: 0, Y: 20},
						},
					},
				},
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Quantity:  1,
						Geometry: domaingeometry.Polygon{
							Exterior: domaingeometry.Ring{
								Points: []domaingeometry.Point{
									{X: 0, Y: 0},
									{X: 10, Y: 0},
									{X: 10, Y: 20},
									{X: 0, Y: 20},
								},
							},
						},
					},
					{
						PatternID: 2,
						Quantity:  1,
						Geometry: domaingeometry.Polygon{
							Exterior: domaingeometry.Ring{
								Points: []domaingeometry.Point{
									{X: 0, Y: 0},
									{X: 30, Y: 0},
									{X: 30, Y: 20},
									{X: 0, Y: 20},
								},
							},
						},
					},
				},
				AllowedRotations: []float64{0},
			},
			wantResult: Result{
				Placements: []Placement{
					{
						PatternID: 2,
						X:         15,
						Y:         10,
						Rotation:  0,
					},
					{
						PatternID: 1,
						X:         35,
						Y:         10,
						Rotation:  0,
					},
				},
				Unplaced: []UnplacedPattern{},
				Metrics: Metrics{
					RequestedCount: 2,
					PlacedCount:    2,
					SurfaceArea:    800,
					PlacedArea:     800,
					Utilization:    1,
				},
			},
		},
		{
			name: "equal area preserves input order",
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
						Quantity:  1,
						Geometry: domaingeometry.Polygon{
							Exterior: domaingeometry.Ring{
								Points: []domaingeometry.Point{
									{X: 0, Y: 0},
									{X: 20, Y: 0},
									{X: 20, Y: 10},
									{X: 0, Y: 10},
								},
							},
						},
					},
					{
						PatternID: 2,
						Quantity:  1,
						Geometry: domaingeometry.Polygon{
							Exterior: domaingeometry.Ring{
								Points: []domaingeometry.Point{
									{X: 0, Y: 0},
									{X: 10, Y: 0},
									{X: 10, Y: 20},
									{X: 0, Y: 20},
								},
							},
						},
					},
				},
				AllowedRotations: []float64{0},
			},
			wantResult: Result{
				Placements: []Placement{
					{
						PatternID: 1,
						X:         10,
						Y:         5,
						Rotation:  0,
					},
					{
						PatternID: 2,
						X:         25,
						Y:         10,
						Rotation:  0,
					},
				},
				Unplaced: []UnplacedPattern{},
				Metrics: Metrics{
					RequestedCount: 2,
					PlacedCount:    2,
					SurfaceArea:    10000,
					PlacedArea:     400,
					Utilization:    0.04,
				},
			},
		},
		{
			name: "existing obstacle respected",
			problem: Problem{
				Surface: domaingeometry.Polygon{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 0, Y: 0},
							{X: 60, Y: 0},
							{X: 60, Y: 20},
							{X: 0, Y: 20},
						},
					},
				},
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Quantity:  1,
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
					},
				},
				Obstacles: []domaingeometry.Polygon{
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
				AllowedRotations: []float64{0},
			},
			wantResult: Result{
				Placements: []Placement{
					{
						PatternID: 1,
						X:         30,
						Y:         10,
						Rotation:  0,
					},
				},
				Unplaced: []UnplacedPattern{},
				Metrics: Metrics{
					RequestedCount: 1,
					PlacedCount:    1,
					SurfaceArea:    1200,
					PlacedArea:     400,
					Utilization:    400.0 / 1200.0,
				},
			},
		},
		{
			name: "obstacles excluded from metrics",
			problem: Problem{
				Surface: domaingeometry.Polygon{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 0, Y: 0},
							{X: 100, Y: 0},
							{X: 100, Y: 20},
							{X: 0, Y: 20},
						},
					},
				},
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Quantity:  1,
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
					},
				},
				Obstacles: []domaingeometry.Polygon{
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
				AllowedRotations: []float64{0},
			},
			wantResult: Result{
				Placements: []Placement{
					{
						PatternID: 1,
						X:         30,
						Y:         10,
						Rotation:  0,
					},
				},
				Unplaced: []UnplacedPattern{},
				Metrics: Metrics{
					RequestedCount: 1,
					PlacedCount:    1,
					SurfaceArea:    2000,
					PlacedArea:     400,
					Utilization:    0.2,
				},
			},
		},
		{
			name: "rotation required",
			problem: Problem{
				Surface: domaingeometry.Polygon{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 0, Y: 0},
							{X: 30, Y: 0},
							{X: 30, Y: 50},
							{X: 0, Y: 50},
						},
					},
				},
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Quantity:  1,
						Geometry: domaingeometry.Polygon{
							Exterior: domaingeometry.Ring{
								Points: []domaingeometry.Point{
									{X: 0, Y: 0},
									{X: 40, Y: 0},
									{X: 40, Y: 20},
									{X: 0, Y: 20},
								},
							},
						},
					},
				},
				AllowedRotations: []float64{0, 90},
			},
			wantResult: Result{
				Placements: []Placement{
					{
						PatternID: 1,
						X:         10,
						Y:         20,
						Rotation:  90,
					},
				},
				Unplaced: []UnplacedPattern{},
				Metrics: Metrics{
					RequestedCount: 1,
					PlacedCount:    1,
					SurfaceArea:    1500,
					PlacedArea:     800,
					Utilization:    800.0 / 1500.0,
				},
			},
		},
		{
			name: "invalid problem",
			problem: Problem{
				Surface: domaingeometry.Polygon{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 0, Y: 0},
							{X: 30, Y: 0},
							{X: 30, Y: 50},
							{X: 0, Y: 50},
						},
					},
				},
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Quantity:  0,
						Geometry: domaingeometry.Polygon{
							Exterior: domaingeometry.Ring{
								Points: []domaingeometry.Point{
									{X: 0, Y: 0},
									{X: 40, Y: 0},
									{X: 40, Y: 20},
									{X: 0, Y: 20},
								},
							},
						},
					},
				},
				AllowedRotations: []float64{0, 90},
			},
			wantResult: Result{},
			wantErr:    ErrInvalidQuantity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := optimizer.Optimize(
				context.Background(),
				tt.problem,
			)

			if tt.wantErr != nil {
				assert.Zero(t, result)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)

			require.Equal(
				t,
				len(tt.wantResult.Placements),
				len(result.Placements),
			)

			require.Equal(
				t,
				len(tt.wantResult.Unplaced),
				len(result.Unplaced),
			)

			l := len(tt.wantResult.Placements)

			for i := 0; i < l; i++ {
				assert.Equal(
					t,
					tt.wantResult.Placements[i].PatternID,
					result.Placements[i].PatternID,
				)

				assert.InDelta(
					t,
					tt.wantResult.Placements[i].X,
					result.Placements[i].X,
					config.Epsilon,
				)

				assert.InDelta(
					t,
					tt.wantResult.Placements[i].Y,
					result.Placements[i].Y,
					config.Epsilon,
				)

				assert.InDelta(
					t,
					tt.wantResult.Placements[i].Rotation,
					result.Placements[i].Rotation,
					config.Epsilon,
				)
			}

			l = len(result.Unplaced)

			for i := 0; i < l; i++ {
				assert.Equal(
					t,
					tt.wantResult.Unplaced[i].PatternID,
					result.Unplaced[i].PatternID,
				)

				assert.Equal(
					t,
					tt.wantResult.Unplaced[i].Quantity,
					result.Unplaced[i].Quantity,
				)
			}

			assert.Equal(
				t,
				tt.wantResult.Metrics.RequestedCount,
				result.Metrics.RequestedCount,
			)

			assert.Equal(
				t,
				tt.wantResult.Metrics.PlacedCount,
				result.Metrics.PlacedCount,
			)

			assert.InDelta(
				t,
				tt.wantResult.Metrics.SurfaceArea,
				result.Metrics.SurfaceArea,
				config.Epsilon,
			)

			assert.InDelta(
				t,
				tt.wantResult.Metrics.PlacedArea,
				result.Metrics.PlacedArea,
				config.Epsilon,
			)

			assert.InDelta(
				t,
				tt.wantResult.Metrics.Utilization,
				result.Metrics.Utilization,
				config.Epsilon,
			)
		})
	}
}

func TestBaselineOptimizerCanceledContext(t *testing.T) {
	sfEngine := simplefeatures.NewEngine()
	optimizer := NewBaselineOptimizer(
		sfEngine,
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := optimizer.Optimize(
		ctx,
		Problem{
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
			AllowedRotations: []float64{0},
		},
	)

	assert.Zero(t, result)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestBaselineOptimizerDoesNotMutatePatterns(t *testing.T) {
	sfEngine := simplefeatures.NewEngine()
	optimizer := NewBaselineOptimizer(sfEngine)

	problem := Problem{
		Surface: domaingeometry.Polygon{
			Exterior: domaingeometry.Ring{
				Points: []domaingeometry.Point{
					{X: 0, Y: 0},
					{X: 40, Y: 0},
					{X: 40, Y: 20},
					{X: 0, Y: 20},
				},
			},
		},
		Patterns: []PatternItem{
			{
				PatternID: 1,
				Quantity:  1,
				Geometry: domaingeometry.Polygon{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 0, Y: 0},
							{X: 10, Y: 0},
							{X: 10, Y: 20},
							{X: 0, Y: 20},
						},
					},
				},
			},
			{
				PatternID: 2,
				Quantity:  1,
				Geometry: domaingeometry.Polygon{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 0, Y: 0},
							{X: 30, Y: 0},
							{X: 30, Y: 20},
							{X: 0, Y: 20},
						},
					},
				},
			},
		},
		AllowedRotations: []float64{0},
	}

	wantPatterns := []PatternItem{
		{
			PatternID: 1,
			Quantity:  1,
			Geometry: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 10, Y: 0},
						{X: 10, Y: 20},
						{X: 0, Y: 20},
					},
				},
			},
		},
		{
			PatternID: 2,
			Quantity:  1,
			Geometry: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 30, Y: 0},
						{X: 30, Y: 20},
						{X: 0, Y: 20},
					},
				},
			},
		},
	}

	_, err := optimizer.Optimize(
		context.Background(),
		problem,
	)

	require.NoError(t, err)

	assert.Equal(
		t,
		wantPatterns,
		problem.Patterns,
	)
}

func TestBaselineOptimizerDoesNotMutateObstacles(t *testing.T) {
	sfEngine := simplefeatures.NewEngine()
	optimizer := NewBaselineOptimizer(sfEngine)

	problem := Problem{
		Surface: domaingeometry.Polygon{
			Exterior: domaingeometry.Ring{
				Points: []domaingeometry.Point{
					{X: 0, Y: 0},
					{X: 40, Y: 0},
					{X: 40, Y: 20},
					{X: 0, Y: 20},
				},
			},
		},
		Patterns: []PatternItem{
			{
				PatternID: 1,
				Quantity:  1,
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
			},
		},
		Obstacles: []domaingeometry.Polygon{
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
		AllowedRotations: []float64{0},
	}

	wantObstacles := []domaingeometry.Polygon{
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
	}

	_, err := optimizer.Optimize(
		context.Background(),
		problem,
	)

	require.NoError(t, err)

	assert.Equal(
		t,
		wantObstacles,
		problem.Obstacles,
	)
}
