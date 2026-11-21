package config

import (
	"net/url"
	"testing"
)

func TestStage17DSNCredentialsRoundTrip(t *testing.T) {
	c := DBConfig{Host: "::1", Port: 5432, User: "test user", Password: "synthetic:@/?#% value", DBName: "nesting_test", SSLMode: "verify-full"}
	parsed, err := url.Parse(c.DSN())
	if err != nil {
		t.Fatal("DSN cannot encode arbitrary fixture credentials")
	}
	password, ok := parsed.User.Password()
	if !ok || password != c.Password || parsed.User.Username() != c.User {
		t.Fatal("DSN changes credentials")
	}
}
