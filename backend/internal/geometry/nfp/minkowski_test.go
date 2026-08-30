package nfp

import (
	"server_nesting_optimizer/internal/config"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMinkowskiRectangles(t *testing.T) {
	// arrange
	sfEngine := simplefeatures.NewEngine()

	first := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 0, Y: 0},
				{X: 4, Y: 0},
				{X: 4, Y: 3},
				{X: 0, Y: 3},
			},
		},
	}

	second := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 0, Y: 0},
				{X: 2, Y: 0},
				{X: 2, Y: 1},
				{X: 0, Y: 1},
			},
		},
	}

	// act
	minkowskiSum, err := convexMinkowskiSum(
		first,
		second,
	)

	// assert
	require.NoError(t, err)

	bounds, err := sfEngine.Bounds(minkowskiSum)
	require.NoError(t, err)

	area, err := sfEngine.Area(minkowskiSum)
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

	assert.InDelta(
		t,
		24,
		area,
		config.Epsilon,
	)
}

func TestMinkowskiTriangles(t *testing.T) {
	// arrange
	sfEngine := simplefeatures.NewEngine()

	first := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 0, Y: 0},
				{X: 2, Y: 0},
				{X: 0, Y: 2},
			},
		},
	}

	second := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 0, Y: 0},
				{X: 1, Y: 0},
				{X: 0, Y: 1},
			},
		},
	}

	// act
	minkowskiSum, err := convexMinkowskiSum(
		first,
		second,
	)

	// assert
	require.NoError(t, err)

	bounds, err := sfEngine.Bounds(minkowskiSum)
	require.NoError(t, err)

	area, err := sfEngine.Area(minkowskiSum)
	require.NoError(t, err)

	assert.Equal(
		t,
		domaingeometry.Bounds{
			MinX: 0,
			MinY: 0,
			MaxX: 3,
			MaxY: 3,
		},
		bounds,
	)

	assert.InDelta(
		t,
		4.5,
		area,
		config.Epsilon,
	)
}

func TestMinkowskiCommutativity(t *testing.T) {
	sfEngine := simplefeatures.NewEngine()

	first := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 1, Y: 2},
				{X: 4, Y: 2},
				{X: 2, Y: 5},
			},
		},
	}

	second := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: -2, Y: -1},
				{X: 1, Y: -1},
				{X: 1, Y: 1},
				{X: -2, Y: 1},
			},
		},
	}

	firstSecond, err := convexMinkowskiSum(first, second)
	require.NoError(t, err)

	secondFirst, err := convexMinkowskiSum(second, first)
	require.NoError(t, err)

	firstBounds, err := sfEngine.Bounds(firstSecond)
	require.NoError(t, err)

	secondBounds, err := sfEngine.Bounds(secondFirst)
	require.NoError(t, err)

	assert.Equal(
		t,
		firstBounds,
		secondBounds,
	)

	firstArea, err := sfEngine.Area(firstSecond)
	require.NoError(t, err)

	secondArea, err := sfEngine.Area(secondFirst)
	require.NoError(t, err)

	assert.InDelta(
		t,
		firstArea,
		secondArea,
		config.Epsilon,
	)
}

func TestMinkowskiWithHoleReturnsError(t *testing.T) {
	first := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 0, Y: 0},
				{X: 10, Y: 0},
				{X: 10, Y: 10},
				{X: 0, Y: 10},
			},
		},
		Holes: []domaingeometry.Ring{
			{
				Points: []domaingeometry.Point{
					{X: 2, Y: 2},
					{X: 4, Y: 2},
					{X: 4, Y: 4},
					{X: 2, Y: 4},
				},
			},
		},
	}

	second := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 0, Y: 0},
				{X: 1, Y: 0},
				{X: 0, Y: 1},
			},
		},
	}

	_, err := convexMinkowskiSum(
		first,
		second,
	)

	require.Error(t, err)
}
