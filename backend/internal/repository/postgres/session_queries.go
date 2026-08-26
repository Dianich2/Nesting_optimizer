package postgres

const upsertSessionQuery = `
	INSERT INTO sessions (
		session_id, 
		user_id, 
		refresh_token_hash, 
		expires_at
	)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (user_id) DO UPDATE
	SET
		session_id = EXCLUDED.session_id,
		refresh_token_hash = EXCLUDED.refresh_token_hash,
		created_at = NOW(),
		expires_at = EXCLUDED.expires_at
	RETURNING 
		session_id, 
		user_id, 
		refresh_token_hash, 
		created_at, 
		expires_at
`

const rotateSessionQuery = `
	UPDATE sessions
	SET
		session_id = $1,
		refresh_token_hash = $2,
		created_at = NOW(),
		expires_at = $3
	WHERE session_id = $4
		AND refresh_token_hash = $5
	RETURNING 
		session_id, 
		user_id, 
		refresh_token_hash, 
		created_at, 
		expires_at
`

const getSessionBySessionIDQuery = `
	SELECT 
		session_id, 
		user_id, 
		refresh_token_hash, 
		created_at, 
		expires_at
	FROM sessions
	WHERE session_id = $1
`
const getSessionByUserIDQuery = `
	SELECT 
		session_id, 
		user_id, 
		refresh_token_hash, 
		created_at, 
		expires_at
	FROM sessions
	WHERE user_id = $1
`

const deleteSessionBySessionIDQuery = `
	DELETE FROM sessions
	WHERE session_id = $1
`

const deleteExpiredSessionsQuery = `
	DELETE FROM sessions
	WHERE expires_at <= NOW()
`

const deleteSessionByUserIDQuery = `
	DELETE FROM sessions
	WHERE user_id = $1
`
