package integration

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"server_nesting_optimizer/internal/app/container"
	"server_nesting_optimizer/internal/config"
	dg "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
	sf "server_nesting_optimizer/internal/geometry/simplefeatures"
	opt "server_nesting_optimizer/internal/nesting"
	"server_nesting_optimizer/internal/transport/http/dto"
	httperror "server_nesting_optimizer/internal/transport/http/errors"
	"server_nesting_optimizer/internal/transport/http/handler"
	"server_nesting_optimizer/internal/transport/http/middleware"
	uc "server_nesting_optimizer/internal/usecase/nesting"
)

func stage17Poly(xy ...float64) dg.Polygon {
	p := dg.Polygon{}
	for i := 0; i < len(xy); i += 2 {
		p.Exterior.Points = append(p.Exterior.Points, dg.Point{X: xy[i], Y: xy[i+1]})
	}
	return p
}
func stage17Dense(p dg.Polygon, n int) dg.Polygon {
	points := p.Exterior.Points
	p.Exterior.Points = nil
	for i, a := range points {
		b := points[(i+1)%len(points)]
		for j := 0; j < n; j++ {
			f := float64(j) / float64(n)
			p.Exterior.Points = append(p.Exterior.Points, dg.Point{X: a.X + (b.X-a.X)*f, Y: a.Y + (b.Y-a.Y)*f})
		}
	}
	return p
}
func TestStage17RealComplexityHTTP(t *testing.T) {
	f := newNestingFixture(t)
	engine := sf.NewEngine()
	codec := sf.NewCodec()
	a := stage17Dense(stage17Poly(0, 0, 10, 0, 10, 2, 2, 2, 2, 8, 10, 8, 10, 10, 0, 10), 128)
	b := stage17Dense(stage17Poly(0, 0, 2, 0, 2, 2, 0, 2), 256)
	shifted, err := engine.Transform(b, 5, 5, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = engine.InteriorsIntersect(a, shifted); !errors.Is(err, geometry.ErrComplexity) {
		t.Fatalf("missing real complexity: %v", err)
	}
	wa, _ := codec.EncodeWKB(a)
	wb, _ := codec.EncodeWKB(b)
	if _, err = f.db.Exec("UPDATE project_patterns SET geometry=ST_GeomFromWKB($1) WHERE id=1;", wa); err != nil {
		t.Fatal(err)
	}
	if _, err = f.db.Exec("INSERT INTO project_patterns(id,project_id,name,geometry) VALUES(2,1,'B',ST_GeomFromWKB($1))", wb); err != nil {
		t.Fatal(err)
	}
	var valid bool
	if err = f.db.Get(&valid, "SELECT bool_and(ST_IsValid(geometry)) FROM project_patterns"); err != nil || !valid {
		t.Fatal("invalid complexity fixture")
	}
	deps, err := container.New(f.db, config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	app := fiber.New(fiber.Config{ErrorHandler: httperror.Handle})
	app.Use(func(c fiber.Ctx) error { c.Locals(middleware.UserIDLocalKey, int64(1)); return c.Next() })
	h := handler.NewPlacementHandler(deps.CreatePlacementUseCase, deps.GetPlacementByIDUseCase, deps.ListPlacementsUseCase, deps.UpdatePlacementUseCase, deps.DeletePlacementUseCase)
	app.Post("/projects/:project_id/surfaces/:project_surface_id/placements", h.Create)
	app.Patch("/projects/:project_id/placements/:placement_id", h.Update)
	observed := []int{}
	request := func(target *fiber.App, method, path string, data any, key string) int {
		t.Helper()
		raw, _ := json.Marshal(data)
		req := httptest.NewRequest(method, path, bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		if key != "" {
			req.Header.Set("Idempotency-Key", key)
		}
		res, e := target.Test(req)
		if e != nil {
			t.Fatal(e)
		}
		defer func() { _ = res.Body.Close() }()
		return res.StatusCode
	}
	pose := func(id int, x, y float64) any {
		return map[string]any{"project_pattern_id": id, "x": x, "y": y, "rotation": 0}
	}
	if code := request(app, "POST", "/projects/1/surfaces/1/placements", pose(1, 5, 5), ""); code != 201 {
		t.Fatalf("seed create%d", code)
	}
	observed = append(observed, request(app, "POST", "/projects/1/surfaces/1/placements", pose(2, 5, 5), ""))
	f.counts(t, 1, 0)
	if code := request(app, "POST", "/projects/1/surfaces/1/placements", pose(2, 50, 50), ""); code != 201 {
		t.Fatal(code)
	}
	var id int
	requireProcessGet(t, f.db, &id, "SELECT id FROM placements WHERE project_pattern_id=2")
	observed = append(observed, request(app, "PATCH", fmt.Sprintf("/projects/1/placements/%d", id), map[string]float64{"x": 5, "y": 5}, ""))
	var x float64
	requireProcessGet(t, f.db, &x, "SELECT x FROM placements WHERE project_pattern_id=2")
	if x != 50 {
		t.Fatal("partial update")
	}
	if _, err = f.db.Exec("DELETE FROM placements WHERE project_pattern_id=2"); err != nil {
		t.Fatal(err)
	}
	for _, keep := range []bool{false, true} {
		result := opt.Result{Placements: []opt.Placement{{PatternID: 2, X: 5, Y: 5}}, Metrics: opt.Metrics{RequestedCount: 1, PlacedCount: 1, SurfaceArea: 10000, PlacedArea: 4, Utilization: .0004}}
		patterns := []dto.RunNestingPatternRequest{{ProjectPatternID: 2, Quantity: 1}}
		if !keep {
			result.Placements = append([]opt.Placement{{PatternID: 1, X: 5, Y: 5}}, result.Placements...)
			result.Metrics.RequestedCount = 2
			result.Metrics.PlacedCount = 2
			result.Metrics.PlacedArea = 56
			result.Metrics.Utilization = .0056
			patterns = append(patterns, dto.RunNestingPatternRequest{ProjectPatternID: 1, Quantity: 1})
		}
		solver := &strictRejectedOptimizer{result: result}
		run := nativeAlgorithmUseCase(t, f, opt.BaselineAlgorithm, solver)
		nh := handler.NewNestingHandler(run)
		target := fiber.New(fiber.Config{ErrorHandler: httperror.Handle})
		target.Use(func(c fiber.Ctx) error { c.Locals(middleware.UserIDLocalKey, int64(1)); return c.Next() })
		target.Post("/projects/:project_id/surfaces/:project_surface_id/nesting", nh.Run)
		target.Post("/projects/:project_id/surfaces/:project_surface_id/nesting/preview", nh.Preview)
		body := dto.RunNestingRequest{Algorithm: string(opt.BaselineAlgorithm), Patterns: patterns, AllowedRotations: []float64{0}, KeepExisting: keep}
		for _, suffix := range []string{"", "/preview"} {
			key := uuid.NewString()
			for i := 0; i < 2; i++ {
				observed = append(observed, request(target, "POST", "/projects/1/surfaces/1/nesting"+suffix, body, key))
				f.counts(t, 1, 0)
			}
		}
		input := f.input
		input.KeepExisting = keep
		input.Patterns = nil
		for _, p := range patterns {
			input.Patterns = append(input.Patterns, uc.RunNestingPatternInput{ProjectPatternID: p.ProjectPatternID, Quantity: p.Quantity})
		}
		directKey := uuid.NewString()
		for replay := 0; replay < 2; replay++ {
			_, cause := run.ExecuteWithKey(t.Context(), input, directKey)
			if !errors.Is(cause, geometry.ErrComplexity) {
				t.Errorf("durable complexity cause lost, replay=%d", replay)
			}
		}
		if solver.calls.Load() != 3 {
			t.Fatal("budget failure replay recomputed")
		}
	}
	t.Logf("REAL_COMPLEXITY_HTTP statuses=%v", observed)
	for _, status := range observed {
		if status != 400 {
			t.Errorf("complexity classified as%d, expected400", status)
		}
	}
}

func TestStage17CoveragePostGIS(t *testing.T) {
	path := os.Getenv("NESTING_COVERAGE_CORPUS")
	if path == "" {
		t.Skip("explicit exact inclusion matrix")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Name         string
		Inner, Outer dg.Polygon
		Want         bool
	}
	if err = json.Unmarshal(data, &rows); err != nil || len(rows) < 141 {
		t.Fatal("incomplete coverage matrix")
	}
	db := newTestSchema(t)
	applyAllMigrations(t, db)
	codec := sf.NewCodec()
	for i, row := range rows {
		a, _ := codec.EncodeWKB(row.Inner)
		b, _ := codec.EncodeWKB(row.Outer)
		var valid, covered bool
		if err = db.QueryRow("SELECT ST_IsValid(ST_GeomFromWKB($1)) AND ST_IsValid(ST_GeomFromWKB($2)),ST_CoveredBy(ST_GeomFromWKB($1),ST_GeomFromWKB($2))", a, b).Scan(&valid, &covered); err != nil || !valid || covered != row.Want {
			t.Fatalf("coverage oracle mismatch case%d/%s valid%v got%v exact%v err%v", i, row.Name, valid, covered, row.Want, err)
		}
	}
	t.Logf("COVERAGE_POSTGIS cases=%d", len(rows))
}

func TestStage17CoverageHTTP(t *testing.T) {
	f := newNestingFixture(t)
	codec := sf.NewCodec()
	engine := sf.NewEngine()
	data, err := os.ReadFile("../geometry/simplefeatures/testdata/stage17_coverage_counterexample.json")
	if err != nil {
		t.Fatal(err)
	}
	var minimal struct{ Inner, Outer dg.Polygon }
	if err = json.Unmarshal(data, &minimal); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile("../nesting/testdata/stage12_property12_counterexample.json")
	if err != nil {
		t.Fatal(err)
	}
	var original struct {
		Problem opt.Problem
		Result  opt.Result
	}
	if err = json.Unmarshal(data, &original); err != nil {
		t.Fatal(err)
	}
	pattern := original.Problem.Patterns[0].Geometry
	pose := original.Result.Placements[1]
	actual, err := engine.Transform(pattern, pose.X, pose.Y, pose.Rotation)
	if err != nil {
		t.Fatal(err)
	}
	wa, _ := codec.EncodeWKB(actual)
	wo, _ := codec.EncodeWKB(minimal.Outer)
	wp, _ := codec.EncodeWKB(pattern)
	var covered bool
	if err = f.db.Get(&covered, "SELECT ST_CoveredBy(ST_GeomFromWKB($1),ST_GeomFromWKB($2))", wa, wo); err != nil || covered {
		t.Fatal("actual transformed coverage proof missing")
	}
	if _, err = f.db.Exec("UPDATE project_surfaces SET geometry=ST_GeomFromWKB($1) WHERE id=1", wo); err != nil {
		t.Fatal(err)
	}
	if _, err = f.db.Exec("UPDATE project_patterns SET geometry=ST_GeomFromWKB($1) WHERE id=1", wp); err != nil {
		t.Fatal(err)
	}
	deps, err := container.New(f.db, config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	app := fiber.New(fiber.Config{ErrorHandler: httperror.Handle})
	app.Use(func(c fiber.Ctx) error { c.Locals(middleware.UserIDLocalKey, int64(1)); return c.Next() })
	pl := handler.NewPlacementHandler(deps.CreatePlacementUseCase, deps.GetPlacementByIDUseCase, deps.ListPlacementsUseCase, deps.UpdatePlacementUseCase, deps.DeletePlacementUseCase)
	app.Post("/projects/:project_id/surfaces/:project_surface_id/placements", pl.Create)
	app.Patch("/projects/:project_id/placements/:placement_id", pl.Update)
	send := func(method, path string, x, y, rotation float64, want int) {
		t.Helper()
		raw, _ := json.Marshal(map[string]any{"project_pattern_id": 1, "x": x, "y": y, "rotation": rotation})
		req := httptest.NewRequest(method, path, bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		res, e := app.Test(req)
		if e != nil {
			t.Fatal(e)
		}
		defer func() { _ = res.Body.Close() }()
		if res.StatusCode != want {
			t.Fatalf("coverage HTTP %d want%d", res.StatusCode, want)
		}
	}
	send("POST", "/projects/1/surfaces/1/placements", pose.X, pose.Y, pose.Rotation, 409)
	f.counts(t, 0, 0)
	send("POST", "/projects/1/surfaces/1/placements", 20, 0, 0, 201)
	var id int
	requireProcessGet(t, f.db, &id, "SELECT id FROM placements")
	send("PATCH", fmt.Sprintf("/projects/1/placements/%d", id), pose.X, pose.Y, pose.Rotation, 409)
	var x float64
	requireProcessGet(t, f.db, &x, "SELECT x FROM placements")
	if x != 20 {
		t.Fatal("coverage rejection changed saved pose")
	}
	f.counts(t, 1, 0)
	t.Log("COVERAGE actual Transform/WKB + manual create/update rejection preserves SQL PASS")
}
