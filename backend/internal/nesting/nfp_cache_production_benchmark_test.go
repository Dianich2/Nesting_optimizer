package nesting

import (
	"context"
	"testing"

	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/internal/geometry/nfp"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
)

func BenchmarkProductionNFPCacheHybrid(b *testing.B) {
	var problem Problem
	found := false

	for _, benchmarkCase := range benchmarkProblems() {
		if benchmarkCase.name == "constrained_space" {
			problem = benchmarkCase.problem
			found = true
			break
		}
	}

	if !found {
		b.Fatal("constrained_space not found")
	}

	for _, useCache := range []bool{false, true} {
		name := "without_cache"
		if useCache {
			name = "with_cache"
		}

		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.StopTimer()

			for i := 0; i < b.N; i++ {
				engine := simplefeatures.NewEngine()

				var builder geometry.NFPBuilder = nfp.NewBuilder(engine)
				if useCache {
					builder = nfp.NewCachedBuilder(builder)
				}

				optimizer := newNFPCacheBenchmarkHybrid(
					b,
					engine,
					builder,
				)

				b.StartTimer()

				result, err := optimizer.Optimize(
					context.Background(),
					problem,
				)

				b.StopTimer()

				if err != nil {
					b.Fatal(err)
				}

				if result.Metrics.PlacedCount != 8 {
					b.Fatalf(
						"unexpected placed count: %d",
						result.Metrics.PlacedCount,
					)
				}
			}
		})
	}
}
