package nfp

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
	"math"
	"unsafe"

	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
)

const (
	defaultCacheMaxEntries = 2048
	defaultCacheMaxBytes   = 8 * 1024 * 1024
)

type cacheEntry struct {
	stationary domaingeometry.Polygon
	moving     domaingeometry.Polygon
	result     domaingeometry.MultiPolygon
}

// CachedBuilder caches exact BuildExternal inputs.
//
// Each instance belongs to one optimization run.
// Concurrent use of the same instance is not supported.
type CachedBuilder struct {
	delegate geometry.NFPBuilder
	entries  map[[32]byte][]cacheEntry

	count     int
	usedBytes uint64
	hits      uint64
	misses    uint64
}

var _ geometry.NFPBuilder = (*CachedBuilder)(nil)
var _ geometry.ContextNFPBuilder = (*CachedBuilder)(nil)

func NewCachedBuilder(delegate geometry.NFPBuilder) *CachedBuilder {
	return &CachedBuilder{
		delegate: delegate,
		entries:  make(map[[32]byte][]cacheEntry),
	}
}

func (c *CachedBuilder) BuildExternal(
	stationary domaingeometry.Polygon,
	moving domaingeometry.Polygon,
) (domaingeometry.MultiPolygon, error) {
	return c.BuildExternalContext(context.Background(), stationary, moving)
}

func (c *CachedBuilder) BuildExternalContext(ctx context.Context, stationary, moving domaingeometry.Polygon) (domaingeometry.MultiPolygon, error) {
	if err := ctx.Err(); err != nil {
		return domaingeometry.MultiPolygon{}, err
	}
	if c == nil || c.delegate == nil {
		return domaingeometry.MultiPolygon{},
			fmt.Errorf("cached NFP builder is not initialized")
	}

	key := exactCacheKey(stationary, moving)

	for _, entry := range c.entries[key] {
		if !samePolygonBits(stationary, entry.stationary) ||
			!samePolygonBits(moving, entry.moving) {
			continue
		}

		result := cloneMultiPolygon(entry.result)
		if err := ctx.Err(); err != nil {
			return domaingeometry.MultiPolygon{}, err
		}
		c.hits++
		return result, nil
	}

	c.misses++

	result, err := geometry.BuildExternal(ctx, c.delegate, stationary, moving)
	if err != nil {
		return result, err
	}
	if err := ctx.Err(); err != nil {
		return domaingeometry.MultiPolygon{}, err
	}

	if c.count >= defaultCacheMaxEntries {
		return result, nil
	}

	entrySize := cacheEntryBytes(stationary, moving, result)

	if entrySize > uint64(defaultCacheMaxBytes)-c.usedBytes {
		return result, nil
	}

	entry := cacheEntry{
		stationary: clonePolygon(stationary),
		moving:     clonePolygon(moving),
		result:     cloneMultiPolygon(result),
	}
	if err := ctx.Err(); err != nil {
		return domaingeometry.MultiPolygon{}, err
	}
	c.entries[key] = append(c.entries[key], entry)

	c.count++
	c.usedBytes += entrySize

	return result, nil
}

func exactCacheKey(
	stationary domaingeometry.Polygon,
	moving domaingeometry.Polygon,
) [32]byte {
	h := sha256.New()

	writePolygonBits(h, stationary)
	writePolygonBits(h, moving)

	var key [32]byte
	copy(key[:], h.Sum(nil))

	return key
}

func writeUint64(h hash.Hash, value uint64) {
	var buffer [8]byte
	binary.LittleEndian.PutUint64(buffer[:], value)
	_, _ = h.Write(buffer[:])
}

func writeRingBits(h hash.Hash, ring domaingeometry.Ring) {
	writeUint64(h, uint64(len(ring.Points)))

	for _, point := range ring.Points {
		writeUint64(h, math.Float64bits(point.X))
		writeUint64(h, math.Float64bits(point.Y))
	}
}

func writePolygonBits(
	h hash.Hash,
	polygon domaingeometry.Polygon,
) {
	writeRingBits(h, polygon.Exterior)
	writeUint64(h, uint64(len(polygon.Holes)))

	for _, hole := range polygon.Holes {
		writeRingBits(h, hole)
	}
}

func sameRingBits(a, b domaingeometry.Ring) bool {
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

func samePolygonBits(
	a, b domaingeometry.Polygon,
) bool {
	if !sameRingBits(a.Exterior, b.Exterior) ||
		len(a.Holes) != len(b.Holes) {
		return false
	}

	for i := range a.Holes {
		if !sameRingBits(a.Holes[i], b.Holes[i]) {
			return false
		}
	}

	return true
}

func cloneRing(ring domaingeometry.Ring) domaingeometry.Ring {
	result := ring

	if ring.Points != nil {
		result.Points = append(
			[]domaingeometry.Point{},
			ring.Points...,
		)
	}

	return result
}

func clonePolygon(
	polygon domaingeometry.Polygon,
) domaingeometry.Polygon {
	result := polygon
	result.Exterior = cloneRing(polygon.Exterior)

	if polygon.Holes != nil {
		result.Holes = make(
			[]domaingeometry.Ring,
			len(polygon.Holes),
		)

		for i, hole := range polygon.Holes {
			result.Holes[i] = cloneRing(hole)
		}
	}

	return result
}

func cloneMultiPolygon(
	multi domaingeometry.MultiPolygon,
) domaingeometry.MultiPolygon {
	result := multi

	if multi.Polygons != nil {
		result.Polygons = make(
			[]domaingeometry.Polygon,
			len(multi.Polygons),
		)

		for i, polygon := range multi.Polygons {
			result.Polygons[i] = clonePolygon(polygon)
		}
	}

	return result
}

// All size calculations are capped at maxBytes+1.
// They estimate retained geometry data, not actual Go heap usage.
func addLimited(a, b uint64) uint64 {
	limit := uint64(defaultCacheMaxBytes)

	if a > limit || b > limit-a {
		return limit + 1
	}

	return a + b
}

func multiplyLimited(count int, elementSize uintptr) uint64 {
	limit := uint64(defaultCacheMaxBytes)
	size := uint64(elementSize)

	if size == 0 {
		return 0
	}

	if uint64(count) > limit/size {
		return limit + 1
	}

	return uint64(count) * size
}

func ringDataBytes(ring domaingeometry.Ring) uint64 {
	return multiplyLimited(
		len(ring.Points),
		unsafe.Sizeof(domaingeometry.Point{}),
	)
}

func polygonDataBytes(
	polygon domaingeometry.Polygon,
) uint64 {
	size := ringDataBytes(polygon.Exterior)

	size = addLimited(
		size,
		multiplyLimited(
			len(polygon.Holes),
			unsafe.Sizeof(domaingeometry.Ring{}),
		),
	)

	for _, hole := range polygon.Holes {
		size = addLimited(size, ringDataBytes(hole))
		if size > defaultCacheMaxBytes {
			return size
		}
	}

	return size
}

func multiPolygonDataBytes(
	multi domaingeometry.MultiPolygon,
) uint64 {
	size := multiplyLimited(
		len(multi.Polygons),
		unsafe.Sizeof(domaingeometry.Polygon{}),
	)

	for _, polygon := range multi.Polygons {
		size = addLimited(size, polygonDataBytes(polygon))
		if size > defaultCacheMaxBytes {
			return size
		}
	}

	return size
}

func cacheEntryBytes(
	stationary domaingeometry.Polygon,
	moving domaingeometry.Polygon,
	result domaingeometry.MultiPolygon,
) uint64 {
	size := uint64(32) + uint64(unsafe.Sizeof(cacheEntry{}))

	size = addLimited(size, polygonDataBytes(stationary))
	size = addLimited(size, polygonDataBytes(moving))
	size = addLimited(size, multiPolygonDataBytes(result))

	return size
}
