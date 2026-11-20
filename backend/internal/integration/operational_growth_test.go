package integration

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"os"
	"runtime"
	"server_nesting_optimizer/internal/repository/postgres"
	usecase "server_nesting_optimizer/internal/usecase/nesting"
	"server_nesting_optimizer/pkg/apperror"
	"sort"
	"testing"
	"time"
)

func TestOperationalGrowth(t *testing.T) {
	if os.Getenv("NESTING_OPERATIONAL_EXPERIMENT") != "1" {
		t.Skip("explicit bounded operational experiment")
	}
	f := newNestingFixture(t)
	f.db.SetMaxOpenConns(8)
	f.db.SetMaxIdleConns(4)
	close(f.gate.release)
	ctx, cancel := context.WithTimeout(t.Context(), 240*time.Second)
	defer cancel()
	beforeGoroutines := runtime.NumGoroutine()
	started := time.Now()
	var schema string
	requireProcessGet(t, f.db, &schema, "SELECT current_schema()")
	t.Log("OPERATION_SCHEMA " + schema)
	durations := make([]float64, 0, 10000)
	var firstPreview, firstSave, latestPreview, latestSave string
	latestInput := f.input
	var latestResult usecase.RunNestingOutput
	var firstResult usecase.RunNestingOutput
	hash, err := usecase.Fingerprint(f.input)
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 5000; i++ {
		previewKey := uuid.NewString()
		begin := time.Now()
		preview, err := f.uc.PreviewWithKey(ctx, f.input, previewKey)
		if err != nil {
			t.Fatalf("preview %d: %v", i, err)
		}
		durations = append(durations, float64(time.Since(begin))/float64(time.Millisecond))
		input := f.input
		if i%10 == 0 {
			input.PreviewID = preview.PreviewID
		}
		saveKey := uuid.NewString()
		begin = time.Now()
		result, err := f.uc.ExecuteWithKey(ctx, input, saveKey)
		if err != nil {
			t.Fatalf("save %d: %v", i, err)
		}
		durations = append(durations, float64(time.Since(begin))/float64(time.Millisecond))
		if i == 1 {
			firstPreview = previewKey
			firstSave = saveKey
			firstResult = result
		}
		if i == 5000 {
			latestPreview, latestSave, latestInput, latestResult = previewKey, saveKey, input, result
		}
		if i%50 == 0 {
			key, token := uuid.New(), uuid.New()
			_, err = f.uc.PreviewOperations.Reserve(ctx, 1, key, hash[:], token, time.Now().Add(-time.Second))
			if err != nil {
				t.Fatal(err)
			}
			_, state, err := f.uc.PreviewOperations.Find(ctx, 1, key)
			if err != nil || state == nil || state.Status != "failed" {
				t.Fatal("orphan recovery failed")
			}
		}
		if i == 100 || i == 1000 || i == 5000 {
			operationalSnapshot(t, f, i, started, durations)
			var ready int
			requireProcessGet(t, f.db, &ready, "SELECT count(*) FROM nesting_previews")
			if ready > 10 {
				t.Fatal("ready quota exceeded")
			}
		}
	}
	f.counts(t, 1, 5000)
	if f.gate.calls.Load() != 9500 {
		t.Fatalf("actual Optimize count %d want9500 (500 exact-preview saves)", f.gate.calls.Load())
	}
	beforeCalls := f.gate.calls.Load()
	// Old successful save is historical; replay must not overwrite the newer layout.
	replay, err := f.uc.ExecuteWithKey(ctx, f.input, firstSave)
	if err != nil {
		t.Fatal(err)
	}
	originalJSON, _ := json.Marshal(firstResult)
	replayJSON, _ := json.Marshal(replay)
	if string(originalJSON) != string(replayJSON) {
		t.Fatal("permanent save replay changed")
	}
	if f.gate.calls.Load() != beforeCalls {
		t.Fatal("old save recalculated")
	}
	_, err = f.uc.PreviewWithKey(ctx, f.input, firstPreview)
	expectCode(t, err, apperror.Code("preview_expired"))
	if f.gate.calls.Load() != beforeCalls {
		t.Fatal("evicted preview recalculated")
	}
	f.counts(t, 1, 5000)
	// Synthetic expired payloads deliberately bypass per-user Store quota ONLY to
	// exercise a global backlog. They are not claimed as 5000 API-created previews.
	_, err = f.db.ExecContext(ctx, `INSERT INTO nesting_previews(id,user_id,project_id,project_surface_id,payload,saved_result,expires_at)
 SELECT gen_random_uuid(),p.user_id,p.project_id,p.project_surface_id,p.payload,p.saved_result,clock_timestamp()-interval '1 second'
 FROM nesting_previews p CROSS JOIN generate_series(1,5000) WHERE p.id=(SELECT id FROM nesting_previews WHERE saved_result IS NOT NULL LIMIT 1)`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.db.ExecContext(ctx, "UPDATE nesting_previews SET expires_at=clock_timestamp()-interval '1 second'"); err != nil {
		t.Fatal(err)
	}
	// Synthetic crashed-worker claims, no terminal row is deleted or reused.
	_, err = f.db.ExecContext(ctx, `INSERT INTO nesting_preview_operations(project_surface_id,operation_key,request_hash,response,state,lease_token,lease_expires_at)
 SELECT 1,gen_random_uuid(),$1,'{"Status":"running"}'::jsonb,'running',gen_random_uuid(),clock_timestamp()-interval '1 second' FROM generate_series(1,2500)`, hash[:])
	if err != nil {
		t.Fatal(err)
	}
	// Hold only our fixture row: SKIP LOCKED must drain the rest without waiting.
	locked, err := f.db.BeginTxx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = locked.Rollback() }()
	if _, err = locked.ExecContext(ctx, "SELECT id FROM nesting_previews ORDER BY id LIMIT 1 FOR UPDATE"); err != nil {
		t.Fatal(err)
	}
	repo := postgres.NewNestingPreviewRepository(f.scoped)
	deleted := int64(0)
	batch := []float64{}
	for {
		start := time.Now()
		n, err := repo.PurgeExpired(ctx, 1000)
		if err != nil {
			t.Fatal(err)
		}
		batch = append(batch, float64(time.Since(start))/float64(time.Millisecond))
		deleted += n
		if n > 1000 {
			t.Fatal("unbounded cleanup batch")
		}
		if n == 0 {
			break
		}
	}
	if deleted != 5009 {
		t.Fatalf("locked cleanup deleted=%d want5009", deleted)
	}
	var remaining int
	requireProcessGet(t, f.db, &remaining, "SELECT count(*) FROM nesting_previews")
	if remaining != 1 {
		t.Fatalf("locked row was lost: %d", remaining)
	}
	if err = locked.Rollback(); err != nil {
		t.Fatal(err)
	}
	n, err := repo.PurgeExpired(ctx, 1000)
	if err != nil || n != 1 {
		t.Fatalf("unlocked cleanup n=%d err=%v", n, err)
	}
	deleted += n
	if deleted != 5010 {
		t.Fatalf("expired payloads deleted=%d want5010", deleted)
	}
	var running, operations int
	requireProcessGet(t, f.db, &running, "SELECT count(*) FROM nesting_preview_operations WHERE state='running'")
	requireProcessGet(t, f.db, &operations, "SELECT count(*) FROM nesting_preview_operations")
	if running != 0 || operations != 7600 {
		t.Fatalf("cleanup damaged tombstones/running: %d/%d", operations, running)
	}
	// Payload TTL must neither invalidate a committed save nor resurrect preview.
	_, err = f.uc.PreviewWithKey(ctx, f.input, latestPreview)
	expectCode(t, err, apperror.Code("preview_expired"))
	lastReplay, err := f.uc.ExecuteWithKey(ctx, latestInput, latestSave)
	if err != nil {
		t.Fatal(err)
	}
	lastJSON, err := json.Marshal(latestResult)
	if err != nil {
		t.Fatal(err)
	}
	lastReplayJSON, err := json.Marshal(lastReplay)
	if err != nil || string(lastJSON) != string(lastReplayJSON) {
		t.Fatal("save changed after preview TTL")
	}
	f.counts(t, 1, 5000)
	operationalSnapshot(t, f, 5000, started, durations)
	if f.gate.calls.Load() != beforeCalls {
		t.Fatal("cleanup invoked optimizer")
	}
	if runtime.NumGoroutine() > beforeGoroutines+10 {
		t.Fatal("operation goroutines leaked")
	}
	encoded, _ := json.Marshal(map[string]any{"deleted_payloads": deleted, "batch_ms": batch, "preview_tombstones": operations, "remaining_running": running, "actual_native_optimize_calls": beforeCalls, "total_ms": float64(time.Since(started)) / float64(time.Millisecond), "goroutines_before": beforeGoroutines, "goroutines_after": runtime.NumGoroutine()})
	t.Log("CLEANUP_ROW " + string(encoded))
}

func operationalSnapshot(t *testing.T, f nestingFixture, n int, started time.Time, durations []float64) {
	t.Helper()
	tables := map[string]any{}
	for _, name := range []string{"nesting_operations", "nesting_preview_operations"} {
		type countRow struct {
			State string `db:"state"`
			Count int    `db:"count"`
		}
		var states []countRow
		if err := f.db.Select(&states, "SELECT state,count(*) AS count FROM "+name+" GROUP BY state ORDER BY state"); err != nil {
			t.Fatal(err)
		}
		var logical, stored, total, index int64
		requireProcessGet(t, f.db, &logical, "SELECT coalesce(sum(octet_length(response::text)),0) FROM "+name)
		requireProcessGet(t, f.db, &stored, "SELECT coalesce(sum(pg_column_size(response)),0) FROM "+name)
		requireProcessGet(t, f.db, &total, "SELECT pg_total_relation_size($1::regclass)", name)
		requireProcessGet(t, f.db, &index, "SELECT pg_indexes_size($1::regclass)", name)
		tables[name] = map[string]any{"states": states, "logical_json_bytes": logical, "stored_json_bytes": stored, "total_relation_bytes": total, "index_bytes": index}
	}
	var ready int
	var payload, saved int64
	requireProcessGet(t, f.db, &ready, "SELECT count(*) FROM nesting_previews")
	requireProcessGet(t, f.db, &payload, "SELECT coalesce(sum(pg_column_size(payload)),0) FROM nesting_previews")
	requireProcessGet(t, f.db, &saved, "SELECT coalesce(sum(pg_column_size(saved_result)),0) FROM nesting_previews")
	values := append([]float64(nil), durations...)
	sort.Float64s(values)
	stats := f.db.Stats()
	row := map[string]any{"checkpoint_per_kind": n, "elapsed_ms": float64(time.Since(started)) / float64(time.Millisecond), "operations": tables, "ready_count": ready, "ready_payload_bytes": payload, "ready_saved_result_bytes": saved, "uc_latency_p50_ms": values[len(values)/2], "uc_latency_p95_ms": values[len(values)*95/100], "uc_latency_max_ms": values[len(values)-1], "pool_open": stats.OpenConnections, "pool_in_use": stats.InUse, "pool_wait_count": stats.WaitCount, "goroutines": runtime.NumGoroutine()}
	encoded, _ := json.Marshal(row)
	t.Log("GROWTH_ROW " + string(encoded))
}
