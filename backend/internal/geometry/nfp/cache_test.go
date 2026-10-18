package nfp

import (
	"errors"
	"testing"

	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"

	"github.com/stretchr/testify/require"
)

type cacheTestBuilder struct {
	calls int
	err   error
}

var _ geometry.NFPBuilder = (*cacheTestBuilder)(nil)

func (b *cacheTestBuilder) BuildExternal(
	stationary domaingeometry.Polygon,
	_ domaingeometry.Polygon,
) (domaingeometry.MultiPolygon, error) {
	b.calls++

	if b.err != nil {
		return domaingeometry.MultiPolygon{}, b.err
	}

	return domaingeometry.MultiPolygon{
		Polygons: []domaingeometry.Polygon{
			clonePolygon(stationary),
		},
	}, nil
}

func cacheTestRectangle(x float64) domaingeometry.Polygon {
	return domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: x, Y: 0},
				{X: x + 20, Y: 0},
				{X: x + 20, Y: 10},
				{X: x, Y: 10},
			},
		},
	}
}

func TestCachedBuilderReusesIdenticalInputs(t *testing.T) {
	delegate := &cacheTestBuilder{}
	cache := NewCachedBuilder(delegate)

	stationary := cacheTestRectangle(0)
	moving := cacheTestRectangle(30)

	first, err := cache.BuildExternal(stationary, moving)
	require.NoError(t, err)

	second, err := cache.BuildExternal(stationary, moving)
	require.NoError(t, err)

	require.Equal(t, 1, delegate.calls)
	require.Equal(t, uint64(1), cache.hits)
	require.Equal(t, uint64(1), cache.misses)

	require.Equal(t, first, second)
}

func TestCachedBuilderProtectsStoredData(t *testing.T) {
	delegate := &cacheTestBuilder{}
	cache := NewCachedBuilder(delegate)

	stationary := cacheTestRectangle(0)
	moving := cacheTestRectangle(30)

	originalStationary := clonePolygon(stationary)
	originalMoving := clonePolygon(moving)

	first, err := cache.BuildExternal(stationary, moving)
	require.NoError(t, err)

	expected := cloneMultiPolygon(first)

	stationary.Exterior.Points[0].X = 999
	moving.Exterior.Points[0].Y = 999
	first.Polygons[0].Exterior.Points[0].X = 999

	second, err := cache.BuildExternal(
		originalStationary,
		originalMoving,
	)
	require.NoError(t, err)

	require.Equal(t, 1, delegate.calls)
	require.Equal(t, expected, second)

	second.Polygons[0].Exterior.Points[0].Y = 999

	third, err := cache.BuildExternal(
		originalStationary,
		originalMoving,
	)
	require.NoError(t, err)

	require.Equal(t, expected, third)
	require.Equal(t, 1, delegate.calls)
}

func TestCachedBuilderDoesNotCacheErrors(t *testing.T) {
	sentinel := errors.New("NFP failure")
	delegate := &cacheTestBuilder{err: sentinel}
	cache := NewCachedBuilder(delegate)

	stationary := cacheTestRectangle(0)
	moving := cacheTestRectangle(30)

	_, err := cache.BuildExternal(stationary, moving)
	require.ErrorIs(t, err, sentinel)

	_, err = cache.BuildExternal(stationary, moving)
	require.ErrorIs(t, err, sentinel)

	require.Equal(t, 2, delegate.calls)
	require.Zero(t, cache.count)
	require.Zero(t, cache.hits)
}

func TestCachedBuilderInstancesAreIndependent(t *testing.T) {
	delegateA := &cacheTestBuilder{}
	delegateB := &cacheTestBuilder{}

	cacheA := NewCachedBuilder(delegateA)
	cacheB := NewCachedBuilder(delegateB)

	stationary := cacheTestRectangle(0)
	moving := cacheTestRectangle(30)

	_, err := cacheA.BuildExternal(stationary, moving)
	require.NoError(t, err)

	_, err = cacheA.BuildExternal(stationary, moving)
	require.NoError(t, err)

	_, err = cacheB.BuildExternal(stationary, moving)
	require.NoError(t, err)

	require.Equal(t, 1, delegateA.calls)
	require.Equal(t, 1, delegateB.calls)
	require.Equal(t, uint64(1), cacheA.hits)
	require.Zero(t, cacheB.hits)
}

func TestCachedBuilderRespectsByteBudget(t *testing.T) {
	delegate := &cacheTestBuilder{}
	cache := NewCachedBuilder(delegate)

	stationary := cacheTestRectangle(0)
	moving := cacheTestRectangle(30)

	// Fill the cache's accounted budget with a single
	// oversized result produced by a test builder.
	large := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: make(
				[]domaingeometry.Point,
				600_000,
			),
		},
	}

	_, err := cache.BuildExternal(large, moving)
	require.NoError(t, err)

	require.Zero(t, cache.count)
	require.Zero(t, cache.usedBytes)

	_, err = cache.BuildExternal(stationary, moving)
	require.NoError(t, err)

	require.Equal(t, 1, cache.count)
	require.Positive(t, cache.usedBytes)
	require.LessOrEqual(
		t,
		cache.usedBytes,
		uint64(defaultCacheMaxBytes),
	)
}
