package app

import (
	"context"
	"fmt"
	"log/slog"
	"server_nesting_optimizer/internal/app/container"
	"server_nesting_optimizer/internal/config"
	"server_nesting_optimizer/internal/repository/postgres"
	httptransport "server_nesting_optimizer/internal/transport/http"
	"server_nesting_optimizer/pkg/logger"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jmoiron/sqlx"
)

type App struct {
	cfg       config.Config
	log       *slog.Logger
	server    *fiber.App
	db        *sqlx.DB
	container *container.Container
}

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	log := logger.New(cfg.Log.Level)

	ctx := context.Background()

	db, err := postgres.New(ctx, cfg.DB)

	if err != nil {
		return nil, fmt.Errorf("create db connection pool: %w", err)
	}

	deps, err := container.New(db, cfg)
	if err != nil {
		if err := db.Close(); err != nil {
			log.Error("failed to close DB", "error", err)
		}

		return nil, fmt.Errorf("create dependencies: %w", err)
	}

	server := httptransport.NewRouter(deps, cfg)

	return &App{
		cfg:       cfg,
		log:       log,
		server:    server,
		db:        db,
		container: deps,
	}, nil
}

func (a *App) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	cleanupDone := make(chan struct{})
	go func() { defer close(cleanupDone); a.cleanupPreviews(ctx) }()
	defer func() { cancel(); <-cleanupDone }()
	addr := fmt.Sprintf(":%d", a.cfg.HTTP.Port)

	a.log.Info(
		"starting http server",
		"addr", addr,
		"env", a.cfg.App.Env,
	)

	return a.server.Listen(addr)
}

func (a *App) Stop() {
	if a.server != nil {
		if err := a.server.Shutdown(); err != nil {
			a.log.Error("failed to shutdown server", "error", err)
		}
	}
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			a.log.Error("failed to close DB", "error", err)
		}
	}

	a.log.Info("application stopped")
}

// Cleanup is tied to the server lifetime and runs immediately after a restart.
// Each pass is bounded in rows and time and never waits on an applying preview.
func (a *App) cleanupPreviews(ctx context.Context) {
	previews := postgres.NewNestingPreviewRepository(a.db)
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		cleanup, cancel := context.WithTimeout(ctx, 5*time.Second)
		var err error
		for cleanup.Err() == nil {
			var count int64
			count, err = previews.PurgeExpired(cleanup, 1000)
			if err != nil || count < 1000 {
				break
			}
		}
		cancel()
		if err != nil && ctx.Err() == nil {
			a.log.Error("purge expired nesting previews", "error", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
