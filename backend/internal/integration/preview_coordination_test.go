package integration

import (
	"context"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"net/http/httptest"
	"server_nesting_optimizer/internal/repository/postgres"
	httperror "server_nesting_optimizer/internal/transport/http/errors"
	"server_nesting_optimizer/internal/transport/http/handler"
	"server_nesting_optimizer/internal/transport/http/middleware"
	usecase "server_nesting_optimizer/internal/usecase/nesting"
	"server_nesting_optimizer/pkg/apperror"
	"testing"
	"time"
)

func TestPreviewCrossInstanceCancellationAndLease(t *testing.T) {
	for _, mode := range []string{"cancel", "lease"} {
		t.Run(mode, func(t *testing.T) {
			f := newNestingFixture(t)
			second := newFixtureUseCase(t, f.scoped, f.gate)
			key := uuid.NewString()
			done := make(chan error, 1)
			go func() { _, err := f.uc.PreviewWithKey(context.Background(), f.input, key); done <- err }()
			awaitStart(t, f.gate)
			status, err := second.GetPreviewOperation(t.Context(), 1, 1, 1, key)
			if err != nil || status.Status != "running" {
				t.Fatalf("status %v %v", status, err)
			}
			if _, err = second.GetPreviewOperation(t.Context(), 2, 1, 1, key); err == nil {
				t.Fatal("foreign read allowed")
			}
			if ok, err := second.CancelPreviewOperation(t.Context(), 2, 1, 1, key); err == nil || ok {
				t.Fatal("foreign cancel allowed")
			}
			// Same UUID in save namespace is independent.
			hash, _ := usecase.Fingerprint(f.input)
			_, err = f.uc.Operations.Reserve(t.Context(), 1, uuid.MustParse(key), hash[:], uuid.New(), time.Now().Add(time.Minute))
			if err != nil {
				t.Fatal(err)
			}
			if mode == "cancel" {
				app := fiber.New(fiber.Config{ErrorHandler: httperror.Handle})
				app.Use(func(c fiber.Ctx) error { c.Locals(middleware.UserIDLocalKey, int64(1)); return c.Next() })
				h := handler.NewNestingHandler(second)
				app.Delete("/projects/:project_id/surfaces/:project_surface_id/nesting/operations/:operation_key", h.Cancel)
				response, err := app.Test(httptest.NewRequest("DELETE", "/projects/1/surfaces/1/nesting/operations/"+key+"?kind=preview", nil))
				if err != nil {
					t.Fatal(err)
				}
				_ = response.Body.Close()
				if response.StatusCode != 204 {
					t.Fatalf("cross handler DELETE: %d", response.StatusCode)
				}
			} else {
				_, err = f.db.Exec("UPDATE nesting_preview_operations SET lease_expires_at=clock_timestamp()-interval '1 second'")
				if err != nil {
					t.Fatal(err)
				}
				_, err = second.GetPreviewOperation(t.Context(), 1, 1, 1, key)
				if err != nil {
					t.Fatal(err)
				}
			}
			select {
			case err := <-done:
				if err == nil {
					t.Fatal("cancelled worker succeeded")
				}
			case <-time.After(3 * time.Second):
				t.Fatal("worker did not stop")
			}
			status, err = second.GetPreviewOperation(t.Context(), 1, 1, 1, key)
			if err != nil || status.Status != "failed" || status.Failure == nil {
				t.Fatalf("terminal %v %v", status, err)
			}
			saved, err := second.GetOperation(t.Context(), 1, 1, 1, key)
			if err != nil || saved.Status != "running" {
				t.Fatal("save namespace changed")
			}
			_, err = second.PreviewWithKey(t.Context(), f.input, key)
			if err == nil || f.gate.calls.Load() != 1 {
				t.Fatal("terminal key recomputed")
			}
			f.counts(t, 0, 0)
		})
	}
}
func TestPreviewExpiryDoesNotRecompute(t *testing.T) {
	f := newNestingFixture(t)
	close(f.gate.release)
	key := uuid.NewString()
	result, err := f.uc.PreviewWithKey(t.Context(), f.input, key)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.db.Exec("UPDATE nesting_previews SET expires_at=clock_timestamp()-interval '1 second' WHERE id=$1", result.PreviewID); err != nil {
		t.Fatal(err)
	}
	_, err = f.uc.PreviewWithKey(t.Context(), f.input, key)
	expectCode(t, err, apperror.Code("preview_expired"))
	status, err := f.uc.GetPreviewOperation(t.Context(), 1, 1, 1, key)
	if err != nil || status.Failure == nil || status.Failure.Code != apperror.Code("preview_expired") {
		t.Fatalf("expired %v %v", status, err)
	}
	if f.gate.calls.Load() != 1 {
		t.Fatal("expired key recomputed")
	}
}

func TestPreviewCancelCommitRace(t *testing.T) {
	for i := 0; i < 5; i++ {
		t.Run(uuid.NewString(), func(t *testing.T) {
			f := newNestingFixture(t)
			second := newFixtureUseCase(t, f.scoped, f.gate)
			key := uuid.NewString()
			done := make(chan error, 1)
			go func() { _, err := f.uc.PreviewWithKey(context.Background(), f.input, key); done <- err }()
			awaitStart(t, f.gate)
			cancelDone := make(chan error, 1)
			go func() { _, err := second.CancelPreviewOperation(context.Background(), 1, 1, 1, key); cancelDone <- err }()
			close(f.gate.release)
			err := <-done
			if e := <-cancelDone; e != nil {
				t.Fatal(e)
			}
			status, e := second.GetPreviewOperation(t.Context(), 1, 1, 1, key)
			if e != nil {
				t.Fatal(e)
			}
			var count int
			if e = f.db.Get(&count, "SELECT count(*) FROM nesting_previews"); e != nil {
				t.Fatal(e)
			}
			switch status.Status {
			case "completed":
				if err != nil || count != 1 || status.Result == nil {
					t.Fatalf("inconsistent completed: %v count=%d", err, count)
				}
			case "failed":
				if err == nil || count != 0 {
					t.Fatalf("inconsistent failure: %v count=%d", err, count)
				}
			default:
				t.Fatalf("nonterminal %v", status)
			}
			f.counts(t, 0, 0)
		})
	}
}

func TestPreviewScopesLegacyAndCleanup(t *testing.T) {
	f := newNestingFixture(t)
	close(f.gate.release)
	key := uuid.NewString()
	first, err := f.uc.PreviewWithKey(t.Context(), f.input, key)
	if err != nil {
		t.Fatal(err)
	}
	other := f.input
	other.ProjectSurfaceID = 2
	second, err := f.uc.PreviewWithKey(t.Context(), other, key)
	if err != nil || second.PreviewID == first.PreviewID {
		t.Fatalf("surface scope %v", err)
	}
	// Simulate a ready record from before the new operation table existed.
	if _, err = f.db.Exec("DELETE FROM nesting_preview_operations WHERE project_surface_id=1"); err != nil {
		t.Fatal(err)
	}
	replay, err := f.uc.PreviewWithKey(t.Context(), f.input, key)
	if err != nil || replay.PreviewID != first.PreviewID || f.gate.calls.Load() != 2 {
		t.Fatalf("legacy replay %v", err)
	}
	if _, err = f.db.Exec("UPDATE nesting_previews SET expires_at=clock_timestamp()-interval '1 second' WHERE id=$1", first.PreviewID); err != nil {
		t.Fatal(err)
	}
	tx, err := f.db.BeginTxx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err = tx.Exec("SELECT id FROM nesting_previews WHERE id=$1 FOR UPDATE", first.PreviewID); err != nil {
		t.Fatal(err)
	}
	repo := postgres.NewNestingPreviewRepository(f.scoped)
	count, err := repo.PurgeExpired(t.Context(), 1000)
	if err != nil || count != 0 {
		t.Fatalf("cleanup did not skip locked row %d %v", count, err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
	count, err = repo.PurgeExpired(t.Context(), 1000)
	if err != nil || count != 1 {
		t.Fatalf("cleanup %d %v", count, err)
	}
	f.counts(t, 0, 0)
}
