package integration

import (
	"context"
	"encoding/json"
	"os"
	dg "server_nesting_optimizer/internal/domain/geometry"
	sf "server_nesting_optimizer/internal/geometry/simplefeatures"
	opt "server_nesting_optimizer/internal/nesting"
	"testing"
	"time"
)

func TestGeometryCorpusPostGISOracle(t *testing.T) {
	path := os.Getenv("NESTING_INVARIANT_CORPUS")
	if path == "" {
		t.Skip("explicit generated Go invariant corpus")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Name      string
		Algorithm opt.Algorithm
		Problem   opt.Problem
		Result    opt.Result
		Formed    []dg.Polygon
	}
	if err = json.Unmarshal(data, &rows); err != nil || len(rows) < 500 {
		t.Fatalf("incomplete corpus count=%d err=%v", len(rows), err)
	}
	db := newTestSchema(t)
	applyAllMigrations(t, db)
	db.SetMaxOpenConns(4)
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	codec := sf.NewCodec()
	checked, pairs := 0, 0
	for _, row := range rows {
		surface, e := codec.EncodeWKB(row.Problem.Surface)
		if e != nil {
			t.Fatal(e)
		}
		var valid bool
		if e = db.GetContext(ctx, &valid, "SELECT ST_IsValid(ST_GeomFromWKB($1))", surface); e != nil || !valid {
			t.Fatalf("surface %s/%s %v", row.Name, row.Algorithm, e)
		}
		if len(row.Formed) != len(row.Result.Placements) {
			t.Fatal("missing actual Go geometry")
		}
		occupied := [][]byte{}
		for _, p := range row.Problem.Obstacles {
			w, e := codec.EncodeWKB(p)
			if e != nil {
				t.Fatal(e)
			}
			occupied = append(occupied, w)
		}
		for i, p := range row.Formed {
			w, e := codec.EncodeWKB(p)
			if e != nil {
				t.Fatal(e)
			}
			if e = db.GetContext(ctx, &valid, `SELECT ST_IsValid(ST_GeomFromWKB($1)) AND ST_CoveredBy(ST_GeomFromWKB($1),ST_GeomFromWKB($2))`, w, surface); e != nil || !valid {
				failed, _ := json.Marshal(row)
				if out := os.Getenv("NESTING_ORACLE_COUNTEREXAMPLE"); out != "" {
					_ = os.WriteFile(out, failed, 0600)
				}
				t.Fatalf("PostGIS coverage rejects actual Go geometry: case=%s algorithm=%s placement=%d err=%v", row.Name, row.Algorithm, i, e)
			}
			for _, other := range occupied {
				var overlap bool
				if e = db.GetContext(ctx, &overlap, `SELECT ST_Relate(ST_GeomFromWKB($1),ST_GeomFromWKB($2),'2********')`, w, other); e != nil || overlap {
					failed, _ := json.Marshal(row)
					if out := os.Getenv("NESTING_ORACLE_COUNTEREXAMPLE"); out != "" {
						_ = os.WriteFile(out, failed, 0600)
					}
					t.Fatalf("PostGIS overlap: %s/%s placement%d err=%v", row.Name, row.Algorithm, i, e)
				}
				pairs++
			}
			occupied = append(occupied, w)
			checked++
		}
	}
	t.Logf("POSTGIS_ORACLE cases=%d actual_polygons=%d interior_pairs=%d", len(rows), checked, pairs)
	// The independent oracle itself must reject a valid polygon outside the sheet
	// and detect identical interiors, while allowing boundary-only contact.
	var outside, overlap, touch bool
	if err = db.GetContext(ctx, &outside, `SELECT ST_CoveredBy(ST_GeomFromText('POLYGON((20 20,21 20,21 21,20 21,20 20))'),ST_GeomFromText('POLYGON((0 0,10 0,10 10,0 10,0 0))'))`); err != nil {
		t.Fatal(err)
	}
	if err = db.GetContext(ctx, &overlap, `SELECT ST_Relate(ST_MakeEnvelope(0,0,1,1),ST_MakeEnvelope(0,0,1,1),'2********')`); err != nil {
		t.Fatal(err)
	}
	if err = db.GetContext(ctx, &touch, `SELECT ST_Relate(ST_MakeEnvelope(0,0,1,1),ST_MakeEnvelope(1,0,2,1),'2********')`); err != nil {
		t.Fatal(err)
	}
	if outside || !overlap || touch {
		t.Fatal("oracle negative/control predicates broken")
	}
}
