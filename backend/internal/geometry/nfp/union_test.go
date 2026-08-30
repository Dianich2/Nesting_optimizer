package nfp

import (
	"server_nesting_optimizer/internal/config"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnionPolygonsOverlapping(t *testing.T) {
	first := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 0, Y: 0},
				{X: 4, Y: 0},
				{X: 4, Y: 4},
				{X: 0, Y: 4},
			},
		},
	}

	second := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 2, Y: 0},
				{X: 6, Y: 0},
				{X: 6, Y: 4},
				{X: 2, Y: 4},
			},
		},
	}

	result, err := unionPolygons(
		[]domaingeometry.Polygon{first, second},
	)

	require.NoError(t, err)
	require.Len(t, result.Polygons, 1)

	engine := simplefeatures.NewEngine()

	area, err := engine.Area(result.Polygons[0])
	require.NoError(t, err)

	assert.InDelta(
		t,
		24.0,
		area,
		config.Epsilon,
	)

	bounds, err := engine.Bounds(result.Polygons[0])
	require.NoError(t, err)

	assert.Equal(
		t,
		domaingeometry.Bounds{
			MinX: 0,
			MinY: 0,
			MaxX: 6,
			MaxY: 4,
		},
		bounds,
	)
}

func TestUnionPolygonsDisjoint(t *testing.T) {
	first := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 0, Y: 0},
				{X: 2, Y: 0},
				{X: 2, Y: 2},
				{X: 0, Y: 2},
			},
		},
	}

	second := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 10, Y: 10},
				{X: 12, Y: 10},
				{X: 12, Y: 12},
				{X: 10, Y: 12},
			},
		},
	}

	result, err := unionPolygons(
		[]domaingeometry.Polygon{first, second},
	)

	require.NoError(t, err)

	assert.Len(
		t,
		result.Polygons,
		2,
	)
}

func TestUnionPolygonsEmpty(t *testing.T) {
	_, err := unionPolygons(nil)

	require.Error(t, err)
}
