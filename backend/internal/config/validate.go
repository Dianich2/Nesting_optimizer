package config

import (
	"fmt"
	"net/url"
	"strings"
)

// Validate is a readiness gate, not a claim of deployment-wide security. Errors
// name fields only and never echo credentials or a connection string.
func (c Config) Validate() error {
	if c.Nesting.Workers < 1 || c.Nesting.Workers > 32 || c.Nesting.Queue < 0 || c.Nesting.Queue > 128 {
		return fmt.Errorf("NESTING_WORKERS must be 1..32 and NESTING_QUEUE must be 0..128")
	}
	if c.HTTP.Port < 1 || c.HTTP.Port > 65535 || c.DB.Port < 1 || c.DB.Port > 65535 {
		return fmt.Errorf("HTTP_PORT and DB_PORT must be 1..65535")
	}
	if c.JWT.AccessTTLMin < 1 || c.JWT.AccessTTLMin > 1440 || c.JWT.RefreshTTLMin < c.JWT.AccessTTLMin || c.JWT.RefreshTTLMin > 525600 {
		return fmt.Errorf("JWT TTLs must be positive, access <= 1440, refresh >= access and <= 525600 minutes")
	}
	if c.Security.BcryptCost < 4 || c.Security.BcryptCost > 31 {
		return fmt.Errorf("SECURITY_BCRYPT_COST must be 4..31")
	}
	origin, err := url.Parse(c.HTTP.BrowserOrigin)
	if err != nil || origin.Hostname() == "" || origin.User != nil || origin.RawQuery != "" || origin.Fragment != "" || origin.Path != "" || (origin.Scheme != "http" && origin.Scheme != "https") {
		return fmt.Errorf("HTTP_BROWSER_ORIGIN must be an HTTP(S) origin without a path")
	}
	switch c.App.Env {
	case "local", "dev", "development", "test":
		return nil
	case "production", "prod", "staging":
	default:
		return fmt.Errorf("APP_ENV must be local/dev/development/test or production/prod/staging")
	}
	if len(strings.TrimSpace(c.JWT.AccessSecret)) < 32 || len(strings.TrimSpace(c.JWT.RefreshSecret)) < 32 || c.JWT.AccessSecret == c.JWT.RefreshSecret {
		return fmt.Errorf("nonlocal JWT secrets must be distinct and at least 32 bytes")
	}
	if strings.Contains(c.JWT.AccessSecret, "dev-access-secret") || strings.Contains(c.JWT.RefreshSecret, "dev-refresh-secret") {
		return fmt.Errorf("nonlocal JWT secrets must not use development defaults")
	}
	if len(strings.TrimSpace(c.DB.Password)) < 16 || c.DB.Password == "postgres" {
		return fmt.Errorf("nonlocal DB_PASSWORD must be explicit and at least 16 bytes")
	}
	if c.DB.SSLMode != "verify-full" {
		return fmt.Errorf("nonlocal DB_SSL_MODE must be verify-full")
	}
	if strings.TrimSpace(c.DB.Host) == "" || strings.TrimSpace(c.DB.User) == "" || strings.TrimSpace(c.DB.DBName) == "" {
		return fmt.Errorf("nonlocal DB_HOST, DB_USER and DB_NAME must be nonempty")
	}
	if origin.Scheme != "https" {
		return fmt.Errorf("nonlocal HTTP_BROWSER_ORIGIN must use HTTPS")
	}
	if c.Security.BcryptCost < 10 || c.Security.BcryptCost > 16 {
		return fmt.Errorf("nonlocal SECURITY_BCRYPT_COST must be 10..16")
	}
	return nil
}
