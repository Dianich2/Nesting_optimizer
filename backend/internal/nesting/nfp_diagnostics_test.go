package nesting

import (
	"crypto/sha256"
	"encoding/binary"
	"hash"
	"math"
	"sync"

	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
)

type nfpDiagnosticStats struct {
	TotalCalls    uint64
	UniquePairs   uint64
	RepeatedCalls uint64
}

type diagnosticNFPBuilder struct {
	delegate geometry.NFPBuilder

	mu    sync.Mutex
	seen  map[[32]byte]struct{}
	stats nfpDiagnosticStats
}

var _ geometry.NFPBuilder = (*diagnosticNFPBuilder)(nil)

func newDiagnosticNFPBuilder(
	delegate geometry.NFPBuilder,
) *diagnosticNFPBuilder {
	return &diagnosticNFPBuilder{
		delegate: delegate,
		seen:     make(map[[32]byte]struct{}),
	}
}

func (b *diagnosticNFPBuilder) BuildExternal(
	stationary domaingeometry.Polygon,
	moving domaingeometry.Polygon,
) (domaingeometry.MultiPolygon, error) {
	key := diagnosticNFPKey(stationary, moving)

	b.mu.Lock()

	b.stats.TotalCalls++

	if _, exists := b.seen[key]; exists {
		b.stats.RepeatedCalls++
	} else {
		b.seen[key] = struct{}{}
		b.stats.UniquePairs++
	}

	b.mu.Unlock()

	// Always execute the real implementation.
	return b.delegate.BuildExternal(stationary, moving)
}

func (b *diagnosticNFPBuilder) Stats() nfpDiagnosticStats {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.stats
}

func diagnosticNFPKey(
	stationary domaingeometry.Polygon,
	moving domaingeometry.Polygon,
) [32]byte {
	h := sha256.New()

	writeDiagnosticPolygon(h, stationary)
	writeDiagnosticPolygon(h, moving)

	var key [32]byte
	copy(key[:], h.Sum(nil))

	return key
}

func writeDiagnosticPolygon(
	h hash.Hash,
	polygon domaingeometry.Polygon,
) {
	var anchor domaingeometry.Point

	if len(polygon.Exterior.Points) > 0 {
		anchor = polygon.Exterior.Points[0]
	}

	writeDiagnosticRing(h, polygon.Exterior, anchor)
	writeDiagnosticUint64(h, uint64(len(polygon.Holes)))

	for _, hole := range polygon.Holes {
		writeDiagnosticRing(h, hole, anchor)
	}
}

func writeDiagnosticRing(
	h hash.Hash,
	ring domaingeometry.Ring,
	anchor domaingeometry.Point,
) {
	writeDiagnosticUint64(h, uint64(len(ring.Points)))

	for _, point := range ring.Points {
		writeDiagnosticFloat64(h, point.X-anchor.X)
		writeDiagnosticFloat64(h, point.Y-anchor.Y)
	}
}

func writeDiagnosticFloat64(
	h hash.Hash,
	value float64,
) {
	// Normalize signed zero for diagnostic comparison.
	if value == 0 {
		value = 0
	}

	writeDiagnosticUint64(h, math.Float64bits(value))
}

func writeDiagnosticUint64(
	h hash.Hash,
	value uint64,
) {
	var data [8]byte

	binary.LittleEndian.PutUint64(data[:], value)
	_, _ = h.Write(data[:])
}
