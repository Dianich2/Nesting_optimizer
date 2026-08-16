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
