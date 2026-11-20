package handler

import (
	"context"
	nestingcase "server_nesting_optimizer/internal/usecase/nesting"
	"server_nesting_optimizer/pkg/apperror"
)

// In-flight preview computation is shared within this API process. The ready
// preview is kept in PostgreSQL, where its key is unique and survives restarts.
func (executions *nestingExecutions) preview(ctx context.Context, key nestingExecution, hash [32]byte, limiter *nestingcase.ComputeLimiter, run func() (nestingcase.RunNestingOutput, error)) (output nestingcase.RunNestingOutput, executionErr error) {
	executions.mu.Lock()
	job := executions.active[key]
	if job == nil {
		executions.mu.Unlock()
		return output, apperror.Internal("preview execution is not registered", nil)
	}
	owner := job.previewDone == nil
	if owner {
		job.previewDone = make(chan struct{})
		job.previewHash = hash
	} else if job.previewHash != hash {
		executions.mu.Unlock()
		return output, apperror.Conflict("idempotency key already used for different preview parameters")
	}
	done := job.previewDone
	executions.mu.Unlock()
	if !owner {
		release, err := limiter.AwaitResponse(ctx)
		if err != nil {
			return output, err
		}
		defer release()
		select {
		case <-ctx.Done():
			return output, ctx.Err()
		case <-done:
			return job.previewResult, job.previewErr
		}
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			output = nestingcase.RunNestingOutput{}
			executionErr = apperror.Internal("preview execution panicked", nil)
		}
		executions.mu.Lock()
		job.previewResult = output
		job.previewErr = executionErr
		close(done)
		executions.mu.Unlock()
	}()
	return run()
}
