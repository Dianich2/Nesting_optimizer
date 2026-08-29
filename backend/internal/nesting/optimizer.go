package nesting

import "context"

type Optimizer interface {
	Optimize(
		ctx context.Context,
		problem Problem,
	) (Result, error)
}
