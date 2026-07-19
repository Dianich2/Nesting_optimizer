package postgres

const existsUserByLoginQuery = `
	SELECT exists(
		SELECT 1 
		FROM users 
		WHERE login = $1
	)
`

const existsUserByEmailQuery = `
	SELECT exists(
		SELECT 1 
		FROM users 
		WHERE email = $1
	)
`

const createUserQuery = `
	INSERT INTO users(
		login, 
		email, 
		password_hash, 
		first_name, 
		last_name
	)
	VALUES($1, $2, $3, $4, $5)
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
