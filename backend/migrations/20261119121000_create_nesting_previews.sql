-- +goose Up
CREATE TABLE nesting_previews (
    id UUID PRIMARY KEY,
    operation_key UUID,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    project_surface_id BIGINT NOT NULL REFERENCES project_surfaces(id) ON DELETE CASCADE,
    payload JSONB NOT NULL CHECK (octet_length(payload::text) <= 8388608),
    saved_result JSONB CHECK (saved_result IS NULL OR octet_length(saved_result::text) <= 16777216),
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    expires_at TIMESTAMPTZ NOT NULL
);
CREATE UNIQUE INDEX nesting_previews_operation ON nesting_previews(user_id,project_surface_id,operation_key) WHERE operation_key IS NOT NULL;
CREATE INDEX nesting_previews_expiry ON nesting_previews(expires_at);
CREATE INDEX nesting_previews_user_expiry ON nesting_previews(user_id, expires_at);

-- +goose Down
DROP TABLE nesting_previews;
