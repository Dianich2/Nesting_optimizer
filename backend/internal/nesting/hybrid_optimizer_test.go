package nesting

import (
	"context"
	"errors"
	"testing"

	"server_nesting_optimizer/internal/config"
	"server_nesting_optimizer/internal/geometry/simplefeatures"

	"github.com/stretchr/testify/require"
)

type stubOptimizer struct {
	result Result
	err    error
}

func (o stubOptimizer) Optimize(
	ctx context.Context,
	problem Problem,
) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	if o.err != nil {
		return Result{}, o.err
	}

	return o.result, nil
}

func TestHybridOptimizerOptimizeChoosesBetterCandidate(
	t *testing.T,
) {
	engine := simplefeatures.NewEngine()

	problem := Problem{
		Surface: rectangle(100, 100),
		Patterns: []PatternItem{
			{
				PatternID: 1,
				Quantity:  2,
				Geometry:  rectangle(20, 20),
			},
		},
		AllowedRotations: []float64{0},
	}

	geneticOptimizer := stubOptimizer{
		result: Result{
			Placements: []Placement{
				{
					PatternID: 1,
					X:         70,
					Y:         70,
					Rotation:  0,
				},
			},
		},
	}

	annealingOptimizer := stubOptimizer{
		result: Result{
			Placements: []Placement{
				{
					PatternID: 1,
					X:         70,
					Y:         70,
					Rotation:  0,
				},
				{
					PatternID: 1,
					X:         90,
					Y:         70,
					Rotation:  0,
				},
			},
		},
	}

	optimizer, err := NewHybridOptimizer(
		engine,
		geneticOptimizer,
		annealingOptimizer,
		CompactionConfig{
			MaxPasses:         2,
			InitialStepFactor: 0.1,
			MinStepFactor:     0.0001,
		},
	)
	require.NoError(t, err)

	result, err := optimizer.Optimize(
		context.Background(),
		problem,
	)
	require.NoError(t, err)

	require.Len(
		t,
		result.Placements,
		2,
	)

	require.Empty(
		t,
		result.Unplaced,
	)

	require.Equal(
		t,
		2,
		result.Metrics.RequestedCount,
	)

	require.Equal(
		t,
		2,
		result.Metrics.PlacedCount,
	)

	require.InDelta(
		t,
		10000,
		result.Metrics.SurfaceArea,
		config.Epsilon,
	)

	require.InDelta(
		t,
		800,
		result.Metrics.PlacedArea,
		config.Epsilon,
	)

	require.InDelta(
		t,
		0.08,
		result.Metrics.Utilization,
		config.Epsilon,
	)
}

func TestHybridOptimizerAppliesCompaction(
	t *testing.T,
) {
	engine := simplefeatures.NewEngine()

	problem := Problem{
		Surface: rectangle(100, 100),
		Patterns: []PatternItem{
			{
				PatternID: 1,
				Quantity:  1,
				Geometry:  rectangle(20, 20),
			},
		},
		AllowedRotations: []float64{0},
	}

	candidate := Result{
		Placements: []Placement{
			{
				PatternID: 1,
				X:         70,
				Y:         70,
				Rotation:  0,
			},
		},
	}

	optimizer, err := NewHybridOptimizer(
		engine,
		stubOptimizer{
			result: candidate,
		},
		stubOptimizer{
			result: candidate,
		},
		CompactionConfig{
			MaxPasses:         2,
			InitialStepFactor: 0.1,
			MinStepFactor:     0.0001,
		},
	)
	require.NoError(t, err)

	result, err := optimizer.Optimize(
		context.Background(),
		problem,
	)
	require.NoError(t, err)

	require.Len(
		t,
		result.Placements,
		1,
	)

	assertPlacementInDelta(
		t,
		Placement{
			PatternID: 1,
			X:         10,
			Y:         10,
			Rotation:  0,
		},
		result.Placements[0],
		config.Epsilon,
	)

	require.Equal(
		t,
		1,
		result.Metrics.RequestedCount,
	)

	require.Equal(
		t,
		1,
		result.Metrics.PlacedCount,
	)
}

func TestHybridOptimizerReturnsGeneticOnEqualFitness(
	t *testing.T,
) {
	engine := simplefeatures.NewEngine()

	problem := Problem{
		Surface: rectangle(100, 100),
		Patterns: []PatternItem{
			{
				PatternID: 1,
				Quantity:  1,
				Geometry:  rectangle(20, 20),
			},
		},
		AllowedRotations: []float64{0},
	}

	geneticOptimizer := stubOptimizer{
		result: Result{
			Placements: []Placement{
				{
					PatternID: 1,
					X:         70,
					Y:         70,
					Rotation:  0,
				},
			},
		},
	}

	annealingOptimizer := stubOptimizer{
		result: Result{
			Placements: []Placement{
				{
					PatternID: 1,
					X:         80,
					Y:         80,
					Rotation:  0,
				},
			},
		},
	}

	optimizer, err := NewHybridOptimizer(
		engine,
		geneticOptimizer,
		annealingOptimizer,
		DefaultCompactionConfig(),
	)
	require.NoError(t, err)

	result, err := optimizer.Optimize(
		context.Background(),
		problem,
	)
	require.NoError(t, err)

	require.Len(
		t,
		result.Placements,
		1,
	)

	assertPlacementInDelta(
		t,
		Placement{
			PatternID: 1,
			X:         10,
			Y:         10,
			Rotation:  0,
		},
		result.Placements[0],
		config.Epsilon,
	)
}

func TestHybridOptimizerChildErrors(
	t *testing.T,
) {
	engine := simplefeatures.NewEngine()

	problem := Problem{
		Surface: rectangle(100, 100),
		Patterns: []PatternItem{
			{
				PatternID: 1,
				Quantity:  1,
				Geometry:  rectangle(20, 20),
			},
		},
		AllowedRotations: []float64{0},
	}

	childErr := errors.New(
		"optimizer failed",
	)

	tests := []struct {
		name      string
		genetic   Optimizer
		annealing Optimizer
	}{
		{
			name: "genetic optimizer error",
			genetic: stubOptimizer{
				err: childErr,
			},
			annealing: stubOptimizer{},
		},
		{
			name: "annealing optimizer error",
			genetic: stubOptimizer{
				result: Result{
					Placements: []Placement{
						{
							PatternID: 1,
							X:         10,
							Y:         10,
							Rotation:  0,
						},
					},
				},
			},
			annealing: stubOptimizer{
				err: childErr,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			optimizer, err := NewHybridOptimizer(
				engine,
				tt.genetic,
				tt.annealing,
				DefaultCompactionConfig(),
			)
			require.NoError(t, err)

			_, err = optimizer.Optimize(
				context.Background(),
				problem,
			)

			require.ErrorIs(
				t,
				err,
				childErr,
			)
		})
	}
}

func TestHybridOptimizerCanceledContext(
	t *testing.T,
) {
	engine := simplefeatures.NewEngine()

	optimizer, err := NewHybridOptimizer(
		engine,
		stubOptimizer{},
		stubOptimizer{},
		DefaultCompactionConfig(),
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

func TestHybridOptimizerInvalidProblem(
	t *testing.T,
) {
	engine := simplefeatures.NewEngine()

	optimizer, err := NewHybridOptimizer(
		engine,
		stubOptimizer{},
		stubOptimizer{},
		DefaultCompactionConfig(),
	)
	require.NoError(t, err)

	_, err = optimizer.Optimize(
		context.Background(),
		Problem{},
	)

	require.Error(t, err)
}

func TestNewHybridOptimizerInvalidConfig(
	t *testing.T,
) {
	engine := simplefeatures.NewEngine()

	tests := []struct {
		name               string
		geneticOptimizer   Optimizer
		annealingOptimizer Optimizer
		config             CompactionConfig
	}{
		{
			name:               "nil genetic optimizer",
			geneticOptimizer:   nil,
			annealingOptimizer: stubOptimizer{},
			config:             DefaultCompactionConfig(),
		},
		{
			name:               "nil annealing optimizer",
			geneticOptimizer:   stubOptimizer{},
			annealingOptimizer: nil,
			config:             DefaultCompactionConfig(),
		},
		{
			name:               "invalid compaction config",
			geneticOptimizer:   stubOptimizer{},
			annealingOptimizer: stubOptimizer{},
			config: CompactionConfig{
				MaxPasses:         0,
				InitialStepFactor: 0.1,
				MinStepFactor:     0.001,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewHybridOptimizer(
				engine,
				tt.geneticOptimizer,
				tt.annealingOptimizer,
				tt.config,
			)

			require.Error(t, err)
		})
	}
}

func TestCalculatePlacementsFitness(
	t *testing.T,
) {
	engine := simplefeatures.NewEngine()

	problem := Problem{
		Surface: rectangle(100, 100),
		Patterns: []PatternItem{
			{
				PatternID: 1,
				Quantity:  3,
				Geometry:  rectangle(20, 20),
			},
		},
	}

	fitness, err := calculatePlacementsFitness(
		engine,
		problem,
		[]Placement{
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
	)
	require.NoError(t, err)

	require.Equal(
		t,
		1,
		fitness.UnplacedCount,
	)

	require.InDelta(
		t,
		800,
		fitness.PlacedArea,
		config.Epsilon,
	)

	require.InDelta(
		t,
		800,
		fitness.UsedArea,
		config.Epsilon,
	)

	require.InDelta(
		t,
		20,
		fitness.UsedHeight,
		config.Epsilon,
	)
}
