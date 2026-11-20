package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"math"
	"net/http/httptest"
	"reflect"
	dg "server_nesting_optimizer/internal/domain/geometry"
	sf "server_nesting_optimizer/internal/geometry/simplefeatures"
	"server_nesting_optimizer/internal/transport/http/dto"
	httperror "server_nesting_optimizer/internal/transport/http/errors"
	"server_nesting_optimizer/internal/transport/http/handler"
	"server_nesting_optimizer/internal/transport/http/middleware"
	"server_nesting_optimizer/pkg/apperror"
	"testing"
	"time"
)

func TestRadialPreviewSavedExactly(t *testing.T) {
	f := newNestingFixture(t)
	pattern := dg.Polygon{}
	for i := 0; i < 12; i++ {
		r := 5.0
		if i%2 == 1 {
			r = 3
		}
		a := 2 * math.Pi * float64(i) / 12
		pattern.Exterior.Points = append(pattern.Exterior.Points, dg.Point{X: 6 + r*math.Cos(a), Y: 6 + r*math.Sin(a)})
	}
	data, err := sf.NewCodec().EncodeWKB(pattern)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.db.Exec("UPDATE project_patterns SET geometry=ST_GeomFromWKB($1) WHERE id=1", data); err != nil {
		t.Fatal(err)
	}
	native := &countedNativeBaseline{}
	uc := newFixtureUseCase(t, f.scoped, native)
	app := fiber.New(fiber.Config{ErrorHandler: httperror.Handle})
	app.Use(func(c fiber.Ctx) error { c.Locals(middleware.UserIDLocalKey, int64(1)); return c.Next() })
	h := handler.NewNestingHandler(uc)
	path := "/projects/1/surfaces/1/nesting"
	app.Post("/projects/:project_id/surfaces/:project_surface_id/nesting/preview", h.Preview)
	app.Post("/projects/:project_id/surfaces/:project_surface_id/nesting", h.Run)
	post := func(url, body, key string) dto.RunNestingResponse {
		t.Helper()
		req := httptest.NewRequest("POST", url, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", key)
		res, e := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
		if e != nil {
			t.Fatal(e)
		}
		defer func() { _ = res.Body.Close() }()
		if res.StatusCode != 200 {
			t.Fatalf("radial POST status %d", res.StatusCode)
		}
		var output dto.RunNestingResponse
		if e = json.NewDecoder(res.Body).Decode(&output); e != nil {
			t.Fatal(e)
		}
		return output
	}
	body := `"algorithm":"baseline","patterns":[{"project_pattern_id":1,"quantity":1}],"allowed_rotations":[0],"keep_existing":false`
	previewKey := uuid.NewString()
	httpPreview := post(path+"/preview", "{"+body+"}", previewKey)
	preview, err := uc.PreviewWithKey(t.Context(), f.input, previewKey)
	if httpPreview.PreviewID != preview.PreviewID || len(httpPreview.Placements) != 1 {
		t.Fatal("HTTP preview differs from durable replay")
	}
	if err != nil {
		t.Fatal(err)
	}
	if len(preview.Placements) != 1 {
		t.Fatal("real SQL radial preview is empty")
	}
	input := f.input
	input.PreviewID = preview.PreviewID
	key := uuid.NewString()
	httpSaved := post(path, fmt.Sprintf(`{%s,"preview_id":%q}`, body, preview.PreviewID), key)
	saved, err := uc.ExecuteWithKey(t.Context(), input, key)
	if len(httpSaved.Placements) != 1 || httpSaved.Placements[0].ID != saved.Placements[0].ID {
		t.Fatal("HTTP save differs from durable replay")
	}
	if err != nil {
		t.Fatal(err)
	}
	if native.calls.Load() != 1 || len(saved.Placements) != 1 {
		t.Fatal("preview was recalculated")
	}
	a, b := preview.Placements[0], saved.Placements[0]
	if a.X != b.X || a.Y != b.Y || a.Rotation != b.Rotation || !reflect.DeepEqual(a.Geometry, b.Geometry) {
		t.Fatal("exact preview geometry changed")
	}
	replay, err := uc.ExecuteWithKey(t.Context(), input, key)
	if err != nil || !reflect.DeepEqual(saved, replay) || native.calls.Load() != 1 {
		t.Fatalf("save replay: %v", err)
	}
	f.counts(t, 1, 1)
	// Validate the exact engine-transformed output, not a SQL reimplementation
	// whose centroid and arithmetic order may round boundary contacts differently.
	encoded, err := sf.NewCodec().EncodeWKB(b.Geometry)
	if err != nil {
		t.Fatal(err)
	}
	var valid bool
	requireProcessGet(t, f.db, &valid, `SELECT ST_IsValid(ST_GeomFromWKB($1)) AND ST_CoveredBy(ST_GeomFromWKB($1),geometry) FROM project_surfaces WHERE id=1`, encoded)
	if !valid {
		t.Fatal("PostGIS rejects exact returned geometry")
	}
	stalePreview, err := uc.PreviewWithKey(t.Context(), f.input, uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.db.Exec("UPDATE placements SET x=x+1 WHERE project_surface_id=1"); err != nil {
		t.Fatal(err)
	}
	input.PreviewID = stalePreview.PreviewID
	_, err = uc.ExecuteWithKey(t.Context(), input, uuid.NewString())
	expectCode(t, err, apperror.AppCodeStale)
	// Cancellation before work cannot publish a new partial layout.
	cancelled, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = uc.PreviewWithKey(cancelled, f.input, uuid.NewString())
	if err == nil {
		t.Fatal("cancelled preview accepted")
	}
	f.counts(t, 1, 1)
}
