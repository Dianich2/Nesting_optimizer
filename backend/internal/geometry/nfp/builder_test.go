package nfp

import (
	"server_nesting_optimizer/internal/config"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildExternalRectangles(t *testing.T) {
	engine := simplefeatures.NewEngine()
	builder := NewBuilder(engine)

	stationary := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 0, Y: 0},
				{X: 4, Y: 0},
				{X: 4, Y: 3},
				{X: 0, Y: 3},
			},
		},
	}

	moving := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 0, Y: 0},
				{X: 2, Y: 0},
				{X: 2, Y: 1},
				{X: 0, Y: 1},
			},
		},
	}

	result, err := builder.BuildExternal(
		stationary,
		moving,
	)

	require.NoError(t, err)
	require.Len(t, result.Polygons, 1)

	nfpPolygon := result.Polygons[0]

	require.NoError(
		t,
		engine.ValidatePolygon(nfpPolygon),
	)

	bounds, err := engine.Bounds(nfpPolygon)
	require.NoError(t, err)

	assert.Equal(
		t,
		domaingeometry.Bounds{
			MinX: -1,
			MinY: -0.5,
			MaxX: 5,
			MaxY: 3.5,
		},
		bounds,
	)

	area, err := engine.Area(nfpPolygon)
	require.NoError(t, err)

	assert.InDelta(
		t,
		24.0,
		area,
		config.Epsilon,
	)
}

func TestBuildExternalConcave(t *testing.T) {
	engine := simplefeatures.NewEngine()
	builder := NewBuilder(engine)

	stationary := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 0, Y: 0},
				{X: 100, Y: 0},
				{X: 100, Y: 50},
				{X: 50, Y: 50},
				{X: 50, Y: 100},
				{X: 0, Y: 100},
			},
		},
	}

	moving := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 0, Y: 0},
				{X: 20, Y: 0},
				{X: 20, Y: 20},
				{X: 0, Y: 20},
			},
		},
	}

	result, err := builder.BuildExternal(
		stationary,
		moving,
	)

	require.NoError(t, err)
	require.NotEmpty(t, result.Polygons)

	for i, polygon := range result.Polygons {
		if err := engine.ValidatePolygon(polygon); err != nil {
			t.Fatalf(
				"NFP polygon %d is invalid: %v",
				i,
				err,
			)
		}

		area, err := engine.Area(polygon)
		require.NoError(t, err)
		assert.Greater(t, area, 0.0)
	}
}

func TestBuildExternalWithHole(t *testing.T) {
	engine := simplefeatures.NewEngine()
	builder := NewBuilder(engine)

	stationary := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 0, Y: 0},
				{X: 100, Y: 0},
				{X: 100, Y: 100},
				{X: 0, Y: 100},
			},
		},
		Holes: []domaingeometry.Ring{
			{
				Points: []domaingeometry.Point{
					{X: 25, Y: 25},
					{X: 25, Y: 75},
					{X: 75, Y: 75},
					{X: 75, Y: 25},
				},
			},
		},
	}

	moving := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 0, Y: 0},
				{X: 10, Y: 0},
				{X: 10, Y: 10},
				{X: 0, Y: 10},
			},
		},
	}

	result, err := builder.BuildExternal(
		stationary,
		moving,
	)

	require.NoError(t, err)
	require.NotEmpty(t, result.Polygons)

	require.Len(
		t,
		result.Polygons,
		1,
	)

	nfpPolygon := result.Polygons[0]

	require.Len(
		t,
		nfpPolygon.Holes,
		1,
	)

	bounds, err := engine.Bounds(nfpPolygon)
	require.NoError(t, err)

	assert.Equal(
		t,
		domaingeometry.Bounds{
			MinX: -5,
			MinY: -5,
			MaxX: 105,
			MaxY: 105,
		},
		bounds,
	)

	holePolygon := domaingeometry.Polygon{
		Exterior: nfpPolygon.Holes[0],
	}

	holeBounds, err := engine.Bounds(holePolygon)
	require.NoError(t, err)

	assert.Equal(
		t,
		domaingeometry.Bounds{
			MinX: 30,
			MinY: 30,
			MaxX: 70,
			MaxY: 70,
		},
		holeBounds,
	)

	area, err := engine.Area(nfpPolygon)
	require.NoError(t, err)

	assert.InDelta(
		t,
		10500.0,
		area,
		config.Epsilon,
	)

	for _, polygon := range result.Polygons {
		require.NoError(
			t,
			engine.ValidatePolygon(polygon),
		)
	}
}
