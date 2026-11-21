package integration

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"
)

// The Python lab guardian alone restarts its recorded database container.
// This test has no Docker access and cannot select a database for restart.
func TestStage19PGRecovery(t *testing.T) {
	if os.Getenv("NESTING_STAGE19_PG_RESTART") != "1" {
		t.Skip("requires guarded own-lab restart coordinator")
	}
	f := newNestingFixture(t)
	hash, e := bcrypt.GenerateFromPassword([]byte("test-recovery-password"), 4)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = f.db.Exec("UPDATE users SET password_hash=$1", string(hash)); e != nil {
		t.Fatal(e)
	}
	if _, e = f.db.Exec(`CREATE TABLE process_gate(released BOOLEAN NOT NULL,commit_hold BOOLEAN NOT NULL,reserve_hold BOOLEAN NOT NULL DEFAULT FALSE);INSERT INTO process_gate VALUES(TRUE,FALSE,FALSE);CREATE TABLE process_calls(id BIGSERIAL PRIMARY KEY,pid INT NOT NULL,started TIMESTAMPTZ DEFAULT clock_timestamp(),ended TIMESTAMPTZ);CREATE TABLE process_commits(pid INT NOT NULL)`); e != nil {
		t.Fatal(e)
	}
	var schema string
	requireProcessGet(t, f.db, &schema, "SELECT current_schema()")
	dsn, e := url.Parse(os.Getenv("NESTING_TEST_DSN"))
	if e != nil {
		t.Fatal("invalid DSN")
	}
	q := dsn.Query()
	q.Set("search_path", schema+",public")
	dsn.RawQuery = q.Encode()
	a, b := startProcess(t, dsn.String(), false), startProcess(t, dsn.String(), false)
	login := processRequest(a.base, "", "POST", "/auth/login", "", `{"identifier":"owner","password":"test-recovery-password"}`)
	requireProcessStatus(t, login, 200)
	token := login.body["access_token"].(string)
	path := "/projects/1/surfaces/1/nesting"
	body := `{"algorithm":"baseline","patterns":[{"project_pattern_id":1,"quantity":1}],"allowed_rotations":[0],"keep_existing":false}`
	completed := uuid.NewString()
	requireProcessStatus(t, processRequest(a.base, token, "POST", path, completed, body), 200)
	if _, e = f.db.Exec("UPDATE process_gate SET released=FALSE"); e != nil {
		t.Fatal(e)
	}
	pending := uuid.NewString()
	done := make(chan processReply, 1)
	go func() { done <- processRequest(a.base, token, "POST", path, pending, body) }()
	processEventually(t, func() bool {
		var n int
		requireProcessGet(t, f.db, &n, "SELECT count(*) FROM process_calls WHERE ended IS NULL")
		return n == 1
	})
	fmt.Println("OWN_PG_RESTART_READY")
	deadline := time.Now().Add(45 * time.Second)
	observedDown := false
	for time.Now().Before(deadline) {
		err := f.db.Ping()
		if err != nil {
			observedDown = true
		}
		if observedDown && err == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !observedDown {
		t.Fatal("coordinator did not produce database outage")
	}
	if e = f.db.Ping(); e != nil {
		t.Fatal("own database did not recover")
	}
	r := <-done
	if r.status != 500 && r.status != 408 {
		t.Fatalf("pending interruption status=%d error=%v", r.status, r.err)
	}
	requireProcessStatus(t, processRequest(b.base, token, "POST", path, completed, body), 200)
	var runs int
	requireProcessGet(t, f.db, &runs, "SELECT count(*) FROM nesting_runs")
	if runs != 1 {
		t.Fatal("completed replay duplicated or interrupted operation committed")
	}
	if _, e = f.db.Exec("UPDATE nesting_operations SET lease_expires_at=clock_timestamp()-interval '1 second' WHERE operation_key=$1", pending); e != nil {
		t.Fatal(e)
	}
	status := processRequest(b.base, token, "GET", path+"/operations/"+pending, "", "")
	requireProcessStatus(t, status, 200)
	if status.body["status"] != "failed" {
		t.Fatal("interrupted operation stuck or falsely completed")
	}
	if _, e = f.db.Exec("UPDATE process_gate SET released=TRUE"); e != nil {
		t.Fatal(e)
	}
	requireProcessStatus(t, processRequest(b.base, token, "POST", path, uuid.NewString(), body), 200)
	auditLoadState(t, f)
}

func TestStage19PendingIdentityChanges(t *testing.T) {
	if os.Getenv("NESTING_STAGE18") != "1" {
		t.Skip("opt-in real-process pending identity scenarios")
	}
	for _, kind := range []string{"logout", "password", "project-delete", "account-delete"} {
		t.Run(kind, func(t *testing.T) {
			f := newNestingFixture(t)
			hash, e := bcrypt.GenerateFromPassword([]byte("test-recovery-password"), 4)
			if e != nil {
				t.Fatal(e)
			}
			if _, e = f.db.Exec("UPDATE users SET password_hash=$1", string(hash)); e != nil {
				t.Fatal(e)
			}
			if _, e = f.db.Exec(`CREATE TABLE process_gate(released BOOLEAN NOT NULL,commit_hold BOOLEAN NOT NULL,reserve_hold BOOLEAN NOT NULL DEFAULT FALSE);INSERT INTO process_gate VALUES(FALSE,FALSE,FALSE);CREATE TABLE process_calls(id BIGSERIAL PRIMARY KEY,pid INT NOT NULL,started TIMESTAMPTZ DEFAULT clock_timestamp(),ended TIMESTAMPTZ);CREATE TABLE process_commits(pid INT NOT NULL)`); e != nil {
				t.Fatal(e)
			}
			var schema string
			requireProcessGet(t, f.db, &schema, "SELECT current_schema()")
			dsn, e := url.Parse(os.Getenv("NESTING_TEST_DSN"))
			if e != nil {
				t.Fatal("invalid DSN")
			}
			q := dsn.Query()
			q.Set("search_path", schema+",public")
			dsn.RawQuery = q.Encode()
			a, b := startProcess(t, dsn.String(), false), startProcess(t, dsn.String(), false)
			login := processRequest(a.base, "", "POST", "/auth/login", "", `{"identifier":"owner","password":"test-recovery-password"}`)
			requireProcessStatus(t, login, 200)
			token := login.body["access_token"].(string)
			refresh := login.body["refresh_token"].(string)
			path := "/projects/1/surfaces/1/nesting"
			body := `{"algorithm":"baseline","patterns":[{"project_pattern_id":1,"quantity":1}],"allowed_rotations":[0],"keep_existing":false}`
			key := uuid.NewString()
			done := make(chan processReply, 1)
			go func() { done <- processRequest(a.base, token, "POST", path, key, body) }()
			processEventually(t, func() bool {
				var n int
				requireProcessGet(t, f.db, &n, "SELECT count(*) FROM process_calls WHERE ended IS NULL")
				return n == 1
			})
			password := "test-recovery-password"
			switch kind {
			case "logout":
				requireProcessStatus(t, processRequest(b.base, token, "POST", "/auth/logout", "", ""), 204)
			case "password":
				password = "changed-recovery-password"
				requireProcessStatus(t, processRequest(b.base, token, "PATCH", "/users/me/password", "", `{"old_password":"test-recovery-password","new_password":"changed-recovery-password","repeat_new_password":"changed-recovery-password"}`), 204)
			case "project-delete":
				requireProcessStatus(t, processRequest(b.base, token, "DELETE", "/projects/1", "", ""), 204)
			case "account-delete":
				requireProcessStatus(t, processRequest(b.base, token, "POST", "/users/me/delete", "", `{"password":"test-recovery-password"}`), 204)
			}
			if kind != "project-delete" {
				requireProcessStatus(t, processRequest(b.base, token, "GET", "/users/me", "", ""), 401)
				requireProcessStatus(t, processRequest(b.base, "", "POST", "/auth/refresh", "", `{"refresh_token":"`+refresh+`"}`), 401)
			}
			if _, e = f.db.Exec("UPDATE process_gate SET released=TRUE"); e != nil {
				t.Fatal(e)
			}
			result := <-done
			if kind == "project-delete" || kind == "account-delete" {
				requireProcessStatus(t, result, 404)
				f.counts(t, 0, 0)
				var operations int
				requireProcessGet(t, f.db, &operations, "SELECT count(*) FROM nesting_operations")
				if operations != 0 {
					t.Fatal("cascade left orphan operation")
				}
				return
			}
			// Revoking a session blocks future HTTP requests; a previously accepted durable
			// worker may finish. The owner recovers its UUID through a new authenticated session.
			requireProcessStatus(t, result, 200)
			relogin := processRequest(b.base, "", "POST", "/auth/login", "", fmt.Sprintf(`{"identifier":"owner","password":%q}`, password))
			requireProcessStatus(t, relogin, 200)
			newToken := relogin.body["access_token"].(string)
			terminal := processRequest(b.base, newToken, "GET", path+"/operations/"+key, "", "")
			requireProcessStatus(t, terminal, 200)
			if terminal.body["status"] != "completed" {
				t.Fatal("new session failed to recover durable result")
			}
			f.counts(t, 1, 1)
			auditLoadState(t, f)
		})
	}
}

func TestStage19StatusTransportFaults(t *testing.T) {
	if os.Getenv("NESTING_STAGE18") != "1" {
		t.Skip("opt-in bounded loopback transport faults")
	}
	f := newNestingFixture(t)
	hash, e := bcrypt.GenerateFromPassword([]byte("test-recovery-password"), 4)
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
	dsn.RawQuery = q.Encode()
	a, b := startProcess(t, dsn.String(), true), startProcess(t, dsn.String(), true)
	login := processRequest(a.base, "", "POST", "/auth/login", "", `{"identifier":"owner","password":"test-recovery-password"}`)
	requireProcessStatus(t, login, 200)
	token := login.body["access_token"].(string)
	key := uuid.NewString()
	path := "/projects/1/surfaces/1/nesting"
	// Cancellation before Reserve creates no phantom terminal claim.
	requireProcessStatus(t, processRequest(b.base, token, "DELETE", path+"/operations/"+key, "", ""), 404)
	requireProcessStatus(t, processRequest(a.base, token, "POST", path, key, `{"algorithm":"baseline","patterns":[{"project_pattern_id":1,"quantity":1}],"allowed_rotations":[0],"keep_existing":false}`), 200)
	// After COMMIT, cancellation must not undo or relabel the completed layout.
	requireProcessStatus(t, processRequest(b.base, token, "DELETE", path+"/operations/"+key, "", ""), 404)
	for _, mode := range []string{"delayed-status-cancel", "reset-status", "drop-status"} {
		t.Run(mode, func(t *testing.T) {
			reached, release := make(chan struct{}, 1), make(chan struct{})
			defer close(release)
			proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				request, err := http.NewRequestWithContext(r.Context(), "GET", b.base+path+"/operations/"+key, nil)
				if err != nil {
					t.Error(err)
					return
				}
				request.Header.Set("Authorization", "Bearer "+token)
				response, err := (&http.Client{Timeout: 2 * time.Second}).Do(request)
				if err != nil {
					http.Error(w, "test upstream unavailable", http.StatusBadGateway)
					return
				}
				_, _ = io.Copy(io.Discard, response.Body)
				_ = response.Body.Close()
				if response.StatusCode != 200 {
					t.Error("upstream status failed")
				}
				select {
				case reached <- struct{}{}:
				default:
				}
				if mode == "delayed-status-cancel" {
					select {
					case <-release:
					case <-r.Context().Done():
					}
					return
				}
				hijacker, ok := w.(http.Hijacker)
				if !ok {
					t.Error("no test hijacker")
					return
				}
				conn, _, err := hijacker.Hijack()
				if err != nil {
					t.Error(err)
					return
				}
				if mode == "reset-status" {
					if tcp, ok := conn.(*net.TCPConn); ok {
						_ = tcp.SetLinger(0)
					}
				}
				_ = conn.Close()
			}))
			defer proxy.Close()
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			request, e := http.NewRequestWithContext(ctx, "GET", proxy.URL, nil)
			if e != nil {
				t.Fatal(e)
			}
			done := make(chan error, 1)
			go func() {
				response, err := (&http.Client{Timeout: 3 * time.Second}).Do(request)
				if response != nil {
					_ = response.Body.Close()
				}
				done <- err
			}()
			select {
			case <-reached:
			case <-time.After(3 * time.Second):
				t.Fatal("transport barrier timeout")
			}
			if mode == "delayed-status-cancel" {
				cancel()
			}
			if err := <-done; err == nil {
				t.Fatal("fault did not reach client")
			}
			terminal := processRequest(b.base, token, "GET", path+"/operations/"+key, "", "")
			requireProcessStatus(t, terminal, 200)
			if terminal.body["status"] != "completed" {
				t.Fatal("transport fault changed durable status")
			}
			f.counts(t, 1, 1)
		})
	}
	auditLoadState(t, f)
}
