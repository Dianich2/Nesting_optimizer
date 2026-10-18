package nesting

import (
	"testing"

	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry/nfp"
	"server_nesting_optimizer/internal/geometry/simplefeatures"

	"github.com/stretchr/testify/require"
)

type nfpCacheMemoryStats struct {
	Entries       int
	InputVertices int
	NFPVertices   int
	NFPPolygons   int
	NFPHoles      int
}

func countNFPRingVertices(ring domaingeometry.Ring) int {
	return len(ring.Points)
}

func countNFPPolygonVertices(
	polygon domaingeometry.Polygon,
) int {
	count := countNFPRingVertices(polygon.Exterior)

	for _, hole := range polygon.Holes {
		count += countNFPRingVertices(hole)
	}

	return count
}

func collectNFPCacheMemoryStats(
	cache *experimentalNFPCache,
) nfpCacheMemoryStats {
	var stats nfpCacheMemoryStats

	for _, bucket := range cache.entries {
		for _, entry := range bucket {
			stats.Entries++

			stats.InputVertices += countNFPPolygonVertices(
				entry.stationary,
			)
			stats.InputVertices += countNFPPolygonVertices(
				entry.moving,
			)

			for _, polygon := range entry.result.Polygons {
				stats.NFPPolygons++
				stats.NFPVertices += countNFPPolygonVertices(
					polygon,
				)
				stats.NFPHoles += len(polygon.Holes)
			}
		}
	}

	return stats
}

func TestExperimentalNFPCacheMemoryProfile(t *testing.T) {
	for _, benchmarkCase := range benchmarkProblems() {
		t.Run(benchmarkCase.name, func(t *testing.T) {
			engine := simplefeatures.NewEngine()

			cache := newExperimentalNFPCache(
				nfp.NewBuilder(engine),
			)

			result := runExperimentalHybrid(
				t,
				engine,
				cache,
				benchmarkCase.problem,
			)

			stats := collectNFPCacheMemoryStats(cache)

			require.Equal(t, cache.count, stats.Entries)
			require.LessOrEqual(
				t,
				stats.Entries,
				experimentalNFPCacheLimit,
			)

			// Each point contains two float64 coordinates.
			coordinateBytes := int64(
				stats.InputVertices+stats.NFPVertices,
			) * 16

			t.Logf(
				"entries=%d, hits=%d, misses=%d",
				stats.Entries,
				cache.hits,
				cache.misses,
			)

			t.Logf(
				"input vertices=%d, NFP vertices=%d",
				stats.InputVertices,
				stats.NFPVertices,
			)

			t.Logf(
				"NFP polygons=%d, holes=%d",
				stats.NFPPolygons,
				stats.NFPHoles,
			)

			t.Logf(
				"coordinate storage lower bound: %.2f KiB",
				float64(coordinateBytes)/1024,
			)

			t.Logf(
				"placed=%d",
				result.Metrics.PlacedCount,
			)
		})
	}
}
