-- +goose Up
CREATE TABLE sessions (
    session_id UUID, 
    user_id BIGINT NOT NULL,
    refresh_token_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,

    CONSTRAINT sessions_pkey PRIMARY KEY (session_id),

    CONSTRAINT sessions_user_id_unique UNIQUE (user_id),

    CONSTRAINT sessions_refresh_token_hash_unique UNIQUE (refresh_token_hash),

    CONSTRAINT sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS sessions;
