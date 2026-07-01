package postgres

import (
	"context"
	"fmt"
	"server_nesting_optimizer/internal/config"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func New(ctx context.Context, cfg config.DBConfig) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", cfg.DSN())

	if err != nil {
		return nil, fmt.Errorf("postgres open: %w", err)
	}

	configurePool(db)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("postgres ping: %w", err)
	}

	return db, nil
}

func configurePool(db *sqlx.DB) {
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Minute)
}
