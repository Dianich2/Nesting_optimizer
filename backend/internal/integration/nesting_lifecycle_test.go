package integration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v3"
	"net/http/httptest"
	"os"
	domainplacement "server_nesting_optimizer/internal/domain/placement"
	httperror "server_nesting_optimizer/internal/transport/http/errors"
	"server_nesting_optimizer/internal/transport/http/handler"
	"server_nesting_optimizer/internal/transport/http/middleware"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	sf "server_nesting_optimizer/internal/geometry/simplefeatures"
	optimizer "server_nesting_optimizer/internal/nesting"
	"server_nesting_optimizer/internal/repository/postgres"
	usecase "server_nesting_optimizer/internal/usecase/nesting"
	"server_nesting_optimizer/pkg/apperror"
)

type gatedNesting struct {
	started chan struct{}
	release chan struct{}
	calls   atomic.Int64
}

func (g *gatedNesting) Optimize(ctx context.Context, p optimizer.Problem) (optimizer.Result, error) {
	g.calls.Add(1)
	select {
	case g.started <- struct{}{}:
	default:
	}
	select {
	case <-ctx.Done():
		return optimizer.Result{}, ctx.Err()
	case <-g.release:
	}
	return optimizer.NewBaselineOptimizer(sf.NewEngine()).Optimize(ctx, p)
}

type nestingFixture struct {
	db     *sqlx.DB
	scoped *postgres.ScopedDB
	gate   *gatedNesting
	uc     *usecase.RunNestingUseCase
	input  usecase.RunNestingInput
}

func newNestingFixture(t *testing.T) nestingFixture {
	t.Helper()
	db := newTestSchema(t)
	applyAllMigrations(t, db)
	_, err := db.Exec(`INSERT INTO users(id,login,email,password_hash,first_name,last_name) VALUES(1,'owner','owner@test','x','Owner','Test'),(2,'other','other@test','x','Other','Test');
        INSERT INTO projects(id,user_id,name) VALUES(1,1,'P'),(2,2,'Other');
        INSERT INTO project_surfaces(id,project_id,name,geometry) VALUES(1,1,'S',ST_GeomFromText('POLYGON((0 0,100 0,100 100,0 100,0 0))')),(2,1,'S2',ST_GeomFromText('POLYGON((0 0,100 0,100 100,0 100,0 0))'));
        INSERT INTO project_patterns(id,project_id,name,geometry) VALUES(1,1,'P',ST_GeomFromText('POLYGON((0 0,10 0,10 10,0 10,0 0))'));`)
	if err != nil {
		t.Fatal(err)
	}
	scoped := &postgres.ScopedDB{DB: db}
	gate := &gatedNesting{started: make(chan struct{}, 8), release: make(chan struct{})}
	uc := newFixtureUseCase(t, scoped, gate)
	return nestingFixture{db: db, scoped: scoped, gate: gate, uc: uc, input: usecase.RunNestingInput{
		UserID: 1, ProjectID: 1, ProjectSurfaceID: 1, Algorithm: optimizer.BaselineAlgorithm,
		Patterns: []usecase.RunNestingPatternInput{{ProjectPatternID: 1, Quantity: 1}}, AllowedRotations: []float64{0},
	}}
}
func newFixtureUseCase(t *testing.T, scoped *postgres.ScopedDB, opt optimizer.Optimizer, overrides ...usecase.UnitOfWork) *usecase.RunNestingUseCase {
	t.Helper()
	registry, err := optimizer.NewOptimizerRegistry(map[optimizer.Algorithm]optimizer.Optimizer{optimizer.BaselineAlgorithm: opt})
	if err != nil {
		t.Fatal(err)
	}
	codec := sf.NewCodec()
	var unit usecase.UnitOfWork = postgres.NewNestingUnitOfWork(scoped.DB, codec)
	if len(overrides) > 0 {
		unit = overrides[0]
	}
	uc := usecase.NewRunNestingUseCase(postgres.NewProjectSurfaceRepository(scoped, codec), postgres.NewProjectPatternRepository(scoped, codec), postgres.NewPlacementRepository(scoped, codec), sf.NewEngine(), registry, unit)
	uc.Transactions = scoped
	uc.Budgets = scoped
	uc.Operations = postgres.NewNestingOperationRepository(scoped)
	uc.Previews = postgres.NewNestingPreviewRepository(scoped)
	return uc
}
func (f nestingFixture) counts(t *testing.T, placements, runs int) {
	t.Helper()
	var p, r int
	if err := f.db.Get(&p, "SELECT count(*) FROM placements"); err != nil {
		t.Fatal(err)
	}
	if err := f.db.Get(&r, "SELECT count(*) FROM nesting_runs"); err != nil {
		t.Fatal(err)
	}
	if p != placements || r != runs {
		t.Fatalf("placements=%d runs=%d, expected %d/%d", p, r, placements, runs)
	}
}
func awaitStart(t *testing.T, g *gatedNesting) {
	t.Helper()
	select {
	case <-g.started:
	case <-time.After(3 * time.Second):
		t.Fatal("optimizer did not start")
	}
}
func expectCode(t *testing.T, err error, code apperror.Code) {
	t.Helper()
	got, ok := apperror.As(err)
	if !ok || got.Code != code {
		t.Fatalf("got %v, expected code %s", err, code)
	}
}

func TestNestingSnapshotRejectsConcurrentChanges(t *testing.T) {
	for _, keep := range []bool{false, true} {
		for _, change := range []string{"surface", "pattern", "add", "delete", "project", "other_surface"} {
			t.Run(fmt.Sprintf("keep=%v/%s", keep, change), func(t *testing.T) {
				f := newNestingFixture(t)
				// A retained transaction would occupy this sole connection and
				// prevent the independent mutation from completing.
				f.db.SetMaxOpenConns(1)
				f.input.KeepExisting = keep
				if change == "delete" {
					if _, err := f.db.Exec("INSERT INTO placements(project_surface_id,project_pattern_id,x,y,rotation) VALUES(1,1,80,80,0)"); err != nil {
						t.Fatal(err)
					}
				}
				done := make(chan error, 1)
				go func() { _, err := f.uc.ExecuteWithKey(t.Context(), f.input, uuid.NewString()); done <- err }()
				awaitStart(t, f.gate)
				ctx, cancel := context.WithTimeout(t.Context(), 700*time.Millisecond)
				defer cancel()
				// This must complete while the optimizer is deliberately blocked.
				err := f.scoped.WithinProject(ctx, 1, 1, func(tx context.Context) error {
					queries := map[string]string{
						"surface":       "UPDATE project_surfaces SET geometry=ST_Scale(geometry,1.1,1.1) WHERE id=1",
						"pattern":       "UPDATE project_patterns SET geometry=ST_Scale(geometry,1.1,1.1) WHERE id=1",
						"add":           "INSERT INTO placements(project_surface_id,project_pattern_id,x,y,rotation) VALUES(1,1,80,80,0)",
						"delete":        "DELETE FROM placements",
						"project":       "DELETE FROM projects WHERE id=1",
						"other_surface": "UPDATE project_surfaces SET geometry=ST_Scale(geometry,1.1,1.1) WHERE id=2",
					}
					_, err := f.scoped.ExecContext(tx, queries[change])
					return err
				})
				if err != nil {
					t.Fatalf("project lock remained held while computing: %v", err)
				}
				close(f.gate.release)
				err = <-done
				switch change {
				case "project":
					expectCode(t, err, apperror.AppCodeNotFound)
					f.counts(t, 0, 0)
				case "other_surface":
					if err != nil {
						t.Fatal(err)
					}
					f.counts(t, 1, 1)
				default:
					expectCode(t, err, apperror.AppCodeStale)
					count := 0
					if change == "add" {
						count = 1
					}
					f.counts(t, count, 0)
				}
			})
		}
	}
}

func TestNestingDurableIdempotencyAndRestart(t *testing.T) {
	f := newNestingFixture(t)
	key := uuid.NewString()
	type outcome struct {
		result usecase.RunNestingOutput
		err    error
	}
	done := make(chan outcome, 2)
	go func() { r, e := f.uc.ExecuteWithKey(t.Context(), f.input, key); done <- outcome{r, e} }()
	awaitStart(t, f.gate)
	go func() { r, e := f.uc.ExecuteWithKey(t.Context(), f.input, key); done <- outcome{r, e} }()
	changed := f.input
	changed.Patterns = []usecase.RunNestingPatternInput{{ProjectPatternID: 1, Quantity: 2}}
	_, err := f.uc.ExecuteWithKey(t.Context(), changed, key)
	expectCode(t, err, apperror.AppCodeConflict)
	status, err := f.uc.GetOperation(t.Context(), 1, 1, 1, key)
	if err != nil || status.Status != "running" {
		t.Fatalf("running status: %v %v", status, err)
	}
	close(f.gate.release)
	a, b := <-done, <-done
	if a.err != nil || b.err != nil || a.result.Placements[0].ID != b.result.Placements[0].ID || f.gate.calls.Load() != 1 {
		t.Fatalf("duplicate computation or save: %v %v calls=%d", a.err, b.err, f.gate.calls.Load())
	}
	f.counts(t, 1, 1)
	// Reconstruct the service with no in-memory executions or optimizer state.
	restarted := newFixtureUseCase(t, f.scoped, &gatedNesting{started: make(chan struct{}, 1), release: make(chan struct{})})
	replay, err := restarted.ExecuteWithKey(t.Context(), f.input, key)
	if err != nil || replay.Placements[0].ID != a.result.Placements[0].ID {
		t.Fatalf("lost response/restart replay: %v", err)
	}
	f.counts(t, 1, 1)
}

func TestNestingCancelledAndExpiredOperationsAreTerminal(t *testing.T) {
	for _, kind := range []string{"cancel", "deadline", "crash"} {
		t.Run(kind, func(t *testing.T) {
			f := newNestingFixture(t)
			key := uuid.NewString()
			if kind == "crash" {
				encoded, err := json.Marshal(f.input)
				if err != nil {
					t.Fatal(err)
				}
				hash := sha256.Sum256(encoded)
				if _, err := f.db.Exec(`INSERT INTO nesting_operations(project_surface_id,operation_key,request_hash,response,state,lease_token,lease_expires_at) VALUES(1,$1,$3,'{"Status":"running"}','running',$2,clock_timestamp()-interval '1 second')`, key, uuid.NewString(), hash[:]); err != nil {
					t.Fatal(err)
				}
				status, err := f.uc.GetOperation(t.Context(), 1, 1, 1, key)
				if err != nil || status.Status != "failed" || status.Failure.Code != apperror.AppCodeTimeout {
					t.Fatalf("expired worker remained running: %v %v", status, err)
				}
				_, err = f.uc.ExecuteWithKey(t.Context(), f.input, key)
				expectCode(t, err, apperror.AppCodeTimeout)
				if f.gate.calls.Load() != 0 {
					t.Fatal("crashed operation restarted with the same key")
				}
				f.counts(t, 0, 0)
				return
			}
			limit := 200 * time.Millisecond
			if kind == "cancel" {
				limit = 3 * time.Second
			}
			ctx, cancel := context.WithTimeout(t.Context(), limit)
			defer cancel()
			done := make(chan error, 1)
			go func() { _, err := f.uc.ExecuteWithKey(ctx, f.input, key); done <- err }()
			awaitStart(t, f.gate)
			expected := apperror.AppCodeTimeout
			if kind == "cancel" {
				if _, err := f.uc.CancelOperation(t.Context(), 2, 1, 1, key); err == nil {
					t.Fatal("other owner cancelled operation")
				}
				ok, err := f.uc.CancelOperation(t.Context(), 1, 1, 1, key)
				if err != nil || !ok {
					t.Fatalf("cancel active: %v %v", ok, err)
				}
				expected = apperror.AppCodeCancelled
			}
			expectCode(t, <-done, expected)
			_, err := f.uc.ExecuteWithKey(t.Context(), f.input, key)
			expectCode(t, err, expected)
			if f.gate.calls.Load() != 1 {
				t.Fatal("terminal operation executed again")
			}
			if ok, err := f.uc.CancelOperation(t.Context(), 1, 1, 1, key); err != nil || ok {
				t.Fatalf("completed cancellation: %v %v", ok, err)
			}
			f.counts(t, 0, 0)
		})
	}
}

func TestNestingSavingFailureRollsBackReplacement(t *testing.T) {
	f := newNestingFixture(t)
	close(f.gate.release)
	if _, err := f.db.Exec(`INSERT INTO placements(project_surface_id,project_pattern_id,x,y,rotation) VALUES(1,1,80,80,0);
        CREATE FUNCTION reject_run() RETURNS trigger LANGUAGE plpgsql AS 'BEGIN RAISE EXCEPTION ''injected save failure''; END';
        CREATE TRIGGER reject_run BEFORE INSERT ON nesting_runs FOR EACH ROW EXECUTE FUNCTION reject_run();`); err != nil {
		t.Fatal(err)
	}
	key := uuid.NewString()
	_, err := f.uc.ExecuteWithKey(t.Context(), f.input, key)
	if err == nil {
		t.Fatal("save failure not propagated")
	}
	f.counts(t, 1, 0)
	var x float64
	if err = f.db.Get(&x, "SELECT x FROM placements"); err != nil || x != 80 {
		t.Fatalf("replacement did not roll back: %v %v", x, err)
	}
	_, err = f.uc.ExecuteWithKey(t.Context(), f.input, key)
	if err == nil || f.gate.calls.Load() != 1 {
		t.Fatal("failed save retried its computation")
	}
}

func TestNestingSavesExactPreviewOnce(t *testing.T) {
	f := newNestingFixture(t)
	type reply struct {
		value usecase.RunNestingOutput
		err   error
	}
	calculated := make(chan reply, 1)
	go func() { value, err := f.uc.Preview(t.Context(), f.input); calculated <- reply{value, err} }()
	awaitStart(t, f.gate)
	time.Sleep(30 * time.Millisecond)
	close(f.gate.release)
	response := <-calculated
	preview, err := response.value, response.err
	if err != nil || preview.PreviewID == "" {
		t.Fatalf("preview: %v", err)
	}
	input := f.input
	input.PreviewID = preview.PreviewID
	saved, err := f.uc.ExecuteWithKey(t.Context(), input, uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}
	if f.gate.calls.Load() != 1 || saved.Placements[0].ID == 0 || saved.Placements[0].X != preview.Placements[0].X || saved.Placements[0].Y != preview.Placements[0].Y {
		t.Fatal("preview was recalculated or changed")
	}
	replay, err := f.uc.ExecuteWithKey(t.Context(), input, uuid.NewString())
	if err != nil || replay.Placements[0].ID != saved.Placements[0].ID || f.gate.calls.Load() != 1 {
		t.Fatalf("preview saved twice: %v", err)
	}
	f.counts(t, 1, 1)
	var runtime int64
	var sourcePreview string
	if err = f.db.Get(&runtime, "SELECT duration_ms FROM nesting_runs"); err != nil || runtime < 30 {
		t.Fatalf("history lost original optimization duration: %d %v", runtime, err)
	}
	if err = f.db.Get(&sourcePreview, "SELECT parameters->>'PreviewID' FROM nesting_runs"); err != nil || sourcePreview != preview.PreviewID {
		t.Fatalf("history lost original parameters: %q %v", sourcePreview, err)
	}
	input.UserID = 2
	if _, err = f.uc.ExecuteWithKey(t.Context(), input, uuid.NewString()); err == nil {
		t.Fatal("foreign preview saved")
	}
	input = f.input
	input.PreviewID = preview.PreviewID
	input.ProjectSurfaceID = 2
	if _, err = f.uc.ExecuteWithKey(t.Context(), input, uuid.NewString()); err == nil {
		t.Fatal("preview applied to another surface")
	}
	input = f.input
	input.PreviewID = preview.PreviewID
	if _, err = f.db.Exec("UPDATE nesting_previews SET expires_at=clock_timestamp()-interval '1 second'"); err != nil {
		t.Fatal(err)
	}
	if _, err = f.uc.ExecuteWithKey(t.Context(), input, uuid.NewString()); err == nil {
		t.Fatal("expired preview saved")
	}
	f.counts(t, 1, 1)
}

func TestConcurrentNestingFinalCommits(t *testing.T) {
	for _, sameSurface := range []bool{true, false} {
		t.Run(fmt.Sprintf("same_surface=%v", sameSurface), func(t *testing.T) {
			f := newNestingFixture(t)
			second := f.input
			if !sameSurface {
				second.ProjectSurfaceID = 2
			}
			done := make(chan error, 2)
			go func() { _, err := f.uc.ExecuteWithKey(t.Context(), f.input, uuid.NewString()); done <- err }()
			go func() { _, err := f.uc.ExecuteWithKey(t.Context(), second, uuid.NewString()); done <- err }()
			awaitStart(t, f.gate)
			awaitStart(t, f.gate)
			close(f.gate.release)
			a, b := <-done, <-done
			if sameSurface {
				if a == nil && b == nil {
					t.Fatal("two results based on the same snapshot both committed")
				}
				if a == nil {
					expectCode(t, b, apperror.AppCodeStale)
				} else if b == nil {
					expectCode(t, a, apperror.AppCodeStale)
				} else {
					t.Fatalf("both failed: %v %v", a, b)
				}
				f.counts(t, 1, 1)
			} else {
				if a != nil || b != nil {
					t.Fatalf("independent surface commits: %v %v", a, b)
				}
				f.counts(t, 2, 2)
			}
		})
	}
}

func TestPreviewStalenessAndBoundedRetention(t *testing.T) {
	f := newNestingFixture(t)
	close(f.gate.release)
	preview, err := f.uc.Preview(t.Context(), f.input)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.scoped.WithinProject(t.Context(), 1, 1, func(ctx context.Context) error {
		_, err := f.scoped.ExecContext(ctx, "UPDATE project_patterns SET geometry=ST_Scale(geometry,1.1,1.1) WHERE id=1")
		return err
	}); err != nil {
		t.Fatal(err)
	}
	input := f.input
	input.PreviewID = preview.PreviewID
	_, err = f.uc.ExecuteWithKey(t.Context(), input, uuid.NewString())
	expectCode(t, err, apperror.AppCodeStale)
	f.counts(t, 0, 0)
	for range 11 {
		if _, err = f.uc.Preview(t.Context(), f.input); err != nil {
			t.Fatal(err)
		}
	}
	var count int
	if err = f.db.Get(&count, "SELECT count(*) FROM nesting_previews"); err != nil || count != 10 {
		t.Fatalf("preview retention: %d %v", count, err)
	}
	_, err = f.uc.ExecuteWithKey(t.Context(), input, uuid.NewString())
	expectCode(t, err, apperror.AppCodeNotFound)
	f.counts(t, 0, 0)
	if _, err = f.db.Exec("UPDATE nesting_previews SET expires_at=clock_timestamp()-interval '1 second'"); err != nil {
		t.Fatal(err)
	}
	purged, err := postgres.NewNestingPreviewRepository(f.scoped).PurgeExpired(t.Context(), 1000)
	if err != nil || purged != 10 {
		t.Fatalf("physical TTL cleanup: %d %v", purged, err)
	}
}

type pausedPlacementWrite struct {
	usecase.PlacementWriter
	started chan struct{}
}

func (p *pausedPlacementWrite) Create(ctx context.Context, input domainplacement.Placement, projectID, userID int64) (domainplacement.Placement, error) {
	row, err := p.PlacementWriter.Create(ctx, input, projectID, userID)
	if err != nil {
		return row, err
	}
	close(p.started)
	<-ctx.Done()
	return row, ctx.Err()
}

type pausedNestingSave struct {
	usecase.UnitOfWork
	started chan struct{}
}

func (p pausedNestingSave) WithinTransaction(ctx context.Context, fn func(usecase.TransactionRepositories) error) error {
	return p.UnitOfWork.WithinTransaction(ctx, func(repos usecase.TransactionRepositories) error {
		repos.Placements = &pausedPlacementWrite{PlacementWriter: repos.Placements, started: p.started}
		return fn(repos)
	})
}
func TestCancellationDuringNestingWriteRollsBack(t *testing.T) {
	f := newNestingFixture(t)
	close(f.gate.release)
	if _, err := f.db.Exec("INSERT INTO placements(project_surface_id,project_pattern_id,x,y,rotation) VALUES(1,1,80,80,0)"); err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	unit := pausedNestingSave{UnitOfWork: postgres.NewNestingUnitOfWork(f.db, sf.NewCodec()), started: started}
	uc := newFixtureUseCase(t, f.scoped, f.gate, unit)
	key := uuid.NewString()
	done := make(chan error, 1)
	go func() { _, err := uc.ExecuteWithKey(t.Context(), f.input, key); done <- err }()
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("write phase not reached")
	}
	if ok, err := uc.CancelOperation(t.Context(), 1, 1, 1, key); err != nil || !ok {
		t.Fatalf("cancel during write: %v %v", ok, err)
	}
	expectCode(t, <-done, apperror.AppCodeCancelled)
	f.counts(t, 1, 0)
	var x float64
	if err := f.db.Get(&x, "SELECT x FROM placements"); err != nil || x != 80 {
		t.Fatal("old layout was not restored by rollback")
	}
	if _, err := f.uc.ExecuteWithKey(t.Context(), f.input, uuid.NewString()); err != nil {
		t.Fatalf("new execution after cancellation: %v", err)
	}
	f.counts(t, 1, 1)
}

type panickingNesting struct{}

func (panickingNesting) Optimize(context.Context, optimizer.Problem) (optimizer.Result, error) {
	panic("injected optimizer panic")
}
func TestOptimizerPanicDoesNotLeakSlots(t *testing.T) {
	f := newNestingFixture(t)
	uc := newFixtureUseCase(t, f.scoped, panickingNesting{})
	for range 4 {
		_, err := uc.ExecuteWithKey(t.Context(), f.input, uuid.NewString())
		expectCode(t, err, apperror.AppCodeInternal)
	}
	f.counts(t, 0, 0)
}

func TestNestingSourceBudgetsPrecedeOptimizer(t *testing.T) {
	for _, kind := range []string{"vertices", "placements"} {
		t.Run(kind, func(t *testing.T) {
			f := newNestingFixture(t)
			query := "UPDATE project_patterns SET geometry=ST_Buffer(ST_Point(0,0),100,1025) WHERE id=1"
			if kind == "placements" {
				query = `UPDATE project_surfaces SET geometry=ST_Scale(geometry,100,100) WHERE id=1;
                    INSERT INTO placements(project_surface_id,project_pattern_id,x,y,rotation)
                    SELECT 1,1,(g%80)*12+6,(g/80)*12+6,0 FROM generate_series(1,2001) g`
			}
			if _, err := f.db.Exec(query); err != nil {
				t.Fatal(err)
			}
			_, err := f.uc.ExecuteWithKey(t.Context(), f.input, uuid.NewString())
			expectCode(t, err, apperror.AppCodeValidation)
			if f.gate.calls.Load() != 0 {
				t.Fatal("oversized source reached optimizer")
			}
		})
	}
}

func TestNestingLifecycleMigrationPreservesLegacyRecords(t *testing.T) {
	f := newNestingFixture(t)
	_, err := f.db.Exec(`ALTER TABLE nesting_operations DROP CONSTRAINT nesting_operations_state_check,
        DROP COLUMN state,DROP COLUMN lease_token,DROP COLUMN lease_expires_at,DROP COLUMN cancel_requested;
        ALTER TABLE nesting_runs DROP COLUMN parameters;
        INSERT INTO nesting_operations(project_surface_id,operation_key,request_hash,response)
        VALUES(1,'11111111-1111-4111-8111-111111111111',decode(repeat('12',32),'hex'),'{"Status":"failed","Failure":{"Code":"conflict_error","Message":"legacy"}}')`)
	if err != nil {
		t.Fatal(err)
	}
	var before, after string
	if err = f.db.Get(&before, "SELECT response::text FROM nesting_operations"); err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("../../migrations/20261119120000_nesting_operation_lifecycle.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.db.Exec(strings.Split(string(source), "-- +goose Down")[0]); err != nil {
		t.Fatal(err)
	}
	if err = f.db.Get(&after, "SELECT response::text FROM nesting_operations"); err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatal("legacy saved response was rewritten")
	}
	status, err := f.uc.GetOperation(t.Context(), 1, 1, 1, "11111111-1111-4111-8111-111111111111")
	if err != nil || status.Status != "failed" || status.Failure.Message != "legacy" {
		t.Fatalf("legacy status after upgrade: %v %v", status, err)
	}
}

func TestPreviewHTTPKeySharesComputationAndSurvivesRestart(t *testing.T) {
	f := newNestingFixture(t)
	router := func(uc *usecase.RunNestingUseCase) *fiber.App {
		app := fiber.New(fiber.Config{ErrorHandler: httperror.Handle})
		app.Use(func(c fiber.Ctx) error { c.Locals(middleware.UserIDLocalKey, int64(1)); return c.Next() })
		app.Post("/projects/:project_id/surfaces/:project_surface_id/nesting/preview", handler.NewNestingHandler(uc).Preview)
		return app
	}
	app := router(f.uc)
	key := uuid.NewString()
	body := `{"algorithm":"baseline","patterns":[{"project_pattern_id":1,"quantity":1}],"allowed_rotations":[0],"keep_existing":false}`
	type reply struct {
		status int
		data   map[string]any
		err    error
	}
	call := func(app *fiber.App, body string) reply {
		req := httptest.NewRequest("POST", "/projects/1/surfaces/1/nesting/preview", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", key)
		res, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
		if err != nil {
			return reply{err: err}
		}
		defer func() { _ = res.Body.Close() }()
		var data map[string]any
		err = json.NewDecoder(res.Body).Decode(&data)
		return reply{status: res.StatusCode, data: data, err: err}
	}
	done := make(chan reply, 2)
	go func() { done <- call(app, body) }()
	awaitStart(t, f.gate)
	go func() { done <- call(app, body) }()
	conflict := call(app, strings.Replace(body, `"quantity":1`, `"quantity":2`, 1))
	if conflict.err != nil || conflict.status != 409 {
		t.Fatalf("preview parameter conflict: %v", conflict)
	}
	close(f.gate.release)
	a, b := <-done, <-done
	if a.err != nil || b.err != nil || a.status != 200 || b.status != 200 || a.data["preview_id"] != b.data["preview_id"] || f.gate.calls.Load() != 1 {
		t.Fatalf("duplicate preview computation: %v %v calls=%d", a, b, f.gate.calls.Load())
	}
	restartedGate := &gatedNesting{started: make(chan struct{}, 1), release: make(chan struct{})}
	restarted := router(newFixtureUseCase(t, f.scoped, restartedGate))
	cached := call(restarted, body)
	if cached.err != nil || cached.status != 200 || cached.data["preview_id"] != a.data["preview_id"] || restartedGate.calls.Load() != 0 {
		t.Fatalf("ready preview did not survive restart: %v", cached)
	}
	f.counts(t, 0, 0)
}
