package nesting

import (
	"context"
	"server_nesting_optimizer/internal/config"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry/nfp"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertPlacementInDelta(
	t *testing.T,
	want Placement,
	got Placement,
	delta float64,
) {
	t.Helper()

	assert.Equal(t, want.PatternID, got.PatternID)
	assert.InDelta(t, want.X, got.X, delta)
	assert.InDelta(t, want.Y, got.Y, delta)
	assert.InDelta(t, want.Rotation, got.Rotation, delta)
}

func assertUnplaced(
	t *testing.T,
	want UnplacedPattern,
	got UnplacedPattern,
) {
	t.Helper()

	assert.Equal(
		t,
		want.PatternID,
		got.PatternID,
	)

	assert.Equal(
		t,
		want.Quantity,
		got.Quantity,
	)
}

func assertMetrics(
	t *testing.T,
	want Metrics,
	got Metrics,
	delta float64,
) {
	t.Helper()

	assert.Equal(t, want.RequestedCount, got.RequestedCount)
	assert.Equal(t, want.PlacedCount, got.PlacedCount)

	assert.InDelta(t, want.SurfaceArea, got.SurfaceArea, delta)
	assert.InDelta(t, want.PlacedArea, got.PlacedArea, delta)

	assert.InDelta(t, want.Utilization, got.Utilization, delta)
}

func TestNFPGreedyOptimizerOptimize(t *testing.T) {
	engine := simplefeatures.NewEngine()
	nfpBuilder := nfp.NewBuilder(engine)
	optimizer := NewNFPGreedyOptimizer(
		engine,
		nfpBuilder,
	)

	tests := []struct {
		name       string
		problem    Problem
		wantResult Result
		wantErr    bool
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
					SurfaceArea: 10000,
				},
			},
		},
		{
			name: "multiple quantity packs compactly",
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
						Quantity:  5,
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
					{PatternID: 1, X: 10, Y: 10, Rotation: 0},
					{PatternID: 1, X: 30, Y: 10, Rotation: 0},
					{PatternID: 1, X: 50, Y: 10, Rotation: 0},
					{PatternID: 1, X: 70, Y: 10, Rotation: 0},
					{PatternID: 1, X: 90, Y: 10, Rotation: 0},
				},
				Unplaced: []UnplacedPattern{},
				Metrics: Metrics{
					RequestedCount: 5,
					PlacedCount:    5,
					SurfaceArea:    2000,
					PlacedArea:     2000,
					Utilization:    1,
				},
			},
		},
		{
			name: "partially unplaced",
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
					{PatternID: 1, X: 10, Y: 10, Rotation: 0},
					{PatternID: 1, X: 30, Y: 10, Rotation: 0},
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
				Surface:          domaingeometry.Polygon{},
				AllowedRotations: []float64{0},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := optimizer.Optimize(
				context.Background(),
				tt.problem,
			)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, result.Placements, len(tt.wantResult.Placements))

			for i := 0; i < len(tt.wantResult.Placements); i++ {
				assertPlacementInDelta(
					t,
					tt.wantResult.Placements[i],
					result.Placements[i],
					config.Epsilon,
				)
			}

			assert.Len(t, result.Unplaced, len(tt.wantResult.Unplaced))

			for i := 0; i < len(tt.wantResult.Unplaced); i++ {
				assertUnplaced(
					t,
					tt.wantResult.Unplaced[i],
					result.Unplaced[i],
				)
			}

			assertMetrics(
				t,
				tt.wantResult.Metrics,
				result.Metrics,
				config.Epsilon,
			)

		})
	}
}

func rectangle(width, height float64) domaingeometry.Polygon {
	return domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 0, Y: 0},
				{X: width, Y: 0},
				{X: width, Y: height},
				{X: 0, Y: height},
			},
		},
	}
}

func TestNFPGreedyOptimizerLargestPatternGoesFirst(t *testing.T) {
	engine := simplefeatures.NewEngine()
	optimizer := NewNFPGreedyOptimizer(
		engine,
		nfp.NewBuilder(engine),
	)

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
				Geometry:  rectangle(10, 20),
			},
			{
				PatternID: 2,
				Quantity:  1,
				Geometry:  rectangle(30, 20),
			},
		},
		AllowedRotations: []float64{0},
	}

	result, err := optimizer.Optimize(
		context.Background(),
		problem,
	)

	require.NoError(t, err)
	require.Len(t, result.Placements, 2)

	assert.Equal(t, int64(2), result.Placements[0].PatternID)
	assert.Equal(t, int64(1), result.Placements[1].PatternID)

	assert.Equal(t, 2, result.Metrics.RequestedCount)
	assert.Equal(t, 2, result.Metrics.PlacedCount)
	assert.InDelta(t, 800.0, result.Metrics.PlacedArea, config.Epsilon)
	assert.InDelta(t, 1.0, result.Metrics.Utilization, config.Epsilon)
}

func TestNFPGreedyOptimizerCanceledContext(t *testing.T) {
	engine := simplefeatures.NewEngine()
	optimizer := NewNFPGreedyOptimizer(
		engine,
		nfp.NewBuilder(engine),
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := optimizer.Optimize(
		ctx,
		Problem{},
	)

	assert.Equal(t, Result{}, result)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestNFPGreedyOptimizerRespectsExistingObstacle(t *testing.T) {
	engine := simplefeatures.NewEngine()
	optimizer := NewNFPGreedyOptimizer(
		engine,
		nfp.NewBuilder(engine),
	)

	problem := Problem{
		Surface: rectangle(60, 20),

		Patterns: []PatternItem{
			{
				PatternID: 1,
				Quantity:  1,
				Geometry:  rectangle(20, 20),
			},
		},

		Obstacles: []domaingeometry.Polygon{
			rectangle(20, 20),
		},

		AllowedRotations: []float64{0},
	}

	result, err := optimizer.Optimize(
		context.Background(),
		problem,
	)

	require.NoError(t, err)
	require.Len(t, result.Placements, 1)

	assertPlacementInDelta(
		t,
		Placement{
			PatternID: 1,
			X:         30,
			Y:         10,
			Rotation:  0,
		},
		result.Placements[0],
		config.Epsilon,
	)

	assertMetrics(
		t,
		Metrics{
			RequestedCount: 1,
			PlacedCount:    1,
			SurfaceArea:    1200,
			PlacedArea:     400,
			Utilization:    400.0 / 1200.0,
		},
		result.Metrics,
		config.Epsilon,
	)
}

func TestNFPGreedyOptimizerDoesNotMutateObstacles(t *testing.T) {
	engine := simplefeatures.NewEngine()
	optimizer := NewNFPGreedyOptimizer(
		engine,
		nfp.NewBuilder(engine),
	)

	obstacle := rectangle(20, 20)

	problem := Problem{
		Surface:          rectangle(60, 20),
		Patterns:         []PatternItem{},
		Obstacles:        []domaingeometry.Polygon{obstacle},
		AllowedRotations: []float64{0},
	}

	before := problem.Obstacles[0]

	_, err := optimizer.Optimize(
		context.Background(),
		problem,
	)

	require.NoError(t, err)

	assert.Equal(
		t,
		before,
		problem.Obstacles[0],
	)
}
