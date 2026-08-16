package simplefeatures

import (
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCodecRoundTrip(t *testing.T) {
	sfCodec := NewCodec()

	tests := []struct {
		name    string
		polygon domaingeometry.Polygon
	}{
		{
			name: "valid rectangle",
			polygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 100, Y: 0},
						{X: 100, Y: 50},
						{X: 0, Y: 50},
					},
				},
			},
		},
		{
			name: "valid polygon with hole",
			polygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 100, Y: 0},
						{X: 100, Y: 100},
						{X: 0, Y: 100},
					},
				},
				Holes: []domaingeometry.Ring{
					domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 30, Y: 30},
							{X: 30, Y: 60},
							{X: 60, Y: 60},
							{X: 60, Y: 30},
						},
					},
				},
			},
		},
		{
			name: "polygon with fractional coordinates",
			polygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0.5, Y: 0.25},
						{X: 10.75, Y: 0.25},
						{X: 10.75, Y: 5.5},
						{X: 0.5, Y: 5.5},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := sfCodec.EncodeWKB(tt.polygon)
			require.NoError(t, err)

			decoded, err := sfCodec.DecodeWKB(encoded)
			require.NoError(t, err)

			assert.Equal(
				t,
				tt.polygon,
				decoded,
			)
		})
	}
}

func TestDecodeWKB_InvalidData(t *testing.T) {
	// arrange
	sfCodec := NewCodec()
	data := []byte{
		0x01,
		0x02,
		0x03,
	}

	// act
	polygon, err := sfCodec.DecodeWKB(data)

	// assert
	assert.Zero(t, polygon)

	assert.ErrorIs(
		t,
		err,
		geometry.ErrInvalidWKB,
	)
}

func TestDecodeWKB_NonPolygonGeometry(t *testing.T) {
	// arrange
	sfCodec := NewCodec()
	point := domaingeometry.Point{
		X: 2,
		Y: 2,
	}
	sfPoint := geom.NewPointXY(point.X, point.Y)
	pointAsWKB := sfPoint.AsGeometry().AsBinary()

	// act
	polygon, err := sfCodec.DecodeWKB(pointAsWKB)

	// assert
	assert.Zero(t, polygon)

	assert.ErrorIs(
		t,
		err,
		geometry.ErrInvalidWKB,
	)
}
