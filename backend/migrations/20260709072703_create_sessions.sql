-- +goose Up
create table sessions(
    session_id uuid primary key, 
    user_id bigint not null,
    refresh_token_hash text not null,
    created_at timestamptz not null default now(),
    expires_at timestamptz not null,

    CONSTRAINT sessions_pkey PRIMARY KEY (session_id),

    CONSTRAINT sessions_user_id_unique UNIQUE (user_id),

    CONSTRAINT sessions_refresh_token_hash_unique UNIQUE (refresh_token_hash),

    CONSTRAINT sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
drop table if exists sessions;
