package postgres

import (
	"context"
	"github.com/lib/pq"
	nestingcase "server_nesting_optimizer/internal/usecase/nesting"
	"server_nesting_optimizer/pkg/apperror"
)

// Preflight counts vertices in PostGIS before materializing repeated placement
// geometries in Go. It runs under the same short source-snapshot transaction.
func (s *ScopedDB) CheckNestingBudget(ctx context.Context, input nestingcase.RunNestingInput) error {
	ids := make([]int64, 0, len(input.Patterns))
	for _, pattern := range input.Patterns {
		ids = append(ids, pattern.ProjectPatternID)
	}
	var row struct {
		Vertices   int64 `db:"vertices"`
		Largest    int   `db:"largest"`
		Holes      int   `db:"holes"`
		Placements int   `db:"placements"`
	}
	err := s.GetContext(ctx, &row, `
        WITH source AS (
            SELECT ps.geometry,FALSE AS placement FROM project_surfaces ps
            JOIN projects p ON p.id=ps.project_id WHERE p.user_id=$1 AND p.id=$2 AND ps.id=$3
            UNION ALL
            SELECT pp.geometry,FALSE FROM project_patterns pp
            JOIN projects p ON p.id=pp.project_id WHERE p.user_id=$1 AND p.id=$2 AND pp.id=ANY($4)
            UNION ALL
            SELECT pp.geometry,TRUE FROM placements pl
            JOIN project_surfaces ps ON ps.id=pl.project_surface_id
            JOIN projects p ON p.id=ps.project_id
            JOIN project_patterns pp ON pp.id=pl.project_pattern_id
            WHERE p.user_id=$1 AND p.id=$2 AND ps.id=$3
        )
        SELECT coalesce(sum(ST_NPoints(geometry)),0) AS vertices,
            coalesce(max(ST_NPoints(geometry)-ST_NumInteriorRings(geometry)-1),0) AS largest,
            coalesce(max(ST_NumInteriorRings(geometry)),0) AS holes,
            count(*) FILTER(WHERE placement) AS placements FROM source`, input.UserID, input.ProjectID, input.ProjectSurfaceID, pq.Array(ids))
	if err != nil {
		return apperror.Internal("check nesting source complexity", err)
	}
	// ST_NPoints includes an explicit closing vertex per ring.
	if row.Vertices > 100000 || row.Largest > 4096 || row.Holes > 64 || row.Placements > 2000 {
		return apperror.Validation("nesting source complexity budget exceeded", apperror.NewFieldError("geometry", apperror.FieldCodeTooLong, "source snapshot exceeds 100000 vertices, 4096 open-ring vertices per polygon, 64 holes or 2000 existing placements"))
	}
	return nil
}
