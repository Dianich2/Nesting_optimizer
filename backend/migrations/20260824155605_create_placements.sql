-- +goose Up
CREATE TABLE placements(
    id BIGSERIAL PRIMARY KEY,
    project_surface_id BIGINT NOT NULL,
    project_pattern_id BIGINT NOT NULL,
    x DOUBLE PRECISION NOT NULL,
    y DOUBLE PRECISION NOT NULL,
    rotation DOUBLE PRECISION NOT NULL CHECK (rotation >= 0 AND rotation < 360),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT placements_project_surface_id_fkey FOREIGN KEY (project_surface_id) REFERENCES project_surfaces(id) ON DELETE CASCADE,
    CONSTRAINT placements_project_pattern_id_fkey FOREIGN KEY (project_pattern_id) REFERENCES project_patterns(id) ON DELETE CASCADE
);

CREATE INDEX idx_placements_active_project_surface_id
ON placements(project_surface_id)
WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS placements;