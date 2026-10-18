package nesting

import (
	"errors"
	"math"
	"testing"

	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/internal/geometry/nfp"
	"server_nesting_optimizer/internal/geometry/simplefeatures"

	"github.com/stretchr/testify/require"
)

type countingNFPBuilder struct {
	delegate geometry.NFPBuilder
	calls    int
}

func (b *countingNFPBuilder) BuildExternal(
	stationary domaingeometry.Polygon,
	moving domaingeometry.Polygon,
) (domaingeometry.MultiPolygon, error) {
	b.calls++
	return b.delegate.BuildExternal(stationary, moving)
}

func TestExperimentalNFPCacheProtectsStoredData(t *testing.T) {
	engine := simplefeatures.NewEngine()

	delegate := &countingNFPBuilder{
		delegate: nfp.NewBuilder(engine),
	}

	cache := newExperimentalNFPCache(delegate)

	stationary := benchmarkRectangle(30, 20)
	moving := benchmarkRectangle(10, 10)

	// Preserve independent copies of the original inputs.
	originalStationary := cloneExperimentalPolygon(stationary)
	originalMoving := cloneExperimentalPolygon(moving)

	first, err := cache.BuildExternal(stationary, moving)
	require.NoError(t, err)
	require.Equal(t, 1, delegate.calls)
	require.NotEmpty(t, first.Polygons)

	// Preserve the expected result independently.
	expected := cloneExperimentalMultiPolygon(first)

	// Mutate the original input slices after cache insertion.
	stationary.Exterior.Points[0].X = 999
	moving.Exterior.Points[0].Y = 999

	// Also mutate the result returned on the cache miss.
	first.Polygons[0].Exterior.Points[0].X = 999

	second, err := cache.BuildExternal(
		originalStationary,
		originalMoving,
	)
	require.NoError(t, err)

	// No additional geometric construction should occur.
	require.Equal(t, 1, delegate.calls)
	require.Equal(t, uint64(1), cache.hits)

	// Cached geometry must remain unchanged.
	require.True(
		t,
		sameNFPMultiPolygonBits(expected, second),
	)

	// Mutate the result returned on the cache hit.
	second.Polygons[0].Exterior.Points[0].Y = 999

	third, err := cache.BuildExternal(
		originalStationary,
		originalMoving,
	)
	require.NoError(t, err)

	require.Equal(t, 1, delegate.calls)
	require.Equal(t, uint64(2), cache.hits)

	// Each hit must return an independent copy.
	require.True(
		t,
		sameNFPMultiPolygonBits(expected, third),
	)
}

type failingOnceNFPBuilder struct {
	calls int
	err   error
}

func (b *failingOnceNFPBuilder) BuildExternal(
	stationary domaingeometry.Polygon,
	moving domaingeometry.Polygon,
) (domaingeometry.MultiPolygon, error) {
	b.calls++

	if b.calls == 1 {
		return domaingeometry.MultiPolygon{}, b.err
	}

	return domaingeometry.MultiPolygon{
		Polygons: []domaingeometry.Polygon{
			cloneExperimentalPolygon(stationary),
		},
	}, nil
}

func TestExperimentalNFPCacheDoesNotStoreErrors(t *testing.T) {
	sentinel := errors.New("temporary NFP failure")

	delegate := &failingOnceNFPBuilder{
		err: sentinel,
	}

	cache := newExperimentalNFPCache(delegate)

	stationary := benchmarkRectangle(30, 20)
	moving := benchmarkRectangle(10, 10)

	_, err := cache.BuildExternal(stationary, moving)
	require.ErrorIs(t, err, sentinel)

	require.Equal(t, 0, cache.count)
	require.Equal(t, 1, delegate.calls)

	second, err := cache.BuildExternal(stationary, moving)
	require.NoError(t, err)
	require.NotEmpty(t, second.Polygons)

	require.Equal(t, 1, cache.count)
	require.Equal(t, 2, delegate.calls)

	third, err := cache.BuildExternal(stationary, moving)
	require.NoError(t, err)

	require.True(
		t,
		sameNFPMultiPolygonBits(second, third),
	)

	require.Equal(t, 2, delegate.calls)
	require.Equal(t, uint64(1), cache.hits)
}

func TestExperimentalNFPCacheInstancesAreIndependent(t *testing.T) {
	engine := simplefeatures.NewEngine()

	firstDelegate := &countingNFPBuilder{
		delegate: nfp.NewBuilder(engine),
	}
	secondDelegate := &countingNFPBuilder{
		delegate: nfp.NewBuilder(engine),
	}

	firstCache := newExperimentalNFPCache(firstDelegate)
	secondCache := newExperimentalNFPCache(secondDelegate)

	stationary := benchmarkRectangle(30, 20)
	moving := benchmarkRectangle(10, 10)

	_, err := firstCache.BuildExternal(stationary, moving)
	require.NoError(t, err)

	_, err = firstCache.BuildExternal(stationary, moving)
	require.NoError(t, err)

	_, err = secondCache.BuildExternal(stationary, moving)
	require.NoError(t, err)

	require.Equal(t, 1, firstDelegate.calls)
	require.Equal(t, 1, secondDelegate.calls)

	require.Equal(t, uint64(1), firstCache.hits)
	require.Equal(t, uint64(0), secondCache.hits)

	require.Equal(t, 1, firstCache.count)
	require.Equal(t, 1, secondCache.count)
}

type deterministicCountingNFPBuilder struct {
	calls int
}

func (b *deterministicCountingNFPBuilder) BuildExternal(
	stationary domaingeometry.Polygon,
	moving domaingeometry.Polygon,
) (domaingeometry.MultiPolygon, error) {
	b.calls++

	return domaingeometry.MultiPolygon{
		Polygons: []domaingeometry.Polygon{
			cloneExperimentalPolygon(stationary),
		},
	}, nil
}

func TestExperimentalNFPCacheRespectsEntryLimit(t *testing.T) {
	delegate := &deterministicCountingNFPBuilder{}
	cache := newExperimentalNFPCache(delegate)

	moving := benchmarkRectangle(10, 10)

	// Insert exactly the maximum permitted number of unique pairs.
	for i := 0; i < experimentalNFPCacheLimit; i++ {
		stationary := benchmarkRectangle(20, 20)

		for j := range stationary.Exterior.Points {
			stationary.Exterior.Points[j].X += float64(i)
		}

		_, err := cache.BuildExternal(stationary, moving)
		require.NoError(t, err)
	}

	require.Equal(
		t,
		experimentalNFPCacheLimit,
		cache.count,
	)
	require.Equal(
		t,
		experimentalNFPCacheLimit,
		delegate.calls,
	)

	// Repeat a pair that was saved before reaching the limit.
	firstStationary := benchmarkRectangle(20, 20)

	_, err := cache.BuildExternal(
		firstStationary,
		moving,
	)
	require.NoError(t, err)

	// The cached result should be reused.
	require.Equal(
		t,
		experimentalNFPCacheLimit,
		delegate.calls,
	)
	require.Equal(t, uint64(1), cache.hits)

	// Request a new pair after the cache is full.
	newStationary := benchmarkRectangle(20, 20)

	for j := range newStationary.Exterior.Points {
		newStationary.Exterior.Points[j].X += 100000
	}

	firstUncached, err := cache.BuildExternal(
		newStationary,
		moving,
	)
	require.NoError(t, err)

	require.Equal(
		t,
		experimentalNFPCacheLimit+1,
		delegate.calls,
	)

	// The full cache must not store this new pair.
	require.Equal(
		t,
		experimentalNFPCacheLimit,
		cache.count,
	)

	// Repeating that new pair must invoke the delegate again.
	secondUncached, err := cache.BuildExternal(
		newStationary,
		moving,
	)
	require.NoError(t, err)

	require.Equal(
		t,
		experimentalNFPCacheLimit+2,
		delegate.calls,
	)

	// Results must still be correct.
	require.True(
		t,
		sameNFPMultiPolygonBits(
			firstUncached,
			secondUncached,
		),
	)

	// Previously cached pairs must remain available.
	_, err = cache.BuildExternal(
		firstStationary,
		moving,
	)
	require.NoError(t, err)

	require.Equal(
		t,
		experimentalNFPCacheLimit+2,
		delegate.calls,
	)

	require.Equal(t, uint64(2), cache.hits)
	require.Equal(
		t,
		experimentalNFPCacheLimit,
		cache.count,
	)
}

func TestExperimentalNFPCacheLargeEntry(t *testing.T) {
	const vertexCount = 100_000

	stationary := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: make(
				[]domaingeometry.Point,
				vertexCount,
			),
		},
	}

	for i := range stationary.Exterior.Points {
		angle := 2 * math.Pi * float64(i) /
			float64(vertexCount)

		stationary.Exterior.Points[i] = domaingeometry.Point{
			X: 100 + 50*math.Cos(angle),
			Y: 100 + 50*math.Sin(angle),
		}
	}

	moving := benchmarkRectangle(10, 10)

	delegate := &deterministicCountingNFPBuilder{}
	cache := newExperimentalNFPCache(delegate)

	_, err := cache.BuildExternal(stationary, moving)
	require.NoError(t, err)

	stats := collectNFPCacheMemoryStats(cache)

	require.Equal(t, 1, cache.count)
	require.Equal(t, 1, delegate.calls)

	require.Equal(t, vertexCount, stats.NFPVertices)
	require.Equal(
		t,
		vertexCount+len(moving.Exterior.Points),
		stats.InputVertices,
	)

	coordinateBytes := int64(
		stats.InputVertices+stats.NFPVertices,
	) * 16

	t.Logf("Stored entries: %d", stats.Entries)
	t.Logf("Input vertices: %d", stats.InputVertices)
	t.Logf("NFP vertices: %d", stats.NFPVertices)

	t.Logf(
		"Coordinate storage lower bound: %.2f MiB",
		float64(coordinateBytes)/(1024*1024),
	)

	require.Greater(
		t,
		coordinateBytes,
		int64(3*1024*1024),
	)
}

func TestExperimentalNFPCacheRespectsByteBudget(
	t *testing.T,
) {
	delegate := &deterministicCountingNFPBuilder{}
	cache := newExperimentalNFPCache(delegate)

	// Small budget specifically for this test.
	cache.maxBytes = 1024

	largeStationary := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: make(
				[]domaingeometry.Point,
				100,
			),
		},
	}

	for i := range largeStationary.Exterior.Points {
		angle := 2 * math.Pi * float64(i) / 100

		largeStationary.Exterior.Points[i] =
			domaingeometry.Point{
				X: 50 * math.Cos(angle),
				Y: 50 * math.Sin(angle),
			}
	}

	moving := benchmarkRectangle(10, 10)

	// The result is too large for the cache budget,
	// but BuildExternal must still succeed.
	first, err := cache.BuildExternal(
		largeStationary,
		moving,
	)
	require.NoError(t, err)

	require.Equal(t, 1, delegate.calls)
	require.Zero(t, cache.count)
	require.Zero(t, cache.usedBytes)

	// The oversized pair was not cached.
	second, err := cache.BuildExternal(
		largeStationary,
		moving,
	)
	require.NoError(t, err)

	require.Equal(t, 2, delegate.calls)
	require.Zero(t, cache.count)
	require.Zero(t, cache.hits)

	require.True(
		t,
		sameNFPMultiPolygonBits(first, second),
	)

	// A smaller entry must still be cacheable.
	smallStationary := benchmarkRectangle(20, 20)

	_, err = cache.BuildExternal(
		smallStationary,
		moving,
	)
	require.NoError(t, err)

	require.Equal(t, 1, cache.count)
	require.Positive(t, cache.usedBytes)
	require.LessOrEqual(
		t,
		cache.usedBytes,
		cache.maxBytes,
	)

	callsAfterInsert := delegate.calls

	// Exact repetition must use the saved result.
	_, err = cache.BuildExternal(
		smallStationary,
		moving,
	)
	require.NoError(t, err)

	require.Equal(
		t,
		callsAfterInsert,
		delegate.calls,
	)
	require.Equal(t, uint64(1), cache.hits)
	require.Equal(t, 1, cache.count)
}

func TestExperimentalNFPCacheExactByteBudgetBoundary(
	t *testing.T,
) {
	delegate := &deterministicCountingNFPBuilder{}
	cache := newExperimentalNFPCache(delegate)

	moving := benchmarkRectangle(10, 10)

	first := benchmarkRectangle(20, 20)

	second := cloneExperimentalPolygon(first)
	for i := range second.Exterior.Points {
		second.Exterior.Points[i].X += 30
	}

	third := cloneExperimentalPolygon(first)
	for i := range third.Exterior.Points {
		third.Exterior.Points[i].X += 60
	}

	// Our deterministic builder returns stationary
	// as the NFP result.
	expectedResult := domaingeometry.MultiPolygon{
		Polygons: []domaingeometry.Polygon{
			cloneExperimentalPolygon(first),
		},
	}

	entryBytes := estimateExperimentalEntryBytes(
		first,
		moving,
		expectedResult,
	)

	require.Positive(t, entryBytes)

	// Exactly enough space for two entries.
	cache.maxBytes = entryBytes * 2

	_, err := cache.BuildExternal(first, moving)
	require.NoError(t, err)

	require.Equal(t, int64(entryBytes), cache.usedBytes)
	require.Equal(t, 1, cache.count)

	_, err = cache.BuildExternal(second, moving)
	require.NoError(t, err)

	require.Equal(t, cache.maxBytes, cache.usedBytes)
	require.Equal(t, 2, cache.count)
	require.Equal(t, 2, delegate.calls)

	// Third entry must not be stored.
	_, err = cache.BuildExternal(third, moving)
	require.NoError(t, err)

	require.Equal(t, 2, cache.count)
	require.Equal(t, cache.maxBytes, cache.usedBytes)
	require.Equal(t, 3, delegate.calls)

	// The first entry is still cached.
	_, err = cache.BuildExternal(first, moving)
	require.NoError(t, err)

	require.Equal(t, 3, delegate.calls)
	require.Equal(t, uint64(1), cache.hits)

	// The third entry was not saved.
	_, err = cache.BuildExternal(third, moving)
	require.NoError(t, err)

	require.Equal(t, 4, delegate.calls)
	require.Equal(t, 2, cache.count)
	require.Equal(t, cache.maxBytes, cache.usedBytes)
}
