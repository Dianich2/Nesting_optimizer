package handler

import (
	"context"
	"sync"
	"testing"
)

func TestNestingExecutionsCancelConcurrentCallers(t *testing.T) {
	var executions nestingExecutions
	key := nestingExecution{userID: 1, projectID: 2, surfaceID: 3, key: "operation"}
	primary, finishPrimary := executions.begin(t.Context(), key)
	defer finishPrimary()
	var callers sync.WaitGroup
	started := make(chan context.Context, 8)
	cancelCalled := make(chan struct{})
	for range cap(started) {
		callers.Go(func() {
			ctx, finish := executions.begin(t.Context(), key)
			defer finish()
			started <- ctx
			<-cancelCalled
			if ctx.Err() != context.Canceled {
				t.Errorf("attached execution got %v, want context.Canceled", ctx.Err())
			}
		})
	}
	for range cap(started) {
		if attached := <-started; attached != primary {
			t.Error("replay did not attach to the active execution")
		}
	}
	unrelated := key
	unrelated.userID++
	if executions.cancel(unrelated) || primary.Err() != nil {
		t.Error("execution can be cancelled by another user")
	}
	if !executions.cancel(key) {
		t.Error("active execution was not found")
	}
	close(cancelCalled)
	callers.Wait()
}
