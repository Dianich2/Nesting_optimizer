package nesting

import (
	"context"
	"testing"

	"server_nesting_optimizer/internal/config"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry/nfp"
	"server_nesting_optimizer/internal/geometry/simplefeatures"

	"github.com/stretchr/testify/require"
)

func TestSimulatedAnnealingOptimizerOptimize(t *testing.T) {
	engine := simplefeatures.NewEngine()
	nfpBuilder := nfp.NewBuilder(engine)

	optimizer, err := NewSimulatedAnnealingOptimizer(
		engine,
		nfpBuilder,
		AnnealingConfig{
			InitialTemperature: 1,
			MinTemperature:     0.1,
			CoolingRate:        0.5,
			MaxIterations:      2,
		},
	)
	require.NoError(t, err)

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
				Surface: rectangle(100, 20),
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Quantity:  5,
						Geometry:  rectangle(20, 20),
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
					{
						PatternID: 1,
						X:         70,
						Y:         10,
						Rotation:  0,
					},
					{
						PatternID: 1,
						X:         90,
						Y:         10,
						Rotation:  0,
					},
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
				Surface: rectangle(40, 20),
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Quantity:  4,
						Geometry:  rectangle(20, 20),
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
			name: "rotation required",
			problem: Problem{
				Surface: rectangle(30, 50),
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Quantity:  1,
						Geometry:  rectangle(40, 20),
					},
				},
				AllowedRotations: []float64{90},
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

			require.Len(
				t,
				result.Placements,
				len(tt.wantResult.Placements),
			)

			for i := range tt.wantResult.Placements {
				assertPlacementInDelta(
					t,
					tt.wantResult.Placements[i],
					result.Placements[i],
					config.Epsilon,
				)
			}

			require.Len(
				t,
				result.Unplaced,
				len(tt.wantResult.Unplaced),
			)

			for i := range tt.wantResult.Unplaced {
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

func TestSimulatedAnnealingOptimizerCanceledContext(
	t *testing.T,
) {
	engine := simplefeatures.NewEngine()
	nfpBuilder := nfp.NewBuilder(engine)

	optimizer, err := NewSimulatedAnnealingOptimizer(
		engine,
		nfpBuilder,
		AnnealingConfig{
			InitialTemperature: 1,
			MinTemperature:     0.1,
			CoolingRate:        0.5,
			MaxIterations:      2,
		},
	)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	_, err = optimizer.Optimize(
		ctx,
		Problem{
			Surface: rectangle(100, 100),
			Patterns: []PatternItem{
				{
					PatternID: 1,
					Quantity:  1,
					Geometry:  rectangle(20, 20),
				},
			},
			AllowedRotations: []float64{0},
		},
	)

	require.ErrorIs(
		t,
		err,
		context.Canceled,
	)
}

func TestNewSimulatedAnnealingOptimizerInvalidConfig(
	t *testing.T,
) {
	engine := simplefeatures.NewEngine()
	nfpBuilder := nfp.NewBuilder(engine)

	tests := []struct {
		name   string
		config AnnealingConfig
	}{
		{
			name: "zero initial temperature",
			config: AnnealingConfig{
				InitialTemperature: 0,
				MinTemperature:     0.1,
				CoolingRate:        0.5,
				MaxIterations:      10,
			},
		},
		{
			name: "zero minimum temperature",
			config: AnnealingConfig{
				InitialTemperature: 1,
				MinTemperature:     0,
				CoolingRate:        0.5,
				MaxIterations:      10,
			},
		},
		{
			name: "minimum equals initial temperature",
			config: AnnealingConfig{
				InitialTemperature: 1,
				MinTemperature:     1,
				CoolingRate:        0.5,
				MaxIterations:      10,
			},
		},
		{
			name: "minimum greater than initial temperature",
			config: AnnealingConfig{
				InitialTemperature: 1,
				MinTemperature:     2,
				CoolingRate:        0.5,
				MaxIterations:      10,
			},
		},
		{
			name: "zero cooling rate",
			config: AnnealingConfig{
				InitialTemperature: 1,
				MinTemperature:     0.1,
				CoolingRate:        0,
				MaxIterations:      10,
			},
		},
		{
			name: "cooling rate equals one",
			config: AnnealingConfig{
				InitialTemperature: 1,
				MinTemperature:     0.1,
				CoolingRate:        1,
				MaxIterations:      10,
			},
		},
		{
			name: "cooling rate greater than one",
			config: AnnealingConfig{
				InitialTemperature: 1,
				MinTemperature:     0.1,
				CoolingRate:        1.1,
				MaxIterations:      10,
			},
		},
		{
			name: "zero max iterations",
			config: AnnealingConfig{
				InitialTemperature: 1,
				MinTemperature:     0.1,
				CoolingRate:        0.5,
				MaxIterations:      0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSimulatedAnnealingOptimizer(
				engine,
				nfpBuilder,
				tt.config,
			)

			require.Error(t, err)
		})
	}
}
