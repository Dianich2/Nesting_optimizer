-- +goose Up
CREATE TABLE nesting_runs (
    id BIGSERIAL PRIMARY KEY,
    project_surface_id BIGINT NOT NULL,
    algorithm VARCHAR(50) NOT NULL,
    keep_existing BOOLEAN NOT NULL,
    requested_count INT NOT NULL CHECK(requested_count >= 0),
    placed_count INT NOT NULL CHECK(placed_count >= 0 AND placed_count <= requested_count),
    surface_area DOUBLE PRECISION NOT NULL CHECK(surface_area > 0),
    placed_area DOUBLE PRECISION NOT NULL CHECK(placed_area >= 0),
    utilization DOUBLE PRECISION NOT NULL CHECK(utilization >= 0 AND utilization <= 1),
    duration_ms BIGINT NOT NULL CHECK(duration_ms >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT nesting_runs_project_surface_id_fkey FOREIGN KEY (project_surface_id) REFERENCES project_surfaces(id) ON DELETE CASCADE
);

CREATE INDEX idx_nesting_runs_project_surface_id
ON nesting_runs (project_surface_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS nesting_runs;