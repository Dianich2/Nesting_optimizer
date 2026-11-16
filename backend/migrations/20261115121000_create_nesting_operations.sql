-- +goose Up
CREATE TABLE nesting_operations (
    project_surface_id BIGINT NOT NULL REFERENCES project_surfaces(id) ON DELETE CASCADE,
    operation_key UUID NOT NULL,
    request_hash BYTEA NOT NULL CHECK (octet_length(request_hash) = 32),
    response JSONB NOT NULL,
    PRIMARY KEY (project_surface_id, operation_key)
);

-- +goose Down
DROP TABLE nesting_operations;
