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

const listProjectsQuery = `
	WITH filtered AS(
		SELECT
			id,
			user_id,
			name,
			description,
			created_at,
			updated_at,
			COUNT(*) OVER() as total
		FROM projects
		WHERE user_id = $1
			AND deleted_at IS NULL
	),
	paged AS(
		SELECT 
			id,
			user_id,
			name,
			description,
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
		p.description,
		p.created_at,
		p.updated_at,
		COALESCE(p.total, m.total) AS total
	FROM meta m LEFT JOIN paged p ON TRUE
	ORDER BY 
		p.updated_at DESC NULLS LAST,
		p.id DESC NULLS LAST
`

const updateProjectQuery = `
	UPDATE projects
	SET
		name = COALESCE($1, name),
		description = COALESCE($2, description),
		updated_at = NOW()
	WHERE id = $3
		AND user_id = $4
		AND deleted_at IS NULL
	RETURNING 
		id, 
		user_id,
		name,
		description,
		created_at, 
		updated_at
`
