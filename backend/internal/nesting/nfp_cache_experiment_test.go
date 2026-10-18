package nesting

import (
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
)

const experimentalNFPCacheLimit = 2048
const experimentalNFPCacheMaxBytes int64 = 8 * 1024 * 1024

type experimentalNFPCacheEntry struct {
	stationary domaingeometry.Polygon
	moving     domaingeometry.Polygon
	result     domaingeometry.MultiPolygon
}

type experimentalNFPCache struct {
	delegate  geometry.NFPBuilder
	entries   map[[32]byte][]experimentalNFPCacheEntry
	count     int
	hits      uint64
	misses    uint64
	maxBytes  int64
	usedBytes int64
}

var _ geometry.NFPBuilder = (*experimentalNFPCache)(nil)

func newExperimentalNFPCache(
	delegate geometry.NFPBuilder,
) *experimentalNFPCache {
	return &experimentalNFPCache{
		delegate: delegate,
		entries:  make(map[[32]byte][]experimentalNFPCacheEntry),
		maxBytes: experimentalNFPCacheMaxBytes,
	}
}

func cloneExperimentalRing(
	ring domaingeometry.Ring,
) domaingeometry.Ring {
	result := ring

	if ring.Points != nil {
		result.Points = make(
			[]domaingeometry.Point,
			len(ring.Points),
		)
		copy(result.Points, ring.Points)
	}

	return result
}

func cloneExperimentalPolygon(
	polygon domaingeometry.Polygon,
) domaingeometry.Polygon {
	result := polygon
	result.Exterior = cloneExperimentalRing(
		polygon.Exterior,
	)

	if polygon.Holes != nil {
		result.Holes = make(
			[]domaingeometry.Ring,
			len(polygon.Holes),
		)

		for i, hole := range polygon.Holes {
			result.Holes[i] = cloneExperimentalRing(hole)
		}
	}

	return result
}

func cloneExperimentalMultiPolygon(
	multi domaingeometry.MultiPolygon,
) domaingeometry.MultiPolygon {
	result := multi

	if multi.Polygons != nil {
		result.Polygons = make(
			[]domaingeometry.Polygon,
			len(multi.Polygons),
		)

		for i, polygon := range multi.Polygons {
			result.Polygons[i] = cloneExperimentalPolygon(
				polygon,
			)
		}
	}

	return result
}

func (c *experimentalNFPCache) BuildExternal(
	stationary domaingeometry.Polygon,
	moving domaingeometry.Polygon,
) (domaingeometry.MultiPolygon, error) {
	key := exactNFPInputKey(stationary, moving)

	for _, entry := range c.entries[key] {
		if !sameNFPPolygonBits(
			stationary,
			entry.stationary,
		) {
			continue
		}

		if !sameNFPPolygonBits(
			moving,
			entry.moving,
		) {
			continue
		}

		c.hits++

		return cloneExperimentalMultiPolygon(
			entry.result,
		), nil
	}

	c.misses++

	result, err := c.delegate.BuildExternal(
		stationary,
		moving,
	)
	if err != nil {
		return result, err
	}

	entryBytes := estimateExperimentalEntryBytes(
		stationary,
		moving,
		result,
	)

	if c.count < experimentalNFPCacheLimit &&
		entryBytes >= 0 &&
		c.usedBytes >= 0 &&
		c.usedBytes <= c.maxBytes &&
		entryBytes <= c.maxBytes-c.usedBytes {

		c.entries[key] = append(
			c.entries[key],
			experimentalNFPCacheEntry{
				stationary: cloneExperimentalPolygon(
					stationary,
				),
				moving: cloneExperimentalPolygon(
					moving,
				),
				result: cloneExperimentalMultiPolygon(
					result,
				),
			},
		)

		c.count++
		c.usedBytes += entryBytes
	}

	return result, nil
}

func estimateExperimentalRingBytes(
	ring domaingeometry.Ring,
) int64 {
	// Two float64 values per point.
	return int64(len(ring.Points)) * 16
}

func estimateExperimentalPolygonBytes(
	polygon domaingeometry.Polygon,
) int64 {
	// Polygon contains two slice-based fields:
	// exterior ring and holes.
	const polygonHeaderBytes int64 = 48
	const ringHeaderBytes int64 = 24

	size := polygonHeaderBytes

	size += estimateExperimentalRingBytes(
		polygon.Exterior,
	)

	for _, hole := range polygon.Holes {
		size += ringHeaderBytes
		size += estimateExperimentalRingBytes(hole)
	}

	return size
}

func estimateExperimentalMultiPolygonBytes(
	multi domaingeometry.MultiPolygon,
) int64 {
	// Slice header for the polygon collection.
	const multiHeaderBytes int64 = 24

	size := multiHeaderBytes

	for _, polygon := range multi.Polygons {
		size += estimateExperimentalPolygonBytes(polygon)
	}

	return size
}

func estimateExperimentalEntryBytes(
	stationary domaingeometry.Polygon,
	moving domaingeometry.Polygon,
	result domaingeometry.MultiPolygon,
) int64 {
	// Approximate SHA-256 key storage.
	const keyBytes int64 = 32

	return keyBytes +
		estimateExperimentalPolygonBytes(stationary) +
		estimateExperimentalPolygonBytes(moving) +
		estimateExperimentalMultiPolygonBytes(result)
}
