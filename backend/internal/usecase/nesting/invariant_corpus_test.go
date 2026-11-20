package nesting

import (
	"encoding/json"
	"os"
	sf "server_nesting_optimizer/internal/geometry/simplefeatures"
	opt "server_nesting_optimizer/internal/nesting"
	"testing"
)

func TestInvariantCorpusServerValidation(t *testing.T) {
	path := os.Getenv("NESTING_INVARIANT_CORPUS")
	if path == "" {
		t.Skip("explicit bounded generated corpus")
	}
	data, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Name      string
		Algorithm opt.Algorithm
		Problem   opt.Problem
		Result    opt.Result
	}
	if e = json.Unmarshal(data, &rows); e != nil || len(rows) < 500 {
		t.Fatal("incomplete corpus")
	}
	engine := sf.NewEngine()
	negative := 0
	for _, row := range rows {
		if _, e = validateResult(t.Context(), engine, 1, row.Problem, row.Result); e != nil {
			t.Fatalf("server validation %s/%s: %v", row.Name, row.Algorithm, e)
		}
		// Copy only the mutable result slice; never corrupt the positive corpus.
		if len(row.Result.Placements) > 0 {
			bad := row.Result
			bad.Placements = append([]opt.Placement(nil), bad.Placements...)
			bad.Placements[0].X = 1e8
			if _, e = validateResult(t.Context(), engine, 1, row.Problem, bad); e == nil {
				t.Fatal("outside placement accepted")
			}
			bad = row.Result
			bad.Metrics.PlacedCount++
			if _, e = validateResult(t.Context(), engine, 1, row.Problem, bad); e == nil {
				t.Fatal("invalid metrics accepted")
			}
			negative += 2
		}
	}
	t.Logf("SERVER_CORPUS cases=%d negative=%d", len(rows), negative)
}
