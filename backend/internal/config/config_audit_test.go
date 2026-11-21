package config

import (
	"os"
	"os/exec"
	"testing"
)

func TestStage17ConfigChild(t *testing.T) {
	if os.Getenv("NESTING_CONFIG_AUDIT_CHILD") != "1" {
		t.Skip("isolated environment child")
	}
	_, err := Load()
	wantError := os.Getenv("NESTING_CONFIG_WANT_ERROR") == "1"
	if (err != nil) != wantError {
		t.Fatalf("configuration acceptance mismatch: rejected=%v wantRejected=%v", err != nil, wantError)
	}
}
func TestStage17ProductionRejectsDefaults(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(executable, "-test.run=^TestStage17ConfigChild$", "-test.count=1")
	cmd.Dir = t.TempDir()
	cmd.Env = []string{"APP_ENV=production", "NESTING_CONFIG_AUDIT_CHILD=1", "NESTING_CONFIG_WANT_ERROR=1"}
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("production accepted unsafe defaults: child=%s", out)
	}
}

func TestStage17ConfigMatrix(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	valid := map[string]string{"APP_ENV": "production", "JWT_ACCESS_SECRET": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "JWT_REFRESH_SECRET": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "DB_PASSWORD": "only-synthetic-db-password-for-config-test", "DB_SSL_MODE": "verify-full", "HTTP_BROWSER_ORIGIN": "https://nesting.example.test"}
	cases := []struct {
		name      string
		overrides map[string]string
		bad       bool
	}{
		{"production-valid", nil, false}, {"local", map[string]string{"APP_ENV": "local"}, false},
		{"equal", map[string]string{"JWT_REFRESH_SECRET": valid["JWT_ACCESS_SECRET"]}, true},
		{"blank", map[string]string{"JWT_ACCESS_SECRET": " "}, true}, {"short", map[string]string{"JWT_REFRESH_SECRET": "short"}, true},
		{"database-default", map[string]string{"DB_PASSWORD": "postgres"}, true}, {"ssl-disabled", map[string]string{"DB_SSL_MODE": "disable"}, true},
		{"ssl-no-host-verification", map[string]string{"DB_SSL_MODE": "require"}, true}, {"http-origin", map[string]string{"HTTP_BROWSER_ORIGIN": "http://nesting.example.test"}, true},
		{"origin-userinfo", map[string]string{"HTTP_BROWSER_ORIGIN": "https://user:private@nesting.example.test"}, true},
		{"origin-trailing-path-slash", map[string]string{"HTTP_BROWSER_ORIGIN": "https://nesting.example.test/"}, true},
		{"http-port", map[string]string{"HTTP_PORT": "0"}, true}, {"db-port", map[string]string{"DB_PORT": "65536"}, true},
		{"ttl", map[string]string{"JWT_ACCESS_TTL_MIN": "0"}, true}, {"refresh-order", map[string]string{"JWT_REFRESH_TTL_MIN": "1"}, true},
		{"bcrypt", map[string]string{"SECURITY_BCRYPT_COST": "4"}, true}, {"workers", map[string]string{"NESTING_WORKERS": "0"}, true},
		{"queue", map[string]string{"NESTING_QUEUE": "129"}, true}, {"unknown-env", map[string]string{"APP_ENV": "productoin"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			values := map[string]string{}
			if tc.name != "local" {
				for k, v := range valid {
					values[k] = v
				}
			}
			for k, v := range tc.overrides {
				values[k] = v
			}
			env := []string{"NESTING_CONFIG_AUDIT_CHILD=1"}
			if tc.bad {
				env = append(env, "NESTING_CONFIG_WANT_ERROR=1")
			}
			for k, v := range values {
				env = append(env, k+"="+v)
			}
			cmd := exec.Command(executable, "-test.run=^TestStage17ConfigChild$", "-test.count=1")
			cmd.Dir = t.TempDir()
			cmd.Env = env
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("isolated configuration gate mismatch: %s", out)
			}
		})
	}
}
