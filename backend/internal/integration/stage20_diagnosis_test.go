package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"net/url"
	"os"
	"server_nesting_optimizer/internal/repository/postgres"
	"server_nesting_optimizer/pkg/apperror"
	"testing"
	"time"
)

func TestStage20RealMatrix(t *testing.T) {
	if os.Getenv("NESTING_STAGE18") != "1" {
		t.Skip("own guarded HTTP lab only")
	}
	f := newNestingFixture(t)
	hash, e := bcrypt.GenerateFromPassword([]byte("test-matrix-password"), 4)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = f.db.Exec(`UPDATE users SET password_hash=$1`, string(hash)); e != nil {
		t.Fatal(e)
	}
	if _, e = f.db.Exec(`UPDATE project_patterns SET geometry=ST_GeomFromText('POLYGON((0 0,40 0,40 30,0 30,0 0))') WHERE id=1`); e != nil {
		t.Fatal(e)
	}
	if e = postgres.ValidateSchema(t.Context(), f.db); e != nil {
		t.Fatal(e)
	}
	var schema string
	requireProcessGet(t, f.db, &schema, "SELECT current_schema()")
	dsn, e := url.Parse(os.Getenv("NESTING_TEST_DSN"))
	if e != nil {
		t.Fatal("invalid lab DSN")
	}
	q := dsn.Query()
	q.Set("search_path", schema+",public")
	dsn.RawQuery = q.Encode()
	api := startProcess(t, dsn.String(), true)
	login := processRequest(api.base, "", "POST", "/auth/login", "", `{"identifier":"owner","password":"test-matrix-password"}`)
	requireProcessStatus(t, login, 200)
	token := login.body["access_token"].(string)
	type scenario struct {
		name, wkt          string
		occupied, quantity int
		keep               bool
		algorithm          string
	}
	rect := "POLYGON((0 0,300 0,300 200,0 200,0 0))"
	concave := "POLYGON((0 0,300 0,300 80,120 80,120 200,0 200,0 0))"
	hole := "POLYGON((0 0,300 0,300 200,0 200,0 0),(120 80,120 120,160 120,160 80,120 80))"
	complex := "POLYGON((0 0,600 0,600 180,400 180,400 400,0 400,0 0),(280 100,280 130,320 130,320 100,280 100))"
	var scenarios []scenario
	for _, algo := range []string{"baseline", "nfp_greedy"} {
		for _, keep := range []bool{false, true} {
			for _, n := range []int{1, 2, 5} {
				scenarios = append(scenarios, scenario{"empty", rect, 0, n, keep, algo})
			}
			scenarios = append(scenarios, scenario{"concave", concave, 0, 2, keep, algo}, scenario{"hole-one", hole, 1, 2, keep, algo}, scenario{"hole-five", hole, 5, 2, keep, algo}, scenario{"complex-24", complex, 24, 2, keep, algo})
		}
	}
	for _, algo := range []string{"baseline", "nfp_greedy", "genetic", "simulated_annealing", "hybrid"} {
		scenarios = append(scenarios, scenario{"occupied", rect, 1, 2, true, algo})
	}
	for _, n := range []int{1, 2} {
		scenarios = append(scenarios, scenario{"full", "POLYGON((0 0,80 0,80 30,0 30,0 0))", 2, n, true, "baseline"}, scenario{"too-small", "POLYGON((0 0,10 0,10 10,0 10,0 0))", 0, n, true, "baseline"})
	}
	id := int64(2)
	for _, sc := range scenarios {
		for _, preview := range []bool{false, true} {
			id++
			sc, preview, id := sc, preview, id
			if !t.Run(fmt.Sprintf("%s/%s/q%d/keep%v/preview%v", sc.name, sc.algorithm, sc.quantity, sc.keep, preview), func(t *testing.T) {
				if _, e := f.db.Exec("INSERT INTO project_surfaces(id,project_id,name,geometry) VALUES($1,1,$2,ST_GeomFromText($3))", id, sc.name, sc.wkt); e != nil {
					t.Fatal(e)
				}
				for i := 0; i < sc.occupied; i++ {
					x, y := 20+42*(i%8), 15+32*(i/8)
					if sc.name == "full" {
						x = 20 + 40*i
					}
					if _, e := f.db.Exec("INSERT INTO placements(project_surface_id,project_pattern_id,x,y,rotation) VALUES($1,1,$2,$3,0)", id, x, y); e != nil {
						t.Fatal(e)
					}
				}
				count := func(table string) int {
					var n int
					requireProcessGet(t, f.db, &n, "SELECT count(*) FROM "+table+" WHERE project_surface_id=$1", id)
					return n
				}
				beforeP, beforeH := count("placements"), count("nesting_runs")
				angles := "[0,90]"
				if sc.name == "complex-24" {
					angles = "[0,90,180,270]"
				}
				if sc.name == "too-small" {
					angles = "[0]"
				}
				body := fmt.Sprintf(`{"algorithm":%q,"patterns":[{"project_pattern_id":1,"quantity":%d}],"allowed_rotations":%s,"keep_existing":%v}`, sc.algorithm, sc.quantity, angles, sc.keep)
				base := fmt.Sprintf("/projects/1/surfaces/%d/nesting", id)
				path := base
				if preview {
					path += "/preview"
				}
				key := uuid.NewString()
				sample, result := loadCall(api.base, token, "POST", path, key, body)
				if sample.Error || sample.Status != 200 {
					t.Fatalf("HTTP=%d network=%v error=%v", sample.Status, sample.Error, result["error"])
				}
				metrics := result["metrics"].(map[string]any)
				requested, placed := int(metrics["requested_count"].(float64)), int(metrics["placed_count"].(float64))
				unplaced := 0
				if rows, ok := result["unplaced"].([]any); ok {
					for _, row := range rows {
						unplaced += int(row.(map[string]any)["quantity"].(float64))
					}
				}
				if requested != sc.quantity || placed+unplaced != requested {
					t.Fatal("result count inconsistency")
				}
				if sc.name == "full" || sc.name == "too-small" {
					if placed != 0 {
						t.Fatal("impossible fit accepted")
					}
				} else if placed != requested {
					t.Fatalf("simple/available fit only %d/%d", placed, requested)
				}
				again, replayed := loadCall(api.base, token, "POST", path, key, body)
				left, _ := json.Marshal(result)
				right, _ := json.Marshal(replayed)
				if again.Status != 200 || string(left) != string(right) {
					t.Fatal("UUID replay changed result")
				}
				previewID, _ := result["preview_id"].(string)
				table := "nesting_operations"
				if preview {
					table = "nesting_preview_operations"
					if count("placements") != beforeP || count("nesting_runs") != beforeH || previewID == "" {
						t.Fatal("preview wrote layout/history or no ID")
					}
					saveBody := body[:len(body)-1] + `,"preview_id":"` + previewID + `"}`
					saved, _ := loadCall(api.base, token, "POST", base, key, saveBody)
					if saved.Status != 200 {
						t.Fatal("exact preview apply failed")
					}
					replay, _ := loadCall(api.base, token, "POST", base, key, saveBody)
					if replay.Status != 200 {
						t.Fatal("exact save replay failed")
					}
				}
				expected := placed
				if sc.keep {
					expected += beforeP
				}
				if count("placements") != expected || count("nesting_runs") != beforeH+1 {
					t.Fatal("partial/duplicate commit")
				}
				var terminal string
				requireProcessGet(t, f.db, &terminal, "SELECT state FROM "+table+" WHERE project_surface_id=$1 AND operation_key=$2", id, key)
				if terminal != "completed" {
					t.Fatal("not terminal completed")
				}
				evidence, _ := json.Marshal(map[string]any{"case": sc.name, "algorithm": sc.algorithm, "preview": preview, "keep": sc.keep, "requested": requested, "placed": placed, "unplaced": unplaced, "elapsed_ms": sample.Milliseconds, "HTTP": sample.Status, "terminal": terminal, "key": key, "preview_id": previewID, "before_placements": beforeP, "after_placements": count("placements"), "before_history": beforeH, "after_history": count("nesting_runs")})
				t.Log("STAGE20_CASE " + string(evidence))
				stats := processRequest(api.control, "", "GET", "/stats", "", "")
				requireProcessStatus(t, stats, 200)
				timings, _ := json.Marshal(stats.body["Optimizer"])
				t.Log("STAGE20_TIMING_CUMULATIVE " + string(timings))
			}) {
				t.Fatal("STOP stage20 matrix; do not continue stages21/22")
			}
		}
	}
	auditLoadState(t, f)
	auditRetainedLoadResults(t, f)
}

// PostgreSQL may return pq/57014 instead of context.Canceled when a context
// interrupts a query waiting for a real row lock. The durable terminal outcome
// must still reflect the accepted cancellation, for both save and preview.
func TestStage20CancelDuringSQLLock(t *testing.T) {
	for _, preview := range []bool{false, true} {
		t.Run(fmt.Sprintf("preview=%v", preview), func(t *testing.T) {
			f := newNestingFixture(t)
			tx, err := f.db.Beginx()
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback() }()
			var holder, id int
			if err = tx.Get(&holder, "SELECT pg_backend_pid()"); err != nil {
				t.Fatal(err)
			}
			if err = tx.Get(&id, "SELECT id FROM projects WHERE id=1 FOR UPDATE"); err != nil {
				t.Fatal(err)
			}
			key := uuid.NewString()
			execute := f.uc.ExecuteWithKey
			cancel := f.uc.CancelOperation
			status := f.uc.GetOperation
			if preview {
				execute = f.uc.PreviewWithKey
				cancel = f.uc.CancelPreviewOperation
				status = f.uc.GetPreviewOperation
			}
			ctx, stop := context.WithTimeout(t.Context(), 5*time.Second)
			defer stop()
			done := make(chan error, 1)
			go func() { _, e := execute(ctx, f.input, key); done <- e }()
			processEventually(t, func() bool {
				var waiting bool
				if e := f.db.Get(&waiting, "SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE $1=ANY(pg_blocking_pids(pid)))", holder); e != nil {
					t.Fatal(e)
				}
				return waiting
			})
			ok, err := cancel(t.Context(), 1, 1, 1, key)
			if err != nil || !ok {
				t.Fatalf("cancel not accepted: %v %v", ok, err)
			}
			expectCode(t, <-done, apperror.AppCodeCancelled)
			if err = tx.Rollback(); err != nil {
				t.Fatal(err)
			}
			outcome, err := status(t.Context(), 1, 1, 1, key)
			if err != nil || outcome.Status != "failed" || outcome.Failure == nil || outcome.Failure.Code != apperror.AppCodeCancelled {
				t.Fatalf("incorrect terminal outcome: %+v %v", outcome, err)
			}
			_, err = execute(t.Context(), f.input, key)
			expectCode(t, err, apperror.AppCodeCancelled)
			if f.gate.calls.Load() != 0 {
				t.Fatal("cancelled locked snapshot reached optimizer")
			}
			f.counts(t, 0, 0)
		})
	}
}

func TestStage20SchemaGuard(t *testing.T) {
	f := newNestingFixture(t)
	if e := postgres.ValidateSchema(t.Context(), f.db); e != nil {
		t.Fatal(e)
	}
	if _, e := f.db.Exec("DROP TABLE nesting_preview_operations"); e != nil {
		t.Fatal(e)
	}
	if e := postgres.ValidateSchema(t.Context(), f.db); e == nil {
		t.Fatal("missing required table accepted")
	}
}

func TestStage20CancelDuringSQLLockHTTP(t *testing.T) {
	for _, preview := range []bool{false, true} {
		t.Run(fmt.Sprintf("preview=%v", preview), func(t *testing.T) {
			f := newNestingFixture(t)
			hash, err := bcrypt.GenerateFromPassword([]byte("test-lock-password"), 4)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = f.db.Exec("UPDATE users SET password_hash=$1 WHERE id=1", string(hash)); err != nil {
				t.Fatal(err)
			}
			var schema string
			requireProcessGet(t, f.db, &schema, "SELECT current_schema()")
			dsn, err := url.Parse(os.Getenv("NESTING_TEST_DSN"))
			if err != nil {
				t.Fatal("invalid test DSN")
			}
			q := dsn.Query()
			q.Set("search_path", schema+",public")
			dsn.RawQuery = q.Encode()
			api := startProcess(t, dsn.String(), true)
			login := processRequest(api.base, "", "POST", "/auth/login", "", `{"identifier":"owner","password":"test-lock-password"}`)
			requireProcessStatus(t, login, 200)
			token := login.body["access_token"].(string)
			tx, err := f.db.Beginx()
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback() }()
			var holder, id int
			if err = tx.Get(&holder, "SELECT pg_backend_pid()"); err != nil {
				t.Fatal(err)
			}
			if err = tx.Get(&id, "SELECT id FROM projects WHERE id=1 FOR UPDATE"); err != nil {
				t.Fatal(err)
			}
			base := "/projects/1/surfaces/1/nesting"
			path, operation := base, base+"/operations/"
			key := uuid.NewString()
			operation += key
			if preview {
				path += "/preview"
				operation += "?kind=preview"
			}
			body := `{"algorithm":"baseline","patterns":[{"project_pattern_id":1,"quantity":1}],"allowed_rotations":[0],"keep_existing":true}`
			done := make(chan processReply, 1)
			go func() { done <- processRequest(api.base, token, "POST", path, key, body) }()
			processEventually(t, func() bool {
				var waiting bool
				requireProcessGet(t, f.db, &waiting, "SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE $1=ANY(pg_blocking_pids(pid)))", holder)
				return waiting
			})
			requireProcessStatus(t, processRequest(api.base, token, "DELETE", operation, "", ""), 204)
			result := <-done
			t.Logf("SQL_LOCK_CANCEL preview=%v HTTP=%d error=%v key=%s", preview, result.status, result.body["error"], key)
			requireProcessStatus(t, result, 409)
			if result.body["error"].(map[string]any)["code"] != "nesting_cancelled" {
				t.Fatal("incorrect cancellation code")
			}
			if err = tx.Rollback(); err != nil {
				t.Fatal(err)
			}
			terminal := processRequest(api.base, token, "GET", operation, "", "")
			requireProcessStatus(t, terminal, 200)
			if terminal.body["status"] != "failed" {
				t.Fatal("cancellation not terminal")
			}
			requireProcessStatus(t, processRequest(api.base, token, "POST", path, key, body), 409)
			f.counts(t, 0, 0)
		})
	}
}
