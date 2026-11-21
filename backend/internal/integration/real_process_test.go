package integration

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
	"server_nesting_optimizer/internal/app/container"
	"server_nesting_optimizer/internal/config"
	sf "server_nesting_optimizer/internal/geometry/simplefeatures"
	optimizer "server_nesting_optimizer/internal/nesting"
	"server_nesting_optimizer/internal/repository/postgres"
	transport "server_nesting_optimizer/internal/transport/http"
	uc "server_nesting_optimizer/internal/usecase/nesting"
)

// All controls exist exclusively in the integration test binary. Each child
// builds the real application, JWT/session middleware and an independent pool.
// SQL gates synchronize different OS processes without Go shared memory.
type processOptimizer struct{ db *sqlx.DB }

func (p processOptimizer) Optimize(ctx context.Context, problem optimizer.Problem) (optimizer.Result, error) {
	var id int
	if err := p.db.GetContext(ctx, &id, "INSERT INTO process_calls(pid) VALUES($1) RETURNING id", os.Getpid()); err != nil {
		return optimizer.Result{}, err
	}
	defer func() {
		c, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_, _ = p.db.ExecContext(c, "UPDATE process_calls SET ended=clock_timestamp() WHERE id=$1", id)
	}()
	for {
		var released bool
		if err := p.db.GetContext(ctx, &released, "SELECT released FROM process_gate"); err != nil {
			return optimizer.Result{}, err
		}
		if released {
			break
		}
		select {
		case <-ctx.Done():
			return optimizer.Result{}, ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
	return optimizer.NewBaselineOptimizer(sf.NewEngine()).Optimize(ctx, problem)
}

type processUOW struct {
	uc.UnitOfWork
	db *sqlx.DB
}

func (p processUOW) WithinTransaction(ctx context.Context, fn func(uc.TransactionRepositories) error) error {
	var hold bool
	if err := p.db.GetContext(ctx, &hold, "SELECT commit_hold FROM process_gate"); err != nil {
		return err
	}
	if hold {
		if _, err := p.db.ExecContext(ctx, "INSERT INTO process_commits(pid) VALUES($1)", os.Getpid()); err != nil {
			return err
		}
		for hold {
			if err := p.db.GetContext(ctx, &hold, "SELECT commit_hold FROM process_gate"); err != nil {
				return err
			}
			if hold {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(5 * time.Millisecond):
				}
			}
		}
	}
	return p.UnitOfWork.WithinTransaction(ctx, fn)
}

type processOperations struct {
	uc.OperationRepository
	db *sqlx.DB
}

func (p processOperations) Reserve(ctx context.Context, surface int64, key uuid.UUID, hash []byte, token uuid.UUID, expires time.Time) (bool, error) {
	owned, err := p.OperationRepository.Reserve(ctx, surface, key, hash, token, expires)
	if err != nil || !owned {
		return owned, err
	}
	for {
		var hold bool
		if err = p.db.GetContext(ctx, &hold, "SELECT reserve_hold FROM process_gate"); err != nil {
			return owned, err
		}
		if !hold {
			return owned, nil
		}
		select {
		case <-ctx.Done():
			return owned, ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
}

// Simulates loss of THIS child's DB network only; unrelated processes and
// PostgreSQL remain untouched. All tracked sockets were created by this helper.
type processDialer struct {
	mu          sync.Mutex
	blocked     bool
	connections []net.Conn
}

func (d *processDialer) Dial(network, address string) (net.Conn, error) {
	return d.DialTimeout(network, address, 3*time.Second)
}
func (d *processDialer) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.blocked {
		return nil, errors.New("test-only SQL network unavailable")
	}
	c, err := net.DialTimeout(network, address, timeout)
	if err == nil {
		d.connections = append(d.connections, c)
	}
	return c, err
}
func (d *processDialer) disconnect() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.blocked = true
	for _, c := range d.connections {
		_ = c.Close()
	}
}

func TestRealProcessWorker(t *testing.T) {
	if os.Getenv("NESTING_PROCESS_CHILD") != "1" {
		t.Skip("subprocess helper")
	}
	connectionURL, err := url.Parse(os.Getenv("NESTING_PROCESS_DSN"))
	if err != nil {
		t.Fatal("invalid child DSN")
	}
	parameters := connectionURL.Query()
	parameters.Set("application_name", parameters.Get("application_name")+"_pid"+strconv.Itoa(os.Getpid()))
	connectionURL.RawQuery = parameters.Encode()
	connector, err := pq.NewConnector(connectionURL.String())
	if err != nil {
		t.Fatal("invalid child connector")
	}
	network := &processDialer{}
	connector.Dialer(network)
	db := sqlx.NewDb(sql.OpenDB(connector), "postgres")
	defer func() { _ = db.Close() }()
	var database, user, postgis string
	if err = db.Get(&database, "SELECT current_database()"); err != nil || database != "nesting_test" {
		t.Fatal("unsafe child database")
	}
	if err = db.Get(&user, "SELECT current_user"); err != nil {
		t.Fatal(err)
	}
	if err = db.Get(&postgis, "SELECT postgis_version()"); err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(4)
	cfg := config.Config{JWT: config.JWTConfig{Issuer: "process-tests", AccessSecret: "process-test-access", RefreshSecret: "process-test-refresh", AccessTTLMin: 15, RefreshTTLMin: 60}, Security: config.SecurityConfig{BcryptCost: 4}, Nesting: config.NestingConfig{Workers: 2, Queue: 4}}
	deps, err := container.New(db, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if os.Getenv("NESTING_PROCESS_NATIVE") != "1" {
		scoped := &postgres.ScopedDB{DB: db}
		deps.RunNestingUseCase = newFixtureUseCase(t, scoped, processOptimizer{db}, processUOW{UnitOfWork: postgres.NewNestingUnitOfWork(db, sf.NewCodec()), db: db})
		deps.RunNestingUseCase.Operations = processOperations{deps.RunNestingUseCase.Operations, db}
		deps.RunNestingUseCase.PreviewOperations = processOperations{deps.RunNestingUseCase.PreviewOperations, db}
	}
	app := transport.NewRouter(deps, cfg)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	var stopProfile func()
	if directory := os.Getenv("NESTING_LAB_PROFILE_DIR"); directory != "" {
		file, e := os.OpenFile(filepath.Join(directory, fmt.Sprintf("cpu-api-%d.pprof", os.Getpid())), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			t.Fatal("test CPU profile output unavailable")
		}
		if e = pprof.StartCPUProfile(file); e != nil {
			t.Fatal(e)
		}
		var once sync.Once
		stopProfile = func() { once.Do(func() { pprof.StopCPUProfile(); _ = file.Close() }) }
		defer stopProfile()
	}
	// Separate test-only diagnostics; never compiled into cmd/api.
	var janitorRuns, janitorDeleted, janitorErrors atomic.Int64
	diagnostic := http.NewServeMux()
	diagnostic.HandleFunc("/stats", func(w http.ResponseWriter, _ *http.Request) {
		var memory runtime.MemStats
		runtime.ReadMemStats(&memory)
		_ = json.NewEncoder(w).Encode(struct {
			sql.DBStats
			Goroutines                                 int
			HeapAlloc                                  uint64
			TotalAlloc                                 uint64
			HeapSys                                    uint64
			NumGC                                      uint32
			PauseTotalNs                               uint64
			SampleUnixMS                               int64
			Optimizer                                  map[string]any
			JanitorRuns, JanitorDeleted, JanitorErrors int64
			FDs                                        int
			Process                                    processResourceUsage
			Limiter                                    map[string]any
		}{db.Stats(), runtime.NumGoroutine(), memory.HeapAlloc, memory.TotalAlloc, memory.HeapSys, memory.NumGC, memory.PauseTotalNs, time.Now().UnixMilli(), readProcessOptimizer(deps.RunNestingUseCase), janitorRuns.Load(), janitorDeleted.Load(), janitorErrors.Load(), processFDCount(), readProcessUsage(), readProcessLimiter(deps.RunNestingUseCase.Limiter)})
	})
	diagnostic.HandleFunc("/profile-stop", func(w http.ResponseWriter, _ *http.Request) {
		if stopProfile != nil {
			stopProfile()
		}
		w.WriteHeader(204)
	})
	diagnostic.HandleFunc("/network-off", func(w http.ResponseWriter, _ *http.Request) { network.disconnect(); w.WriteHeader(204) })
	diagnostic.HandleFunc("/close-pool", func(w http.ResponseWriter, _ *http.Request) { _ = db.Close(); w.WriteHeader(204) })
	control := &http.Server{Handler: diagnostic, ReadHeaderTimeout: time.Second}
	ctl, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() { _ = control.Serve(ctl) }()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	// Match app.cleanupPreviews scheduling/bounds for the production-like soak.
	// The production function is private; this helper calls the SAME repository
	// purge,1000-row batches,5s pass deadline, immediate then once per minute.
	if os.Getenv("NESTING_STAGE18_SOAK") == "1" {
		cleanupCtx, cancel := context.WithCancel(ctx)
		done := make(chan struct{})
		go func() {
			defer close(done)
			ticker := time.NewTicker(time.Minute)
			defer ticker.Stop()
			repo := postgres.NewNestingPreviewRepository(db)
			for {
				pass, stop := context.WithTimeout(cleanupCtx, 5*time.Second)
				janitorRuns.Add(1)
				for pass.Err() == nil {
					n, e := repo.PurgeExpired(pass, 1000)
					if e != nil {
						if cleanupCtx.Err() == nil {
							janitorErrors.Add(1)
						}
						break
					}
					janitorDeleted.Add(n)
					if n < 1000 {
						break
					}
				}
				stop()
				select {
				case <-cleanupCtx.Done():
					return
				case <-ticker.C:
				}
			}
		}()
		defer func() { cancel(); <-done }()
	}
	go func() {
		<-ctx.Done()
		c, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = app.ShutdownWithContext(c)
		_ = control.Shutdown(c)
	}()
	fmt.Printf("PROCESS_READY %d http://%s http://%s %s %s\n", os.Getpid(), listener.Addr(), ctl.Addr(), database, postgis)
	if err = app.Listener(listener); err != nil {
		t.Fatal(err)
	}
}

type processAPI struct {
	cmd           *exec.Cmd
	base, control string
	pid           int
	stopped       bool
}

func startProcess(t *testing.T, dsn string, native bool) *processAPI {
	t.Helper()
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if explicit := os.Getenv("NESTING_TEST_PROCESS_BINARY"); explicit != "" {
		binary = explicit
	}
	childTimeout := "180s"
	if os.Getenv("NESTING_STAGE18") == "1" {
		childTimeout = "30m"
	}
	args := []string{"-test.run=^TestRealProcessWorker$", "-test.timeout=" + childTimeout}
	cmd := exec.Command(binary, args...)
	if loader := os.Getenv("NESTING_TEST_CHILD_LOADER"); loader != "" {
		cmd = exec.Command(loader, append([]string{"--library-path", os.Getenv("NESTING_TEST_CHILD_LIBRARY"), binary}, args...)...)
	}
	cmd.Env = append(os.Environ(), "NESTING_PROCESS_CHILD=1", "NESTING_PROCESS_DSN="+dsn, "GOMAXPROCS=2", "GOMEMLIMIT=384MiB")
	if native {
		cmd.Env = append(cmd.Env, "NESTING_PROCESS_NATIVE=1")
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stderr = &bytes.Buffer{}
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	api := &processAPI{cmd: cmd, pid: cmd.Process.Pid}
	t.Cleanup(func() { api.stop(false) })
	ready := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			if strings.HasPrefix(scanner.Text(), "PROCESS_READY ") {
				ready <- scanner.Text()
			}
		}
	}()
	select {
	case line := <-ready:
		fields := strings.Fields(line)
		api.base = fields[2] + "/api/v1"
		api.control = fields[3]
		t.Logf("independent API pid=%d base=%s database=%s own pool max=8", api.pid, api.base, fields[4])
	case <-time.After(10 * time.Second):
		t.Fatal("child startup timeout (credentials suppressed)")
	}
	return api
}
func (p *processAPI) stop(hard bool) {
	if p.stopped {
		return
	}
	p.stopped = true
	if hard {
		_ = p.cmd.Process.Kill()
	} else {
		_ = p.cmd.Process.Signal(syscall.SIGTERM)
	}
	_ = p.cmd.Wait()
}

type processReply struct {
	status int
	body   map[string]any
	err    error
}

func processRequest(base, token, method, path, key, body string) processReply {
	request, err := http.NewRequest(method, base+path, strings.NewReader(body))
	if err != nil {
		return processReply{err: err}
	}
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	if key != "" {
		request.Header.Set("Idempotency-Key", key)
	}
	client := &http.Client{Timeout: 12 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return processReply{err: err}
	}
	defer func() { _ = response.Body.Close() }()
	result := processReply{status: response.StatusCode, body: map[string]any{}}
	if response.StatusCode != 204 {
		result.err = json.NewDecoder(response.Body).Decode(&result.body)
	}
	return result
}
func requireProcessGet(t *testing.T, db *sqlx.DB, destination any, query string, args ...any) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	if err := db.GetContext(ctx, destination, query, args...); err != nil {
		t.Fatalf("SQL invariant query failed: %v", err)
	}
}

func requireProcessStatus(t *testing.T, r processReply, status int) {
	t.Helper()
	if r.err != nil || r.status != status {
		t.Fatalf("HTTP got=%d want=%d error=%v", r.status, status, r.err)
	}
}
func processEventually(t *testing.T, fn func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("process barrier timeout")
}

func TestRealProcessLifecycle(t *testing.T) {
	suite := t
	f := newNestingFixture(t)
	hash, err := bcrypt.GenerateFromPassword([]byte("test-process-password"), 4)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.db.Exec("UPDATE users SET password_hash=$1 WHERE id=1", string(hash)); err != nil {
		t.Fatal(err)
	}
	_, err = f.db.Exec(`CREATE TABLE process_gate(released BOOLEAN NOT NULL,commit_hold BOOLEAN NOT NULL,reserve_hold BOOLEAN NOT NULL DEFAULT FALSE);INSERT INTO process_gate(released,commit_hold) VALUES(FALSE,FALSE);
 CREATE TABLE process_calls(id BIGSERIAL PRIMARY KEY,pid INT NOT NULL,started TIMESTAMPTZ DEFAULT clock_timestamp(),ended TIMESTAMPTZ);
 CREATE TABLE process_commits(pid INT NOT NULL)`)
	if err != nil {
		t.Fatal(err)
	}
	var schema string
	if err = f.db.Get(&schema, "SELECT current_schema()"); err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(os.Getenv("NESTING_TEST_DSN"))
	if err != nil {
		t.Fatal("invalid DSN")
	}
	q := parsed.Query()
	q.Set("search_path", schema+",public")
	q.Set("application_name", "stage5_"+schema)
	parsed.RawQuery = q.Encode()
	a, b := startProcess(t, parsed.String(), false), startProcess(t, parsed.String(), false)
	if a.pid == b.pid {
		t.Fatal("same OS process")
	}
	login := processRequest(a.base, "", "POST", "/auth/login", "", `{"identifier":"owner","password":"test-process-password"}`)
	requireProcessStatus(t, login, 200)
	token := login.body["access_token"].(string)
	base := "/projects/1/surfaces/1/nesting"
	body := `{"algorithm":"baseline","patterns":[{"project_pattern_id":1,"quantity":1}],"allowed_rotations":[0],"keep_existing":false}`
	count := func() int {
		var n int
		if err := f.db.Get(&n, "SELECT count(*) FROM process_calls"); err != nil {
			t.Fatal(err)
		}
		return n
	}
	gate := func(released, commit bool) {
		if _, err := f.db.Exec("UPDATE process_gate SET released=$1,commit_hold=$2", released, commit); err != nil {
			t.Fatal(err)
		}
	}
	launch := func(api *processAPI, path, key, request string) <-chan processReply {
		done := make(chan processReply, 1)
		go func() { done <- processRequest(api.base, token, "POST", path, key, request) }()
		return done
	}
	status := func(api *processAPI, key string, preview bool) processReply {
		path := base + "/operations/" + key
		if preview {
			path += "?kind=preview"
		}
		return processRequest(api.base, token, "GET", path, "", "")
	}
	for _, preview := range []bool{true, false} {
		for _, keep := range []bool{false, true} {
			t.Run(fmt.Sprintf("duplicate/preview=%v/keep=%v", preview, keep), func(t *testing.T) {
				// Preserve prior layouts for keep_existing; clear only this fixture otherwise.
				gate(false, false)
				before := count()
				key := uuid.NewString()
				path := base
				if preview {
					path += "/preview"
				}
				request := strings.Replace(body, `"keep_existing":false`, `"keep_existing":`+strconv.FormatBool(keep), 1)
				first := launch(a, path, key, request)
				processEventually(t, func() bool { return count() == before+1 })
				running := status(b, key, preview)
				requireProcessStatus(t, running, 200)
				if running.body["status"] != "running" {
					t.Fatal("not running")
				}
				var held int
				if err := f.db.Get(&held, "SELECT count(*) FROM pg_stat_activity WHERE application_name LIKE $1||'%' AND state='idle in transaction'", "stage5_"+schema); err != nil || held != 0 {
					t.Fatalf("transaction held during Optimize: %d %v", held, err)
				}
				second := launch(b, path, key, request)
				conflict := processRequest(b.base, token, "POST", path, key, strings.Replace(request, `"quantity":1`, `"quantity":2`, 1))
				requireProcessStatus(t, conflict, 409)
				gate(true, false)
				r1, r2 := <-first, <-second
				requireProcessStatus(t, r1, 200)
				requireProcessStatus(t, r2, 200)
				j1, _ := json.Marshal(r1.body)
				j2, _ := json.Marshal(r2.body)
				if !bytes.Equal(j1, j2) || count() != before+1 {
					t.Fatal("duplicate result or Optimize")
				}
				replay := processRequest(b.base, token, "POST", path, key, request)
				requireProcessStatus(t, replay, 200)
				if count() != before+1 {
					t.Fatal("replay computed")
				}
				var rows int
				table := "nesting_operations"
				if preview {
					table = "nesting_preview_operations"
				}
				if err = f.db.Get(&rows, "SELECT count(*) FROM "+table+" WHERE operation_key=$1", key); err != nil || rows != 1 {
					t.Fatal("operation count")
				}
				if preview {
					var ready int
					if err = f.db.Get(&ready, "SELECT count(*) FROM nesting_previews WHERE operation_key=$1", key); err != nil || ready != 1 {
						t.Fatal("preview duplicate")
					}
					// Apply the exact ready result through the other process, without Optimize.
					save := strings.TrimSuffix(request, "}") + `,"preview_id":"` + r1.body["preview_id"].(string) + `"}`
					requireProcessStatus(t, processRequest(b.base, token, "POST", base, uuid.NewString(), save), 200)
					if count() != before+1 {
						t.Fatal("preview save recomputed")
					}
				}
			})
		}
	}
	for _, preview := range []bool{true, false} {
		t.Run(fmt.Sprintf("cross-cancel/%v", preview), func(t *testing.T) {
			gate(false, false)
			before := count()
			key := uuid.NewString()
			path := base
			suffix := ""
			if preview {
				path += "/preview"
				suffix = "?kind=preview"
			}
			first := launch(a, path, key, body)
			processEventually(t, func() bool { return count() == before+1 })
			began := time.Now()
			requireProcessStatus(t, processRequest(b.base, token, "DELETE", base+"/operations/"+key+suffix, "", ""), 204)
			requireProcessStatus(t, <-first, 409)
			t.Logf("cross-process cancel to worker response=%s", time.Since(began))
			terminal := status(b, key, preview)
			requireProcessStatus(t, terminal, 200)
			if terminal.body["status"] != "failed" {
				t.Fatal("cancel not terminal")
			}
			requireProcessStatus(t, processRequest(b.base, token, "POST", path, key, body), 409)
			if count() != before+1 {
				t.Fatal("cancel replay computed")
			}
		})
	}
	t.Run("kill-expiry-restart", func(t *testing.T) {
		gate(false, false)
		before := count()
		key := uuid.NewString()
		first := launch(a, base+"/preview", key, body)
		processEventually(t, func() bool { return count() == before+1 })
		a.stop(true)
		<-first
		if _, err = f.db.Exec("UPDATE nesting_preview_operations SET lease_expires_at=clock_timestamp()-interval '1 second' WHERE operation_key=$1", key); err != nil {
			t.Fatal(err)
		}
		failed := status(b, key, true)
		requireProcessStatus(t, failed, 200)
		if failed.body["status"] != "failed" {
			t.Fatal("dead worker not expired")
		}
		requireProcessStatus(t, processRequest(b.base, token, "POST", base+"/preview", key, body), 408)
		a = startProcess(suite, parsed.String(), false)
		gate(true, false)
		requireProcessStatus(t, processRequest(a.base, token, "POST", base+"/preview", uuid.NewString(), body), 200)
		if count() != before+2 {
			t.Fatal("expired key computed")
		}
	})
	t.Run("cancel-at-commit", func(t *testing.T) {
		gate(true, true)
		key := uuid.NewString()
		var before int
		if err = f.db.Get(&before, "SELECT count(*) FROM nesting_runs"); err != nil {
			t.Fatal(err)
		}
		first := launch(a, base, key, body)
		processEventually(t, func() bool {
			var n int
			requireProcessGet(t, f.db, &n, "SELECT count(*) FROM process_commits")
			return n > 0
		})
		cancelled := make(chan processReply, 1)
		go func() { cancelled <- processRequest(b.base, token, "DELETE", base+"/operations/"+key, "", "") }()
		gate(true, false)
		r, del := <-first, <-cancelled
		if del.status != 204 && del.status != 404 {
			t.Fatal("cancel response")
		}
		terminal := status(b, key, false)
		requireProcessStatus(t, terminal, 200)
		var after int
		requireProcessGet(t, f.db, &after, "SELECT count(*) FROM nesting_runs")
		if terminal.body["status"] == "completed" {
			requireProcessStatus(t, r, 200)
			if after != before+1 {
				t.Fatal("partial completed")
			}
		} else {
			requireProcessStatus(t, r, 409)
			if after != before {
				t.Fatal("partial cancelled write")
			}
		}
	})
	t.Run("stale-and-independent-surfaces", func(t *testing.T) {
		gate(false, false)
		before := count()
		k1, k2 := uuid.NewString(), uuid.NewString()
		one := launch(a, base, k1, body)
		two := launch(b, base, k2, body)
		processEventually(t, func() bool { return count() == before+2 })
		gate(true, false)
		r1, r2 := <-one, <-two
		validPair := (r1.status == 200 && r2.status == 409) || (r2.status == 200 && r1.status == 409)
		if !validPair {
			t.Fatalf("same surface outcomes %d/%d", r1.status, r2.status)
		}
		gate(false, false)
		before = count()
		one = launch(a, base, uuid.NewString(), body)
		two = launch(b, strings.Replace(base, "surfaces/1", "surfaces/2", 1), uuid.NewString(), body)
		processEventually(t, func() bool { return count() == before+2 })
		gate(true, false)
		requireProcessStatus(t, <-one, 200)
		requireProcessStatus(t, <-two, 200)
	})
	t.Run("duplicate-storm", func(t *testing.T) {
		gate(false, false)
		before := count()
		key := uuid.NewString()
		const clients = 160
		replies := make(chan processReply, clients)
		for i := 0; i < clients; i++ {
			api := a
			if i%2 == 1 {
				api = b
			}
			go func() { replies <- processRequest(api.base, token, "POST", base+"/preview", key, body) }()
		}
		processEventually(t, func() bool { return count() == before+1 })
		// Excess callers must fail fast; the bounded response waiters stay blocked.
		pending := 0
		for pending < clients-65 {
			select {
			case r := <-replies:
				requireProcessStatus(t, r, 503)
				pending++
			case <-time.After(5 * time.Second):
				t.Fatal("unbounded duplicate waiters")
			}
		}
		began := time.Now()
		requireProcessStatus(t, status(b, key, true), 200)
		getLatency := time.Since(began)
		gate(true, false)
		completed := 0
		for n := pending; n < clients; n++ {
			r := <-replies
			if r.status == 503 {
				pending++
			} else {
				requireProcessStatus(t, r, 200)
				completed++
			}
		}
		if count() != before+1 {
			t.Fatal("duplicate storm computed more than once")
		}
		t.Logf("duplicate load clients=%d completed=%d pending503=%d Optimize=1 GET=%s", clients, completed, pending, getLatency)
	})
	t.Run("local-capacity-two-processes", func(t *testing.T) {
		gate(false, false)
		before := count()
		const jobs = 40
		responses := make(chan processReply, jobs)
		for i := 0; i < jobs; i++ {
			api := a
			if i%2 == 1 {
				api = b
			}
			go func() { responses <- processRequest(api.base, token, "POST", base, uuid.NewString(), body) }()
		}
		processEventually(t, func() bool { return count() == before+4 })
		// 2 running + 4 queued on each process; all other keys get explicit overload.
		overloaded := 0
		for overloaded < jobs-12 {
			select {
			case r := <-responses:
				requireProcessStatus(t, r, 503)
				overloaded++
			case <-time.After(5 * time.Second):
				t.Fatal("capacity rejection stalled")
			}
		}
		for _, api := range []*processAPI{a, b} {
			r := processRequest(api.control, "", "GET", "/stats", "", "")
			requireProcessStatus(t, r, 200)
			t.Logf("pool pid=%d InUse=%v Open=%v WaitCount=%v goroutines=%v heap_bytes=%v", api.pid, r.body["InUse"], r.body["OpenConnections"], r.body["WaitCount"], r.body["Goroutines"], r.body["HeapAlloc"])
		}
		var concurrent int
		if err = f.db.Get(&concurrent, "SELECT count(*) FROM process_calls WHERE ended IS NULL AND pid IN($1,$2)", a.pid, b.pid); err != nil || concurrent != 4 {
			t.Fatalf("global running=%d %v", concurrent, err)
		}

		var activeKey string
		requireProcessGet(t, f.db, &activeKey, "SELECT operation_key::text FROM nesting_operations WHERE state='running' ORDER BY operation_key LIMIT 1")
		began := time.Now()
		requireProcessStatus(t, status(b, activeKey, false), 200)
		requireProcessStatus(t, processRequest(b.base, token, "DELETE", base+"/operations/"+activeKey, "", ""), 204)
		t.Logf("FULL_QUEUE_READ_CANCEL elapsed=%s", time.Since(began))
		gate(true, false)
		accepted := 0
		stale := 0
		for n := overloaded; n < jobs; n++ {
			r := <-responses
			switch r.status {
			case 200:
				accepted++
			case 408:
				stale++ // Explicit cancellation under full queue; no commit allowed.
			case 409:
				stale++
			case 503:
				overloaded++
			default:
				t.Fatalf("load HTTP %d error %v", r.status, r.err)
			}
		}
		t.Logf("capacity jobs=%d total_running=4 completed=%d stale=%d overloaded=%d", jobs, accepted, stale, overloaded)
	})
	t.Run("ready-restart-expiry", func(t *testing.T) {
		gate(true, false)
		key := uuid.NewString()
		before := count()
		ready := processRequest(a.base, token, "POST", base+"/preview", key, body)
		requireProcessStatus(t, ready, 200)
		a.stop(false)
		a = startProcess(suite, parsed.String(), false)
		restored := status(a, key, true)
		requireProcessStatus(t, restored, 200)
		if restored.body["result"].(map[string]any)["preview_id"] != ready.body["preview_id"] {
			t.Fatal("restart changed ready ID")
		}
		if _, err = f.db.Exec("UPDATE nesting_previews SET expires_at=clock_timestamp()-interval '1 second' WHERE operation_key=$1", key); err != nil {
			t.Fatal(err)
		}
		expired := status(b, key, true)
		requireProcessStatus(t, expired, 200)
		if expired.body["status"] != "failed" {
			t.Fatal("expired status")
		}
		requireProcessStatus(t, processRequest(b.base, token, "POST", base+"/preview", key, body), 404)
		if count() != before+1 {
			t.Fatal("expired replay computed")
		}
	})
	t.Run("unavailable-process-pool", func(t *testing.T) {
		broken := startProcess(t, parsed.String(), false)
		requireProcessStatus(t, processRequest(broken.control, "", "GET", "/close-pool", "", ""), 204)
		r := status(broken, uuid.NewString(), false)
		requireProcessStatus(t, r, 500)
		if r.body["status"] == "completed" {
			t.Fatal("false completed on unavailable SQL")
		}
	})

	t.Run("client-disconnect-during-compute", func(t *testing.T) {
		gate(false, false)
		before := count()
		key := uuid.NewString()
		var runs int
		requireProcessGet(t, f.db, &runs, "SELECT count(*) FROM nesting_runs")
		endpoint, _ := url.Parse(a.base)
		connection, err := net.DialTimeout("tcp", endpoint.Host, time.Second)
		if err != nil {
			t.Fatal(err)
		}
		_, err = fmt.Fprintf(connection, "POST /api/v1%s HTTP/1.1\r\nHost: %s\r\nAuthorization: Bearer %s\r\nIdempotency-Key: %s\r\nContent-Type: application/json\r\nContent-Length: %d\r\nConnection: close\r\n\r\n%s", base, endpoint.Host, token, key, len(body), body)
		if err != nil {
			t.Fatal("test socket write failed")
		}
		processEventually(t, func() bool { return count() == before+1 })
		_ = connection.Close()
		gate(true, false)
		processEventually(t, func() bool { return status(b, key, false).body["status"] == "completed" })
		requireProcessStatus(t, processRequest(b.base, token, "POST", base, key, body), 200)
		var after int
		requireProcessGet(t, f.db, &after, "SELECT count(*) FROM nesting_runs")
		if after != runs+1 || count() != before+1 {
			t.Fatal("disconnect cancelled or duplicated committed work")
		}
	})

	t.Run("lost-response-after-real-commit", func(t *testing.T) {
		gate(true, false)
		before := count()
		key := uuid.NewString()
		var runs int
		requireProcessGet(t, f.db, &runs, "SELECT count(*) FROM nesting_runs")
		endpoint, _ := url.Parse(a.base)
		connection, err := net.DialTimeout("tcp", endpoint.Host, time.Second)
		if err != nil {
			t.Fatal(err)
		}
		// Deliberately never read the HTTP response. SQL completion is observed on B.
		_, err = fmt.Fprintf(connection, "POST /api/v1%s HTTP/1.1\r\nHost: %s\r\nAuthorization: Bearer %s\r\nIdempotency-Key: %s\r\nContent-Type: application/json\r\nContent-Length: %d\r\nConnection: close\r\n\r\n%s", base, endpoint.Host, token, key, len(body), body)
		if err != nil {
			t.Fatal("test socket write failed")
		}
		processEventually(t, func() bool { return status(b, key, false).body["status"] == "completed" })
		_ = connection.Close()
		replay := processRequest(b.base, token, "POST", base, key, body)
		requireProcessStatus(t, replay, 200)
		var after int
		requireProcessGet(t, f.db, &after, "SELECT count(*) FROM nesting_runs")
		if after != runs+1 || count() != before+1 {
			t.Fatal("lost response caused duplicate write or Optimize")
		}
	})
	t.Run("crash-after-reserve-before-optimize", func(t *testing.T) {
		gate(false, false)
		before := count()
		key := uuid.NewString()
		if _, err = f.db.Exec("UPDATE process_gate SET reserve_hold=TRUE"); err != nil {
			t.Fatal(err)
		}
		first := launch(a, base, key, body)
		processEventually(t, func() bool {
			var n int
			requireProcessGet(t, f.db, &n, "SELECT count(*) FROM nesting_operations WHERE operation_key=$1", key)
			return n == 1
		})
		a.stop(true)
		<-first
		if count() != before {
			t.Fatal("Optimize started despite Reserve barrier")
		}
		if _, err = f.db.Exec("UPDATE process_gate SET reserve_hold=FALSE; UPDATE nesting_operations SET lease_expires_at=clock_timestamp()-interval '1 second' WHERE operation_key='" + key + "'"); err != nil {
			t.Fatal(err)
		}
		expired := status(b, key, false)
		requireProcessStatus(t, expired, 200)
		if expired.body["status"] != "failed" {
			t.Fatal("unowned claim stuck")
		}
		a = startProcess(suite, parsed.String(), false)
	})
	t.Run("crash-after-optimize-before-commit", func(t *testing.T) {
		gate(true, true)
		var before, barriers int
		requireProcessGet(t, f.db, &before, "SELECT count(*) FROM nesting_runs")
		requireProcessGet(t, f.db, &barriers, "SELECT count(*) FROM process_commits")
		key := uuid.NewString()
		first := launch(a, base, key, body)
		processEventually(t, func() bool {
			var n int
			requireProcessGet(t, f.db, &n, "SELECT count(*) FROM process_commits")
			return n > barriers
		})
		a.stop(true)
		<-first
		if _, err = f.db.Exec("UPDATE nesting_operations SET lease_expires_at=clock_timestamp()-interval '1 second' WHERE operation_key=$1", key); err != nil {
			t.Fatal(err)
		}
		terminal := status(b, key, false)
		requireProcessStatus(t, terminal, 200)
		if terminal.body["status"] != "failed" {
			t.Fatal("crashed precommit operation falsely completed")
		}
		var after int
		requireProcessGet(t, f.db, &after, "SELECT count(*) FROM nesting_runs")
		if before != after {
			t.Fatal("precommit crash wrote history")
		}
		gate(true, false)
		requireProcessStatus(t, processRequest(b.base, token, "POST", base, key, body), 408)
		requireProcessStatus(t, processRequest(b.base, token, "POST", base, uuid.NewString(), body), 200)
		a = startProcess(suite, parsed.String(), false)
	})
	t.Run("timeout-at-commit", func(t *testing.T) {
		gate(true, true)
		key := uuid.NewString()
		var before, commits int
		requireProcessGet(t, f.db, &before, "SELECT count(*) FROM nesting_runs")
		requireProcessGet(t, f.db, &commits, "SELECT count(*) FROM process_commits")
		first := launch(a, base, key, body)
		processEventually(t, func() bool {
			var n int
			requireProcessGet(t, f.db, &n, "SELECT count(*) FROM process_commits")
			return n > commits
		})
		if _, err = f.db.Exec("UPDATE nesting_operations SET lease_expires_at=clock_timestamp()-interval '1 second' WHERE operation_key=$1", key); err != nil {
			t.Fatal(err)
		}
		gate(true, false)
		requireProcessStatus(t, <-first, 408)
		var after int
		requireProcessGet(t, f.db, &after, "SELECT count(*) FROM nesting_runs")
		if before != after {
			t.Fatal("timeout committed partial result")
		}
	})

	t.Run("SQL-network-loss-at-commit", func(t *testing.T) {
		faulted := startProcess(t, parsed.String(), false)
		gate(true, true)
		var before int
		requireProcessGet(t, f.db, &before, "SELECT count(*) FROM nesting_runs")
		key := uuid.NewString()
		first := launch(faulted, base, key, body)
		processEventually(t, func() bool {
			var n int
			requireProcessGet(t, f.db, &n, "SELECT count(*) FROM process_commits WHERE pid=$1", faulted.pid)
			return n > 0
		})
		requireProcessStatus(t, processRequest(faulted.control, "", "GET", "/network-off", "", ""), 204)
		gate(true, false)
		requireProcessStatus(t, <-first, 500)
		requireProcessStatus(t, status(faulted, key, false), 500)
		var after int
		requireProcessGet(t, f.db, &after, "SELECT count(*) FROM nesting_runs")
		if after != before {
			t.Fatal("SQL network failure committed partial history")
		}
		requireProcessStatus(t, status(b, key, false), 200)
		if _, err = f.db.Exec("UPDATE nesting_operations SET lease_expires_at=clock_timestamp()-interval '1 second' WHERE operation_key=$1", key); err != nil {
			t.Fatal(err)
		}
		if status(b, key, false).body["status"] != "failed" {
			t.Fatal("failed network worker stuck")
		}
	})

	t.Run("graceful-active-and-queued", func(t *testing.T) {
		gate(false, false)
		before := count()
		var runs int
		requireProcessGet(t, f.db, &runs, "SELECT count(*) FROM nesting_runs")
		keys := []string{uuid.NewString(), uuid.NewString(), uuid.NewString()}
		replies := make([]<-chan processReply, 0, 3)
		for _, key := range keys {
			replies = append(replies, launch(a, base, key, body))
		}
		processEventually(t, func() bool {
			var n int
			requireProcessGet(t, f.db, &n, "SELECT count(*) FROM nesting_operations WHERE operation_key IN($1,$2,$3)", keys[0], keys[1], keys[2])
			return n == 3 && count() == before+2
		})
		stopped := make(chan struct{})
		go func() { a.stop(false); close(stopped) }()
		select {
		case <-stopped:
		case <-time.After(6 * time.Second):
			_ = a.cmd.Process.Kill()
			<-stopped
			t.Fatal("graceful shutdown exceeded bounded six-second guard")
		}
		for _, r := range replies {
			reply := <-r
			if reply.status == 200 {
				t.Fatal("held optimizer completed during shutdown")
			}
		}
		for _, key := range keys {
			if _, err = f.db.Exec("UPDATE nesting_operations SET lease_expires_at=clock_timestamp()-interval '1 second' WHERE operation_key=$1", key); err != nil {
				t.Fatal(err)
			}
			terminal := status(b, key, false)
			requireProcessStatus(t, terminal, 200)
			if terminal.body["status"] != "failed" {
				t.Fatal("shutdown left stuck or false-completed operation")
			}
		}
		var after int
		requireProcessGet(t, f.db, &after, "SELECT count(*) FROM nesting_runs")
		if after != runs {
			t.Fatal("shutdown committed partial history")
		}
		gate(true, false)
		a = startProcess(suite, parsed.String(), false)
	})
	t.Run("native-smoke", func(t *testing.T) {
		native := startProcess(t, parsed.String(), true)
		requireProcessStatus(t, processRequest(native.base, token, "POST", base+"/preview", uuid.NewString(), body), 200)
	})
}

func processFDCount() int {
	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		return -1
	}
	return len(entries)
}

type processResourceUsage struct {
	UserSeconds, SystemSeconds float64
	MaxRSSKiB                  int64
	Status                     string
}

func readProcessUsage() processResourceUsage {
	var r syscall.Rusage
	_ = syscall.Getrusage(syscall.RUSAGE_SELF, &r)
	data, _ := os.ReadFile("/proc/self/status")
	var status []string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "VmRSS:") || strings.HasPrefix(line, "Threads:") {
			status = append(status, line)
		}
	}
	return processResourceUsage{float64(r.Utime.Sec) + float64(r.Utime.Usec)/1e6, float64(r.Stime.Sec) + float64(r.Stime.Usec)/1e6, r.Maxrss, strings.Join(status, "; ")}
}

// Channel Len is a concurrency-safe observation, without changing limiter code.
func readProcessLimiter(l *uc.ComputeLimiter) map[string]any {
	if l == nil {
		return nil
	}
	v := reflect.ValueOf(l).Elem()
	result := map[string]any{"running": v.FieldByName("running").Len(), "admitted": v.FieldByName("admitted").Len(), "duplicates": v.FieldByName("responses").Len()}
	if method := reflect.ValueOf(l).MethodByName("LabMetrics"); method.IsValid() {
		result["observed"] = method.Call(nil)[0].Interface()
	}
	return result
}

func readProcessOptimizer(u *uc.RunNestingUseCase) map[string]any {
	if method := reflect.ValueOf(u).MethodByName("LabOptimizeMetrics"); method.IsValid() {
		return method.Call(nil)[0].Interface().(map[string]any)
	}
	return nil
}
