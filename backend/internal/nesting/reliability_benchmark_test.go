package nesting

import (
	"context"
	"fmt"
	dg "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry/nfp"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
	"testing"
	"time"
)

func BenchmarkReliabilityWorkloads(b *testing.B) {
	shapes := map[string]dg.Polygon{
		"rectangle": rectangle(10, 10),
		"concave":   {Exterior: dg.Ring{Points: []dg.Point{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 4}, {X: 4, Y: 4}, {X: 4, Y: 10}, {X: 0, Y: 10}}}},
	}
	for _, name := range []string{"rectangle", "concave"} {
		for _, count := range []int{10, 50, 100} {
			for _, algorithm := range []Algorithm{BaselineAlgorithm, NFPGreedyAlgorithm} {
				b.Run(fmt.Sprintf("%s/%d/%s", name, count, algorithm), func(b *testing.B) {
					engine := simplefeatures.NewEngine()
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						b.StopTimer()
						var opt Optimizer = NewBaselineOptimizer(engine)
						if algorithm == NFPGreedyAlgorithm {
							opt = NewNFPGreedyOptimizer(engine, nfp.NewCachedBuilder(nfp.NewBuilder(engine)))
						}
						problem := Problem{Surface: rectangle(1000, 1000), Patterns: []PatternItem{{PatternID: 1, Quantity: count, Geometry: shapes[name]}}, AllowedRotations: []float64{0, 90}}
						statistics := &CandidateStatistics{}
						ctx, cancel := context.WithTimeout(WithCandidateStatistics(context.Background(), statistics), 2*time.Minute)
						b.StartTimer()
						result, err := opt.Optimize(ctx, problem)
						b.StopTimer()
						cancel()
						if err != nil {
							b.Fatal(err)
						}
						placed := make([]dg.Polygon, 0, len(result.Placements))
						for _, p := range result.Placements {
							geometry, err := engine.Transform(shapes[name], p.X, p.Y, p.Rotation)
							if err != nil {
								b.Fatal(err)
							}
							inside, err := engine.CoveredBy(geometry, problem.Surface)
							if err != nil || !inside {
								b.Fatal("placement outside surface")
							}
							for _, other := range placed {
								overlap, err := engine.InteriorsIntersect(geometry, other)
								if err != nil || overlap {
									b.Fatal("overlapping benchmark result")
								}
							}
							placed = append(placed, geometry)
						}
						b.ReportMetric(float64(statistics.Generated.Load()), "candidates/op")
						b.ReportMetric(float64(result.Metrics.PlacedCount), "placed/op")
						b.ReportMetric(result.Metrics.Utilization, "utilization/op")
						b.StartTimer()
					}
				})
			}
		}
	}
}
