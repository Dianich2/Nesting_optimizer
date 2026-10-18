package nesting

import (
	"runtime"
	"testing"

	"server_nesting_optimizer/internal/geometry/nfp"
	"server_nesting_optimizer/internal/geometry/simplefeatures"

	"github.com/stretchr/testify/require"
)

func TestExperimentalNFPCacheRetainedHeap(t *testing.T) {
	var problem Problem
	found := false

	for _, benchmarkCase := range benchmarkProblems() {
		if benchmarkCase.name == "mixed_polygons" {
			problem = benchmarkCase.problem
			found = true
			break
		}
	}

	require.True(t, found, "benchmark problem not found")

	engine := simplefeatures.NewEngine()

	// Collect garbage before creating the cache.
	runtime.GC()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	cache := newExperimentalNFPCache(
		nfp.NewBuilder(engine),
	)

	result := runExperimentalHybrid(
		t,
		engine,
		cache,
		problem,
	)

	require.Equal(t, 9, result.Metrics.PlacedCount)
	require.Equal(t, 1539, cache.count)

	// Collect temporary allocations from optimization.
	// Cached objects must remain reachable.
	runtime.GC()

	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	heapDelta := int64(after.HeapAlloc) -
		int64(before.HeapAlloc)

	objectsDelta := int64(after.HeapObjects) -
		int64(before.HeapObjects)

	t.Logf("Cache entries: %d", cache.count)
	t.Logf("Cache hits: %d", cache.hits)
	t.Logf("Cache misses: %d", cache.misses)

	t.Logf(
		"HeapAlloc before: %.2f MiB",
		float64(before.HeapAlloc)/(1024*1024),
	)

	t.Logf(
		"HeapAlloc after: %.2f MiB",
		float64(after.HeapAlloc)/(1024*1024),
	)

	t.Logf(
		"Approximate retained heap delta: %.2f MiB",
		float64(heapDelta)/(1024*1024),
	)

	t.Logf(
		"HeapObjects delta: %d",
		objectsDelta,
	)

	// Ensure the cache remains live through the measurement.
	runtime.KeepAlive(cache)
	runtime.KeepAlive(engine)
	runtime.KeepAlive(problem)
}
