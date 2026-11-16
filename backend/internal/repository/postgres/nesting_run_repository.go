package postgres

import (
	"context"
	"fmt"
	"time"

	nestingrun "server_nesting_optimizer/internal/domain/nesting_run"
	"server_nesting_optimizer/internal/nesting"
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

type NestingRunRow struct {
	ID               int64     `db:"id"`
	ProjectSurfaceID int64     `db:"project_surface_id"`
	Algorithm        string    `db:"algorithm"`
	KeepExisting     bool      `db:"keep_existing"`
	RequestedCount   int       `db:"requested_count"`
	PlacedCount      int       `db:"placed_count"`
	SurfaceArea      float64   `db:"surface_area"`
	PlacedArea       float64   `db:"placed_area"`
	Utilization      float64   `db:"utilization"`
	DurationMS       int64     `db:"duration_ms"`
	CreatedAt        time.Time `db:"created_at"`
}

func (r *NestingRunRepository) Create(
	ctx context.Context,
	input nestingrun.NestingRun,
) (nestingrun.NestingRun, error) {
	var row NestingRunRow

	if err := r.db.GetContext(
		ctx,
		&row,
		createNestingRun,
		input.ProjectSurfaceID,
		input.Algorithm,
		input.KeepExisting,
		input.RequestedCount,
		input.PlacedCount,
		input.SurfaceArea,
		input.PlacedArea,
		input.Utilization,
		input.Duration.Milliseconds(),
	); err != nil {
		return nestingrun.NestingRun{}, fmt.Errorf(
			"create nesting run: %w",
			err,
		)
	}

	return nestingRunRowToDomain(row), nil
}

func nestingRunRowToDomain(
	row NestingRunRow,
) nestingrun.NestingRun {
	return nestingrun.NestingRun{
		ID:               row.ID,
		ProjectSurfaceID: row.ProjectSurfaceID,
		Algorithm:        nesting.Algorithm(row.Algorithm),
		KeepExisting:     row.KeepExisting,
		RequestedCount:   row.RequestedCount,
		PlacedCount:      row.PlacedCount,
		SurfaceArea:      row.SurfaceArea,
		PlacedArea:       row.PlacedArea,
		Utilization:      row.Utilization,
		Duration:         time.Duration(row.DurationMS) * time.Millisecond,
		CreatedAt:        row.CreatedAt,
	}
}

func (r *NestingRunRepository) List(ctx context.Context, userID, projectID, surfaceID int64, limit, offset int) ([]nestingrun.NestingRun, error) {
	rows := make([]NestingRunRow, 0)
	err := r.db.SelectContext(ctx, &rows, `
 SELECT n.id,n.project_surface_id,n.algorithm,n.keep_existing,n.requested_count,n.placed_count,
        n.surface_area,n.placed_area,n.utilization,n.duration_ms,n.created_at
 FROM nesting_runs n
 JOIN project_surfaces s ON s.id=n.project_surface_id
 JOIN projects p ON p.id=s.project_id
 WHERE p.user_id=$1 AND p.id=$2 AND s.id=$3
 ORDER BY n.created_at DESC, n.id DESC LIMIT $4 OFFSET $5`, userID, projectID, surfaceID, limit, offset)
	if err != nil {
		return nil, err
	}
	result := make([]nestingrun.NestingRun, 0, len(rows))
	for _, row := range rows {
		result = append(result, nestingRunRowToDomain(row))
	}
	return result, nil
}
