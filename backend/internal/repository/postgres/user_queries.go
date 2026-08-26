package postgres

const existsUserByLoginQuery = `
	SELECT EXISTS (
		SELECT 1 
		FROM users 
		WHERE login = $1
	)
`

const existsUserByEmailQuery = `
	SELECT EXISTS (
		SELECT 1 
		FROM users 
		WHERE email = $1
	)
`

const createUserQuery = `
	INSERT INTO users (
		login, 
		email, 
		password_hash, 
		first_name, 
		last_name
	)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING 
		id, 
		login, 
		email, 
		first_name, 
		last_name, 
		created_at, 
		updated_at
`

const getByIdentifierQuery = `
	SELECT 
		id, 
		login, 
		email, 
		password_hash, 
		first_name, 
		last_name, 
		created_at, 
		updated_at, 
		deleted_at
	FROM users
	WHERE (login = $1 OR email = $1)
		AND deleted_at IS NULL
	LIMIT 1
`

const getByIDQuery = `
	SELECT 
		id, 
		login, 
		email,  
		first_name, 
		last_name, 
		password_hash,
		created_at, 
		updated_at, 
		deleted_at
	FROM users
	WHERE id = $1
		AND deleted_at IS NULL
`

const updateProfileQuery = `
	UPDATE users
	SET
		first_name = COALESCE($1, first_name),
		last_name = COALESCE($2, last_name),
		updated_at = NOW()
	WHERE id = $3
		AND deleted_at IS NULL
	RETURNING
		id, 
		login, 
		email,  
		first_name, 
		last_name, 
		created_at, 
		updated_at, 
		deleted_at
`

const updatePasswordQuery = `
	UPDATE users
	SET
		password_hash = $1,
		updated_at = NOW()
	WHERE 
		password_hash = $2
			AND id = $3
			AND deleted_at IS NULL
`

const softDeleteUserQuery = `
	UPDATE users
	SET
		deleted_at = NOW(),
		updated_at = NOW()
	WHERE id = $1
		AND password_hash = $2
		AND deleted_at IS NULL
`
