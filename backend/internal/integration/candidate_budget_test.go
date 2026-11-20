package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"math"
	"net/http/httptest"
	dg "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry/nfp"
	sf "server_nesting_optimizer/internal/geometry/simplefeatures"
	opt "server_nesting_optimizer/internal/nesting"
	"server_nesting_optimizer/internal/repository/postgres"
	"server_nesting_optimizer/internal/transport/http/dto"
	httperror "server_nesting_optimizer/internal/transport/http/errors"
	"server_nesting_optimizer/internal/transport/http/handler"
	"server_nesting_optimizer/internal/transport/http/middleware"
	uc "server_nesting_optimizer/internal/usecase/nesting"
	"sync/atomic"
	"testing"
	"time"
)

func nativeTestOptimizer(t *testing.T, algorithm opt.Algorithm) opt.Optimizer {
	t.Helper()
	engine := sf.NewEngine()
	builder := nfp.NewCachedBuilder(nfp.NewBuilder(engine))
	switch algorithm {
	case opt.BaselineAlgorithm:
		return opt.NewBaselineOptimizer(engine)
	case opt.NFPGreedyAlgorithm:
		return opt.NewNFPGreedyOptimizer(engine, builder)
	case opt.GeneticAlgorithm:
		o, e := opt.NewGeneticOptimizer(engine, builder, opt.DefaultGeneticConfig())
		if e != nil {
			t.Fatal(e)
		}
		return o
	case opt.SimulatedAnnealingAlgorithm:
		o, e := opt.NewSimulatedAnnealingOptimizer(engine, builder, opt.DefaultAnnealingConfig())
		if e != nil {
			t.Fatal(e)
		}
		return o
	case opt.HybridAlgorithm:
		ga, e := opt.NewGeneticOptimizer(engine, builder, opt.DefaultGeneticConfig())
		if e != nil {
			t.Fatal(e)
		}
		sa, e := opt.NewSimulatedAnnealingOptimizer(engine, builder, opt.DefaultAnnealingConfig())
		if e != nil {
			t.Fatal(e)
		}
		o, e := opt.NewHybridOptimizer(engine, ga, sa, opt.DefaultCompactionConfig())
		if e != nil {
			t.Fatal(e)
		}
		return o
	}
	t.Fatal("unknown test optimizer")
	return nil
}
func nativeAlgorithmUseCase(t *testing.T, f nestingFixture, a opt.Algorithm, solver opt.Optimizer) *uc.RunNestingUseCase {
	t.Helper()
	codec := sf.NewCodec()
	registry, e := opt.NewOptimizerRegistry(map[opt.Algorithm]opt.Optimizer{a: solver})
	if e != nil {
		t.Fatal(e)
	}
	run := uc.NewRunNestingUseCase(postgres.NewProjectSurfaceRepository(f.scoped, codec), postgres.NewProjectPatternRepository(f.scoped, codec), postgres.NewPlacementRepository(f.scoped, codec), sf.NewEngine(), registry, postgres.NewNestingUnitOfWork(f.db, codec))
	run.Transactions = f.scoped
	run.Budgets = f.scoped
	run.Previews = postgres.NewNestingPreviewRepository(f.scoped)
	run.History = postgres.NewNestingRunRepository(f.scoped)
	run.Operations = postgres.NewNestingOperationRepository(f.scoped)
	run.PreviewOperations = postgres.NewPreviewOperationRepository(f.scoped)
	return run
}
func regularTestPolygon(n int, x, y, r float64) dg.Polygon {
	p := dg.Polygon{}
	for i := 0; i < n; i++ {
		a := 2 * math.Pi * float64(i) / float64(n)
		p.Exterior.Points = append(p.Exterior.Points, dg.Point{X: x + r*math.Cos(a), Y: y + r*math.Sin(a)})
	}
	return p
}
func TestCandidateBudgetHTTPPreservesLayout(t *testing.T) {
	for _, a := range []opt.Algorithm{opt.BaselineAlgorithm, opt.NFPGreedyAlgorithm, opt.GeneticAlgorithm, opt.SimulatedAnnealingAlgorithm, opt.HybridAlgorithm} {
		for _, keep := range []bool{false, true} {
			t.Run(string(a)+map[bool]string{true: "/keep", false: "/replace"}[keep], func(t *testing.T) {
				f := newNestingFixture(t)
				codec := sf.NewCodec()
				surface, err := codec.EncodeWKB(regularTestPolygon(512, 50, 50, 45))
				if err != nil {
					t.Fatal(err)
				}
				pattern, err := codec.EncodeWKB(regularTestPolygon(512, 0, 0, 1))
				if err != nil {
					t.Fatal(err)
				}
				if _, err = f.db.Exec("UPDATE project_surfaces SET geometry=ST_GeomFromWKB($1) WHERE id=1", surface); err != nil {
					t.Fatal(err)
				}
				if _, err = f.db.Exec("UPDATE project_patterns SET geometry=ST_GeomFromWKB($1) WHERE id=1", pattern); err != nil {
					t.Fatal(err)
				}
				if _, err = f.db.Exec("INSERT INTO placements(id,project_surface_id,project_pattern_id,x,y,rotation)VALUES(99,1,1,50,50,0)"); err != nil {
					t.Fatal(err)
				}
				var calls atomic.Int64
				run := nativeAlgorithmUseCase(t, f, a, browserMeasuredOptimizer{nativeTestOptimizer(t, a), &calls})
				app := fiber.New(fiber.Config{ErrorHandler: httperror.Handle})
				app.Use(func(c fiber.Ctx) error { c.Locals(middleware.UserIDLocalKey, int64(1)); return c.Next() })
				h := handler.NewNestingHandler(run)
				app.Post("/projects/:project_id/surfaces/:project_surface_id/nesting", h.Run)
				app.Post("/projects/:project_id/surfaces/:project_surface_id/nesting/preview", h.Preview)
				body, e := json.Marshal(dto.RunNestingRequest{Algorithm: string(a), Patterns: []dto.RunNestingPatternRequest{{ProjectPatternID: 1, Quantity: 1}}, AllowedRotations: []float64{0, 90, 180, 270}, KeepExisting: keep})
				if e != nil {
					t.Fatal(e)
				}
				for _, suffix := range []string{"/preview", ""} {
					key := uuid.NewString()
					for repeat := 0; repeat < 2; repeat++ {
						req := httptest.NewRequest("POST", "/projects/1/surfaces/1/nesting"+suffix, bytes.NewReader(body))
						req.Header.Set("Content-Type", "application/json")
						req.Header.Set("Idempotency-Key", key)
						response, e := app.Test(req, fiber.TestConfig{Timeout: 10 * time.Second})
						if e != nil {
							t.Fatal(e)
						}
						var data struct {
							Error struct {
								Code string `json:"code"`
							} `json:"error"`
						}
						e = json.NewDecoder(response.Body).Decode(&data)
						_ = response.Body.Close()
						if e != nil || response.StatusCode != 400 || data.Error.Code != "validation_error" {
							t.Fatalf("budget HTTP status=%d code=%s err=%v", response.StatusCode, data.Error.Code, e)
						}
						f.counts(t, 1, 0)
					}
				}
				if calls.Load() != 2 {
					t.Fatalf("idempotent budget failure recomputed: %d", calls.Load())
				}
				var id, ready int
				requireProcessGet(t, f.db, &id, "SELECT id FROM placements")
				requireProcessGet(t, f.db, &ready, "SELECT count(*) FROM nesting_previews")
				if id != 99 || ready != 0 {
					t.Fatal("complexity changed existing layout/ready preview")
				}
				ctx, cancel := context.WithCancel(t.Context())
				cancel()
				input := f.input
				input.Algorithm = a
				input.KeepExisting = keep
				if _, err = run.ExecuteWithKey(ctx, input, uuid.NewString()); err == nil {
					t.Fatal("cancelled SQL operation succeeded")
				}
				f.counts(t, 1, 0)
			})
		}
	}
}
