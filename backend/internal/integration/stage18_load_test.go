package integration

// This is an opt-in bounded loopback HTTP load harness, never part of cmd/api.
import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"io"
	"net/http"
	"net/url"
	"os"
	sf "server_nesting_optimizer/internal/geometry/simplefeatures"
	uc "server_nesting_optimizer/internal/usecase/nesting"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

type loadSample struct {
	StartedUnixMS int64
	Algorithm     string
	Endpoint      string
	Status        int
	Milliseconds  float64
	Error         bool
}
type loadPhase struct {
	StartedUnixMS int64
	Name          string
	Clients, APIs int
	Seconds       float64
	Samples       []loadSample
	Diagnostics   []json.RawMessage
}

func loadCall(api, token, method, path, key, body string) (loadSample, map[string]any) {
	start := time.Now()
	sample := loadSample{StartedUnixMS: start.UnixMilli(), Endpoint: method + " " + strings.Split(path, "?")[0]}
	if method == "POST" && strings.Contains(path, "/nesting") {
		var params struct {
			Algorithm string `json:"algorithm"`
		}
		_ = json.Unmarshal([]byte(body), &params)
		sample.Algorithm = params.Algorithm
	}
	// All target URLs originate in a listener allocated by startProcess.
	u, e := url.Parse(api)
	if e != nil || u.Hostname() != "127.0.0.1" {
		panic("non-loopback load target")
	}
	req, e := http.NewRequest(method, api+path, strings.NewReader(body))
	if e != nil {
		panic(e)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	if key != "" {
		req.Header.Set("Idempotency-Key", key)
	}
	r, e := (&http.Client{Timeout: 130 * time.Second}).Do(req)
	var out map[string]any
	if e != nil {
		sample.Error = true
	} else {
		sample.Status = r.StatusCode
		data, err := io.ReadAll(io.LimitReader(r.Body, 4<<20))
		_ = r.Body.Close()
		if err != nil {
			sample.Error = true
		}
		_ = json.Unmarshal(data, &out)
	}
	sample.Milliseconds = float64(time.Since(start).Microseconds()) / 1000
	return sample, out
}
func TestStage18MixedLoad(t *testing.T) {
	if os.Getenv("NESTING_STAGE18") != "1" {
		t.Skip("opt-in guarded mixed load")
	}
	f := newNestingFixture(t)
	hash, e := bcrypt.GenerateFromPassword([]byte("test-load-password"), 4)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = f.db.Exec("UPDATE users SET password_hash=$1", string(hash)); e != nil {
		t.Fatal(e)
	}
	// Unique surfaces per simulated client; second account remains a real owner-isolation control.
	if _, e = f.db.Exec("INSERT INTO project_surfaces(id,project_id,name,geometry) SELECT n,1,'load',geometry FROM project_surfaces CROSS JOIN generate_series(3,18) n WHERE id=1"); e != nil {
		t.Fatal(e)
	}
	if _, e = f.db.Exec(`SELECT setval(pg_get_serial_sequence('projects','id'),2,true);SELECT setval(pg_get_serial_sequence('project_surfaces','id'),18,true);SELECT setval(pg_get_serial_sequence('project_patterns','id'),1,true);INSERT INTO surfaces(id,user_id,name,geometry) SELECT 1,1,'load library',geometry FROM project_surfaces WHERE id=1;INSERT INTO patterns(id,user_id,name,geometry) SELECT 1,1,'load pattern',geometry FROM project_patterns WHERE id=1`); e != nil {
		t.Fatal(e)
	}
	var schema string
	requireProcessGet(t, f.db, &schema, "SELECT current_schema()")
	t.Log("LOAD_SCHEMA", schema)
	dsn, e := url.Parse(os.Getenv("NESTING_TEST_DSN"))
	if e != nil {
		t.Fatal("invalid test DSN")
	}
	q := dsn.Query()
	q.Set("search_path", schema+",public")
	q.Set("application_name", "stage18_"+schema)
	dsn.RawQuery = q.Encode()
	a, b := startProcess(t, dsn.String(), true), startProcess(t, dsn.String(), true)
	tokens := make([]string, 2)
	ownerRefresh := ""
	for i, user := range []string{"owner", "other"} {
		r := processRequest(a.base, "", "POST", "/auth/login", "", fmt.Sprintf(`{"identifier":%q,"password":"test-load-password"}`, user))
		requireProcessStatus(t, r, 200)
		tokens[i] = r.body["access_token"].(string)
		if i == 0 {
			ownerRefresh = r.body["refresh_token"].(string)
		}
	}
	var authMu sync.RWMutex
	ownerAccess := tokens[0]
	// Simulate tabs of one account using the same active session. Login Upsert is
	// intentionally one-session-per-user; refresh is coordinated like the UI.
	call := func(target, method, path, key, body string) (loadSample, map[string]any) {
		authMu.RLock()
		defer authMu.RUnlock()
		return loadCall(target, ownerAccess, method, path, key, body)
	}
	phases := []struct {
		name                   string
		clients, apis, seconds int
	}{{"smoke", 1, 1, 60}, {"baseline", 2, 1, 90}, {"step4", 4, 1, 30}, {"step8", 8, 1, 30}, {"step16", 16, 1, 30}, {"twoAPI", 8, 2, 30}, {"burst4", 4, 2, 20}, {"burst8", 8, 2, 20}, {"burst4tail", 4, 2, 20}}
	if os.Getenv("NESTING_STAGE18_RESUME_HIGH") == "1" {
		phases = []struct {
			name                   string
			clients, apis, seconds int
		}{{"step16", 16, 1, 30}, {"twoAPI", 8, 2, 30}, {"burst4", 4, 2, 20}, {"burst8", 8, 2, 20}, {"burst4tail", 4, 2, 20}}
	}
	if os.Getenv("NESTING_STAGE18_COMPARE") == "1" {
		phases = []struct {
			name                   string
			clients, apis, seconds int
		}{{"compare1API", 8, 1, 30}, {"compare2API", 8, 2, 30}}
	}
	if os.Getenv("NESTING_STAGE18_SMOKE") == "1" {
		phases = []struct {
			name                   string
			clients, apis, seconds int
		}{{"extended-smoke", 1, 1, 10}}
	}
	if os.Getenv("NESTING_STAGE18_SOAK") == "1" {
		phases = []struct {
			name                   string
			clients, apis, seconds int
		}{{"soak", 4, 2, 600}, {"idle", 0, 2, 300}}
	}
	var all []loadPhase
	for _, p := range phases {
		if !t.Run(p.name, func(t *testing.T) {
			auditLoadState(t, f)
			phase := loadPhase{Name: p.name, Clients: p.clients, APIs: p.apis}
			start := time.Now()
			phase.StartedUnixMS = start.UnixMilli()
			deadline := start.Add(time.Duration(p.seconds) * time.Second)
			var mu sync.Mutex
			var wg sync.WaitGroup
			for c := 0; c < p.clients; c++ {
				wg.Add(1)
				go func(c int) {
					defer wg.Done()
					api := a
					if p.apis == 2 && c%2 == 1 {
						api = b
					}
					base := fmt.Sprintf("/projects/1/surfaces/%d", c+1)
					record := func(s loadSample) { mu.Lock(); phase.Samples = append(phase.Samples, s); mu.Unlock() }
					maxIterations := 600
					if os.Getenv("NESTING_STAGE18_SOAK") == "1" {
						maxIterations = 5000
					}
					for i := 0; time.Now().Before(deadline) && i < maxIterations; i++ {
						method, path, key, body := "GET", base+"/placements", "", ""
						if i%40 == 9 {
							authMu.Lock()
							refreshBody, _ := json.Marshal(map[string]string{"refresh_token": ownerRefresh})
							sr, r := loadCall(api.base, "", "POST", "/auth/refresh", "", string(refreshBody))
							record(sr)
							if sr.Status != 200 {
								authMu.Unlock()
								t.Error("coordinated refresh failed")
								return
							}
							ownerAccess = r["access_token"].(string)
							ownerRefresh = r["refresh_token"].(string)
							authMu.Unlock()
							continue
						}
						if i%40 == 19 {
							sr, r := call(api.base, "POST", "/projects", "", `{"name":"load project copy","description":"synthetic"}`)
							record(sr)
							if sr.Status != 201 {
								t.Error("project create failed")
								return
							}
							project := fmt.Sprintf("/projects/%d", int64(r["id"].(float64)))
							for _, item := range []struct{ path, body string }{{"/surfaces", `{"source_surface_id":1,"scale":1}`}, {"/patterns", `{"source_pattern_id":1,"scale":1}`}} {
								sc, copy := call(api.base, "POST", project+item.path, "", item.body)
								record(sc)
								if sc.Status != 201 {
									t.Error("project instance copy failed")
									return
								}
								patch, _ := call(api.base, "PATCH", fmt.Sprintf("%s%s/%d", project, item.path, int64(copy["id"].(float64))), "", `{"name":"independent load copy"}`)
								record(patch)
								if patch.Status != 200 {
									t.Error("copy rename failed")
									return
								}
							}
							edited, _ := call(api.base, "PATCH", project, "", `{"name":"edited load copy"}`)
							record(edited)
							deleted, _ := call(api.base, "DELETE", project, "", "")
							record(deleted)
							if edited.Status != 200 || deleted.Status != 204 {
								t.Error("project edit/delete failed")
								return
							}
							continue
						}
						switch i % 10 {
						case 0:
							path = "/projects/1"
						case 1:
							path = base + "/nesting/history"
						case 2:
							path = "/users/me"
						case 3:
							path = "/projects/1"
							s, _ := loadCall(api.base, tokens[1], "GET", path, "", "")
							record(s)
							if s.Status != 404 {
								t.Error("owner isolation failed")
							}
							continue
						case 4:
							method = "POST"
							path = base + "/placements"
							body = `{"project_pattern_id":1,"x":80,"y":80,"rotation":0}`
						case 5, 6, 7:
							method = "POST"
							path = base + "/nesting"
							if i%10 != 5 {
								path += "/preview"
							}
							key = uuid.NewString()
							algorithm := "baseline"
							if os.Getenv("NESTING_STAGE18_SOAK") == "1" {
								algorithm = []string{"baseline", "nfp_greedy", "genetic", "simulated_annealing", "hybrid"}[(i/10)%5]
							}
							body = fmt.Sprintf(`{"algorithm":%q,"patterns":[{"project_pattern_id":1,"quantity":%d}],"allowed_rotations":[0,90],"keep_existing":%v}`, algorithm, 1+i%5, i%2 == 0)
						case 8:
							method = "GET"
							path = base + "/nesting/operations/" + uuid.NewString()
						}
						s, r := call(api.base, method, path, key, body)
						record(s)
						if s.Error || s.Status == 401 || s.Status >= 500 && s.Status != 503 {
							t.Errorf("unexpected HTTP outcome endpoint=%s status=%d network=%v", s.Endpoint, s.Status, s.Error)
							return
						}
						if i%10 == 7 && key != "" && s.Status == 200 {
							id, ok := r["preview_id"].(string)
							if !ok {
								t.Error("preview has no id")
								return
							}
							saveBody := strings.TrimSuffix(body, "}") + `,"preview_id":"` + id + `"}`
							saveKey := uuid.NewString()
							saved, _ := call(api.base, "POST", base+"/nesting", saveKey, saveBody)
							record(saved)
							if saved.Status == 503 {
								again, _ := call(api.base, "POST", base+"/nesting", saveKey, saveBody)
								record(again)
								if again.Status != 503 {
									t.Errorf("overload save replay status=%d", again.Status)
									return
								}
							}
							if saved.Status != 200 && saved.Status != 409 && saved.Status != 503 {
								t.Errorf("exact preview save status=%d network=%v", saved.Status, saved.Error)
								return
							}
						}
						if key != "" && s.Status == 200 {
							retryTarget := api.base
							if p.apis == 2 {
								retryTarget = b.base
								if api == b {
									retryTarget = a.base
								}
							}
							duplicate, _ := call(retryTarget, method, path, key, body)
							record(duplicate)
							if duplicate.Status != 200 {
								t.Error("durable replay failed")
							}
						}
						if method == "POST" && strings.HasSuffix(path, "/placements") && s.Status == 201 {
							id := int64(r["id"].(float64))
							d, _ := call(api.base, "DELETE", fmt.Sprintf("/projects/1/placements/%d", id), "", "")
							record(d)
							if d.Status != 204 {
								t.Error("manual delete failed")
							}
						}
						time.Sleep(150 * time.Millisecond)
					}
				}(c)
			}
			// Diagnostics continue during work; samples have no JWT or request geometry.
			for time.Now().Before(deadline) {
				for _, api := range []*processAPI{a, b} {
					res, e := (&http.Client{Timeout: 2 * time.Second}).Get(api.control + "/stats")
					if e != nil {
						t.Fatal("diagnostic unreachable")
					}
					data, _ := io.ReadAll(io.LimitReader(res.Body, 64<<10))
					_ = res.Body.Close()
					phase.Diagnostics = append(phase.Diagnostics, json.RawMessage(data))
				}
				time.Sleep(time.Second)
			}
			wg.Wait()
			// Final diagnostics include calculations that finish just after the deadline.
			for _, api := range []*processAPI{a, b} {
				response, e := (&http.Client{Timeout: 2 * time.Second}).Get(api.control + "/stats")
				if e != nil {
					t.Fatal("final diagnostic unavailable")
				}
				data, _ := io.ReadAll(io.LimitReader(response.Body, 64<<10))
				_ = response.Body.Close()
				phase.Diagnostics = append(phase.Diagnostics, json.RawMessage(data))
			}
			if p.name == "soak" {
				for _, api := range []*processAPI{a, b} {
					r := processRequest(api.control, "", "GET", "/profile-stop", "", "")
					requireProcessStatus(t, r, 204)
				}
			}
			phase.Seconds = time.Since(start).Seconds()
			auditLoadState(t, f)
			data, _ := json.Marshal(phase)
			t.Log("LOAD_PHASE_JSON " + string(data))
			all = append(all, phase)
		}) {
			t.Fatal("STOP: load gate failed; no higher concurrency")
		}
	}
	auditRetainedLoadResults(t, f)
	_ = all
}

func auditLoadState(t *testing.T, f nestingFixture) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	var n int
	for _, query := range []string{`SELECT count(*) FROM (SELECT project_surface_id,operation_key,count(*) FROM nesting_operations GROUP BY 1,2 HAVING count(*)>1) x`, `SELECT count(*) FROM pg_stat_activity WHERE datname=current_database() AND application_name LIKE 'stage18_%' AND state='idle in transaction'`, `SELECT count(*) FROM nesting_operations WHERE state='completed' AND response IS NULL`} {
		if e := f.db.GetContext(ctx, &n, query); e != nil {
			t.Fatal(e)
		}
		if n != 0 {
			t.Fatalf("SQL invariant failed: %d", n)
		}
	}
	rows, e := f.db.QueryxContext(ctx, `SELECT p.project_surface_id,ST_AsBinary(s.geometry),ST_AsBinary(pp.geometry),p.x,p.y,p.rotation FROM placements p JOIN project_surfaces s ON s.id=p.project_surface_id JOIN project_patterns pp ON pp.id=p.project_pattern_id ORDER BY p.project_surface_id,p.id`)
	if e != nil {
		t.Fatal(e)
	}
	type saved struct {
		surface      int64
		outer, inner []byte
		x, y, angle  float64
	}
	var data []saved
	for rows.Next() {
		var r saved
		if e = rows.Scan(&r.surface, &r.outer, &r.inner, &r.x, &r.y, &r.angle); e != nil {
			t.Fatal(e)
		}
		data = append(data, r)
	}
	_ = rows.Close()
	if e = rows.Err(); e != nil {
		t.Fatal(e)
	}
	codec, engine := sf.NewCodec(), sf.NewEngine()
	transformed := map[int64][][]byte{}
	for _, r := range data {
		p, e := codec.DecodeWKB(r.inner)
		if e != nil {
			t.Fatal(e)
		}
		p, e = engine.Transform(p, r.x, r.y, r.angle)
		if e != nil {
			t.Fatal(e)
		}
		w, e := codec.EncodeWKB(p)
		if e != nil {
			t.Fatal(e)
		}
		outer, e := codec.DecodeWKB(r.outer)
		if e != nil {
			t.Fatal(e)
		}
		ok, e := engine.CoveredBy(p, outer)
		if e != nil || !ok {
			t.Fatal("server coverage invariant")
		}
		var valid bool
		if e = f.db.GetContext(ctx, &valid, `SELECT ST_IsValid(ST_GeomFromWKB($1)) AND ST_CoveredBy(ST_GeomFromWKB($1),ST_GeomFromWKB($2))`, w, r.outer); e != nil || !valid {
			t.Fatal("independent PostGIS WKB coverage invariant")
		}
		for _, other := range transformed[r.surface] {
			if e = f.db.GetContext(ctx, &valid, `SELECT NOT ST_Relate(ST_GeomFromWKB($1),ST_GeomFromWKB($2),'T********')`, w, other); e != nil || !valid {
				t.Fatal("PostGIS overlap invariant")
			}
			op, e := codec.DecodeWKB(other)
			if e != nil {
				t.Fatal(e)
			}
			overlap, e := engine.InteriorsIntersect(p, op)
			if e != nil || overlap {
				t.Fatal("server overlap invariant")
			}
		}
		transformed[r.surface] = append(transformed[r.surface], w)
	}
	// Explicitly sample retained durable rows; do not delete tombstones.
	var sizes []struct {
		Name string
		Rows int64
	}
	for _, table := range []string{"nesting_operations", "nesting_preview_operations", "nesting_previews", "nesting_runs"} {
		var count int64
		requireProcessGet(t, f.db, &count, "SELECT count(*) FROM "+table)
		sizes = append(sizes, struct {
			Name string
			Rows int64
		}{table, count})
	}
	sort.Slice(sizes, func(i, j int) bool { return sizes[i].Name < sizes[j].Name })
	j, _ := json.Marshal(sizes)
	t.Logf("SQL_AUDIT placements=%d retained=%s", len(data), j)
}

func TestStage18GeometryProfiles(t *testing.T) {
	if os.Getenv("NESTING_STAGE18") != "1" {
		t.Skip("opt-in native profiles")
	}
	f := newNestingFixture(t)
	hash, e := bcrypt.GenerateFromPassword([]byte("test-load-password"), 4)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = f.db.Exec("UPDATE users SET password_hash=$1", string(hash)); e != nil {
		t.Fatal(e)
	}
	var schema string
	requireProcessGet(t, f.db, &schema, "SELECT current_schema()")
	dsn, e := url.Parse(os.Getenv("NESTING_TEST_DSN"))
	if e != nil {
		t.Fatal(e)
	}
	q := dsn.Query()
	q.Set("search_path", schema+",public")
	q.Set("application_name", "stage18_"+schema)
	dsn.RawQuery = q.Encode()
	api := startProcess(t, dsn.String(), true)
	login := processRequest(api.base, "", "POST", "/auth/login", "", `{"identifier":"owner","password":"test-load-password"}`)
	requireProcessStatus(t, login, 200)
	token := login.body["access_token"].(string)
	for _, profile := range []struct {
		name             string
		quantity, angles int
		algorithms       []string
	}{{"S", 3, 2, []string{"baseline", "nfp_greedy", "genetic", "simulated_annealing", "hybrid"}}, {"M", 20, 4, []string{"baseline", "nfp_greedy", "genetic", "simulated_annealing", "hybrid"}}, {"L", 50, 8, []string{"baseline", "nfp_greedy"}}, {"XL", 100, 16, []string{"baseline", "nfp_greedy"}}} {
		if profile.name != "S" {
			if _, e = f.db.Exec(`DELETE FROM placements; UPDATE project_patterns SET geometry=ST_GeomFromText('POLYGON((0 0,8 0,8 3,5 3,5 8,0 8,0 0),(1 1,1 2,2 2,2 1,1 1))') WHERE id=1`); e != nil {
				t.Fatal(e)
			}
		}
		for _, algorithm := range profile.algorithms {
			if !t.Run(profile.name+"/"+algorithm, func(t *testing.T) {
				angles := make([]float64, profile.angles)
				for i := range angles {
					angles[i] = float64(i) * 360 / float64(profile.angles)
				}
				encoded, _ := json.Marshal(angles)
				body := fmt.Sprintf(`{"algorithm":%q,"patterns":[{"project_pattern_id":1,"quantity":%d}],"allowed_rotations":%s,"keep_existing":false}`, algorithm, profile.quantity, encoded)
				result := make(chan loadSample, 1)
				go func() {
					s, _ := loadCall(api.base, token, "POST", "/projects/1/surfaces/1/nesting", uuid.NewString(), body)
					result <- s
				}()
				ticker := time.NewTicker(time.Second)
				defer ticker.Stop()
				var reads []loadSample
				for {
					select {
					case s := <-result:
						j, _ := json.Marshal(struct {
							Profile, Algorithm string
							Result             loadSample
							Reads              []loadSample
						}{profile.name, algorithm, s, reads})
						t.Log("GEOMETRY_PROFILE_JSON " + string(j))
						if s.Error || s.Status >= 500 && s.Status != 503 {
							t.Fatalf("unexpected profile failure status=%d network=%v", s.Status, s.Error)
						}
						if s.Status != 200 && s.Status != 400 && s.Status != 408 && s.Status != 503 {
							t.Fatalf("profile status %d", s.Status)
						}
						auditLoadState(t, f)
						return
					case <-ticker.C:
						s, _ := loadCall(api.base, token, "GET", "/projects/1/surfaces/1/nesting/history", "", "")
						reads = append(reads, s)
						if s.Error || s.Status != 200 {
							t.Fatal("read unavailable during heavy calculation")
						}
					}
				}
			}) {
				t.Fatal("STOP: geometry profile failed")
			}
		}
	}
}

// Audit every completed durable result, including layouts superseded later.
// Each row is checked with actual Go WKB, not a second SQL reconstruction.
func auditRetainedLoadResults(t *testing.T, f nestingFixture) {
	t.Helper()
	var completed, runs int
	requireProcessGet(t, f.db, &completed, "SELECT count(*) FROM nesting_operations WHERE state='completed'")
	requireProcessGet(t, f.db, &runs, "SELECT count(*) FROM nesting_runs")
	if completed != runs {
		t.Fatalf("atomic durable history: completed=%d runs=%d", completed, runs)
	}
	rows, e := f.db.Queryx(`SELECT o.response,ST_AsBinary(s.geometry) FROM nesting_operations o JOIN project_surfaces s ON s.id=o.project_surface_id WHERE o.state='completed' UNION ALL SELECT o.response,ST_AsBinary(s.geometry) FROM nesting_preview_operations o JOIN project_surfaces s ON s.id=o.project_surface_id WHERE o.state='completed'`)
	if e != nil {
		t.Fatal(e)
	}
	type record struct{ response, outer []byte }
	var records []record
	for rows.Next() {
		var r record
		if e = rows.Scan(&r.response, &r.outer); e != nil {
			t.Fatal(e)
		}
		records = append(records, r)
	}
	_ = rows.Close()
	if e = rows.Err(); e != nil {
		t.Fatal(e)
	}
	codec, engine := sf.NewCodec(), sf.NewEngine()
	placements := 0
	for _, r := range records {
		var state uc.OperationStatus
		if e = json.Unmarshal(r.response, &state); e != nil {
			t.Fatal(e)
		}
		if state.Status != "completed" || state.Result == nil {
			t.Fatal("durable state/response mismatch")
		}
		outer, e := codec.DecodeWKB(r.outer)
		if e != nil {
			t.Fatal(e)
		}
		var wkbs pq.ByteaArray
		for i, p := range state.Result.Placements {
			ok, e := engine.CoveredBy(p.Geometry, outer)
			if e != nil || !ok {
				t.Fatal("retained server coverage")
			}
			for _, q := range state.Result.Placements[:i] {
				overlap, e := engine.InteriorsIntersect(p.Geometry, q.Geometry)
				if e != nil || overlap {
					t.Fatal("retained server overlap")
				}
			}
			w, e := codec.EncodeWKB(p.Geometry)
			if e != nil {
				t.Fatal(e)
			}
			wkbs = append(wkbs, w)
			placements++
		}
		if len(wkbs) == 0 {
			continue
		}
		var valid bool
		e = f.db.Get(&valid, `WITH p AS (SELECT ST_GeomFromWKB(x) g,i FROM unnest($1::bytea[]) WITH ORDINALITY AS a(x,i)) SELECT (SELECT bool_and(ST_IsValid(g) AND ST_CoveredBy(g,ST_GeomFromWKB($2))) FROM p) AND NOT EXISTS(SELECT 1 FROM p a JOIN p b ON a.i<b.i WHERE ST_Relate(a.g,b.g,'T********'))`, wkbs, r.outer)
		if e != nil || !valid {
			t.Fatal("retained independent PostGIS WKB oracle")
		}
	}
	t.Logf("ALL_RETAINED_WKB results=%d placements=%d histories=%d PASS", len(records), placements, runs)
}

func TestStage18RetainedLoadAudit(t *testing.T) {
	if os.Getenv("NESTING_STAGE18") != "1" {
		t.Skip("opt-in own lab read-only retained audit")
	}
	db, e := sqlx.Connect("postgres", os.Getenv("NESTING_TEST_DSN"))
	if e != nil {
		t.Fatal("test connection failed")
	}
	defer func() { _ = db.Close() }()
	var actual, user string
	requireProcessGet(t, db, &actual, "SELECT current_database()")
	requireProcessGet(t, db, &user, "SELECT current_user")
	if actual != "nesting_test" || user != os.Getenv("NESTING_EXPECTED_TEST_USER") || user == "" {
		t.Fatal("unsafe database")
	}
	var names []string
	if e = db.Select(&names, `SELECT nspname FROM pg_namespace WHERE nspname LIKE 'test\_%' ESCAPE '\' ORDER BY nspname`); e != nil {
		t.Fatal(e)
	}
	audited := 0
	for _, name := range names {
		if selected := os.Getenv("NESTING_STAGE18_AUDIT_SCHEMA"); selected != "" && selected != name {
			continue
		}
		var exists bool
		requireProcessGet(t, db, &exists, "SELECT to_regclass($1) IS NOT NULL", name+".project_surfaces")
		if !exists {
			continue
		}
		var count int
		requireProcessGet(t, db, &count, "SELECT count(*) FROM "+pq.QuoteIdentifier(name)+".project_surfaces WHERE name='load'")
		if count != 16 && os.Getenv("NESTING_STAGE18_AUDIT_SCHEMA") == "" {
			continue
		}
		dsn, e := url.Parse(os.Getenv("NESTING_TEST_DSN"))
		if e != nil {
			t.Fatal("invalid DSN")
		}
		q := dsn.Query()
		q.Set("search_path", name+",public")
		q.Set("default_transaction_read_only", "on")
		dsn.RawQuery = q.Encode()
		scoped, e := sqlx.Connect("postgres", dsn.String())
		if e != nil {
			t.Fatal("read-only audit connection failed")
		}
		t.Log("RETAINED_LOAD_SCHEMA", name)
		auditRetainedLoadResults(t, nestingFixture{db: scoped})
		_ = scoped.Close()
		audited++
	}
	if audited == 0 {
		t.Fatal("no retained load fixtures found")
	}
	t.Logf("retained load schemas audited=%d", audited)
}

// Positive two-owner isolation, complementing the mixed load's foreign-owner
// negative traffic. Same UUID and simultaneous requests must remain independent.
func TestStage18IndependentOwners(t *testing.T) {
	if os.Getenv("NESTING_STAGE18") != "1" {
		t.Skip("opt-in real HTTP independent owners")
	}
	f := newNestingFixture(t)
	hash, e := bcrypt.GenerateFromPassword([]byte("test-load-password"), 4)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = f.db.Exec(`UPDATE users SET password_hash=$1`, string(hash)); e != nil {
		t.Fatal(e)
	}
	if _, e = f.db.Exec(`INSERT INTO project_surfaces(id,project_id,name,geometry) SELECT 3,2,'Other owner',geometry FROM project_surfaces WHERE id=1;INSERT INTO project_patterns(id,project_id,name,geometry) SELECT 2,2,'Other owner',geometry FROM project_patterns WHERE id=1`); e != nil {
		t.Fatal(e)
	}
	var schema string
	requireProcessGet(t, f.db, &schema, "SELECT current_schema()")
	dsn, e := url.Parse(os.Getenv("NESTING_TEST_DSN"))
	if e != nil {
		t.Fatal(e)
	}
	q := dsn.Query()
	q.Set("search_path", schema+",public")
	dsn.RawQuery = q.Encode()
	a, b := startProcess(t, dsn.String(), true), startProcess(t, dsn.String(), true)
	tokens := make([]string, 2)
	for i, user := range []string{"owner", "other"} {
		login := processRequest(a.base, "", "POST", "/auth/login", "", fmt.Sprintf(`{"identifier":%q,"password":"test-load-password"}`, user))
		requireProcessStatus(t, login, 200)
		tokens[i] = login.body["access_token"].(string)
	}
	paths := []string{"/projects/1/surfaces/1/nesting", "/projects/2/surfaces/3/nesting"}
	apis := []*processAPI{a, b}
	for _, preview := range []bool{false, true} {
		key := uuid.NewString()
		bodies := make([]string, 2)
		results := make([]processReply, 2)
		var wg sync.WaitGroup
		for i := 0; i < 2; i++ {
			bodies[i] = fmt.Sprintf(`{"algorithm":"baseline","patterns":[{"project_pattern_id":%d,"quantity":2}],"allowed_rotations":[0,90],"keep_existing":%v}`, i+1, preview)
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				path := paths[i]
				if preview {
					path += "/preview"
				}
				results[i] = processRequest(apis[i].base, tokens[i], "POST", path, key, bodies[i])
			}(i)
		}
		wg.Wait()
		for i, r := range results {
			requireProcessStatus(t, r, 200)
			path := paths[i]
			if preview {
				path += "/preview"
			}
			replayed := processRequest(apis[1-i].base, tokens[i], "POST", path, key, bodies[i])
			requireProcessStatus(t, replayed, 200)
			left, _ := json.Marshal(r.body)
			right, _ := json.Marshal(replayed.body)
			if string(left) != string(right) {
				t.Fatal("cross-process owner replay changed result")
			}
			requireProcessStatus(t, processRequest(apis[i].base, tokens[1-i], "GET", paths[i]+"/operations/"+key, "", ""), 404)
			if preview {
				save := strings.TrimSuffix(bodies[i], "}") + `,"preview_id":"` + r.body["preview_id"].(string) + `"}`
				requireProcessStatus(t, processRequest(apis[1-i].base, tokens[i], "POST", paths[i], key, save), 200)
				requireProcessStatus(t, processRequest(apis[i].base, tokens[i], "POST", paths[i], key, save), 200)
			}
		}
	}
	f.counts(t, 8, 4)
	auditLoadState(t, f)
	auditRetainedLoadResults(t, f)
}

func TestStage18HeterogeneousOccupied(t *testing.T) {
	if os.Getenv("NESTING_STAGE18") != "1" {
		t.Skip("opt-in bounded heterogeneous native HTTP")
	}
	f := newNestingFixture(t)
	hash, e := bcrypt.GenerateFromPassword([]byte("test-load-password"), 4)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = f.db.Exec("UPDATE users SET password_hash=$1", string(hash)); e != nil {
		t.Fatal(e)
	}
	if _, e = f.db.Exec(`INSERT INTO project_patterns(id,project_id,name,geometry) VALUES(2,1,'concave-hole',ST_GeomFromText('POLYGON((0 0,8 0,8 3,5 3,5 8,0 8,0 0),(1 1,1 2,2 2,2 1,1 1))')),(3,1,'triangle',ST_GeomFromText('POLYGON((0 0,8 0,0 8,0 0))'))`); e != nil {
		t.Fatal(e)
	}
	var schema string
	requireProcessGet(t, f.db, &schema, "SELECT current_schema()")
	dsn, e := url.Parse(os.Getenv("NESTING_TEST_DSN"))
	if e != nil {
		t.Fatal(e)
	}
	q := dsn.Query()
	q.Set("search_path", schema+",public")
	dsn.RawQuery = q.Encode()
	api := startProcess(t, dsn.String(), true)
	login := processRequest(api.base, "", "POST", "/auth/login", "", `{"identifier":"owner","password":"test-load-password"}`)
	requireProcessStatus(t, login, 200)
	token := login.body["access_token"].(string)
	manual := processRequest(api.base, token, "POST", "/projects/1/surfaces/1/placements", "", `{"project_pattern_id":1,"x":80,"y":80,"rotation":0}`)
	requireProcessStatus(t, manual, 201)
	oldID := int64(manual.body["id"].(float64))
	key := uuid.NewString()
	path := "/projects/1/surfaces/1/nesting"
	body := `{"algorithm":"nfp_greedy","patterns":[{"project_pattern_id":1,"quantity":15},{"project_pattern_id":2,"quantity":20},{"project_pattern_id":3,"quantity":15}],"allowed_rotations":[0,90,180,270],"keep_existing":true}`
	type reply struct {
		s loadSample
		r map[string]any
	}
	done := make(chan reply, 1)
	go func() { s, r := loadCall(api.base, token, "POST", path, key, body); done <- reply{s, r} }()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var reads []loadSample
	for {
		select {
		case result := <-done:
			j, _ := json.Marshal(struct {
				Profile, Algorithm string
				Result             loadSample
				Reads              []loadSample
			}{"L-50-three-shapes-occupied", "nfp_greedy", result.s, reads})
			t.Log("GEOMETRY_PROFILE_JSON " + string(j))
			if result.s.Error || result.s.Status != 200 && result.s.Status != 408 {
				t.Fatalf("heterogeneous outcome=%d network=%v", result.s.Status, result.s.Error)
			}
			if result.s.Status == 200 {
				items := result.r["placements"].([]any)
				f.counts(t, len(items)+1, 1)
			} else {
				f.counts(t, 1, 0)
			}
			var count int
			requireProcessGet(t, f.db, &count, "SELECT count(*) FROM placements WHERE id=$1 AND x=80 AND y=80", oldID)
			if count != 1 {
				t.Fatal("occupied placement was lost")
			}
			replay, _ := loadCall(api.base, token, "POST", path, key, body)
			if replay.Status != result.s.Status {
				t.Fatal("heterogeneous replay changed outcome")
			}
			auditLoadState(t, f)
			auditRetainedLoadResults(t, f)
			return
		case <-ticker.C:
			s, _ := loadCall(api.base, token, "GET", path+"/operations/"+key, "", "")
			reads = append(reads, s)
			if s.Error || s.Status != 200 {
				t.Fatal("status blocked during heterogeneous calculation")
			}
		}
	}
}
