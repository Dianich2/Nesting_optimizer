package integration

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/signal"
	"server_nesting_optimizer/internal/app/container"
	"server_nesting_optimizer/internal/config"
	"server_nesting_optimizer/internal/geometry/nfp"
	sf "server_nesting_optimizer/internal/geometry/simplefeatures"
	opt "server_nesting_optimizer/internal/nesting"
	"server_nesting_optimizer/internal/repository/postgres"
	transport "server_nesting_optimizer/internal/transport/http"
	uc "server_nesting_optimizer/internal/usecase/nesting"
	"strconv"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

type browserMeasuredOptimizer struct {
	opt.Optimizer
	calls *atomic.Int64
}

func (o browserMeasuredOptimizer) Optimize(ctx context.Context, p opt.Problem) (opt.Result, error) {
	o.calls.Add(1)
	return o.Optimizer.Optimize(ctx, p)
}

// Test-only executable server. No config.Load/.env, fake auth or production
// diagnostics endpoints. Each factory uses production defaults and a fresh cache.
func TestBrowserAcceptanceAPI(t *testing.T) {
	if os.Getenv("NESTING_BROWSER_API") != "1" {
		t.Skip("explicit isolated browser API helper")
	}
	origin := os.Getenv("NESTING_E2E_ORIGIN")
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Scheme != "http" || parsed.Hostname() != "localhost" || parsed.Port() == "" {
		t.Fatal("explicit loopback browser origin required")
	}
	ready := os.Getenv("NESTING_E2E_READY_FILE")
	if ready == "" {
		t.Fatal("explicit ready file required")
	}
	db := newTestSchema(t)
	applyAllMigrations(t, db)
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(4)
	cfg := config.Config{HTTP: config.HTTPConfig{BrowserOrigin: origin}, JWT: config.JWTConfig{Issuer: "browser-tests", AccessSecret: "browser-only-test-access", RefreshSecret: "browser-only-test-refresh", AccessTTLMin: 15, RefreshTTLMin: 60}, Security: config.SecurityConfig{BcryptCost: 4}, Nesting: config.NestingConfig{Workers: 2, Queue: 4}}
	deps, err := container.New(db, cfg)
	if err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int64
	factories := map[opt.Algorithm]opt.OptimizerFactory{}
	for _, algorithm := range []opt.Algorithm{opt.BaselineAlgorithm, opt.NFPGreedyAlgorithm, opt.GeneticAlgorithm, opt.SimulatedAnnealingAlgorithm, opt.HybridAlgorithm} {
		factories[algorithm] = func() (opt.Optimizer, error) {
			engine := sf.NewEngine()
			builder := nfp.NewCachedBuilder(nfp.NewBuilder(engine))
			var solver opt.Optimizer
			var e error
			switch algorithm {
			case opt.BaselineAlgorithm:
				solver = opt.NewBaselineOptimizer(engine)
			case opt.NFPGreedyAlgorithm:
				solver = opt.NewNFPGreedyOptimizer(engine, builder)
			case opt.GeneticAlgorithm:
				solver, e = opt.NewGeneticOptimizer(engine, builder, opt.DefaultGeneticConfig())
			case opt.SimulatedAnnealingAlgorithm:
				solver, e = opt.NewSimulatedAnnealingOptimizer(engine, builder, opt.DefaultAnnealingConfig())
			case opt.HybridAlgorithm:
				ga, err := opt.NewGeneticOptimizer(engine, builder, opt.DefaultGeneticConfig())
				if err != nil {
					return nil, err
				}
				sa, err := opt.NewSimulatedAnnealingOptimizer(engine, builder, opt.DefaultAnnealingConfig())
				if err != nil {
					return nil, err
				}
				solver, e = opt.NewHybridOptimizer(engine, ga, sa, opt.DefaultCompactionConfig())
			}
			if e != nil {
				return nil, e
			}
			return browserMeasuredOptimizer{solver, &calls}, nil
		}
	}
	registry, err := opt.NewOptimizerFactoryRegistry(factories)
	if err != nil {
		t.Fatal(err)
	}
	scoped := &postgres.ScopedDB{DB: db}
	codec := sf.NewCodec()
	run := uc.NewRunNestingUseCase(postgres.NewProjectSurfaceRepository(scoped, codec), postgres.NewProjectPatternRepository(scoped, codec), postgres.NewPlacementRepository(scoped, codec), sf.NewEngine(), registry, postgres.NewNestingUnitOfWork(db, codec))
	run.Transactions = scoped
	run.Budgets = scoped
	run.Previews = postgres.NewNestingPreviewRepository(scoped)
	run.History = postgres.NewNestingRunRepository(scoped)
	run.Operations = postgres.NewNestingOperationRepository(scoped)
	run.PreviewOperations = postgres.NewPreviewOperationRepository(scoped)
	deps.RunNestingUseCase = run
	app := transport.NewRouter(deps, cfg)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	stats := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result := map[string]any{"optimize_calls": calls.Load()}
		if raw := r.URL.Query().Get("surface_id"); raw != "" {
			id, err := strconv.ParseInt(raw, 10, 64)
			if err != nil || id <= 0 {
				http.Error(w, "invalid test surface", 400)
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			var count, runs int
			var snapshot string
			if err := db.GetContext(ctx, &count, "SELECT count(*) FROM placements WHERE project_surface_id=$1", id); err != nil {
				http.Error(w, "test query failed", 500)
				return
			}
			if err := db.GetContext(ctx, &runs, "SELECT count(*) FROM nesting_runs WHERE project_surface_id=$1", id); err != nil {
				http.Error(w, "test query failed", 500)
				return
			}
			if err := db.GetContext(ctx, &snapshot, `SELECT COALESCE(jsonb_agg(to_jsonb(p) ORDER BY id),'[]'::jsonb)::text FROM placements p WHERE project_surface_id=$1`, id); err != nil {
				http.Error(w, "test query failed", 500)
				return
			}
			result["placements"] = count
			result["nesting_runs"] = runs
			result["layout_hash"] = fmt.Sprintf("%x", sha256.Sum256([]byte(snapshot)))
		}
		_ = json.NewEncoder(w).Encode(result)
	}))
	defer stats.Close()
	var schema string
	requireProcessGet(t, db, &schema, "SELECT current_schema()")
	metadata, err := json.Marshal(map[string]any{"api": "http://" + listener.Addr().String(), "stats": stats.URL, "database": "nesting_test", "schema": schema, "pid": os.Getpid(), "origin": origin})
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(ready, metadata, 0600); err != nil {
		t.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(t.Context(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = app.ShutdownWithContext(shutdown)
	}()
	if err = app.Listener(listener); err != nil {
		t.Fatal(err)
	}
}
