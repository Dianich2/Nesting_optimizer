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
