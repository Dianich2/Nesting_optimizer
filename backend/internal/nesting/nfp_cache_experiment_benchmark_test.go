package nesting

import (
	"context"
	"testing"

	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/internal/geometry/nfp"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
)

func newNFPCacheBenchmarkHybrid(
	b *testing.B,
	engine geometry.Engine,
	builder geometry.NFPBuilder,
) Optimizer {
	b.Helper()

	profile := comparisonBenchmarkProfile()

	ga, err := NewGeneticOptimizer(
		engine,
		builder,
		profile.geneticConfig,
	)
	if err != nil {
		b.Fatal(err)
	}

	ga.rng = newBenchmarkRandom(benchmarkSeed + 3)

	sa, err := NewSimulatedAnnealingOptimizer(
		engine,
		builder,
		profile.annealingConfig,
	)
	if err != nil {
		b.Fatal(err)
	}

	sa.rng = newBenchmarkRandom(benchmarkSeed + 4)

	hybrid, err := NewHybridOptimizer(
		engine,
		ga,
		sa,
		profile.compactionConfig,
	)
	if err != nil {
		b.Fatal(err)
	}

	return hybrid
}

func BenchmarkExperimentalNFPCacheHybrid(b *testing.B) {
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
		b.Fatal("constrained_space benchmark problem not found")
	}

	for _, useCache := range []bool{false, true} {
		name := "without_cache"
		if useCache {
			name = "with_cache"
		}

		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			b.StopTimer()

			var totalHits uint64
			var totalMisses uint64

			for i := 0; i < b.N; i++ {
				engine := simplefeatures.NewEngine()

				var builder geometry.NFPBuilder = nfp.NewBuilder(engine)

				var cache *experimentalNFPCache

				if useCache {
					cache = newExperimentalNFPCache(builder)
					builder = cache
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

				fitness, err := calculatePlacementsFitness(
					engine,
					problem,
					result.Placements,
				)
				if err != nil {
					b.Fatal(err)
				}

				if result.Metrics.PlacedCount != 8 ||
					fitness.UnplacedCount != 3 {
					b.Fatalf(
						"unexpected result: placed=%d, unplaced=%d",
						result.Metrics.PlacedCount,
						fitness.UnplacedCount,
					)
				}

				if cache != nil {
					totalHits += cache.hits
					totalMisses += cache.misses
				}
			}

			if useCache {
				b.ReportMetric(
					float64(totalHits)/float64(b.N),
					"cache_hits/op",
				)

				b.ReportMetric(
					float64(totalMisses)/float64(b.N),
					"cache_misses/op",
				)
			}
		})
	}
}
