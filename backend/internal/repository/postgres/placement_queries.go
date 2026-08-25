package postgres

const createPlacementQuery = `
	INSERT INTO placements (
	    project_surface_id,
	    project_pattern_id,
	    x,
	    y,
	    rotation
	)
	SELECT
	    ps.id,
	    pp.id,
	    $4,
	    $5,
	    $6
	FROM project_surfaces ps
	JOIN project_patterns pp
	    ON pp.id = $3
	    AND pp.project_id = $1
	    AND pp.deleted_at IS NULL
	JOIN projects p
	    ON p.id = $1
	    AND p.user_id = $7
	    AND p.deleted_at IS NULL
	WHERE ps.id = $2
	    AND ps.project_id = $1
	    AND ps.deleted_at IS NULL
	RETURNING
		id,
		project_surface_id,
		project_pattern_id,
		x,
		y,
		rotation,
		created_at,
		updated_at
`

const listPlacementsForCollisionCheckQuery = `
	SELECT
		pl.id,
		ST_AsBinary(pp.geometry) AS pattern_geometry,
		pl.x,
		pl.y,
		pl.rotation
	FROM placements pl
	INNER JOIN project_surfaces ps
		ON ps.id = pl.project_surface_id
	INNER JOIN project_patterns pp
		ON pp.id = pl.project_pattern_id
	INNER JOIN projects p
		ON p.id = ps.project_id
	WHERE pl.project_surface_id = $1
		AND ps.project_id = $2
		AND pp.project_id = $2
		AND p.user_id = $3
		AND pl.deleted_at IS NULL
		AND ps.deleted_at IS NULL
		AND pp.deleted_at IS NULL
		AND p.deleted_at IS NULL
`

const getPlacementByIDQuery = `
	SELECT
	    pl.id,
	    pl.project_surface_id,
	    pl.project_pattern_id,
	    pl.x,
	    pl.y,
	    pl.rotation,
	    ST_AsBinary(pp.geometry) AS pattern_geometry,
	    pl.created_at,
	    pl.updated_at
	FROM placements pl
	INNER JOIN project_surfaces ps
	    ON ps.id = pl.project_surface_id
	INNER JOIN project_patterns pp
	    ON pp.id = pl.project_pattern_id
			AND pp.project_id = ps.project_id
	INNER JOIN projects p
	    ON p.id = ps.project_id
	WHERE pl.id = $1
	    AND p.id = $2
	    AND p.user_id = $3
	    AND pl.deleted_at IS NULL
	    AND ps.deleted_at IS NULL
	    AND pp.deleted_at IS NULL
	    AND p.deleted_at IS NULL
`

const listPlacementsQuery = `
	SELECT
	    pl.id,
	    pl.project_surface_id,
	    pl.project_pattern_id,
	    pl.x,
	    pl.y,
	    pl.rotation,
	    ST_AsBinary(pp.geometry) AS pattern_geometry,
	    pl.created_at,
	    pl.updated_at
	FROM placements pl
	INNER JOIN project_surfaces ps
	    ON ps.id = pl.project_surface_id
	INNER JOIN project_patterns pp
	    ON pp.id = pl.project_pattern_id
	        AND pp.project_id = ps.project_id
	INNER JOIN projects p
	    ON p.id = ps.project_id
	WHERE pl.project_surface_id = $1
	    AND p.id = $2
	    AND p.user_id = $3
	    AND pl.deleted_at IS NULL
	    AND ps.deleted_at IS NULL
	    AND pp.deleted_at IS NULL
	    AND p.deleted_at IS NULL
	ORDER BY pl.created_at ASC, pl.id ASC
`

const updatePlacementQuery = `
	UPDATE placements pl
	SET
	    x = $4,
	    y = $5,
	    rotation = $6,
	    updated_at = NOW()
	FROM project_surfaces ps,
	     project_patterns pp,
	     projects p
	WHERE pl.id = $1
	    AND pl.project_surface_id = ps.id
	    AND pl.project_pattern_id = pp.id
	    AND ps.project_id = $2
	    AND pp.project_id = $2
	    AND p.id = $2
	    AND p.user_id = $3
	    AND pl.deleted_at IS NULL
	    AND ps.deleted_at IS NULL
	    AND pp.deleted_at IS NULL
	    AND p.deleted_at IS NULL
	RETURNING
		pl.id,
	    pl.project_surface_id,
	    pl.project_pattern_id,
	    pl.x,
	    pl.y,
	    pl.rotation,
	    pl.created_at,
	    pl.updated_at
`

const listPlacementsForCollisionCheckExcludingQuery = `
	SELECT
		pl.id,
		pl.project_surface_id,
		pl.project_pattern_id,
		ST_AsBinary(pp.geometry) AS pattern_geometry,
		pl.x,
		pl.y,
		pl.rotation,
		pl.created_at,
		pl.updated_at
	FROM placements pl
	INNER JOIN project_surfaces ps
		ON ps.id = pl.project_surface_id
	INNER JOIN project_patterns pp
		ON pp.id = pl.project_pattern_id
		AND pp.project_id = ps.project_id
	INNER JOIN projects p
		ON p.id = ps.project_id
	WHERE pl.project_surface_id = $1
		AND p.id = $2
		AND p.user_id = $3
		AND pl.id <> $4
		AND pl.deleted_at IS NULL
		AND ps.deleted_at IS NULL
		AND pp.deleted_at IS NULL
		AND p.deleted_at IS NULL
`
