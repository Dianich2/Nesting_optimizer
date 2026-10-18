package nesting

import (
	"context"
	"sync"
	"testing"

	"server_nesting_optimizer/internal/geometry/nfp"
	"server_nesting_optimizer/internal/geometry/simplefeatures"

	"github.com/stretchr/testify/require"
)

func TestHybridFactoryConcurrentIndependentRuns(t *testing.T) {
	const workers = 4

	var problem Problem
	found := false

	for _, benchmarkCase := range benchmarkProblems() {
		if benchmarkCase.name == "simple_control" {
			problem = benchmarkCase.problem
			found = true
			break
		}
	}

	require.True(t, found, "simple_control problem not found")

	profile := comparisonBenchmarkProfile()

	registry, err := NewOptimizerFactoryRegistry(
		map[Algorithm]OptimizerFactory{
			HybridAlgorithm: func() (Optimizer, error) {
				engine := simplefeatures.NewEngine()

				builder := nfp.NewCachedBuilder(
					nfp.NewBuilder(engine),
				)

				ga, err := NewGeneticOptimizer(
					engine,
					builder,
					profile.geneticConfig,
				)
				if err != nil {
					return nil, err
				}

				ga.rng = newBenchmarkRandom(benchmarkSeed + 3)

				sa, err := NewSimulatedAnnealingOptimizer(
					engine,
					builder,
					profile.annealingConfig,
				)
				if err != nil {
					return nil, err
				}

				sa.rng = newBenchmarkRandom(benchmarkSeed + 4)

				return NewHybridOptimizer(
					engine,
					ga,
					sa,
					profile.compactionConfig,
				)
			},
		},
	)
	require.NoError(t, err)

	// Establish a reference result.
	referenceOptimizer, err := registry.Get(HybridAlgorithm)
	require.NoError(t, err)

	reference, err := referenceOptimizer.Optimize(
		context.Background(),
		problem,
	)
	require.NoError(t, err)

	results := make([]Result, workers)
	errs := make([]error, workers)

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func(index int) {
			defer wg.Done()

			optimizer, err := registry.Get(HybridAlgorithm)
			if err != nil {
				errs[index] = err
				return
			}

			results[index], errs[index] = optimizer.Optimize(
				context.Background(),
				problem,
			)
		}(i)
	}

	wg.Wait()

	for i := 0; i < workers; i++ {
		require.NoError(
			t,
			errs[i],
			"worker %d failed",
			i,
		)

		require.Equal(
			t,
			reference,
			results[i],
			"worker %d result differs",
			i,
		)
	}
}
