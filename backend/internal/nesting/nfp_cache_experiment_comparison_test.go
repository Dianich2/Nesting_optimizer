package nesting

import (
	"context"
	"testing"

	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/internal/geometry/nfp"
	"server_nesting_optimizer/internal/geometry/simplefeatures"

	"github.com/stretchr/testify/require"
)

func runExperimentalHybrid(
	t *testing.T,
	engine geometry.Engine,
	builder geometry.NFPBuilder,
	problem Problem,
) Result {
	t.Helper()

	profile := comparisonBenchmarkProfile()

	ga, err := NewGeneticOptimizer(
		engine,
		builder,
		profile.geneticConfig,
	)
	require.NoError(t, err)

	ga.rng = newBenchmarkRandom(
		benchmarkSeed + 3,
	)

	sa, err := NewSimulatedAnnealingOptimizer(
		engine,
		builder,
		profile.annealingConfig,
	)
	require.NoError(t, err)

	sa.rng = newBenchmarkRandom(
		benchmarkSeed + 4,
	)

	hybrid, err := NewHybridOptimizer(
		engine,
		ga,
		sa,
		profile.compactionConfig,
	)
	require.NoError(t, err)

	result, err := hybrid.Optimize(
		context.Background(),
		problem,
	)
	require.NoError(t, err)

	return result
}

func TestExperimentalNFPCachePreservesHybridResult(
	t *testing.T,
) {
	var problem Problem
	found := false

	for _, benchmarkCase := range benchmarkProblems() {
		if benchmarkCase.name == "constrained_space" {
			problem = benchmarkCase.problem
			found = true
			break
		}
	}

	require.True(t, found)

	engineWithoutCache := simplefeatures.NewEngine()

	originalResult := runExperimentalHybrid(
		t,
		engineWithoutCache,
		nfp.NewBuilder(engineWithoutCache),
		problem,
	)

	engineWithCache := simplefeatures.NewEngine()

	cache := newExperimentalNFPCache(
		nfp.NewBuilder(engineWithCache),
	)

	cachedResult := runExperimentalHybrid(
		t,
		engineWithCache,
		cache,
		problem,
	)

	// Compare actual placement positions and rotations,
	// not just aggregate fitness metrics.
	require.Equal(
		t,
		originalResult.Placements,
		cachedResult.Placements,
	)

	require.Equal(
		t,
		originalResult.Unplaced,
		cachedResult.Unplaced,
	)

	originalFitness, err := calculatePlacementsFitness(
		engineWithoutCache,
		problem,
		originalResult.Placements,
	)
	require.NoError(t, err)

	cachedFitness, err := calculatePlacementsFitness(
		engineWithCache,
		problem,
		cachedResult.Placements,
	)
	require.NoError(t, err)

	require.Equal(
		t,
		originalFitness,
		cachedFitness,
	)

	require.Equal(
		t,
		originalResult.Metrics.PlacedCount,
		cachedResult.Metrics.PlacedCount,
	)

	require.Equal(
		t,
		originalResult.Metrics.PlacedArea,
		cachedResult.Metrics.PlacedArea,
	)

	require.Equal(
		t,
		originalResult.Metrics.Utilization,
		cachedResult.Metrics.Utilization,
	)

	t.Logf("Cache hits: %d", cache.hits)
	t.Logf("Cache misses: %d", cache.misses)
	t.Logf("Stored entries: %d", cache.count)

	require.Greater(t, cache.hits, uint64(0))
	require.LessOrEqual(
		t,
		cache.count,
		experimentalNFPCacheLimit,
	)
}

func TestExperimentalNFPCachePreservesAllBenchmarkResults(
	t *testing.T,
) {
	for _, benchmarkCase := range benchmarkProblems() {
		t.Run(benchmarkCase.name, func(t *testing.T) {
			problem := benchmarkCase.problem

			engineWithoutCache := simplefeatures.NewEngine()

			originalResult := runExperimentalHybrid(
				t,
				engineWithoutCache,
				nfp.NewBuilder(engineWithoutCache),
				problem,
			)

			engineWithCache := simplefeatures.NewEngine()

			cache := newExperimentalNFPCache(
				nfp.NewBuilder(engineWithCache),
			)

			cachedResult := runExperimentalHybrid(
				t,
				engineWithCache,
				cache,
				problem,
			)

			// Compare the complete results, including:
			// placements, order, rotations, unplaced
			// patterns and all metrics.
			require.Equal(
				t,
				originalResult,
				cachedResult,
			)

			originalFitness, err := calculatePlacementsFitness(
				engineWithoutCache,
				problem,
				originalResult.Placements,
			)
			require.NoError(t, err)

			cachedFitness, err := calculatePlacementsFitness(
				engineWithCache,
				problem,
				cachedResult.Placements,
			)
			require.NoError(t, err)

			require.Equal(
				t,
				originalFitness,
				cachedFitness,
			)

			require.LessOrEqual(
				t,
				cache.count,
				experimentalNFPCacheLimit,
			)

			t.Logf(
				"cache hits=%d, misses=%d, entries=%d",
				cache.hits,
				cache.misses,
				cache.count,
			)

			t.Logf(
				"placed=%d, unplaced=%d, usedArea=%.4f",
				cachedResult.Metrics.PlacedCount,
				cachedFitness.UnplacedCount,
				cachedFitness.UsedArea,
			)
		})
	}
}

func TestExperimentalNFPCachePreservesComplexGeometry(
	t *testing.T,
) {
	polygonWithHole := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 65, Y: 45},
				{X: 125, Y: 45},
				{X: 125, Y: 105},
				{X: 65, Y: 105},
			},
		},
		Holes: []domaingeometry.Ring{
			{
				Points: []domaingeometry.Point{
					{X: 80, Y: 60},
					{X: 80, Y: 90},
					{X: 110, Y: 90},
					{X: 110, Y: 60},
				},
			},
		},
	}

	tests := []struct {
		name    string
		problem Problem
	}{
		{
			name: "45 degree rotations with hole obstacle",
			problem: Problem{
				Surface: benchmarkRectangle(180, 150),
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Quantity:  2,
						Geometry:  benchmarkTriangle(25, 20),
					},
					{
						PatternID: 2,
						Quantity:  2,
						Geometry:  benchmarkLShape(),
					},
					{
						PatternID: 3,
						Quantity:  2,
						Geometry:  benchmarkRectangle(25, 15),
					},
				},
				Obstacles: []domaingeometry.Polygon{
					polygonWithHole,
				},
				AllowedRotations: []float64{
					0, 45, 90, 135,
				},
			},
		},
		{
			name: "negative and fractional coordinates",
			problem: Problem{
				Surface: benchmarkRectangleAt(
					-80.5, -40.25, 180, 140,
				),
				Patterns: []PatternItem{
					{
						PatternID: 1,
						Quantity:  2,
						Geometry:  benchmarkTriangle(18.5, 12.25),
					},
					{
						PatternID: 2,
						Quantity:  2,
						Geometry:  benchmarkRectangle(20.5, 10.25),
					},
				},
				Obstacles: []domaingeometry.Polygon{
					benchmarkRectangleAt(
						-20.25, 5.5, 25.5, 20.25,
					),
				},
				AllowedRotations: []float64{
					0, 45, 90,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalEngine := simplefeatures.NewEngine()

			original := runExperimentalHybrid(
				t,
				originalEngine,
				nfp.NewBuilder(originalEngine),
				tt.problem,
			)

			cachedEngine := simplefeatures.NewEngine()

			cache := newExperimentalNFPCache(
				nfp.NewBuilder(cachedEngine),
			)

			cached := runExperimentalHybrid(
				t,
				cachedEngine,
				cache,
				tt.problem,
			)

			require.Equal(t, original, cached)

			originalFitness, err := calculatePlacementsFitness(
				originalEngine,
				tt.problem,
				original.Placements,
			)
			require.NoError(t, err)

			cachedFitness, err := calculatePlacementsFitness(
				cachedEngine,
				tt.problem,
				cached.Placements,
			)
			require.NoError(t, err)

			require.Equal(
				t,
				originalFitness,
				cachedFitness,
			)

			require.LessOrEqual(
				t,
				cache.count,
				experimentalNFPCacheLimit,
			)

			t.Logf(
				"hits=%d, misses=%d, entries=%d",
				cache.hits,
				cache.misses,
				cache.count,
			)

			t.Logf(
				"placed=%d, unplaced=%d",
				cached.Metrics.PlacedCount,
				cachedFitness.UnplacedCount,
			)
		})
	}
}

func TestProductionNFPCachePreservesBenchmarkResults(
	t *testing.T,
) {
	for _, benchmarkCase := range benchmarkProblems() {
		t.Run(benchmarkCase.name, func(t *testing.T) {
			problem := benchmarkCase.problem

			originalEngine := simplefeatures.NewEngine()
			original := runExperimentalHybrid(
				t,
				originalEngine,
				nfp.NewBuilder(originalEngine),
				problem,
			)

			cachedEngine := simplefeatures.NewEngine()

			cachedBuilder := nfp.NewCachedBuilder(
				nfp.NewBuilder(cachedEngine),
			)

			cached := runExperimentalHybrid(
				t,
				cachedEngine,
				cachedBuilder,
				problem,
			)

			require.Equal(t, original, cached)

			originalFitness, err := calculatePlacementsFitness(
				originalEngine,
				problem,
				original.Placements,
			)
			require.NoError(t, err)

			cachedFitness, err := calculatePlacementsFitness(
				cachedEngine,
				problem,
				cached.Placements,
			)
			require.NoError(t, err)

			require.Equal(
				t,
				originalFitness,
				cachedFitness,
			)
		})
	}
}
