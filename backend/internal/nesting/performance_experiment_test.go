package nesting

import (
	"os"
	dg "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
	"testing"
	"time"
)

// Explicit, bounded comparison; production random/config is never modified.
func TestTargetedPerformanceExperiment(t *testing.T) {
	if os.Getenv("NESTING_TARGETED_PERFORMANCE") != "1" {
		t.Skip("explicit capped performance experiment")
	}
	engine := simplefeatures.NewEngine()
	for _, algorithm := range []Algorithm{BaselineAlgorithm, NFPGreedyAlgorithm} {
		for _, name := range []string{"rectangle6", "concave10", "holes10", "occupied50", "occupied100"} {
			surface, pattern := rectangle(100, 100), rectangle(10, 10)
			count := 6
			rotations := []float64{0, 90}
			var obstacles []dg.Polygon
			if name == "concave10" {
				count = 10
				pattern = dg.Polygon{Exterior: dg.Ring{Points: []dg.Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 4}, {X: 4, Y: 4}, {X: 4, Y: 10}, {X: 0, Y: 10}}}}
				rotations = []float64{0, 90, 180}
			}
			if name == "holes10" {
				count = 10
				surface.Holes = []dg.Ring{{Points: []dg.Point{{X: 40, Y: 40}, {X: 60, Y: 40}, {X: 60, Y: 60}, {X: 40, Y: 60}}}, {Points: []dg.Point{{X: 70, Y: 70}, {X: 80, Y: 70}, {X: 80, Y: 80}, {X: 70, Y: 80}}}}
			}
			if name == "occupied50" || name == "occupied100" {
				surface = rectangle(1000, 1000)
				seed, err := engine.Transform(rectangle(10, 10), 200, 200, 0)
				if err != nil {
					t.Fatal(err)
				}
				obstacles = []dg.Polygon{seed}
				count = 50
				pattern = experimentPolygon(12)
				if name == "occupied100" {
					count = 100
					pattern = rectangle(10, 10)
				}
			}
			problem := Problem{Surface: surface, Patterns: []PatternItem{{PatternID: 1, Geometry: pattern, Quantity: count}}, Obstacles: obstacles, AllowedRotations: rotations}
			runGeometryExperiment(t, name, algorithm, problem, 40*time.Second, false)
		}
	}
}
