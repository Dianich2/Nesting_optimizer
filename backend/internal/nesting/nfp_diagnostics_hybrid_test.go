package nesting

import (
	"context"
	"testing"

	"server_nesting_optimizer/internal/geometry/nfp"
	"server_nesting_optimizer/internal/geometry/simplefeatures"

	"github.com/stretchr/testify/require"
)

func TestNFPDiagnosticHybridComparison(t *testing.T) {
	var problem Problem
	foundProblem := false

	for _, benchmarkCase := range benchmarkProblems() {
		if benchmarkCase.name == "constrained_space" {
			problem = benchmarkCase.problem
			foundProblem = true
			break
		}
	}

	require.True(t, foundProblem, "benchmark problem not found")

	engine := simplefeatures.NewEngine()

	realBuilder := nfp.NewBuilder(engine)

	exactBuilder := newExactNFPDiagnosticBuilder(
		realBuilder,
	)

	checkingBuilder := newCheckingNFPBuilder(
		exactBuilder,
	)

	diagnosticBuilder := newDiagnosticNFPBuilder(
		checkingBuilder,
	)

	profile := comparisonBenchmarkProfile()

	geneticOptimizer, err := NewGeneticOptimizer(
		engine,
		diagnosticBuilder,
		profile.geneticConfig,
	)
	require.NoError(t, err)

	geneticOptimizer.rng = newBenchmarkRandom(
		benchmarkSeed + 3,
	)

	annealingOptimizer, err := NewSimulatedAnnealingOptimizer(
		engine,
		diagnosticBuilder,
		profile.annealingConfig,
	)
	require.NoError(t, err)

	annealingOptimizer.rng = newBenchmarkRandom(
		benchmarkSeed + 4,
	)

	hybridOptimizer, err := NewHybridOptimizer(
		engine,
		geneticOptimizer,
		annealingOptimizer,
		profile.compactionConfig,
	)
	require.NoError(t, err)

	result, err := hybridOptimizer.Optimize(
		context.Background(),
		problem,
	)
	require.NoError(t, err)

	stats := diagnosticBuilder.Stats()
	exactHitRate := 100 *
		float64(exactBuilder.repeated) /
		float64(exactBuilder.total)

	t.Logf(
		"Exact input NFP calls: %d",
		exactBuilder.total,
	)

	t.Logf(
		"Exact input unique pairs: %d",
		len(exactBuilder.seen),
	)

	t.Logf(
		"Exact input repeated pairs: %d",
		exactBuilder.repeated,
	)

	t.Logf(
		"Potential exact-input hit rate: %.2f%%",
		exactHitRate,
	)

	checkingBuilder.Report(t)

	require.Greater(t, stats.TotalCalls, uint64(0))

	require.Equal(
		t,
		stats.TotalCalls,
		stats.UniquePairs+stats.RepeatedCalls,
	)

	potentialHitRate := 100 *
		float64(stats.RepeatedCalls) /
		float64(stats.TotalCalls)

	t.Logf("Total NFP calls: %d", stats.TotalCalls)
	t.Logf("Unique relative pairs: %d", stats.UniquePairs)
	t.Logf("Repeated relative pairs: %d", stats.RepeatedCalls)
	t.Logf("Potential exact-key hit rate: %.2f%%", potentialHitRate)

	fitness, err := calculatePlacementsFitness(
		engine,
		problem,
		result.Placements,
	)
	require.NoError(t, err)

	t.Logf("Placed: %d", result.Metrics.PlacedCount)
	t.Logf("Unplaced: %d", fitness.UnplacedCount)
	t.Logf("Used area: %.2f", fitness.UsedArea)
	t.Logf("Used height: %.2f", fitness.UsedHeight)
}
