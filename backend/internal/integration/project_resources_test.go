package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"server_nesting_optimizer/internal/app/container"
	"server_nesting_optimizer/internal/config"
	httperror "server_nesting_optimizer/internal/transport/http/errors"
	"server_nesting_optimizer/internal/transport/http/handler"
	"server_nesting_optimizer/internal/transport/http/middleware"
)

func TestProjectResourcesAPI(t *testing.T) {
	dsn := os.Getenv("NESTING_TEST_DSN")
	if dsn == "" {
		t.Skip("set NESTING_TEST_DSN to an isolated PostGIS database")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var database string
	if err = db.Get(&database, "SELECT current_database()"); err != nil {
		t.Fatal(err)
	}
	if database != "nesting_test" {
		t.Fatal("integration tests require the dedicated nesting_test database")
	}
	schema := fmt.Sprintf("test_%d", time.Now().UnixNano())
	if _, err = db.Exec("CREATE SCHEMA " + schema); err != nil {
		t.Fatal(err)
	}
	// Each run gets an isolated schema, retained for diagnosis.
	scoped, err := sqlx.Connect("postgres", dsn+"&search_path="+schema+",public")
	if err != nil {
		t.Fatal(err)
	}
	defer scoped.Close()
	migrations, err := filepath.Glob("../../migrations/*.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range migrations {
		source, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		up := strings.Split(string(source), "-- +goose Down")[0]
		if _, err = scoped.Exec(up); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
	if _, err = scoped.Exec(`INSERT INTO users(id,login,email,password_hash,first_name,last_name) VALUES (1,'owner','owner@test','test','Owner','Test'),(2,'other','other@test','test','Other','Test'); INSERT INTO projects(id,user_id,name) VALUES(1,1,'Project'),(2,2,'Other')`); err != nil {
		t.Fatal(err)
	}
	deps, err := container.New(scoped, config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	app := fiber.New(fiber.Config{ErrorHandler: httperror.Handle})
	app.Use(func(c fiber.Ctx) error { c.Locals(middleware.UserIDLocalKey, int64(1)); return c.Next() })
	sh := handler.NewSurfaceHandler(deps.CreateSurfaceUseCase, deps.GetSurfaceByIDUseCase, deps.ListSurfacesUseCase, deps.UpdateSurfaceUseCase, deps.DeleteSurfaceUseCase)
	ph := handler.NewPatternHandler(deps.CreatePatternUseCase, deps.GetPatternByIDUseCase, deps.ListPatternsUseCase, deps.UpdatePatternUseCase, deps.DeletePatternUseCase)
	ps := handler.NewProjectSurfaceHandler(deps.CreateProjectSurfaceUseCase, deps.GetProjectSurfaceByIDUseCase, deps.ListProjectSurfacesUseCase, deps.UpdateProjectSurfaceUseCase, deps.DeleteProjectSurfaceUseCase)
	pp := handler.NewProjectPatternHandler(deps.CreateProjectPatternUseCase, deps.GetProjectPatternByIDUseCase, deps.ListProjectPatternsUseCase, deps.UpdateProjectPatternUseCase, deps.DeleteProjectPatternUseCase)
	pl := handler.NewPlacementHandler(deps.CreatePlacementUseCase, deps.GetPlacementByIDUseCase, deps.ListPlacementsUseCase, deps.UpdatePlacementUseCase, deps.DeletePlacementUseCase)
	app.Post("/surfaces", sh.Create)
	app.Patch("/surfaces/:id", sh.Update)
	app.Delete("/surfaces/:id", sh.Delete)
	app.Post("/patterns", ph.Create)
	app.Patch("/patterns/:id", ph.Update)
	app.Delete("/patterns/:id", ph.Delete)
	app.Post("/projects/:project_id/surfaces", ps.Create)
	app.Get("/projects/:project_id/surfaces/:id", ps.GetByID)
	app.Patch("/projects/:project_id/surfaces/:id", ps.Update)
	app.Delete("/projects/:project_id/surfaces/:id", ps.Delete)
	app.Post("/projects/:project_id/patterns", pp.Create)
	app.Get("/projects/:project_id/patterns/:id", pp.GetByID)
	app.Patch("/projects/:project_id/patterns/:id", pp.Update)
	app.Delete("/projects/:project_id/patterns/:id", pp.Delete)
	app.Post("/projects/:project_id/surfaces/:project_surface_id/placements", pl.Create)
	app.Delete("/projects/:project_id/placements/:placement_id", pl.Delete)
	app.Patch("/projects/:project_id/placements/:placement_id", pl.Update)
	app.Get("/projects/:project_id/surfaces/:project_surface_id/placements", pl.ListPlacements)
	nh := handler.NewNestingHandler(deps.RunNestingUseCase)
	app.Post("/projects/:project_id/surfaces/:project_surface_id/nesting", nh.Run)
	request := func(method, path, body string) (int, map[string]any) {
		req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		res, err := app.Test(req, fiber.TestConfig{Timeout: 10 * time.Second})
		if err != nil {
			t.Error(err)
			return 0, nil
		}
		defer res.Body.Close()
		var data map[string]any
		if res.StatusCode != 204 {
			if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
				t.Error(err)
			}
		}
		return res.StatusCode, data
	}
	must := func(method, path, body string, want int) map[string]any {
		status, data := request(method, path, body)
		if status != want {
			t.Fatalf("%s %s: got %d want %d: %v", method, path, status, want, data)
		}
		return data
	}
	polygon := func(size int) string {
		return fmt.Sprintf(`{"exterior":{"points":[{"x":0,"y":0},{"x":%d,"y":0},{"x":%d,"y":%d},{"x":0,"y":%d}]},"holes":[]}`, size, size, size, size)
	}
	ids := map[string][]int{}
	for _, kind := range []string{"surfaces", "patterns"} {
		size := 100
		if kind == "patterns" {
			size = 10
		}
		source := must("POST", "/"+kind, `{"name":"Original","geometry":`+polygon(size)+`}`, 201)
		sourceID := int(source["id"].(float64))
		sourceKey := "source_surface_id"
		if kind == "patterns" {
			sourceKey = "source_pattern_id"
		}
		path := "/projects/1/" + kind
		for range 3 {
			item := must("POST", path, fmt.Sprintf(`{"%s":%d,"scale":1}`, sourceKey, sourceID), 201)
			ids[kind] = append(ids[kind], int(item["id"].(float64)))
		}
		if ids[kind][0] == ids[kind][1] {
			t.Fatal("repeated creation reused ID")
		}
		second := fmt.Sprintf("%s/%d", path, ids[kind][1])
		must("PATCH", second, `{"geometry":`+polygon(size*2)+`}`, 200)
		var sourceArea float64
		if err := scoped.Get(&sourceArea, "SELECT ST_Area(geometry) FROM "+kind+" WHERE id=$1", sourceID); err != nil {
			t.Fatal(err)
		}
		if sourceArea != float64(size*size) {
			t.Fatal("project edit changed library geometry")
		}

		first := must("GET", fmt.Sprintf("%s/%d", path, ids[kind][0]), "", 200)
		originalGeometry, _ := json.Marshal(first["geometry"])
		must("PATCH", fmt.Sprintf("/%s/%d", kind, sourceID), `{"geometry":`+polygon(size*3)+`}`, 200)
		after := must("GET", fmt.Sprintf("%s/%d", path, ids[kind][0]), "", 200)
		afterGeometry, _ := json.Marshal(after["geometry"])
		if string(originalGeometry) != string(afterGeometry) {
			t.Fatal("library edit changed project copy")
		}
		must("DELETE", fmt.Sprintf("/%s/%d", kind, sourceID), "", 204)
		must("GET", second, "", 200)
		must("PATCH", second, `{"geometry":{"exterior":{"points":[]},"holes":[]}}`, 400)
		must("PATCH", second, `{"scale":2,"geometry":`+polygon(size)+`}`, 400)
		must("GET", fmt.Sprintf("/projects/2/%s/%d", kind, ids[kind][0]), "", 404)
		must("DELETE", fmt.Sprintf("%s/%d", path, ids[kind][2]), "", 204)
	}
	surfacePath := fmt.Sprintf("/projects/1/surfaces/%d", ids["surfaces"][0])
	patternPath := fmt.Sprintf("/projects/1/patterns/%d", ids["patterns"][0])
	placementBody := fmt.Sprintf(`{"project_pattern_id":%d,"x":20,"y":20,"rotation":0}`, ids["patterns"][0])
	placement := must("POST", surfacePath+"/placements", placementBody, 201)
	placementPath := fmt.Sprintf("/projects/1/placements/%d", int(placement["id"].(float64)))
	must("PATCH", placementPath, `{"x":30,"y":30,"rotation":90}`, 200)
	must("POST", surfacePath+"/placements", fmt.Sprintf(`{"project_pattern_id":%d,"x":30,"y":30,"rotation":0}`, ids["patterns"][0]), 409)
	must("PATCH", placementPath, `{"x":-100}`, 409)
	restored := must("GET", surfacePath+"/placements", "", 200)
	if len(restored["items"].([]any)) != 1 {
		t.Fatal("saved placement missing")
	}

	for _, path := range []string{surfacePath, patternPath} {
		must("PATCH", path, `{"geometry":`+polygon(20)+`}`, 409)
		must("PATCH", path, `{"scale":2}`, 409)
		must("DELETE", path, "", 409)
		must("PATCH", path, `{"name":"Renamed"}`, 200)
	}
	must("DELETE", fmt.Sprintf("/projects/1/placements/%d", int(placement["id"].(float64))), "", 204)
	// Algorithm results are persisted and isolated to their project surface.
	holeGeometry := `{"exterior":{"points":[{"x":0,"y":0},{"x":100,"y":0},{"x":100,"y":100},{"x":0,"y":100}]},"holes":[{"points":[{"x":40,"y":40},{"x":40,"y":60},{"x":60,"y":60},{"x":60,"y":40}]}]}`
	must("PATCH", surfacePath, `{"geometry":`+holeGeometry+`}`, 200)
	must("POST", surfacePath+"/placements", fmt.Sprintf(`{"project_pattern_id":%d,"x":50,"y":50,"rotation":0}`, ids["patterns"][0]), 409)
	for _, algorithm := range []string{"baseline", "nfp_greedy", "genetic", "simulated_annealing", "hybrid"} {
		body := fmt.Sprintf(`{"algorithm":%q,"patterns":[{"project_pattern_id":%d,"quantity":2},{"project_pattern_id":%d,"quantity":1}],"allowed_rotations":[0,90],"keep_existing":false}`, algorithm, ids["patterns"][0], ids["patterns"][1])
		output := must("POST", surfacePath+"/nesting", body, 200)
		metrics := output["metrics"].(map[string]any)
		if metrics["surface_area"].(float64) != 9600 {
			t.Fatal("surface area must subtract hole")
		}
		saved := must("GET", surfacePath+"/placements", "", 200)
		if len(saved["items"].([]any)) != int(metrics["placed_count"].(float64)) {
			t.Fatal("algorithm results not persisted")
		}
		other := must("GET", fmt.Sprintf("/projects/1/surfaces/%d/placements", ids["surfaces"][1]), "", 200)
		if len(other["items"].([]any)) != 0 {
			t.Fatal("algorithm changed other surface")
		}
	}
	saved := must("GET", surfacePath+"/placements", "", 200)
	prior := len(saved["items"].([]any))
	output := must("POST", surfacePath+"/nesting", fmt.Sprintf(`{"algorithm":"baseline","patterns":[{"project_pattern_id":%d,"quantity":1}],"allowed_rotations":[0],"keep_existing":true}`, ids["patterns"][0]), 200)
	saved = must("GET", surfacePath+"/placements", "", 200)
	if len(saved["items"].([]any)) != prior+int(output["metrics"].(map[string]any)["placed_count"].(float64)) {
		t.Fatal("keep_existing lost placements")
	}
	for _, item := range saved["items"].([]any) {
		must("DELETE", fmt.Sprintf("/projects/1/placements/%d", int(item.(map[string]any)["id"].(float64))), "", 204)
	}

	// Either deletion wins and placement is rejected, or placement wins and deletion conflicts.
	var wg sync.WaitGroup
	var createStatus, deleteStatus int
	wg.Add(2)
	go func() { defer wg.Done(); createStatus, _ = request("POST", surfacePath+"/placements", placementBody) }()
	go func() { defer wg.Done(); deleteStatus, _ = request("DELETE", surfacePath, "") }()
	wg.Wait()
	if !((createStatus == 201 && deleteStatus == 409) || (createStatus == 404 && deleteStatus == 204)) {
		t.Fatalf("unsafe concurrent outcome create=%d delete=%d", createStatus, deleteStatus)
	}
	// A geometry update and a placement must observe each other's committed result.
	secondSurface := fmt.Sprintf("/projects/1/surfaces/%d", ids["surfaces"][1])
	var updateStatus int
	wg.Add(2)
	go func() { defer wg.Done(); createStatus, _ = request("POST", secondSurface+"/placements", placementBody) }()
	go func() {
		defer wg.Done()
		updateStatus, _ = request("PATCH", secondSurface, `{"geometry":`+polygon(5)+`}`)
	}()
	wg.Wait()
	if !((createStatus == 201 && updateStatus == 409) || (createStatus == 409 && updateStatus == 200)) {
		t.Fatalf("unsafe geometry race create=%d update=%d", createStatus, updateStatus)
	}

	var dangling int
	err = scoped.Get(&dangling, `SELECT count(*) FROM placements pl JOIN project_surfaces ps ON ps.id=pl.project_surface_id JOIN project_patterns pp ON pp.id=pl.project_pattern_id WHERE pl.deleted_at IS NULL AND (ps.deleted_at IS NOT NULL OR pp.deleted_at IS NOT NULL)`)
	if err != nil || dangling != 0 {
		t.Fatalf("dangling placements=%d err=%v", dangling, err)
	}
}
