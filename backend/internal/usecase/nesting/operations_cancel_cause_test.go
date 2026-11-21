package nesting

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	domainsurface "server_nesting_optimizer/internal/domain/project_surface"
	optimizer "server_nesting_optimizer/internal/nesting"
	"server_nesting_optimizer/pkg/apperror"
)

type monitorFailureOperationRepo struct {
	terminal OperationStatus
}

func (r *monitorFailureOperationRepo) Reserve(context.Context, int64, uuid.UUID, []byte, uuid.UUID, time.Time) (bool, error) {
	return true, nil
}

func (r *monitorFailureOperationRepo) Find(context.Context, int64, uuid.UUID) ([]byte, *OperationStatus, error) {
	if r.terminal.Status == "" {
		return nil, nil, nil
	}
	saved := r.terminal
	return nil, &saved, nil
}

func (r *monitorFailureOperationRepo) Finish(_ context.Context, _ int64, _, _ uuid.UUID, status OperationStatus) error {
	r.terminal = status
	return nil
}

func (r *monitorFailureOperationRepo) Cancel(context.Context, int64, uuid.UUID) (bool, error) {
	return false, nil
}

func (r *monitorFailureOperationRepo) CancellationRequested(context.Context, int64, uuid.UUID, uuid.UUID) (bool, error) {
	return false, errors.New("simulated cancellation monitor query failure")
}

type monitorFailureSurfaceRepo struct{}

func (monitorFailureSurfaceRepo) GetByID(context.Context, int64, int64, int64) (domainsurface.ProjectSurface, error) {
	return domainsurface.ProjectSurface{}, nil
}

type monitorFailureProjectTransaction struct{}

func (monitorFailureProjectTransaction) WithinProject(ctx context.Context, _, _ int64, fn func(context.Context) error) error {
	return fn(ctx)
}

func TestRunCoordinatedPreservesMonitorFailureWhileQueued(t *testing.T) {
	limiter := NewComputeLimiter(1, 1)
	occupy, err := limiter.Acquire(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer occupy()

	operations := &monitorFailureOperationRepo{}
	uc := &RunNestingUseCase{
		Operations:               operations,
		Limiter:                  limiter,
		Transactions:             monitorFailureProjectTransaction{},
		projectSurfaceRepository: monitorFailureSurfaceRepo{},
	}
	input := RunNestingInput{
		UserID:           1,
		ProjectID:        1,
		ProjectSurfaceID: 1,
		Algorithm:        optimizer.BaselineAlgorithm,
		Patterns:         []RunNestingPatternInput{{ProjectPatternID: 1, Quantity: 1}},
		AllowedRotations: []float64{0},
	}
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	_, err = uc.runCoordinated(ctx, input, uuid.New(), func(context.Context, RunNestingInput, func(context.Context, RunNestingOutput) error) (RunNestingOutput, error) {
		t.Fatal("optimizer must not execute while the worker slot is occupied")
		return RunNestingOutput{}, nil
	})
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.AppCodeInternal {
		t.Fatalf("monitor failure was misreported as cancellation: %v", err)
	}
	if operations.terminal.Status != "failed" || operations.terminal.Failure == nil || operations.terminal.Failure.Code != apperror.AppCodeInternal {
		t.Fatalf("incorrect durable terminal result: %+v", operations.terminal)
	}
}
