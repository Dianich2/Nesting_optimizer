-- +goose Up
CREATE TABLE nesting_preview_operations (
 project_surface_id BIGINT NOT NULL REFERENCES project_surfaces(id) ON DELETE CASCADE,
 operation_key UUID NOT NULL,
 request_hash BYTEA NOT NULL CHECK(octet_length(request_hash)=32),
 response JSONB NOT NULL CHECK(octet_length(response::text)<=16777216),
 state TEXT NOT NULL CHECK(state IN ('running','completed','failed')),
 lease_token UUID NOT NULL,
 lease_expires_at TIMESTAMPTZ NOT NULL,
 cancel_requested BOOLEAN NOT NULL DEFAULT FALSE,
 PRIMARY KEY(project_surface_id,operation_key)
);
CREATE INDEX nesting_preview_operations_running_expiry ON nesting_preview_operations(lease_expires_at) WHERE state='running';
-- +goose Down
DROP TABLE nesting_preview_operations;
