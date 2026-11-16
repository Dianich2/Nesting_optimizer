package integration

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"server_nesting_optimizer/internal/app/container"
	"server_nesting_optimizer/internal/config"
	transport "server_nesting_optimizer/internal/transport/http"
)

func TestAccountDeletionAPI(t *testing.T) {
	db := newTestSchema(t)
	applyAllMigrations(t, db)
	cfg := config.Config{
		JWT:      config.JWTConfig{Issuer: "integration", AccessSecret: "test-access", RefreshSecret: "test-refresh", AccessTTLMin: 15, RefreshTTLMin: 60},
		Security: config.SecurityConfig{BcryptCost: 4},
	}
	deps, err := container.New(db, cfg)
	if err != nil {
		t.Fatal(err)
	}
	app := transport.NewRouter(deps, cfg)
	token := ""
	send := func(method, path, body string, want int) map[string]any {
		t.Helper()
		req := httptest.NewRequest(method, "/api/v1"+path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()
		var result map[string]any
		if resp.StatusCode != 204 {
			_ = json.NewDecoder(resp.Body).Decode(&result)
		}
		if resp.StatusCode != want {
			t.Fatalf("%s %s: got %d want %d: %v", method, path, resp.StatusCode, want, result)
		}
		return result
	}
	const user = `{"login":"owner","email":"owner@example.com","password":"password123","first_name":"Test","last_name":"Owner"}`
	send("GET", "/users/me", "", 401)
	send("POST", "/users", user, 201)
	send("POST", "/users", `{"login":"neighbor","email":"other@example.com","password":"password123","first_name":"Test","last_name":"Neighbor"}`, 201)
	login := send("POST", "/auth/login", `{"identifier":"owner","password":"password123"}`, 200)
	refresh := send("POST", "/auth/refresh", `{"refresh_token":"`+login["refresh_token"].(string)+`"}`, 200)
	token = refresh["access_token"].(string)
	send("PATCH", "/users/me", `{"first_name":"Updated"}`, 200)
	send("PATCH", "/users/me/password", `{"old_password":"password123","new_password":"changed123","repeat_new_password":"changed123"}`, 204)
	login = send("POST", "/auth/login", `{"identifier":"owner","password":"changed123"}`, 200)
	token = login["access_token"].(string)
	send("POST", "/auth/logout", "", 204)
	send("GET", "/users/me", "", 401)
	login = send("POST", "/auth/login", `{"identifier":"owner","password":"changed123"}`, 200)
	token = login["access_token"].(string)
	project := send("POST", "/projects", `{"name":"Test project"}`, 201)
	geometry := `{"exterior":{"points":[{"x":0,"y":0},{"x":20,"y":0},{"x":20,"y":20},{"x":0,"y":20}]},"holes":[]}`
	for _, kind := range []string{"surfaces", "patterns"} {
		send("POST", "/"+kind, `{"name":"Original","geometry":`+geometry+`}`, 201)
	}
	send("POST", "/users/me/delete", `{"password":"wrong-password"}`, 401)
	var updateStatus int
	var updateErr error
	var updates sync.WaitGroup
	updates.Go(func() {
		path := fmt.Sprintf("/api/v1/projects/%d", int(project["id"].(float64)))
		req := httptest.NewRequest("PATCH", path, strings.NewReader(`{"name":"Concurrent rename"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := app.Test(req)
		updateErr = err
		if err == nil {
			updateStatus = resp.StatusCode
			_ = resp.Body.Close()
		}
	})
	send("POST", "/users/me/delete", `{"password":"changed123"}`, 204)
	updates.Wait()
	if updateErr != nil || (updateStatus != 200 && updateStatus != 404 && updateStatus != 401) {
		t.Fatalf("unexpected project update during account deletion: %d, %v", updateStatus, updateErr)
	}
	send("GET", "/users/me", "", 401)
	send("POST", "/auth/refresh", `{"refresh_token":"`+login["refresh_token"].(string)+`"}`, 401)
	for _, table := range []string{"sessions", "projects", "surfaces", "patterns"} {
		var rows int
		if err := db.Get(&rows, "SELECT count(*) FROM "+table); err != nil || rows != 0 {
			t.Fatalf("%s: %d, %v", table, rows, err)
		}
	}
	var count int
	if err = db.Get(&count, "SELECT count(*) FROM users WHERE login='neighbor'"); err != nil || count != 1 {
		t.Fatalf("other user changed: %d, %v", count, err)
	}
	send("POST", "/users", user, 201)
}
