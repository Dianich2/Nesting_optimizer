package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"server_nesting_optimizer/internal/app/container"
	"server_nesting_optimizer/internal/config"
	dg "server_nesting_optimizer/internal/domain/geometry"
	sf "server_nesting_optimizer/internal/geometry/simplefeatures"
	opt "server_nesting_optimizer/internal/nesting"
	"server_nesting_optimizer/internal/transport/http/dto"
	httperror "server_nesting_optimizer/internal/transport/http/errors"
	"server_nesting_optimizer/internal/transport/http/handler"
	"server_nesting_optimizer/internal/transport/http/middleware"
)

type strictRejectedOptimizer struct {
	result opt.Result
	calls  atomic.Int64
}

func (s *strictRejectedOptimizer) Optimize(context.Context, opt.Problem) (opt.Result, error) {
	s.calls.Add(1)
	return s.result, nil
}
func TestStrictTopologyManualAndNestingAPI(t *testing.T) {
	data, err := os.ReadFile("../nesting/testdata/stage12_property12_counterexample.json")
	if err != nil {
		t.Fatal(err)
	}
	var original struct {
		Problem opt.Problem
		Result  opt.Result
		Formed  []dg.Polygon
	}
	if err = json.Unmarshal(data, &original); err != nil {
		t.Fatal(err)
	}
	f := newNestingFixture(t)
	codec := sf.NewCodec()
	engine := sf.NewEngine()
	w, err := codec.EncodeWKB(original.Problem.Patterns[0].Geometry)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.db.Exec("UPDATE project_patterns SET geometry=ST_GeomFromWKB($1) WHERE id=1", w); err != nil {
		t.Fatal(err)
	}
	formed := make([]dg.Polygon, 2)
	for i, p := range original.Result.Placements {
		formed[i], err = engine.Transform(original.Problem.Patterns[0].Geometry, p.X, p.Y, p.Rotation)
		if err != nil {
			t.Fatal(err)
		}
	}
	a, _ := codec.EncodeWKB(formed[0])
	b, _ := codec.EncodeWKB(formed[1])
	var valid, overlap bool
	var relate string
	if err = f.db.QueryRow("SELECT ST_IsValid(ST_GeomFromWKB($1)) AND ST_IsValid(ST_GeomFromWKB($2)),ST_Relate(ST_GeomFromWKB($1),ST_GeomFromWKB($2)),ST_Relate(ST_GeomFromWKB($1),ST_GeomFromWKB($2),'2********')", a, b).Scan(&valid, &relate, &overlap); err != nil || !valid || !overlap {
		t.Fatalf("WKB oracle valid=%v overlap=%v err=%v", valid, overlap, err)
	}
	t.Logf("actual manual Transform WKB: PostGIS relation=%s interiors=%v", relate, overlap)
	deps, err := container.New(f.db, config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	app := fiber.New(fiber.Config{ErrorHandler: httperror.Handle})
	app.Use(func(c fiber.Ctx) error { c.Locals(middleware.UserIDLocalKey, int64(1)); return c.Next() })
	pl := handler.NewPlacementHandler(deps.CreatePlacementUseCase, deps.GetPlacementByIDUseCase, deps.ListPlacementsUseCase, deps.UpdatePlacementUseCase, deps.DeletePlacementUseCase)
	app.Post("/projects/:project_id/surfaces/:project_surface_id/placements", pl.Create)
	app.Patch("/projects/:project_id/placements/:placement_id", pl.Update)
	request := func(method, path string, body any, want int) {
		t.Helper()
		raw, e := json.Marshal(body)
		if e != nil {
			t.Fatal(e)
		}
		req := httptest.NewRequest(method, path, bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		res, e := app.Test(req)
		if e != nil {
			t.Fatal(e)
		}
		defer func() { _ = res.Body.Close() }()
		if res.StatusCode != want {
			var result any
			_ = json.NewDecoder(res.Body).Decode(&result)
			t.Fatalf("%s %s status=%d want=%d body=%v", method, path, res.StatusCode, want, result)
		}
	}
	body := func(p opt.Placement) any {
		return map[string]any{"project_surface_id": 1, "project_pattern_id": 1, "x": p.X, "y": p.Y, "rotation": p.Rotation}
	}
	request("POST", "/projects/1/surfaces/1/placements", body(original.Result.Placements[0]), 201)
	request("POST", "/projects/1/surfaces/1/placements", body(original.Result.Placements[1]), 409)
	f.counts(t, 1, 0)
	request("POST", "/projects/1/surfaces/1/placements", body(opt.Placement{X: 80, Y: 80, Rotation: 90}), 201)
	var id int64
	requireProcessGet(t, f.db, &id, "SELECT id FROM placements WHERE x=80")
	p := original.Result.Placements[1]
	request("PATCH", fmt.Sprintf("/projects/1/placements/%d", id), map[string]float64{"x": p.X, "y": p.Y, "rotation": p.Rotation}, 409)
	var x, y float64
	if err = f.db.QueryRow("SELECT x,y FROM placements WHERE id=$1", id).Scan(&x, &y); err != nil || x != 80 || y != 80 {
		t.Fatalf("rejected update changed pose %g %g %v", x, y, err)
	}
	if _, err = f.db.Exec("DELETE FROM placements WHERE id=$1", id); err != nil {
		t.Fatal(err)
	}
	// A deliberately buggy optimizer returns the old illegal result. Validation
	// must reject before publishing a preview, history or partial placements.
	for _, keep := range []bool{false, true} {
		result := original.Result
		quantity := 2
		if keep {
			quantity = 1
			result.Placements = result.Placements[1:]
			result.Metrics.RequestedCount = 1
			result.Metrics.PlacedCount = 1
			result.Metrics.PlacedArea /= 2
			result.Metrics.Utilization /= 2
		}
		solver := &strictRejectedOptimizer{result: result}
		run := nativeAlgorithmUseCase(t, f, opt.BaselineAlgorithm, solver)
		h := handler.NewNestingHandler(run)
		testapp := fiber.New(fiber.Config{ErrorHandler: httperror.Handle})
		testapp.Use(func(c fiber.Ctx) error { c.Locals(middleware.UserIDLocalKey, int64(1)); return c.Next() })
		testapp.Post("/projects/:project_id/surfaces/:project_surface_id/nesting", h.Run)
		testapp.Post("/projects/:project_id/surfaces/:project_surface_id/nesting/preview", h.Preview)
		raw, _ := json.Marshal(dto.RunNestingRequest{Algorithm: string(opt.BaselineAlgorithm), Patterns: []dto.RunNestingPatternRequest{{ProjectPatternID: 1, Quantity: quantity}}, AllowedRotations: []float64{0, 90}, KeepExisting: keep})
		for _, suffix := range []string{"", "/preview"} {
			key := uuid.NewString()
			for replay := 0; replay < 2; replay++ {
				req := httptest.NewRequest("POST", "/projects/1/surfaces/1/nesting"+suffix, bytes.NewReader(raw))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Idempotency-Key", key)
				res, e := testapp.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
				if e != nil {
					t.Fatal(e)
				}
				var failure struct {
					Error struct {
						Code string `json:"code"`
					} `json:"error"`
				}
				decodeErr := json.NewDecoder(res.Body).Decode(&failure)
				_ = res.Body.Close()
				if decodeErr != nil || res.StatusCode != 500 || failure.Error.Code != "internal_error" {
					t.Fatalf("invalid optimizer response: status%d code%s err%v", res.StatusCode, failure.Error.Code, decodeErr)
				}
				f.counts(t, 1, 0)
			}
		}
		if solver.calls.Load() != 2 {
			t.Fatalf("failed replay recomputed: %d", solver.calls.Load())
		}
		var ready, completed int
		requireProcessGet(t, f.db, &ready, "SELECT count(*) FROM nesting_previews")
		requireProcessGet(t, f.db, &completed, "SELECT (SELECT count(*) FROM nesting_operations WHERE state='completed')+(SELECT count(*) FROM nesting_preview_operations WHERE state='completed')")
		if ready != 0 || completed != 0 {
			t.Fatalf("false completed/ready: %d/%d", completed, ready)
		}
	}
	if _, err = f.db.Exec("DELETE FROM placements; UPDATE project_patterns SET geometry=ST_MakeEnvelope(0,0,2,2) WHERE id=1"); err != nil {
		t.Fatal(err)
	}
	request("POST", "/projects/1/surfaces/1/placements", body(opt.Placement{X: 10, Y: 10}), 201)
	request("POST", "/projects/1/surfaces/1/placements", body(opt.Placement{X: 12, Y: 10}), 201)
	request("POST", "/projects/1/surfaces/1/placements", body(opt.Placement{X: math.Nextafter(8, 10), Y: 10}), 409)
	f.counts(t, 2, 0)
}

func TestStrictTransitionPostGIS(t *testing.T) {
	path := os.Getenv("NESTING_STRICT_TRANSITION")
	if path == "" {
		t.Skip("explicit bounded strict-transition corpus")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var evidence struct {
		Pairs []struct {
			First, Second dg.Polygon
			Old, Strict   bool
		}
		Rows []struct {
			Problem opt.Problem
			Formed  []dg.Polygon
		}
	}
	if err = json.Unmarshal(data, &evidence); err != nil || len(evidence.Pairs) == 0 {
		t.Fatalf("missing evidence: %v", err)
	}
	db := newTestSchema(t)
	applyAllMigrations(t, db)
	codec := sf.NewCodec()
	encode := func(p dg.Polygon) []byte {
		t.Helper()
		w, e := codec.EncodeWKB(p)
		if e != nil {
			t.Fatal(e)
		}
		return w
	}
	check := func(a, b dg.Polygon, want bool) {
		t.Helper()
		var valid, overlap bool
		if e := db.QueryRow("SELECT ST_IsValid(ST_GeomFromWKB($1)) AND ST_IsValid(ST_GeomFromWKB($2)),ST_Relate(ST_GeomFromWKB($1),ST_GeomFromWKB($2),'2********')", encode(a), encode(b)).Scan(&valid, &overlap); e != nil || !valid || overlap != want {
			t.Fatalf("strict transition PostGIS valid=%v overlap=%v want=%v err=%v", valid, overlap, want, e)
		}
	}
	for _, p := range evidence.Pairs {
		check(p.First, p.Second, p.Strict)
	}
	placements, pairs := 0, 0
	for _, row := range evidence.Rows {
		for i, p := range row.Formed {
			var covered bool
			if err = db.Get(&covered, "SELECT ST_CoveredBy(ST_GeomFromWKB($1),ST_GeomFromWKB($2))", encode(p), encode(row.Problem.Surface)); err != nil || !covered {
				t.Fatalf("transition coverage: %v", err)
			}
			for _, other := range row.Problem.Obstacles {
				check(p, other, false)
				pairs++
			}
			for _, other := range row.Formed[:i] {
				check(p, other, false)
				pairs++
			}
			placements++
		}
	}
	t.Logf("TRANSITION_POSTGIS changed_pairs=%d placements=%d disjoint_pairs=%d", len(evidence.Pairs), placements, pairs)
}

func TestStrictMatrixPostGIS(t *testing.T) {
	path := os.Getenv("NESTING_STRICT_MATRIX")
	if path == "" {
		t.Skip("explicit deterministic matrix")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		First, Second dg.Polygon
		Want          bool
	}
	if err = json.Unmarshal(data, &rows); err != nil || len(rows) < 18 {
		t.Fatalf("matrix incomplete: %v", err)
	}
	db := newTestSchema(t)
	applyAllMigrations(t, db)
	codec := sf.NewCodec()
	for i, p := range rows {
		a, _ := codec.EncodeWKB(p.First)
		b, _ := codec.EncodeWKB(p.Second)
		var valid, overlap bool
		if err = db.QueryRow("SELECT ST_IsValid(ST_GeomFromWKB($1)) AND ST_IsValid(ST_GeomFromWKB($2)),ST_Relate(ST_GeomFromWKB($1),ST_GeomFromWKB($2),'2********')", a, b).Scan(&valid, &overlap); err != nil || !valid || overlap != p.Want {
			t.Fatalf("matrix case%d valid=%v overlap=%v want=%v err=%v", i, valid, overlap, p.Want, err)
		}
	}
	t.Logf("MATRIX_POSTGIS pairs=%d", len(rows))
}
