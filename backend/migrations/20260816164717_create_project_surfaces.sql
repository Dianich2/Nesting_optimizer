-- +goose Up
CREATE TABLE project_surfaces(
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL,
    source_surface_id BIGINT,
    name VARCHAR(150) NOT NULL CHECK(LENGTH(trim(name)) > 0),
    geometry GEOMETRY(POLYGON) NOT NULL CHECK (NOT ST_IsEmpty(geometry)),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT project_surfaces_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    CONSTRAINT project_surfaces_source_surface_id_fkey FOREIGN KEY (source_surface_id) REFERENCES surfaces(id) ON DELETE SET NULL
);

CREATE INDEX idx_project_surfaces_active_project_id ON project_surfaces(project_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_project_surfaces_source_surface_id ON project_surfaces(source_surface_id);

-- +goose Down
DROP TABLE IF EXISTS project_surfaces;

