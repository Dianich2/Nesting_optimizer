package postgres

const createProjectPatternQuery = `
	INSERT INTO project_patterns (
		project_id,
		source_pattern_id,
		name,
		geometry
	)
	SELECT
		p.id,
		pa.id,
		$4,
		ST_GeomFromWKB($5)
	FROM projects p
	JOIN patterns pa
		ON pa.id = $3
		AND pa.user_id = $1
		AND pa.deleted_at IS NULL
	WHERE p.id = $2
		AND p.user_id = $1
		AND p.deleted_at IS NULL
	RETURNING
		id,
		project_id,
		source_pattern_id,
		name,
		ST_AsBinary(geometry) AS geometry,
		created_at,
		updated_at
`

const getProjectPatternByIDQuery = `
	SELECT
		pp.id,
		pp.project_id,
		pp.source_pattern_id,
		pp.name,
		ST_AsBinary(pp.geometry) AS geometry,
		pp.created_at,
		pp.updated_at
	FROM project_patterns pp
	INNER JOIN projects p
		ON p.id = pp.project_id
	WHERE pp.id = $1
		AND pp.project_id = $2
		AND p.user_id = $3
		AND pp.deleted_at IS NULL
		AND p.deleted_at IS NULL
`

const listProjectPatternsQuery = `
	WITH accessible_project AS (
		SELECT id
		FROM projects
		WHERE id = $1
			AND user_id = $2
			AND deleted_at IS NULL
	),
	filtered AS(
		SELECT
			pp.id,
    		pp.project_id,
    		pp.source_pattern_id,
    		pp.name,
    		ST_AsBinary(pp.geometry) AS geometry,
    		pp.created_at,
    		pp.updated_at,
			COUNT(*) OVER() as total
		FROM project_patterns pp
		INNER JOIN accessible_project p
		    ON p.id = pp.project_id
		WHERE pp.deleted_at IS NULL
	),
	paged AS(
		SELECT 
			id,
			project_id,
			source_pattern_id,
			name,
			geometry,
			created_at,
			updated_at,
			total
		FROM filtered
		ORDER BY updated_at DESC, id DESC
		LIMIT $3
		OFFSET $4
	),
	meta AS (
		SELECT
			EXISTS (
				SELECT 1
				FROM accessible_project
			) AS project_exists,
			COALESCE(MAX(total), 0) AS total
		FROM filtered
	)
	SELECT
		p.id,
		p.project_id,
		p.source_pattern_id,
		p.name,
		p.geometry,
		p.created_at,
		p.updated_at,
		COALESCE(p.total, m.total) AS total,
		m.project_exists
	FROM meta m LEFT JOIN paged p ON TRUE
	ORDER BY 
		p.updated_at DESC NULLS LAST,
		p.id DESC NULLS LAST
`

const updateProjectPatternQuery = `
	UPDATE project_patterns pp
	SET
		name = COALESCE($4, pp.name),
		geometry = COALESCE(ST_GeomFromWKB($5), pp.geometry),
		updated_at = NOW()
	FROM projects p
	WHERE pp.id = $1
		AND pp.project_id = $2
		AND p.id = pp.project_id
		AND p.user_id = $3
		AND pp.deleted_at IS NULL
		AND p.deleted_at IS NULL
	RETURNING
		id,
		project_id,
		source_pattern_id,
		name,
		ST_AsBinary(geometry) AS geometry,
		created_at,
		updated_at
`

const softDeleteProjectPatternQuery = `
	UPDATE project_patterns pp
	SET 
		deleted_at = NOW(),
		updated_at = NOW()
	FROM projects p
	WHERE pp.id = $1
		AND pp.project_id = $2
		AND p.id = pp.project_id
		AND p.user_id = $3
		AND pp.deleted_at IS NULL
		AND p.deleted_at IS NULL
`

const hasActivePlacementsByProjectPatternIDQuery = `
	SELECT EXISTS (
		SELECT 1
		FROM placements pl
		INNER JOIN project_patterns pp
			ON pp.id = pl.project_pattern_id
		INNER JOIN project_surfaces ps
			ON ps.id = pl.project_surface_id
			AND ps.project_id = pp.project_id
		INNER JOIN projects p
			ON p.id = pp.project_id
		WHERE pp.id = $1
			AND pp.project_id = $2
			AND p.user_id = $3
			AND pl.deleted_at IS NULL
			AND pp.deleted_at IS NULL
			AND ps.deleted_at IS NULL
			AND p.deleted_at IS NULL
	)
`
