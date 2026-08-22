package postgres

const createSurfaceQuery = `
	INSERT INTO surfaces (
		user_id,
		name,
		geometry
	)
	SELECT
		u.id,
		$2,
		ST_GeomFromWKB($3)
	FROM users u
	WHERE u.id = $1
	  AND u.deleted_at IS NULL
	RETURNING
		id,
		user_id,
		name,
		ST_AsBinary(geometry) AS geometry,
		created_at,
		updated_at,
		deleted_at
`

const getSurfaceByIDQuery = `
	SELECT
		id,
		user_id,
		name,
		ST_AsBinary(geometry) AS geometry,
		created_at,
		updated_at,
		deleted_at
	FROM surfaces
	WHERE id = $1
	  AND user_id = $2
	  AND deleted_at IS NULL
`

const listSurfacesQuery = `
	WITH filtered AS(
		SELECT
			id,
			user_id,
			name,
			geometry,
			created_at,
			updated_at,
			COUNT(*) OVER() as total
		FROM surfaces
		WHERE user_id = $1
			AND deleted_at IS NULL
	),
	paged AS(
		SELECT 
			id,
			user_id,
			name,
			geometry,
			created_at,
			updated_at,
			total
		FROM filtered
		ORDER BY updated_at DESC, id DESC
		LIMIT $2
		OFFSET $3
	),
	meta AS (
		SELECT COALESCE(MAX(total), 0) AS total
		FROM filtered
	)
	SELECT
		p.id,
		p.user_id,
		p.name,
		ST_AsBinary(p.geometry) AS geometry,
		p.created_at,
		p.updated_at,
		COALESCE(p.total, m.total) AS total
	FROM meta m LEFT JOIN paged p ON TRUE
	ORDER BY 
		p.updated_at DESC NULLS LAST,
		p.id DESC NULLS LAST
`

const updateSurfaceQuery = `
	UPDATE surfaces
	SET
		name = COALESCE($3, name),
		geometry = COALESCE(ST_GeomFromWKB($4), geometry),
		updated_at = NOW()
	WHERE id = $1
	  AND user_id = $2
	  AND deleted_at IS NULL
	RETURNING
		id,
		user_id,
		name,
		ST_AsBinary(geometry) AS geometry,
		created_at,
		updated_at,
		deleted_at
`
