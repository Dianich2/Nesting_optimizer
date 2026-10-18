package nesting

import (
	"crypto/sha256"
	"encoding/binary"
	"hash"
	"math"

	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
)

type exactNFPDiagnosticBuilder struct {
	delegate geometry.NFPBuilder

	seen     map[[32]byte]struct{}
	total    uint64
	repeated uint64
}

var _ geometry.NFPBuilder = (*exactNFPDiagnosticBuilder)(nil)

func newExactNFPDiagnosticBuilder(
	delegate geometry.NFPBuilder,
) *exactNFPDiagnosticBuilder {
	return &exactNFPDiagnosticBuilder{
		delegate: delegate,
		seen:     make(map[[32]byte]struct{}),
	}
}

func (b *exactNFPDiagnosticBuilder) BuildExternal(
	stationary domaingeometry.Polygon,
	moving domaingeometry.Polygon,
) (domaingeometry.MultiPolygon, error) {
	key := exactNFPInputKey(stationary, moving)

	b.total++

	if _, exists := b.seen[key]; exists {
		b.repeated++
	} else {
		b.seen[key] = struct{}{}
	}

	return b.delegate.BuildExternal(stationary, moving)
}

func exactNFPInputKey(
	stationary domaingeometry.Polygon,
	moving domaingeometry.Polygon,
) [32]byte {
	h := sha256.New()

	writeExactNFPPolygon(h, stationary)
	writeExactNFPPolygon(h, moving)

	var key [32]byte
	copy(key[:], h.Sum(nil))

	return key
}

func writeExactNFPPolygon(
	h hash.Hash,
	polygon domaingeometry.Polygon,
) {
	writeExactNFPRing(h, polygon.Exterior)

	writeExactNFPUint64(
		h,
		uint64(len(polygon.Holes)),
	)

	for _, hole := range polygon.Holes {
		writeExactNFPRing(h, hole)
	}
}

func writeExactNFPRing(
	h hash.Hash,
	ring domaingeometry.Ring,
) {
	writeExactNFPUint64(
		h,
		uint64(len(ring.Points)),
	)

	for _, point := range ring.Points {
		writeExactNFPUint64(
			h,
			math.Float64bits(point.X),
		)

		writeExactNFPUint64(
			h,
			math.Float64bits(point.Y),
		)
	}
}

func writeExactNFPUint64(
	h hash.Hash,
	value uint64,
) {
	var data [8]byte

	binary.LittleEndian.PutUint64(
		data[:],
		value,
	)

	_, _ = h.Write(data[:])
}
