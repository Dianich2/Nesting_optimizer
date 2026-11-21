package postgres

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"time"
)

// ValidateSchema is read-only and never migrates. Check the active schema rather
// than accepting a similarly named table in another schema on search_path.
func ValidateSchema(ctx context.Context, db *sqlx.DB) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var ready bool
	err := db.GetContext(ctx, &ready, `WITH required(table_name,column_name) AS (VALUES
 ('users','id'),('sessions','session_id'),('projects','id'),('surfaces','geometry'),('patterns','geometry'),
 ('project_surfaces','geometry'),('project_patterns','geometry'),('placements','rotation'),('nesting_runs','parameters'),
 ('nesting_operations','state'),('nesting_operations','lease_token'),('nesting_operations','cancel_requested'),
 ('nesting_previews','payload'),('nesting_previews','saved_result'),('nesting_preview_operations','state'),('nesting_preview_operations','lease_expires_at'))
 SELECT NOT EXISTS(SELECT 1 FROM required r WHERE NOT EXISTS(SELECT 1 FROM information_schema.columns c WHERE c.table_schema=current_schema() AND c.table_name=r.table_name AND c.column_name=r.column_name))`)
	if err != nil {
		return fmt.Errorf("database schema readiness check failed; check connection and apply pending migrations")
	}
	if !ready {
		return fmt.Errorf("database schema incompatible: apply pending migrations before starting API (make migrate-up); automatic migrations are disabled")
	}
	return nil
}
