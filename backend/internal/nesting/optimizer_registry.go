package nesting

import (
	"fmt"
	"reflect"
)

// OptimizerFactory creates an independent optimizer instance.
type OptimizerFactory func() (Optimizer, error)

type OptimizerRegistry struct {
	optimizers map[Algorithm]Optimizer
	factories  map[Algorithm]OptimizerFactory
}

// NewOptimizerRegistry preserves the existing API.
//
// Optimizers registered this way are shared instances.
// New production code should use NewOptimizerFactoryRegistry.
func NewOptimizerRegistry(
	optimizers map[Algorithm]Optimizer,
) (*OptimizerRegistry, error) {
	if len(optimizers) == 0 {
		return nil, fmt.Errorf("optimizers must not be empty")
	}

	registryMap := make(
		map[Algorithm]Optimizer,
		len(optimizers),
	)

	for algorithm, optimizer := range optimizers {
		if algorithm == "" {
			return nil, fmt.Errorf(
				"algorithm must not be empty",
			)
		}

		if isNilOptimizer(optimizer) {
			return nil, fmt.Errorf(
				"optimizer must not be nil",
			)
		}

		registryMap[algorithm] = optimizer
	}

	return &OptimizerRegistry{
		optimizers: registryMap,
	}, nil
}

// NewOptimizerFactoryRegistry registers factories instead
// of sharing optimizer instances between requests.
func NewOptimizerFactoryRegistry(
	factories map[Algorithm]OptimizerFactory,
) (*OptimizerRegistry, error) {
	if len(factories) == 0 {
		return nil, fmt.Errorf(
			"optimizer factories must not be empty",
		)
	}

	registryMap := make(
		map[Algorithm]OptimizerFactory,
		len(factories),
	)

	for algorithm, factory := range factories {
		if algorithm == "" {
			return nil, fmt.Errorf(
				"algorithm must not be empty",
			)
		}

		if factory == nil {
			return nil, fmt.Errorf(
				"optimizer factory must not be nil",
			)
		}

		registryMap[algorithm] = factory
	}

	return &OptimizerRegistry{
		factories: registryMap,
	}, nil
}

func (r *OptimizerRegistry) Get(
	algorithm Algorithm,
) (Optimizer, error) {
	if r == nil {
		return nil, fmt.Errorf(
			"optimizer registry is nil",
		)
	}

	if r.factories != nil {
		factory, exists := r.factories[algorithm]
		if !exists {
			return nil, fmt.Errorf(
				"optimizer for algorithm '%s' not found",
				algorithm,
			)
		}

		optimizer, err := factory()
		if err != nil {
			return nil, fmt.Errorf(
				"create optimizer for algorithm '%s': %w",
				algorithm,
				err,
			)
		}

		if isNilOptimizer(optimizer) {
			return nil, fmt.Errorf(
				"optimizer factory for algorithm '%s' returned nil",
				algorithm,
			)
		}

		return optimizer, nil
	}

	optimizer, exists := r.optimizers[algorithm]
	if !exists {
		return nil, fmt.Errorf(
			"optimizer for algorithm '%s' not found",
			algorithm,
		)
	}

	return optimizer, nil
}

func isNilOptimizer(optimizer Optimizer) bool {
	if optimizer == nil {
		return true
	}

	value := reflect.ValueOf(optimizer)

	switch value.Kind() {
	case reflect.Chan,
		reflect.Func,
		reflect.Interface,
		reflect.Map,
		reflect.Pointer,
		reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
