package postgres

const createNestingRun = `
    INSERT INTO nesting_runs (
        project_surface_id, algorithm, keep_existing,
        requested_count, placed_count, surface_area, placed_area,
        utilization, duration_ms, parameters
    )
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULLIF($10,'')::jsonb)
    RETURNING
        id, project_surface_id, algorithm, keep_existing,
        requested_count, placed_count, surface_area, placed_area,
        utilization, duration_ms, created_at, parameters
`
