package integration

import (
	"encoding/json"
	"fmt"
	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"server_nesting_optimizer/internal/app/container"
	"server_nesting_optimizer/internal/config"
	transport "server_nesting_optimizer/internal/transport/http"
	"strings"
	"testing"
	"time"
)

func TestStage17SecurityBoundaries(t *testing.T) {
	f := newNestingFixture(t)
	hash, err := bcrypt.GenerateFromPassword([]byte("OnlySyntheticAuditPassword"), 4)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.db.Exec("UPDATE users SET password_hash=$1", string(hash)); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{App: config.AppConfig{Env: "production"}, HTTP: config.HTTPConfig{BrowserOrigin: "https://nesting.example.test"}, JWT: config.JWTConfig{Issuer: "audit", AccessSecret: strings.Repeat("a", 40), RefreshSecret: strings.Repeat("b", 40), AccessTTLMin: 15, RefreshTTLMin: 60}, Security: config.SecurityConfig{BcryptCost: 4}, Nesting: config.NestingConfig{Workers: 2, Queue: 4}}
	deps, err := container.New(f.db, cfg)
	if err != nil {
		t.Fatal(err)
	}
	app := transport.NewRouter(deps, cfg)
	token := ""
	send := func(method, path, body string, want int, headers map[string]string) (map[string]any, *http.Response) {
		t.Helper()
		req := httptest.NewRequest(method, "/api/v1"+path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		res, e := app.Test(req)
		if e != nil {
			t.Fatal(e)
		}
		var result map[string]any
		if res.StatusCode != 204 {
			_ = json.NewDecoder(res.Body).Decode(&result)
		}
		_ = res.Body.Close()
		if res.StatusCode != want {
			t.Fatalf("%s %s status%d want%d (response suppressed)", method, path, res.StatusCode, want)
		}
		return result, res
	}
	login := func(user string) string {
		t.Helper()
		r, _ := send("POST", "/auth/login", fmt.Sprintf(`{"identifier":%q,"password":"OnlySyntheticAuditPassword"}`, user), 200, nil)
		return r["access_token"].(string)
	}
	send("GET", "/users/me", "", 401, nil)
	owner := login("owner")
	stranger := login("other")
	token = owner
	claims, err := deps.JWTManager.ParseAccessToken(owner)
	if err != nil {
		t.Fatal("test access token invalid")
	}
	claims.ExpiresAt = jwtlib.NewNumericDate(time.Now().Add(-time.Hour))
	token, err = jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims).SignedString([]byte(cfg.JWT.AccessSecret))
	if err != nil {
		t.Fatal(err)
	}
	send("GET", "/users/me", "", 401, nil)
	claims.ExpiresAt = jwtlib.NewNumericDate(time.Now().Add(time.Hour))
	token, _ = jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims).SignedString([]byte(strings.Repeat("c", 40)))
	send("GET", "/users/me", "", 401, nil)
	token = owner
	placement, _ := send("POST", "/projects/1/surfaces/1/placements", `{"project_pattern_id":1,"x":50,"y":50,"rotation":0}`, 201, nil)
	id := int(placement["id"].(float64))
	key := uuid.NewString()
	body := `{"algorithm":"baseline","patterns":[{"project_pattern_id":1,"quantity":1}],"allowed_rotations":[0],"keep_existing":true}`
	preview, _ := send("POST", "/projects/1/surfaces/1/nesting/preview", body, 200, map[string]string{"Idempotency-Key": key})
	token = stranger
	for _, path := range []string{"/projects/1", "/projects/1/surfaces/1", "/projects/1/patterns/1", fmt.Sprintf("/projects/1/placements/%d", id), "/projects/1/surfaces/1/placements", "/projects/1/surfaces/1/nesting/history", "/projects/1/surfaces/1/nesting/operations/" + key + "?kind=preview"} {
		send("GET", path, "", 404, nil)
	}
	send("PATCH", fmt.Sprintf("/projects/1/placements/%d", id), `{"x":70}`, 404, nil)
	send("DELETE", fmt.Sprintf("/projects/1/placements/%d", id), "", 404, nil)
	for _, suffix := range []string{"", "/preview"} {
		send("POST", "/projects/1/surfaces/1/nesting"+suffix, body, 404, map[string]string{"Idempotency-Key": key})
	}
	send("DELETE", "/projects/1/surfaces/1/nesting/operations/"+key+"?kind=preview", "", 404, nil)
	f.counts(t, 1, 0)
	token = owner
	save := strings.TrimSuffix(body, "}") + `,"preview_id":"` + preview["preview_id"].(string) + `"}`
	send("POST", "/projects/1/surfaces/1/nesting", save, 200, map[string]string{"Idempotency-Key": key})
	send("POST", "/projects/1/surfaces/1/nesting", save, 200, map[string]string{"Idempotency-Key": key})
	f.counts(t, 2, 1)
	send("POST", "/projects/1/surfaces/2/nesting", body, 200, map[string]string{"Idempotency-Key": key})
	f.counts(t, 3, 2)
	for _, bad := range []string{`{"algorithm":"baseline","patterns":[{"project_pattern_id":1,"quantity":501}],"allowed_rotations":[0]}`, `{"algorithm":"baseline","patterns":[{"project_pattern_id":1,"quantity":1}],"allowed_rotations":[0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36]}`, `{"x":NaN}`} {
		send("POST", "/projects/1/surfaces/1/nesting/preview", bad, 400, map[string]string{"Idempotency-Key": uuid.NewString()})
	}
	f.counts(t, 3, 2)
	token = ""
	loginBody := `{"identifier":"owner","password":"OnlySyntheticAuditPassword"}`
	send("POST", "/auth/browser/login", loginBody, 403, map[string]string{"Origin": "https://foreign.example.test"})
	browser, res := send("POST", "/auth/browser/login", loginBody, 200, map[string]string{"Origin": cfg.HTTP.BrowserOrigin})
	cookies := res.Cookies()
	if len(cookies) != 1 || !cookies[0].Secure || !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteStrictMode || res.Header.Get("Cache-Control") != "no-store" {
		t.Fatal("unsafe nonlocal browser cookie/cache attributes")
	}
	cookie := cookies[0].Name + "=" + cookies[0].Value
	send("POST", "/auth/browser/refresh", "", 403, map[string]string{"Origin": "https://foreign.example.test", "Cookie": cookie})
	token = browser["access_token"].(string)
	send("POST", "/auth/browser/logout", "", 204, map[string]string{"Origin": cfg.HTTP.BrowserOrigin, "Cookie": cookie})
	token = ""
	send("POST", "/auth/browser/refresh", "", 401, map[string]string{"Origin": cfg.HTTP.BrowserOrigin, "Cookie": cookie})
	t.Log("SECURITY actual JWT/expiry/signature/owner/kind-scope/surface-scope/production-cookie/CSRF/logout PASS")
}

func TestStage17OwnerChangeCannotCommit(t *testing.T) {
	f := newNestingFixture(t)
	if _, err := f.db.Exec("INSERT INTO placements(id,project_surface_id,project_pattern_id,x,y,rotation) VALUES(99,1,1,80,80,0)"); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { _, err := f.uc.ExecuteWithKey(t.Context(), f.input, uuid.NewString()); done <- err }()
	awaitStart(t, f.gate)
	if _, err := f.db.Exec("UPDATE projects SET user_id=2 WHERE id=1"); err != nil {
		t.Fatal(err)
	}
	close(f.gate.release)
	if err := <-done; err == nil {
		t.Fatal("old owner committed after ownership change")
	}
	f.counts(t, 1, 0)
	var id int
	requireProcessGet(t, f.db, &id, "SELECT id FROM placements")
	if id != 99 {
		t.Fatal("ownership race replaced layout")
	}
}
