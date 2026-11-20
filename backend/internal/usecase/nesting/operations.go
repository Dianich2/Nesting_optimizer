package nesting

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	domainprojectsurface "server_nesting_optimizer/internal/domain/project_surface"
	"time"

	"github.com/google/uuid"
	"server_nesting_optimizer/pkg/apperror"
)

type OperationRepository interface {
	Find(context.Context, int64, uuid.UUID) ([]byte, *OperationStatus, error)
	Reserve(context.Context, int64, uuid.UUID, []byte, uuid.UUID, time.Time) (bool, error)
	Finish(context.Context, int64, uuid.UUID, uuid.UUID, OperationStatus) error
	Cancel(context.Context, int64, uuid.UUID) (bool, error)
	CancellationRequested(context.Context, int64, uuid.UUID, uuid.UUID) (bool, error)
}

func parseOperationKey(raw string) (uuid.UUID, error) {
	key, err := uuid.Parse(raw)
	if err != nil || key == uuid.Nil || key.String() != raw {
		return uuid.Nil, apperror.Validation("invalid idempotency key",
			apperror.NewFieldError("idempotency_key", apperror.FieldCodeInvalid, "expected a non-nil UUID in canonical form"))
	}
	return key, nil
}

func (uc *RunNestingUseCase) runOperation(ctx context.Context, input RunNestingInput, key uuid.UUID) (output RunNestingOutput, operationErr error) {
	if details := input.Validate(); len(details) > 0 {
		return output, apperror.Validation("validation failed", details...)
	}
	if uc.Operations == nil || uc.Transactions == nil {
		return output, apperror.Internal("nesting operations unavailable", nil)
	}
	if _, err := uc.projectSurfaceRepository.GetByID(ctx, input.UserID, input.ProjectID, input.ProjectSurfaceID); err != nil {
		return output, operationSurfaceError(err)
	}
	hash, err := Fingerprint(input)
	if err != nil {
		return output, apperror.Internal("encode nesting parameters", err)
	}
	token := uuid.New()
	deadline := time.Now().Add(ExecutionTimeout)
	if expires, ok := ctx.Deadline(); ok && expires.Before(deadline) {
		deadline = expires
	}
	owned, err := uc.Operations.Reserve(ctx, input.ProjectSurfaceID, key, hash[:], token, deadline)
	if err != nil {
		return output, apperror.Internal("reserve nesting operation", err)
	}
	if !owned {
		return uc.awaitOperation(ctx, input.ProjectSurfaceID, key, hash[:])
	}
	requestContext := ctx
	// Persist terminal failures using a separate bounded context after rollback.
	// A completed operation is never overwritten if COMMIT's response was lost.
	defer func() {
		if failure := recover(); failure != nil {
			operationErr = apperror.Internal("nesting execution panicked", fmt.Errorf("%v", failure))
		}
		if operationErr == nil {
			return
		}
		failure := terminalFailure(operationErr, requestContext)
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = uc.Operations.Finish(cleanup, input.ProjectSurfaceID, key, token, OperationStatus{Status: "failed", Failure: failure})
		_, saved, readErr := uc.Operations.Find(cleanup, input.ProjectSurfaceID, key)
		if readErr == nil && saved != nil && saved.Status == "completed" && saved.Result != nil {
			output = *saved.Result
			operationErr = nil
			return
		}
		if readErr == nil && saved != nil && saved.Failure != nil {
			failure = saved.Failure
		}
		operationErr = &apperror.AppError{Code: failure.Code, Message: failure.Message, Details: failure.Details}
	}()
	ctx, stop := uc.watchCancellation(ctx, input.ProjectSurfaceID, key, token)
	defer stop()
	release, err := uc.Limiter.Acquire(ctx)
	if err != nil {
		return output, err
	}
	defer release()
	output, operationErr = uc.executeAndCommit(ctx, input, func(tx context.Context, result RunNestingOutput) error {
		return uc.Operations.Finish(tx, input.ProjectSurfaceID, key, token, OperationStatus{Status: "completed", Result: &result})
	})
	if cause := context.Cause(ctx); operationErr != nil && cause != nil && !errors.Is(cause, context.Canceled) && !errors.Is(cause, context.DeadlineExceeded) {
		operationErr = cause
	}
	return output, operationErr
}

func terminalFailure(err error, ctx context.Context) *OperationFailure {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return &OperationFailure{Code: apperror.AppCodeTimeout, Message: "nesting execution timed out"}
	}
	if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
		return &OperationFailure{Code: apperror.AppCodeCancelled, Message: "nesting was cancelled"}
	}
	if failure, ok := apperror.As(err); ok {
		return &OperationFailure{Code: failure.Code, Message: failure.Message, Details: failure.Details}
	}
	return &OperationFailure{Code: apperror.AppCodeInternal, Message: "nesting execution failed"}
}

func (uc *RunNestingUseCase) awaitOperation(ctx context.Context, surfaceID int64, key uuid.UUID, hash []byte) (RunNestingOutput, error) {
	var release func()
	defer func() {
		if release != nil {
			release()
		}
	}()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		savedHash, saved, err := uc.Operations.Find(ctx, surfaceID, key)
		if err != nil {
			return RunNestingOutput{}, apperror.Internal("read nesting operation", err)
		}
		if saved == nil {
			return RunNestingOutput{}, apperror.NotFound("nesting operation not found")
		}
		if !bytes.Equal(hash, savedHash) {
			return RunNestingOutput{}, apperror.Conflict("idempotency key already used for different nesting parameters")
		}
		if saved.Failure != nil {
			return RunNestingOutput{}, &apperror.AppError{Code: saved.Failure.Code, Message: saved.Failure.Message, Details: saved.Failure.Details}
		}
		if saved.Status == "completed" && saved.Result != nil {
			return *saved.Result, nil
		}
		if release == nil {
			release, err = uc.Limiter.AwaitResponse(ctx)
			if err != nil {
				return RunNestingOutput{}, err
			}
		}
		select {
		case <-ctx.Done():
			return RunNestingOutput{}, ctx.Err()
		case <-ticker.C:
		}
	}
}

func (uc *RunNestingUseCase) watchCancellation(parent context.Context, surfaceID int64, key, token uuid.UUID) (context.Context, func()) {
	ctx, cancel := context.WithCancelCause(parent)
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				requested, err := uc.Operations.CancellationRequested(ctx, surfaceID, key, token)
				if err != nil {
					if ctx.Err() == nil {
						cancel(apperror.Internal("monitor nesting cancellation", err))
					}
					return
				}
				if requested {
					cancel(context.Canceled)
					return
				}
			}
		}
	}()
	return ctx, func() { cancel(context.Canceled); <-done }
}

func (uc *RunNestingUseCase) CancelOperation(ctx context.Context, userID, projectID, surfaceID int64, raw string) (bool, error) {
	key, err := parseOperationKey(raw)
	if err != nil {
		return false, err
	}
	if _, err := uc.projectSurfaceRepository.GetByID(ctx, userID, projectID, surfaceID); err != nil {
		return false, operationSurfaceError(err)
	}
	return uc.Operations.Cancel(ctx, surfaceID, key)
}

func Fingerprint(input RunNestingInput) ([32]byte, error) {
	input.AllowedRotations = NormalizeAllowedRotations(input.AllowedRotations)
	encoded, err := json.Marshal(input)
	return sha256.Sum256(encoded), err
}

type OperationStatus struct {
	Status  string
	Result  *RunNestingOutput
	Failure *OperationFailure
}

type OperationFailure struct {
	Code    apperror.Code
	Message string
	Details []apperror.FieldError
}

func (uc *RunNestingUseCase) GetOperation(ctx context.Context, userID, projectID, surfaceID int64, rawKey string) (OperationStatus, error) {
	key, err := parseOperationKey(rawKey)
	if err != nil {
		return OperationStatus{}, err
	}
	if _, err = uc.projectSurfaceRepository.GetByID(ctx, userID, projectID, surfaceID); err != nil {
		return OperationStatus{}, operationSurfaceError(err)
	}
	_, saved, err := uc.Operations.Find(ctx, surfaceID, key)
	if err != nil {
		return OperationStatus{}, apperror.Internal("read nesting operation", err)
	}
	if saved != nil {
		return *saved, nil
	}

	return OperationStatus{Status: "unknown"}, nil
}

func operationSurfaceError(err error) error {
	if errors.Is(err, domainprojectsurface.ErrNotFound) {
		return apperror.NotFound("project surface not found")
	}
	return apperror.Internal("read project surface", err)
}
