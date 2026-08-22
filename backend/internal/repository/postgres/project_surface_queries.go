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
		AND s.deleted_at IS NULL
	WHERE p.id = $2
		AND p.user_id = $1
		AND p.deleted_at IS NULL
	RETURNING
		id,
		project_id,
		source_surface_id,
		name,
		ST_AsBinary(geometry) AS geometry,
		created_at,
		updated_at
`
