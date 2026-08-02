package postgres

const createProjectQuery = `
	INSERT INTO projects(
		user_id,
		name,
		description
	)
	SELECT $1, $2, $3
	FROM users u
	WHERE u.id = $1
		AND u.deleted_at IS NULL
	RETURNING 
		id, 
		user_id,
		name,
		description,
		created_at, 
		updated_at
`

const getProjectByIDQuery = `
	SELECT
		id,
		user_id,
		name,
		description,
		created_at,
		updated_at,
		deleted_at
	FROM projects
	WHERE id = $1 
		AND user_id = $2
		AND deleted_at IS NULL
`
