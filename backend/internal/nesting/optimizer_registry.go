package nesting

import "fmt"

type OptimizerRegistry struct {
	optimizers map[Algorithm]Optimizer
}

func NewOptimizerRegistry(
	optimizers map[Algorithm]Optimizer,
) (*OptimizerRegistry, error) {
	if len(optimizers) == 0 {
		return nil, fmt.Errorf(
			"optimizers must not be empty",
		)
	}

	registryMap := make(map[Algorithm]Optimizer, len(optimizers))

	for algorithm, optimizer := range optimizers {
		if algorithm == "" {
			return nil, fmt.Errorf(
				"algorithm must not be empty",
			)
		}

		if optimizer == nil {
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

func (r *OptimizerRegistry) Get(
	algorithm Algorithm,
) (Optimizer, error) {
	optimizer, exists := r.optimizers[algorithm]
	if !exists {
		return nil, fmt.Errorf(
			"optimizer for algorithm '%s' not found",
			algorithm,
		)
	}

	return optimizer, nil
}
