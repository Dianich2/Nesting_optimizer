package projecttransaction

import "context"

// Executor serializes project mutations and keeps validation and persistence atomic.
type Executor interface {
	WithinProject(context.Context, int64, int64, func(context.Context) error) error
}

func Run[T any](ctx context.Context, executor Executor, projectID, userID int64, fn func(context.Context) (T, error)) (T, error) {
	if executor == nil {
		return fn(ctx)
	}
	var result T
	err := executor.WithinProject(ctx, projectID, userID, func(txContext context.Context) error {
		var err error
		result, err = fn(txContext)
		return err
	})
	if err != nil {
		var zero T
		return zero, err
	}
	return result, nil
}
