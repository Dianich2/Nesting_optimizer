-- +goose Up
CREATE TABLE project_patterns(
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL,
    source_pattern_id BIGINT,
    name VARCHAR(150) NOT NULL CHECK(LENGTH(trim(name)) > 0),
    geometry GEOMETRY(POLYGON) NOT NULL CHECK (NOT ST_IsEmpty(geometry)),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT project_patterns_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    CONSTRAINT project_patterns_source_pattern_id_fkey FOREIGN KEY (source_pattern_id) REFERENCES patterns(id) ON DELETE SET NULL
);

CREATE INDEX idx_project_patterns_active_project_id ON project_patterns(project_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_project_patterns_source_pattern_id ON project_patterns(source_pattern_id);

-- +goose Down
DROP TABLE IF EXISTS project_patterns;

