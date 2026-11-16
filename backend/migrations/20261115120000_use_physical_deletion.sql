-- +goose Up
-- Run with the old application stopped. Cascades purge dependent records.
ALTER TABLE surfaces DROP CONSTRAINT surfaces_user_id_fkey,
    ADD CONSTRAINT surfaces_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE patterns DROP CONSTRAINT patterns_user_id_fkey,
    ADD CONSTRAINT patterns_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

DELETE FROM users WHERE deleted_at IS NOT NULL;
DELETE FROM projects WHERE deleted_at IS NOT NULL;
DELETE FROM surfaces WHERE deleted_at IS NOT NULL;
DELETE FROM patterns WHERE deleted_at IS NOT NULL;
DELETE FROM project_surfaces WHERE deleted_at IS NOT NULL;
DELETE FROM project_patterns WHERE deleted_at IS NOT NULL;
DELETE FROM placements WHERE deleted_at IS NOT NULL;

ALTER TABLE users DROP COLUMN deleted_at;
ALTER TABLE projects DROP COLUMN deleted_at;
ALTER TABLE surfaces DROP COLUMN deleted_at;
ALTER TABLE patterns DROP COLUMN deleted_at;
ALTER TABLE project_surfaces DROP COLUMN deleted_at;
ALTER TABLE project_patterns DROP COLUMN deleted_at;
ALTER TABLE placements DROP COLUMN deleted_at;
CREATE INDEX idx_projects_by_user_updated_at ON projects (user_id, updated_at DESC, id DESC);
CREATE INDEX idx_surfaces_user_id ON surfaces (user_id);
CREATE INDEX idx_project_surfaces_project_id ON project_surfaces (project_id);
CREATE INDEX idx_patterns_user_id ON patterns (user_id);
CREATE INDEX idx_project_patterns_project_id ON project_patterns (project_id);
CREATE INDEX idx_placements_project_surface_id ON placements (project_surface_id);
CREATE INDEX idx_placements_project_pattern_id ON placements (project_pattern_id);

-- +goose Down
-- Restore the old schema only. Physically removed rows require a backup.
ALTER TABLE surfaces DROP CONSTRAINT surfaces_user_id_fkey,
    ADD CONSTRAINT surfaces_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE patterns DROP CONSTRAINT patterns_user_id_fkey,
    ADD CONSTRAINT patterns_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE users ADD COLUMN deleted_at TIMESTAMPTZ;
ALTER TABLE projects ADD COLUMN deleted_at TIMESTAMPTZ;
ALTER TABLE surfaces ADD COLUMN deleted_at TIMESTAMPTZ;
ALTER TABLE patterns ADD COLUMN deleted_at TIMESTAMPTZ;
ALTER TABLE project_surfaces ADD COLUMN deleted_at TIMESTAMPTZ;
ALTER TABLE project_patterns ADD COLUMN deleted_at TIMESTAMPTZ;
ALTER TABLE placements ADD COLUMN deleted_at TIMESTAMPTZ;
DROP INDEX idx_projects_by_user_updated_at;
CREATE INDEX idx_projects_active_by_user_updated_at ON projects (user_id, updated_at DESC, id DESC) WHERE deleted_at IS NULL;
DROP INDEX idx_surfaces_user_id;
CREATE INDEX idx_surfaces_active_user_id ON surfaces (user_id) WHERE deleted_at IS NULL;
DROP INDEX idx_project_surfaces_project_id;
CREATE INDEX idx_project_surfaces_active_project_id ON project_surfaces (project_id) WHERE deleted_at IS NULL;
DROP INDEX idx_patterns_user_id;
CREATE INDEX idx_patterns_active_user_id ON patterns (user_id) WHERE deleted_at IS NULL;
DROP INDEX idx_project_patterns_project_id;
CREATE INDEX idx_project_patterns_active_project_id ON project_patterns (project_id) WHERE deleted_at IS NULL;
DROP INDEX idx_placements_project_surface_id;
CREATE INDEX idx_placements_active_project_surface_id ON placements (project_surface_id) WHERE deleted_at IS NULL;
DROP INDEX idx_placements_project_pattern_id;
CREATE INDEX idx_placements_active_project_pattern_id ON placements (project_pattern_id) WHERE deleted_at IS NULL;
