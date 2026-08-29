package nesting

import (
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRingVertices(t *testing.T) {
	tests := []struct {
		name         string
		polygon      domaingeometry.Polygon
		wantVertices []domaingeometry.Point
	}{
		{
			name: "closed rectangle",
			polygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 20, Y: 0},
						{X: 20, Y: 20},
						{X: 0, Y: 20},
						{X: 0, Y: 0},
					},
				},
			},
			wantVertices: []domaingeometry.Point{
				{
					X: 0,
					Y: 0,
				},
				{
					X: 20,
					Y: 0,
				},
				{
					X: 20,
					Y: 20,
				},
				{
					X: 0,
					Y: 20,
				},
			},
		},
		{
			name: "unclosed rectangle",
			polygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 20, Y: 0},
						{X: 20, Y: 20},
						{X: 0, Y: 20},
					},
				},
			},
			wantVertices: []domaingeometry.Point{
				{
					X: 0,
					Y: 0,
				},
				{
					X: 20,
					Y: 0,
				},
				{
					X: 20,
					Y: 20,
				},
				{
					X: 0,
					Y: 20,
				},
			},
		},
		{
			name: "empty ring",
			polygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{},
				},
			},
			wantVertices: nil,
		},
		{
			name: "single point",
			polygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
					},
				},
			},
			wantVertices: []domaingeometry.Point{
				{
					X: 0,
					Y: 0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			assert.Equal(
				t,
				tt.wantVertices,
				ringVertices(tt.polygon.Exterior),
			)

		})
	}
}

func TestRingVerticesReturnsIndependentSlice(t *testing.T) {
	// arrange
	polygon := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 0, Y: 0},
				{X: 100, Y: 0},
				{X: 100, Y: 100},
				{X: 0, Y: 100},
			},
		},
	}

	firstPointBeforeVertices := polygon.Exterior.Points[0]

	vertices := ringVertices(polygon.Exterior)
	vertices[0].X = 10
	vertices[0].Y = 15

	assert.Equal(
		t,
		firstPointBeforeVertices,
		polygon.Exterior.Points[0],
	)
}
