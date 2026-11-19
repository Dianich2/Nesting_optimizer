package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
	optimizer "server_nesting_optimizer/internal/nesting"
	"server_nesting_optimizer/internal/repository/postgres"
	nestingcase "server_nesting_optimizer/internal/usecase/nesting"
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

func newTestSchema(t *testing.T) *sqlx.DB {
	t.Helper()
	dsn := os.Getenv("NESTING_TEST_DSN")
	if dsn == "" {
		t.Skip("set NESTING_TEST_DSN to an isolated PostGIS database")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
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
	t.Cleanup(func() { _ = scoped.Close() })
	return scoped
}

func applyAllMigrations(t *testing.T, scoped *sqlx.DB) {
	t.Helper()
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
}

func TestProjectResourcesAPI(t *testing.T) {
	scoped := newTestSchema(t)
	applyAllMigrations(t, scoped)

	if _, err := scoped.Exec(`INSERT INTO users(id,login,email,password_hash,first_name,last_name) VALUES (1,'owner','owner@test','test','Owner','Test'),(2,'other','other@test','test','Other','Test'); INSERT INTO projects(id,user_id,name) VALUES(1,1,'Project'),(2,2,'Other')`); err != nil {
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
	projects := handler.NewProjectHandler(deps.CreateProjectUseCase, deps.GetProjectByIDUseCase, deps.ListProjectsUseCase, deps.UpdateProjectUseCase, deps.DeleteProjectUseCase)
	app.Delete("/projects/:id", projects.DeleteProject)

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
	app.Post("/projects/:project_id/surfaces/:project_surface_id/nesting/preview", nh.Preview)
	app.Get("/projects/:project_id/surfaces/:project_surface_id/nesting/operations/:operation_key", nh.GetOperation)
	app.Get("/projects/:project_id/surfaces/:project_surface_id/nesting/history", nh.History)
	request := func(method, path, body string, operationKey ...string) (int, map[string]any) {
		req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		if len(operationKey) > 0 {
			req.Header.Set("Idempotency-Key", operationKey[0])
			if len(operationKey) > 1 {
				req.Header.Set("X-Placement-Version", operationKey[1])
			}
		}
		res, err := app.Test(req, fiber.TestConfig{Timeout: 10 * time.Second})
		if err != nil {
			t.Error(err)
			return 0, nil
		}
		defer func() { _ = res.Body.Close() }()
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
		if copy := must("GET", second, "", 200); copy[sourceKey] != nil {
			t.Fatal("source must be cleared on physical deletion")
		}
		var removed int
		if err := scoped.Get(&removed, "SELECT count(*) FROM "+kind+" WHERE id=$1", sourceID); err != nil || removed != 0 {
			t.Fatalf("source still present: %d, %v", removed, err)
		}

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
	for _, method := range []string{"PATCH", "DELETE"} {
		stale, _ := request(method, placementPath, `{"x":35}`, "", placement["updated_at"].(string))
		if stale != 409 {
			t.Fatalf("%s accepted stale version: %d", method, stale)
		}
	}

	must("POST", surfacePath+"/placements", fmt.Sprintf(`{"project_pattern_id":%d,"x":30,"y":30,"rotation":0}`, ids["patterns"][0]), 409)
	must("PATCH", placementPath, `{"x":-100}`, 409)
	restored := must("GET", surfacePath+"/placements", "", 200)
	if len(restored["items"].([]any)) != 1 {
		t.Fatal("saved placement missing")
	}

	scopedDB := &postgres.ScopedDB{DB: scoped}
	codec := simplefeatures.NewCodec()
	registry, err := optimizer.NewOptimizerRegistry(map[optimizer.Algorithm]optimizer.Optimizer{
		optimizer.BaselineAlgorithm: invalidLayoutOptimizer{},
	})
	if err != nil {
		t.Fatal(err)
	}
	invalidRun := nestingcase.NewRunNestingUseCase(
		postgres.NewProjectSurfaceRepository(scopedDB, codec),
		postgres.NewProjectPatternRepository(scopedDB, codec),
		postgres.NewPlacementRepository(scopedDB, codec),
		simplefeatures.NewEngine(), registry, postgres.NewNestingUnitOfWork(scoped, codec),
	)
	invalidRun.Transactions = scopedDB
	app.Post("/invalid/projects/:project_id/surfaces/:project_surface_id/nesting", handler.NewNestingHandler(invalidRun).Run)
	must("POST", "/invalid"+surfacePath+"/nesting", fmt.Sprintf(`{"algorithm":"baseline","patterns":[{"project_pattern_id":%d,"quantity":1}],"allowed_rotations":[0],"keep_existing":false}`, ids["patterns"][0]), 500)
	if after := must("GET", surfacePath+"/placements", "", 200); !reflect.DeepEqual(after, restored) {
		t.Fatalf("invalid replacement changed saved geometry: %v", after)
	}
	var runCount int
	if err := scoped.Get(&runCount, "SELECT count(*) FROM nesting_runs"); err != nil || runCount != 0 {
		t.Fatalf("invalid result created a run: %d, %v", runCount, err)
	}

	cancelOptimizer := waitingLayoutOptimizer{started: make(chan struct{}, 1)}
	cancelRegistry, err := optimizer.NewOptimizerRegistry(map[optimizer.Algorithm]optimizer.Optimizer{
		optimizer.BaselineAlgorithm: cancelOptimizer,
	})
	if err != nil {
		t.Fatal(err)
	}
	cancelRun := nestingcase.NewRunNestingUseCase(
		postgres.NewProjectSurfaceRepository(scopedDB, codec),
		postgres.NewProjectPatternRepository(scopedDB, codec),
		postgres.NewPlacementRepository(scopedDB, codec),
		simplefeatures.NewEngine(), cancelRegistry, postgres.NewNestingUnitOfWork(scoped, codec),
	)
	cancelRun.Transactions = scopedDB
	cancelRun.Operations = postgres.NewNestingOperationRepository(scopedDB)
	cancelHandler := handler.NewNestingHandler(cancelRun)
	cancelBase := "/cancel/projects/:project_id/surfaces/:project_surface_id/nesting"
	app.Post(cancelBase, cancelHandler.Run)
	app.Post(cancelBase+"/preview", cancelHandler.Preview)
	app.Delete(cancelBase+"/operations/:operation_key", cancelHandler.Cancel)
	body := fmt.Sprintf(`{"algorithm":"baseline","patterns":[{"project_pattern_id":%d,"quantity":1}],"allowed_rotations":[0],"keep_existing":false}`, ids["patterns"][0])
	for i, suffix := range []string{"/preview", ""} {
		key := fmt.Sprintf("d3faa873-370c-47ad-824a-5a3756c5767%d", i)
		cancelPath := "/cancel" + surfacePath + "/nesting"
		operationPath := cancelPath + "/operations/" + key
		done := make(chan bool, 1)
		go func() {
			status, response := request("POST", cancelPath+suffix, body, key)
			errorBody, _ := response["error"].(map[string]any)
			done <- status == 409 && errorBody["code"] == "nesting_cancelled"
		}()
		select {
		case <-cancelOptimizer.started:
		case <-time.After(3 * time.Second):
			t.Fatal("optimizer did not start")
		}
		must("DELETE", operationPath, "", 204)
		select {
		case cancelled := <-done:
			if !cancelled {
				t.Fatalf("execution %q did not return nesting_cancelled", suffix)
			}
		case <-time.After(3 * time.Second):
			t.Fatal("optimizer did not stop after server cancellation")
		}
		if after := must("GET", surfacePath+"/placements", "", 200); !reflect.DeepEqual(after, restored) {
			t.Fatalf("cancelled execution changed placements: %v", after)
		}
	}
	if err := scoped.Get(&runCount, "SELECT count(*) FROM nesting_runs"); err != nil || runCount != 0 {
		t.Fatalf("cancelled execution created history: %d, %v", runCount, err)
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
	for expectedRuns, algorithm := range []string{"baseline", "nfp_greedy", "genetic", "simulated_annealing", "hybrid"} {
		body := fmt.Sprintf(`{"algorithm":%q,"patterns":[{"project_pattern_id":%d,"quantity":2},{"project_pattern_id":%d,"quantity":1}],"allowed_rotations":[0,90],"keep_existing":false}`, algorithm, ids["patterns"][0], ids["patterns"][1])
		before := must("GET", surfacePath+"/placements", "", 200)
		previewStarted := time.Now()
		preview := must("POST", surfacePath+"/nesting/preview", body, 200)
		previewDuration := time.Since(previewStarted)
		if len(preview["placements"].([]any)) == 0 {
			t.Fatalf("%s preview returned no candidate placements", algorithm)
		}
		if after := must("GET", surfacePath+"/placements", "", 200); !reflect.DeepEqual(after, before) {
			t.Fatalf("%s preview modified saved placements", algorithm)
		}
		if err := scoped.Get(&runCount, "SELECT count(*) FROM nesting_runs"); err != nil || runCount != expectedRuns {
			t.Fatalf("%s preview wrote history: %d, %v", algorithm, runCount, err)
		}
		saveStarted := time.Now()
		output := must("POST", surfacePath+"/nesting", body, 200)
		t.Logf("%s preview=%s saved=%s", algorithm, previewDuration, time.Since(saveStarted))
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
	for i, keep := range []bool{true, false} {
		key := fmt.Sprintf("bd0051ae-5ea0-4600-8ee6-93266700000%d", i)
		operationPath := surfacePath + "/nesting/operations/" + key
		if status := must("GET", operationPath, "", 200); status["status"] != "unknown" {
			t.Fatal(status)
		}
		body := fmt.Sprintf(`{"algorithm":"baseline","patterns":[{"project_pattern_id":%d,"quantity":1}],"allowed_rotations":[0],"keep_existing":%t}`, ids["patterns"][0], keep)
		var before, after int
		if err := scoped.Get(&before, "SELECT count(*) FROM nesting_runs"); err != nil {
			t.Fatal(err)
		}
		var parallel sync.WaitGroup
		var responses [2]map[string]any
		var statuses [2]int
		for j := range statuses {
			parallel.Go(func() { statuses[j], responses[j] = request("POST", surfacePath+"/nesting", body, key) })
		}
		parallel.Wait()
		if statuses != [2]int{200, 200} || !reflect.DeepEqual(responses[0], responses[1]) {
			t.Fatalf("invalid concurrent replay: %v %v", statuses, responses)
		}
		if err := scoped.Get(&after, "SELECT count(*) FROM nesting_runs"); err != nil || after != before+1 {
			t.Fatalf("run count: %d → %d: %v", before, after, err)
		}
		status := must("GET", operationPath, "", 200)
		if status["status"] != "completed" || !reflect.DeepEqual(status["result"], responses[0]) {
			t.Fatalf("unexpected status response: %v", status)
		}
		conflict, _ := request("POST", surfacePath+"/nesting", strings.Replace(body, "\"quantity\":1", "\"quantity\":2", 1), key)
		if conflict != 409 {
			t.Fatalf("same key, changed request: %d", conflict)
		}
	}
	saved = must("GET", surfacePath+"/placements", "", 200)

	const rejectedKey = "bd0051ae-5ea0-4600-8ee6-932667000099"
	rejectedBody := `{"algorithm":"baseline","patterns":[{"project_pattern_id":99999,"quantity":1}],"allowed_rotations":[0],"keep_existing":false}`
	for range 2 {
		rejected, _ := request("POST", surfacePath+"/nesting", rejectedBody, rejectedKey)
		if rejected != 404 {
			t.Fatalf("expected durable rejection: %d", rejected)
		}
	}
	if status := must("GET", surfacePath+"/nesting/operations/"+rejectedKey, "", 200); status["status"] != "failed" {
		t.Fatalf("missing failed status: %v", status)
	}

	historyFirst := must("GET", surfacePath+"/nesting/history?page=1&page_size=2", "", 200)
	historySecond := must("GET", surfacePath+"/nesting/history?page=2&page_size=2", "", 200)
	firstPage := historyFirst["items"].([]any)
	secondPage := historySecond["items"].([]any)
	if len(firstPage) != 2 || len(secondPage) != 2 || historyFirst["has_next"] != true ||
		firstPage[1].(map[string]any)["id"].(float64) <= secondPage[0].(map[string]any)["id"].(float64) {
		t.Fatalf("invalid history pagination: %v %v", historyFirst, historySecond)
	}
	emptyHistory := must("GET", fmt.Sprintf("/projects/1/surfaces/%d/nesting/history", ids["surfaces"][1]), "", 200)
	if len(emptyHistory["items"].([]any)) != 0 {
		t.Fatal("other surface leaked history")
	}

	for _, item := range saved["items"].([]any) {
		must("DELETE", fmt.Sprintf("/projects/1/placements/%d", int(item.(map[string]any)["id"].(float64))), "", 204)
	}

	var placementRace sync.WaitGroup
	var attempts [2]int
	var created [2]map[string]any
	for i := range attempts {
		placementRace.Go(func() { attempts[i], created[i] = request("POST", surfacePath+"/placements", placementBody) })
	}
	placementRace.Wait()
	if attempts[0]+attempts[1] != 201+409 || attempts[0] == attempts[1] {
		t.Fatalf("overlap race: %v", attempts)
	}
	var winner map[string]any
	for i, status := range attempts {
		if status == 201 {
			winner = created[i]
		}
	}
	racePath := fmt.Sprintf("/projects/1/placements/%d", int(winner["id"].(float64)))
	version := winner["updated_at"].(string)
	placementRace.Go(func() { attempts[0], _ = request("PATCH", racePath, `{"x":25}`, "", version) })
	placementRace.Go(func() { attempts[1], _ = request("DELETE", racePath, "", "", version) })
	placementRace.Wait()
	switch attempts {
	case [2]int{200, 409}:
		must("DELETE", racePath, "", 204)
	case [2]int{404, 204}:
	default:
		t.Fatalf("unsafe move/delete race: %v", attempts)
	}

	// Either deletion wins and placement is rejected, or placement wins and deletion conflicts.
	var wg sync.WaitGroup
	var createStatus, deleteStatus int
	wg.Add(2)
	go func() { defer wg.Done(); createStatus, _ = request("POST", surfacePath+"/placements", placementBody) }()
	go func() { defer wg.Done(); deleteStatus, _ = request("DELETE", surfacePath, "") }()
	wg.Wait()
	createdAndProtected := createStatus == 201 && deleteStatus == 409
	deletedAndRejected := createStatus == 404 && deleteStatus == 204
	if !createdAndProtected && !deletedAndRejected {
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
	createdAndGeometryProtected := createStatus == 201 && updateStatus == 409
	updatedAndPlacementRejected := createStatus == 409 && updateStatus == 200
	if !createdAndGeometryProtected && !updatedAndPlacementRejected {
		t.Fatalf("unsafe geometry race create=%d update=%d", createStatus, updateStatus)
	}

	var dangling int
	err = scoped.Get(&dangling, `SELECT count(*) FROM placements pl LEFT JOIN project_surfaces ps ON ps.id=pl.project_surface_id LEFT JOIN project_patterns pp ON pp.id=pl.project_pattern_id WHERE ps.id IS NULL OR pp.id IS NULL OR ps.project_id <> pp.project_id`)
	if err != nil || dangling != 0 {
		t.Fatalf("dangling placements=%d err=%v", dangling, err)
	}
	for _, kind := range []string{"surfaces", "patterns"} {
		source := must("POST", "/"+kind, `{"name":"Race original","geometry":`+polygon(100)+`}`, 201)
		sourceID := int(source["id"].(float64))
		sourceKey := "source_surface_id"
		if kind == "patterns" {
			sourceKey = "source_pattern_id"
		}
		var copyResponse map[string]any
		raceWait := sync.WaitGroup{}
		raceWait.Go(func() {
			createStatus, copyResponse = request("POST", "/projects/1/"+kind, fmt.Sprintf(`{"%s":%d,"scale":1}`, sourceKey, sourceID))
		})
		raceWait.Go(func() { deleteStatus, _ = request("DELETE", fmt.Sprintf("/%s/%d", kind, sourceID), "") })
		raceWait.Wait()
		if deleteStatus != 204 || (createStatus != 201 && createStatus != 404) {
			t.Fatalf("original/copy race for %s: create=%d delete=%d", kind, createStatus, deleteStatus)
		}
		if createStatus == 201 {
			persisted := must("GET", fmt.Sprintf("/projects/1/%s/%d", kind, int(copyResponse["id"].(float64))), "", 200)
			if persisted[sourceKey] != nil || !reflect.DeepEqual(persisted["geometry"], copyResponse["geometry"]) {
				t.Fatalf("copy changed after original delete: %v", persisted)
			}
		}
	}

	exerciseSurface := must("POST", "/surfaces", `{"name":"Concurrent layout","geometry":`+polygon(100)+`}`, 201)
	exerciseCopy := must("POST", "/projects/1/surfaces", fmt.Sprintf(`{"source_surface_id":%d,"scale":1}`, int(exerciseSurface["id"].(float64))), 201)
	exercisePath := fmt.Sprintf("/projects/1/surfaces/%d", int(exerciseCopy["id"].(float64)))
	autoBody := fmt.Sprintf(`{"algorithm":"baseline","patterns":[{"project_pattern_id":%d,"quantity":1}],"allowed_rotations":[0],"keep_existing":true}`, ids["patterns"][1])
	manualBody := fmt.Sprintf(`{"project_pattern_id":%d,"x":70,"y":70,"rotation":0}`, ids["patterns"][1])
	wg.Go(func() { createStatus, _ = request("POST", exercisePath+"/placements", manualBody) })
	wg.Go(func() { updateStatus, _ = request("POST", exercisePath+"/nesting", autoBody) })
	wg.Wait()
	if createStatus != 201 || updateStatus != 200 {
		t.Fatalf("manual/auto race: %d %d", createStatus, updateStatus)
	}
	if layout := must("GET", exercisePath+"/placements", "", 200); len(layout["items"].([]any)) != 2 {
		t.Fatal("layout lost a concurrent write")
	}

	patternUnderLoad := fmt.Sprintf("/projects/1/patterns/%d", ids["patterns"][1])
	freshPlacement := fmt.Sprintf(`{"project_pattern_id":%d,"x":65,"y":20,"rotation":0}`, ids["patterns"][0])
	wg.Go(func() { createStatus, _ = request("POST", exercisePath+"/placements", freshPlacement) })
	wg.Go(func() {
		updateStatus, _ = request("PATCH", fmt.Sprintf("/projects/1/patterns/%d", ids["patterns"][0]), `{"geometry":`+polygon(200)+`}`)
	})
	wg.Wait()
	if createStatus != 201 || updateStatus != 409 {
		allowedAlternative := createStatus == 409 && updateStatus == 200
		if !allowedAlternative {
			t.Fatalf("pattern update/placement race: %d %d", createStatus, updateStatus)
		}
	}
	must("DELETE", patternUnderLoad, "", 409)

	must("DELETE", "/projects/2", "", 404)
	var nestingStatus int
	wg.Go(func() { nestingStatus, _ = request("POST", exercisePath+"/nesting", autoBody) })
	wg.Go(func() { deleteStatus, _ = request("DELETE", "/projects/1", "") })
	wg.Wait()
	if deleteStatus != 204 || (nestingStatus != 200 && nestingStatus != 404) {
		t.Fatalf("project deletion/nesting race: %d %d", deleteStatus, nestingStatus)
	}
	must("DELETE", "/projects/1", "", 404)
	for _, table := range []string{"project_surfaces", "project_patterns", "placements", "nesting_runs"} {
		var rows int
		if err := scoped.Get(&rows, "SELECT count(*) FROM "+table); err != nil || rows != 0 {
			t.Fatalf("%s not cleared by project deletion: %d, %v", table, rows, err)
		}
	}
	var neighbors int
	if err := scoped.Get(&neighbors, "SELECT count(*) FROM projects WHERE id=2"); err != nil || neighbors != 1 {
		t.Fatalf("other project affected: %d, %v", neighbors, err)
	}

}

type invalidLayoutOptimizer struct{}

func (invalidLayoutOptimizer) Optimize(_ context.Context, problem optimizer.Problem) (optimizer.Result, error) {
	return optimizer.Result{
		Placements: []optimizer.Placement{{PatternID: problem.Patterns[0].PatternID, X: -100, Y: 10}},
	}, nil
}

type waitingLayoutOptimizer struct {
	started chan struct{}
}

func (o waitingLayoutOptimizer) Optimize(ctx context.Context, _ optimizer.Problem) (optimizer.Result, error) {
	o.started <- struct{}{}
	<-ctx.Done()
	return optimizer.Result{}, ctx.Err()
}
