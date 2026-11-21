package postgres

import (
	"context"
	"server_nesting_optimizer/internal/config"
	"strings"
	"testing"
	"time"
)

func TestStage17ConnectionErrorDoesNotExposeSecret(t *testing.T) {
	const marker = "SyntheticPrivateMarkerForErrorAuditOnly"
	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()
	_, err := New(ctx, config.DBConfig{Host: "127.0.0.1", Port: -1, User: "audit", Password: marker, DBName: "nesting_test", SSLMode: "verify-full"})
	if err == nil {
		t.Fatal("invalid fixture accepted")
	}
	if strings.Contains(err.Error(), marker) {
		t.Fatal("startup connection error exposes credential marker")
	}
}
