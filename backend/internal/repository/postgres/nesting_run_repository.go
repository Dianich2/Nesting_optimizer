package postgres

import (
	"context"
	"fmt"
	nestingrun "server_nesting_optimizer/internal/domain/nesting_run"
	"time"
)

type NestingRunRepository struct {
	db DBTX
}

func NewNestingRunRepository(
	db DBTX,
) *NestingRunRepository {
	return &NestingRunRepository{
		db: db,
	}
}

func (r *NestingRunRepository) Create(
	ctx context.Context,
	input nestingrun.NestingRun,
) (nestingrun.NestingRun, error) {
	var nestingRun nestingrun.NestingRun

	if err := r.db.GetContext(
		ctx,
		&nestingRun,
		createNestingRun,
		input.ProjectSurfaceID,
		input.Algorithm,
		input.KeepExisting,
		input.RequestedCount,
		input.PlacedCount,
		input.SurfaceArea,
		input.PlacedArea,
		input.Utilization,
		input.Duration/time.Millisecond,
	); err != nil {
		return nestingrun.NestingRun{}, fmt.Errorf(
			"create nesting run: %w",
			err,
		)
	}

	nestingRun.Duration = time.Duration(nestingRun.Duration) * time.Millisecond

	return nestingRun, nil
}
