package postgres

const createProjectSurfaceQuery = `
	INSERT INTO project_surfaces (
		project_id,
		source_surface_id,
		name,
		geometry
	)
	SELECT
		p.id,
		s.id,
		$4,
		ST_GeomFromWKB($5)
	FROM projects p
	JOIN surfaces s
		ON s.id = $3
		AND s.user_id = $1
	WHERE p.id = $2
		AND p.user_id = $1
	FOR KEY SHARE OF s
	RETURNING
		id,
		project_id,
		source_surface_id,
		name,
		ST_AsBinary(geometry) AS geometry,
		created_at,
		updated_at
`

const getProjectSurfaceByIDQuery = `
	SELECT
		ps.id,
		ps.project_id,
		ps.source_surface_id,
		ps.name,
		ST_AsBinary(ps.geometry) AS geometry,
		ps.created_at,
		ps.updated_at
	FROM project_surfaces ps
	INNER JOIN projects p
		ON p.id = ps.project_id
	WHERE ps.id = $1
		AND ps.project_id = $2
		AND p.user_id = $3
`

const listProjectSurfacesQuery = `
	WITH accessible_project AS (
		SELECT id
		FROM projects
		WHERE id = $1
			AND user_id = $2
	),
	filtered AS (
		SELECT
			ps.id,
    		ps.project_id,
    		ps.source_surface_id,
    		ps.name,
    		ST_AsBinary(ps.geometry) AS geometry,
    		ps.created_at,
    		ps.updated_at,
			COUNT(*) OVER() AS total
		FROM project_surfaces ps
		INNER JOIN accessible_project p
		    ON p.id = ps.project_id
	),
	paged AS (
		SELECT 
			id,
			project_id,
			source_surface_id,
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
		p.source_surface_id,
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

const updateProjectSurfaceQuery = `
	UPDATE project_surfaces ps
	SET
		name = COALESCE($4, ps.name),
		geometry = COALESCE(ST_GeomFromWKB($5), ps.geometry),
		updated_at = NOW()
	FROM projects p
	WHERE ps.id = $1
		AND ps.project_id = $2
		AND p.id = ps.project_id
		AND p.user_id = $3
	RETURNING
		ps.id,
		ps.project_id,
		ps.source_surface_id,
		ps.name,
		ST_AsBinary(ps.geometry) AS geometry,
		ps.created_at,
		ps.updated_at
`

const deleteProjectSurfaceQuery = `
	DELETE FROM project_surfaces ps
	USING projects p
	WHERE ps.id = $1
		AND ps.project_id = $2
		AND p.id = ps.project_id
		AND p.user_id = $3
`

const hasActivePlacementsByProjectSurfaceIDQuery = `
	SELECT EXISTS (
		SELECT 1
		FROM placements pl
		INNER JOIN project_surfaces ps
			ON ps.id = pl.project_surface_id
		INNER JOIN project_patterns pp
			ON pp.id = pl.project_pattern_id
			AND pp.project_id = ps.project_id
		INNER JOIN projects p
			ON p.id = ps.project_id
		WHERE ps.id = $1
			AND ps.project_id = $2
			AND p.user_id = $3
	)
`
