-- +goose Up
ALTER TABLE nesting_runs ADD COLUMN parameters JSONB;
ALTER TABLE nesting_operations
    ADD COLUMN state TEXT NOT NULL DEFAULT 'completed',
    ADD COLUMN lease_token UUID,
    ADD COLUMN lease_expires_at TIMESTAMPTZ,
    ADD COLUMN cancel_requested BOOLEAN NOT NULL DEFAULT FALSE;
UPDATE nesting_operations SET state = 'failed' WHERE response->>'Status' = 'failed';
ALTER TABLE nesting_operations ADD CONSTRAINT nesting_operations_state_check CHECK (state IN ('running','completed','failed'));

-- +goose Down
ALTER TABLE nesting_runs DROP COLUMN parameters;
UPDATE nesting_operations SET response='{"Status":"failed","Failure":{"Code":"internal_error","Message":"operation interrupted by schema rollback"}}'::jsonb WHERE state='running';
ALTER TABLE nesting_operations DROP CONSTRAINT nesting_operations_state_check,
    DROP COLUMN state, DROP COLUMN lease_token, DROP COLUMN lease_expires_at, DROP COLUMN cancel_requested;
