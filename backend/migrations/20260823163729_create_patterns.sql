-- +goose Up
CREATE TABLE patterns(
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    name VARCHAR(150) NOT NULL CHECK(LENGTH(trim(name)) > 0),
    geometry GEOMETRY(POLYGON) NOT NULL CHECK (NOT ST_IsEmpty(geometry)),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT patterns_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX idx_patterns_active_user_id ON patterns(user_id) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS patterns;

