package postgres

import (
	"context"
	"server_nesting_optimizer/internal/config"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func New(ctx context.Context, cfg config.DBConfig) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", cfg.DSN())

	if err != nil {
		return nil, connectionError{operation: "open", cause: err}
	}

	configurePool(db)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, connectionError{operation: "ping", cause: err}
	}

	return db, nil
}

func configurePool(db *sqlx.DB) {
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Minute)
}

// Error strings may reach startup logs; driver parse errors can embed a DSN.
// Keep the cause inspectable with errors.Is without including it in messages.
type connectionError struct {
	operation string
	cause     error
}

func (e connectionError) Error() string { return "postgres " + e.operation + ": connection failed" }
func (e connectionError) Unwrap() error { return e.cause }
