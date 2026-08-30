package nfp

import (
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReflect(t *testing.T) {
	// arrange
	polygon := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 2, Y: 1},
				{X: -4, Y: 1},
				{X: -4, Y: -3},
				{X: 2, Y: -3},
			},
		},
		Holes: []domaingeometry.Ring{
			{
				Points: []domaingeometry.Point{
					{X: 0, Y: 0},
					{X: -1, Y: 0},
					{X: -1, Y: -1},
					{X: 0, Y: -1},
				},
			},
		},
	}

	wantPolygon := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: -2, Y: -1},
				{X: 4, Y: -1},
				{X: 4, Y: 3},
				{X: -2, Y: 3},
			},
		},
		Holes: []domaingeometry.Ring{
			{
				Points: []domaingeometry.Point{
					{X: 0, Y: 0},
					{X: 1, Y: 0},
					{X: 1, Y: 1},
					{X: 0, Y: 1},
				},
			},
		},
	}

	// act
	reflectedPolygon := reflectAtOrigin(polygon)

	// assert
	assert.Equal(
		t,
		wantPolygon,
		reflectedPolygon,
	)
}

func TestReflectNoChangeOrigin(t *testing.T) {
	// arrange
	polygon := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 2, Y: 1},
				{X: -4, Y: 1},
				{X: -4, Y: -3},
				{X: 2, Y: -3},
			},
		},
		Holes: []domaingeometry.Ring{
			{
				Points: []domaingeometry.Point{
					{X: 0, Y: 0},
					{X: -1, Y: 0},
					{X: -1, Y: -1},
					{X: 0, Y: -1},
				},
			},
		},
	}

	copyPolygon := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 2, Y: 1},
				{X: -4, Y: 1},
				{X: -4, Y: -3},
				{X: 2, Y: -3},
			},
		},
		Holes: []domaingeometry.Ring{
			{
				Points: []domaingeometry.Point{
					{X: 0, Y: 0},
					{X: -1, Y: 0},
					{X: -1, Y: -1},
					{X: 0, Y: -1},
				},
			},
		},
	}

	// act
	_ = reflectAtOrigin(polygon)

	// assert
	assert.Equal(
		t,
		copyPolygon,
		polygon,
	)
}
