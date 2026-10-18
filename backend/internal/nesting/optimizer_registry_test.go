package nesting

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type registryTestOptimizer struct {
	instanceID int
}

func (*registryTestOptimizer) Optimize(
	ctx context.Context,
	problem Problem,
) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	return Result{}, nil
}

func TestOptimizerFactoryRegistryCreatesIndependentInstances(
	t *testing.T,
) {
	created := 0

	registry, err := NewOptimizerFactoryRegistry(
		map[Algorithm]OptimizerFactory{
			HybridAlgorithm: func() (Optimizer, error) {
				created++
				return &registryTestOptimizer{instanceID: created}, nil
			},
		},
	)
	require.NoError(t, err)

	first, err := registry.Get(HybridAlgorithm)
	require.NoError(t, err)

	second, err := registry.Get(HybridAlgorithm)
	require.NoError(t, err)

	require.NotSame(t, first, second)
	require.Equal(t, 2, created)
}

func TestOptimizerFactoryRegistryRejectsInvalidFactories(
	t *testing.T,
) {
	_, err := NewOptimizerFactoryRegistry(nil)
	require.Error(t, err)

	_, err = NewOptimizerFactoryRegistry(
		map[Algorithm]OptimizerFactory{
			"": func() (Optimizer, error) {
				return &registryTestOptimizer{}, nil
			},
		},
	)
	require.Error(t, err)

	_, err = NewOptimizerFactoryRegistry(
		map[Algorithm]OptimizerFactory{
			HybridAlgorithm: nil,
		},
	)
	require.Error(t, err)
}

func TestOptimizerFactoryRegistryHandlesFactoryErrors(
	t *testing.T,
) {
	sentinel := errors.New("factory failed")

	registry, err := NewOptimizerFactoryRegistry(
		map[Algorithm]OptimizerFactory{
			HybridAlgorithm: func() (Optimizer, error) {
				return nil, sentinel
			},
		},
	)
	require.NoError(t, err)

	_, err = registry.Get(HybridAlgorithm)
	require.ErrorIs(t, err, sentinel)
}

func TestOptimizerFactoryRegistryRejectsNilResult(
	t *testing.T,
) {
	registry, err := NewOptimizerFactoryRegistry(
		map[Algorithm]OptimizerFactory{
			HybridAlgorithm: func() (Optimizer, error) {
				return nil, nil
			},
		},
	)
	require.NoError(t, err)

	_, err = registry.Get(HybridAlgorithm)
	require.Error(t, err)
}

func TestOptimizerFactoryRegistryUnknownAlgorithm(
	t *testing.T,
) {
	registry, err := NewOptimizerFactoryRegistry(
		map[Algorithm]OptimizerFactory{
			HybridAlgorithm: func() (Optimizer, error) {
				return &registryTestOptimizer{}, nil
			},
		},
	)
	require.NoError(t, err)

	_, err = registry.Get(Algorithm("unknown"))
	require.Error(t, err)
}

func TestLegacyOptimizerRegistryPreservesExistingBehavior(
	t *testing.T,
) {
	shared := &registryTestOptimizer{}

	registry, err := NewOptimizerRegistry(
		map[Algorithm]Optimizer{
			HybridAlgorithm: shared,
		},
	)
	require.NoError(t, err)

	first, err := registry.Get(HybridAlgorithm)
	require.NoError(t, err)

	second, err := registry.Get(HybridAlgorithm)
	require.NoError(t, err)

	require.Same(t, shared, first)
	require.Same(t, first, second)
}

func TestOptimizerFactoryRegistryRejectsTypedNil(t *testing.T) {
	registry, err := NewOptimizerFactoryRegistry(
		map[Algorithm]OptimizerFactory{
			HybridAlgorithm: func() (Optimizer, error) {
				var optimizer *GeneticOptimizer
				return optimizer, nil
			},
		},
	)
	require.NoError(t, err)

	optimizer, err := registry.Get(HybridAlgorithm)

	require.Error(t, err)
	require.True(t, isNilOptimizer(optimizer))
}

func TestLegacyOptimizerRegistryRejectsTypedNil(t *testing.T) {
	var optimizer *GeneticOptimizer

	_, err := NewOptimizerRegistry(
		map[Algorithm]Optimizer{
			GeneticAlgorithm: optimizer,
		},
	)

	require.Error(t, err)
}
