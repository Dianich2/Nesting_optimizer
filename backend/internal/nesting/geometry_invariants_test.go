package nesting

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"runtime"
	dg "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
	"testing"
	"time"
)

type invariantCase struct {
	Name      string
	Algorithm Algorithm
	Problem   Problem
	Result    Result
	Formed    []dg.Polygon
}

func invariantShape(points ...float64) dg.Polygon {
	p := dg.Polygon{}
	for i := 0; i < len(points); i += 2 {
		p.Exterior.Points = append(p.Exterior.Points, dg.Point{X: points[i], Y: points[i+1]})
	}
	return p
}
func invariantProblem(name string) Problem {
	p := Problem{Surface: rectangle(100, 100), Patterns: []PatternItem{{PatternID: 1, Geometry: rectangle(6, 4), Quantity: 2}}, AllowedRotations: []float64{0, 90}}
	switch name {
	case "rotated_surface":
		p.Surface = invariantShape(50, 0, 100, 50, 50, 100, 0, 50)
		p.AllowedRotations = []float64{33, 45}
	case "L":
		p.Patterns[0].Geometry = invariantShape(0, 0, 12, 0, 12, 3, 3, 3, 3, 12, 0, 12)
	case "C":
		p.Patterns[0].Geometry = invariantShape(0, 0, 12, 0, 12, 3, 3, 3, 3, 9, 12, 9, 12, 12, 0, 12)
	case "U":
		p.Patterns[0].Geometry = invariantShape(0, 0, 12, 0, 12, 12, 9, 12, 9, 3, 3, 3, 3, 12, 0, 12)
	case "radial":
		p.Patterns[0].Geometry = experimentPolygon(12)
		p.Patterns[0].Quantity = 1
	case "triangle":
		p.Patterns[0].Geometry = invariantShape(20, 40, 28, 40, 23, 46)
		p.AllowedRotations = []float64{33}
	case "holes1", "holes2", "holes3":
		n := map[string]int{"holes1": 1, "holes2": 2, "holes3": 3}[name]
		for i := 0; i < n; i++ {
			v := float64(20 + i*20)
			p.Surface.Holes = append(p.Surface.Holes, invariantShape(v, v, v+10, v, v+10, v+10, v, v+10).Exterior)
		}
	case "neck":
		p.Surface = invariantShape(0, 0, 30, 0, 30, 14, 50, 14, 50, 0, 80, 0, 80, 40, 50, 40, 50, 26, 30, 26, 30, 40, 0, 40)
	case "contact":
		p.Surface = rectangle(12, 4)
		p.AllowedRotations = []float64{0}
	case "near_contact":
		p.Obstacles = []dg.Polygon{invariantShape(0, 0, 6, 0, 6, 4, 0, 4)}
		p.Patterns[0].Geometry = rectangle(5.999999, 4)
	case "occupied":
		p.Obstacles = []dg.Polygon{rectangle(12, 12), invariantShape(40, 40, 50, 40, 50, 50, 40, 50)}
	case "multi_type":
		p.Patterns = append(p.Patterns, PatternItem{PatternID: 2, Geometry: experimentPolygon(12), Quantity: 1})
	case "full":
		p.Obstacles = []dg.Polygon{p.Surface}
	case "oversize":
		p.Patterns[0].Geometry = rectangle(200, 200)
	case "angles4":
		p.AllowedRotations = []float64{0, 90, 180, 270}
	case "angles36":
		p.AllowedRotations = make([]float64, 36)
		for i := range p.AllowedRotations {
			p.AllowedRotations[i] = float64(i) * 10
		}
	}
	return p
}
func invariantMap(p Problem, scale, dx, dy float64, reverse bool) Problem {
	transform := func(p dg.Polygon) dg.Polygon {
		ring := func(r dg.Ring) dg.Ring {
			v := dg.Ring{Points: make([]dg.Point, len(r.Points))}
			for i, x := range r.Points {
				v.Points[i] = dg.Point{X: x.X*scale + dx, Y: x.Y*scale + dy}
			}
			if reverse {
				for i, j := 0, len(v.Points)-1; i < j; i, j = i+1, j-1 {
					v.Points[i], v.Points[j] = v.Points[j], v.Points[i]
				}
			}
			return v
		}
		out := dg.Polygon{Exterior: ring(p.Exterior)}
		for _, h := range p.Holes {
			out.Holes = append(out.Holes, ring(h))
		}
		return out
	}
	p.Surface = transform(p.Surface)
	for i := range p.Obstacles {
		p.Obstacles[i] = transform(p.Obstacles[i])
	}
	for i := range p.Patterns {
		p.Patterns[i].Geometry = transform(p.Patterns[i].Geometry)
	}
	return p
}
func invariantArea(p dg.Polygon) float64 {
	area := func(r dg.Ring) float64 {
		if len(r.Points) == 0 {
			return 0
		}
		origin := r.Points[0]
		v := 0.0
		for i, a := range r.Points {
			b := r.Points[(i+1)%len(r.Points)]
			v += (a.X-origin.X)*(b.Y-origin.Y) - (b.X-origin.X)*(a.Y-origin.Y)
		}
		return math.Abs(v) / 2
	}
	result := area(p.Exterior)
	for _, h := range p.Holes {
		result -= area(h)
	}
	return result
}
func checkInvariantResult(t *testing.T, p Problem, result Result) []dg.Polygon {
	t.Helper()
	validateExperiment(t, p, result)
	engine := simplefeatures.NewEngine()
	requested := 0
	lookup := map[int64]PatternItem{}
	left := map[int64]int{}
	for _, item := range p.Patterns {
		requested += item.Quantity
		lookup[item.PatternID] = item
		left[item.PatternID] = item.Quantity
	}
	formed := []dg.Polygon{}
	placedArea := 0.0
	for _, placement := range result.Placements {
		item, ok := lookup[placement.PatternID]
		if !ok || left[item.PatternID] <= 0 {
			t.Fatal("unknown/excess pattern")
		}
		left[item.PatternID]--
		allowed := false
		for _, a := range p.AllowedRotations {
			allowed = allowed || a == placement.Rotation
		}
		if !allowed {
			t.Fatal("unauthorized angle")
		}
		polygon, e := engine.Transform(item.Geometry, placement.X, placement.Y, placement.Rotation)
		if e != nil {
			t.Fatal(e)
		}
		formed = append(formed, polygon)
		placedArea += invariantArea(polygon)
	}
	seen := map[int64]bool{}
	for _, u := range result.Unplaced {
		if seen[u.PatternID] || u.Quantity <= 0 || left[u.PatternID] != u.Quantity {
			t.Fatal("unplaced invalid")
		}
		seen[u.PatternID] = true
		left[u.PatternID] = 0
	}
	for _, n := range left {
		if n != 0 {
			t.Fatal("missing unplaced")
		}
	}
	close := func(a, b float64) bool {
		return !math.IsNaN(a) && !math.IsInf(a, 0) && math.Abs(a-b) <= 1e-7*math.Max(math.Abs(b), 1e-3)
	}
	if result.Metrics.RequestedCount != requested || result.Metrics.PlacedCount != len(formed) || !close(result.Metrics.SurfaceArea, invariantArea(p.Surface)) || !close(result.Metrics.PlacedArea, placedArea) || !close(result.Metrics.Utilization, placedArea/invariantArea(p.Surface)) {
		t.Fatalf("independent area/counts mismatch %+v expectedarea=%g", result.Metrics, placedArea)
	}
	return formed
}
func TestGeometryInvariantCore(t *testing.T) {
	for _, name := range []string{"rectangle", "C", "U", "holes3", "neck", "contact", "full", "oversize"} {
		for _, a := range []Algorithm{BaselineAlgorithm, NFPGreedyAlgorithm} {
			t.Run(name+"/"+string(a), func(t *testing.T) {
				p := invariantProblem(name)
				o, _, _ := experimentOptimizer(t, a)
				ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
				defer cancel()
				r, e := o.Optimize(ctx, p)
				if e != nil {
					t.Fatal(e)
				}
				checkInvariantResult(t, p, r)
				if (name == "full" || name == "oversize") && len(r.Placements) != 0 {
					t.Fatal("impossible case placed")
				}
			})
		}
	}
}
func TestGeometryInvariantExperiment(t *testing.T) {
	if os.Getenv("NESTING_INVARIANT_EXPERIMENT") != "1" {
		t.Skip("explicit capped invariant/property corpus")
	}
	names := []string{"rectangle", "rotated_surface", "L", "C", "U", "radial", "triangle", "holes1", "holes2", "holes3", "neck", "contact", "near_contact", "occupied", "multi_type", "full", "oversize", "angles4", "angles36"}
	rows := []invariantCase{}
	start := time.Now()
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	run := func(name string, p Problem, a Algorithm) {
		t.Helper()
		t.Logf("INVARIANT_BEGIN %s/%s", name, a)
		o, _, _ := experimentOptimizer(t, a)
		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
		defer cancel()
		r, e := o.Optimize(ctx, p)
		if e != nil {
			t.Fatalf("%s/%s: %v", name, a, e)
		}
		formed := checkInvariantResult(t, p, r)
		rows = append(rows, invariantCase{Name: name, Algorithm: a, Problem: p, Result: r, Formed: formed})
	}
	for _, name := range names {
		for _, variant := range []string{"base", "translated", "reversed", "small", "large"} {
			for _, a := range []Algorithm{BaselineAlgorithm, NFPGreedyAlgorithm, GeneticAlgorithm, SimulatedAnnealingAlgorithm, HybridAlgorithm} {
				p := invariantProblem(name)
				switch variant {
				case "translated":
					p = invariantMap(p, 1, 100000, -100000, false)
				case "reversed":
					p = invariantMap(p, 1, 0, 0, true)
				case "small":
					p = invariantMap(p, .1, 0, 0, false)
				case "large":
					p = invariantMap(p, 10, 0, 0, false)
				}
				run(name+"/"+variant, p, a)
			}
		}
	}
	// Seeded, simple radial rings; no unconstrained live fuzzing in normal go test.
	random := rand.New(rand.NewSource(20261121))
	for i := 0; i < 24; i++ {
		p := invariantProblem("rectangle")
		star := dg.Polygon{}
		n := 5 + random.Intn(8)
		for j := 0; j < n; j++ {
			angle := 2 * math.Pi * float64(j) / float64(n)
			r := 2 + random.Float64()*3
			star.Exterior.Points = append(star.Exterior.Points, dg.Point{X: 6 + r*math.Cos(angle), Y: 6 + r*math.Sin(angle)})
		}
		p.Patterns[0].Geometry = star
		for _, a := range []Algorithm{BaselineAlgorithm, NFPGreedyAlgorithm} {
			run(fmt.Sprintf("property/seed20261121/%d", i), p, a)
		}
	}

	// Preserve the ORIGINAL 523 rows/order first. Additional stage15 examples
	// exercise strict contacts with all five solvers; no old assertions removed.
	if len(rows) != 523 {
		t.Fatalf("original corpus incomplete: %d", len(rows))
	}
	algorithms := []Algorithm{BaselineAlgorithm, NFPGreedyAlgorithm, GeneticAlgorithm, SimulatedAnnealingAlgorithm, HybridAlgorithm}
	data, err := os.ReadFile("testdata/stage12_property12_counterexample.json")
	if err != nil {
		t.Fatal(err)
	}
	var original struct {
		Problem Problem
		Formed  []dg.Polygon
	}
	if err = json.Unmarshal(data, &original); err != nil || len(original.Formed) != 2 {
		t.Fatalf("original fixture: %v", err)
	}
	for _, name := range []string{"original-obstacle", "diagonal-contact", "ulp-width", "hole-boundary"} {
		p := invariantProblem("rectangle")
		switch name {
		case "original-obstacle":
			p = original.Problem
			p.Patterns = append([]PatternItem(nil), p.Patterns...)
			p.Patterns[0].Quantity = 1
			p.Obstacles = []dg.Polygon{original.Formed[0]}
		case "diagonal-contact":
			p.Patterns[0].Geometry = invariantShape(0, 0, 6, 0, 0, 6)
			p.Obstacles = []dg.Polygon{invariantShape(0, 0, 12, 0, 0, 12)}
		case "ulp-width":
			p.Surface = rectangle(12, 4)
			p.AllowedRotations = []float64{0}
			p.Obstacles = []dg.Polygon{rectangle(6, 4)}
			p.Patterns[0].Geometry = rectangle(math.Nextafter(6, 0), 4)
		case "hole-boundary":
			p.Surface = rectangle(12, 12)
			p.Surface.Holes = []dg.Ring{invariantShape(2, 2, 10, 2, 10, 10, 2, 10).Exterior}
			p.Patterns[0].Geometry = rectangle(2, 2)
		}
		for _, a := range algorithms {
			run("stage15/"+name, p, a)
		}
	}
	extraRandom := rand.New(rand.NewSource(20261126))
	for i := 0; i < 4; i++ {
		p := invariantProblem("rectangle")
		p.AllowedRotations = []float64{33, 45, 90}
		star := dg.Polygon{}
		n := 6 + i
		for j := 0; j < n; j++ {
			angle := 2 * math.Pi * float64(j) / float64(n)
			r := 2 + extraRandom.Float64()*3
			star.Exterior.Points = append(star.Exterior.Points, dg.Point{X: 6 + r*math.Cos(angle), Y: 6 + r*math.Sin(angle)})
		}
		p.Patterns[0].Geometry = star
		for _, a := range algorithms {
			run(fmt.Sprintf("stage15/property/seed20261126/%d", i), p, a)
		}
	}
	if len(rows) != 563 {
		t.Fatalf("extended corpus incomplete: %d", len(rows))
	}
	runtime.ReadMemStats(&after)
	path := os.Getenv("NESTING_INVARIANT_OUTPUT")
	if path == "" {
		t.Fatal("explicit output path required")
	}
	data, e := json.Marshal(rows)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(path, data, 0600); e != nil {
		t.Fatal(e)
	}
	t.Logf("INVARIANT_CORPUS rows=%d elapsed=%s total_alloc=%d goroutines=%d", len(rows), time.Since(start), after.TotalAlloc-before.TotalAlloc, runtime.NumGoroutine())
}
