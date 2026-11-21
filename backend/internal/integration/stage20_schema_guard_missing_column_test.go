package integration

import (
	"testing"

	"server_nesting_optimizer/internal/repository/postgres"
)

// The startup guard must reject a missing column used by operation leases.
// newNestingFixture creates an isolated schema in the guarded nesting_test DB.
func TestStage20SchemaGuardRejectsMissingLeaseExpiry(t *testing.T) {
	f := newNestingFixture(t)
	if err := postgres.ValidateSchema(t.Context(), f.db); err != nil {
		t.Fatalf("fresh migrated test schema rejected: %v", err)
	}
	if _, err := f.db.Exec("ALTER TABLE nesting_operations DROP COLUMN lease_expires_at"); err != nil {
		t.Fatal(err)
	}
	if err := postgres.ValidateSchema(t.Context(), f.db); err == nil {
		t.Fatal("missing nesting_operations.lease_expires_at accepted by startup check")
	}
}
