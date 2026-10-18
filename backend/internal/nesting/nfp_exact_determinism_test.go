package nesting

import (
	"context"
	"fmt"
	"math"
	"testing"

	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/internal/geometry/nfp"
	"server_nesting_optimizer/internal/geometry/simplefeatures"

	"github.com/stretchr/testify/require"
)

type exactNFPReference struct {
	stationary domaingeometry.Polygon
	moving     domaingeometry.Polygon
	result     domaingeometry.MultiPolygon
}

type exactNFPDeterminismBuilder struct {
	delegate geometry.NFPBuilder
	first    map[[32]byte][]exactNFPReference

	total      uint64
	repeated   uint64
	mismatches uint64
	firstError string
}

var _ geometry.NFPBuilder = (*exactNFPDeterminismBuilder)(nil)

func newExactNFPDeterminismBuilder(
	delegate geometry.NFPBuilder,
) *exactNFPDeterminismBuilder {
	return &exactNFPDeterminismBuilder{
		delegate: delegate,
		first:    make(map[[32]byte][]exactNFPReference),
	}
}

func sameNFPRingBits(
	a, b domaingeometry.Ring,
) bool {
	if len(a.Points) != len(b.Points) {
		return false
	}

	for i := range a.Points {
		if math.Float64bits(a.Points[i].X) !=
			math.Float64bits(b.Points[i].X) ||
			math.Float64bits(a.Points[i].Y) !=
				math.Float64bits(b.Points[i].Y) {
			return false
		}
	}

	return true
}

func sameNFPPolygonBits(
	a, b domaingeometry.Polygon,
) bool {
	if !sameNFPRingBits(a.Exterior, b.Exterior) ||
		len(a.Holes) != len(b.Holes) {
		return false
	}

	for i := range a.Holes {
		if !sameNFPRingBits(a.Holes[i], b.Holes[i]) {
			return false
		}
	}

	return true
}

func sameNFPMultiPolygonBits(
	a, b domaingeometry.MultiPolygon,
) bool {
	if len(a.Polygons) != len(b.Polygons) {
		return false
	}

	for i := range a.Polygons {
		if !sameNFPPolygonBits(
			a.Polygons[i],
			b.Polygons[i],
		) {
			return false
		}
	}

	return true
}

func (b *exactNFPDeterminismBuilder) BuildExternal(
	stationary domaingeometry.Polygon,
	moving domaingeometry.Polygon,
) (domaingeometry.MultiPolygon, error) {
	// Every call still executes the real builder.
	actual, err := b.delegate.BuildExternal(
		stationary,
		moving,
	)
	if err != nil {
		return actual, err
	}

	b.total++

	key := exactNFPInputKey(stationary, moving)

	for _, reference := range b.first[key] {
		// Verify the actual input, not just its hash.
		if !sameNFPPolygonBits(
			stationary,
			reference.stationary,
		) || !sameNFPPolygonBits(
			moving,
			reference.moving,
		) {
			continue
		}

		b.repeated++

		if !sameNFPMultiPolygonBits(
			reference.result,
			actual,
		) {
			b.mismatches++

			if b.firstError == "" {
				b.firstError = fmt.Sprintf(
					"call %d: identical input produced different NFP",
					b.total,
				)
			}
		}

		return actual, nil
	}

	b.first[key] = append(
		b.first[key],
		exactNFPReference{
			stationary: stationary,
			moving:     moving,
			result:     actual,
		},
	)

	return actual, nil
}

func TestNFPExactInputDeterminismHybrid(t *testing.T) {
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

	engine := simplefeatures.NewEngine()

	builder := newExactNFPDeterminismBuilder(
		nfp.NewBuilder(engine),
	)

	profile := comparisonBenchmarkProfile()

	ga, err := NewGeneticOptimizer(
		engine,
		builder,
		profile.geneticConfig,
	)
	require.NoError(t, err)

	ga.rng = newBenchmarkRandom(benchmarkSeed + 3)

	sa, err := NewSimulatedAnnealingOptimizer(
		engine,
		builder,
		profile.annealingConfig,
	)
	require.NoError(t, err)

	sa.rng = newBenchmarkRandom(benchmarkSeed + 4)

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

	t.Logf("Total successful NFP calls: %d", builder.total)
	t.Logf("Exact-input repetitions: %d", builder.repeated)
	t.Logf("Bitwise result mismatches: %d", builder.mismatches)

	if builder.firstError != "" {
		t.Logf("First mismatch: %s", builder.firstError)
	}

	require.Equal(t, uint64(8258), builder.total)
	require.Equal(t, uint64(7257), builder.repeated)
	require.Zero(t, builder.mismatches)

	fitness, err := calculatePlacementsFitness(
		engine,
		problem,
		result.Placements,
	)
	require.NoError(t, err)

	require.Equal(t, 8, result.Metrics.PlacedCount)
	require.Equal(t, 3, fitness.UnplacedCount)
	require.InDelta(t, 4800, fitness.UsedArea, 1e-8)
	require.InDelta(t, 60, fitness.UsedHeight, 1e-8)
}
