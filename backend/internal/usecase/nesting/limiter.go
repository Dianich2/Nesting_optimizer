package nesting

import (
	"context"
	"server_nesting_optimizer/pkg/apperror"
)

// ComputeLimiter bounds running calculations and the number waiting for a slot.
// One instance is shared by every automatic algorithm and preview in the API.
type ComputeLimiter struct {
	responses chan struct{}
	admitted  chan struct{}
	running   chan struct{}
}

func NewComputeLimiter(active, queued int) *ComputeLimiter {
	if active < 1 {
		active = 2
	}
	if queued < 0 {
		queued = 0
	}
	return &ComputeLimiter{admitted: make(chan struct{}, active+queued), running: make(chan struct{}, active), responses: make(chan struct{}, 32)}
}

// Duplicate POSTs may wait for the original response, but cannot create an
// unbounded population of polling requests or occupy optimizer slots.
func (l *ComputeLimiter) AwaitResponse(ctx context.Context) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	select {
	case l.responses <- struct{}{}:
		return func() { <-l.responses }, nil
	default:
		return nil, &apperror.AppError{Code: apperror.AppCodePending, Message: "operation is already running; read its status"}
	}
}

func (l *ComputeLimiter) Acquire(ctx context.Context) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	select {
	case l.admitted <- struct{}{}:
	default:
		return nil, &apperror.AppError{Code: apperror.AppCodeOverloaded, Message: "nesting capacity is full"}
	}
	select {
	case l.running <- struct{}{}:
		return func() { <-l.running; <-l.admitted }, nil
	case <-ctx.Done():
		<-l.admitted
		return nil, ctx.Err()
	}
}
