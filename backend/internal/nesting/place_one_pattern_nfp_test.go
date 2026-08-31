package nesting

import (
	"context"
	"server_nesting_optimizer/internal/config"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry/nfp"
	"server_nesting_optimizer/internal/geometry/simplefeatures"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlaceOnePatternNFPFirstPattern(t *testing.T) {
	engine := simplefeatures.NewEngine()
	nfpBuilder := nfp.NewBuilder(engine)

	pattern := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 0, Y: 0},
				{X: 20, Y: 0},
				{X: 20, Y: 20},
				{X: 0, Y: 20},
			},
		},
	}

	surface := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 0, Y: 0},
				{X: 100, Y: 0},
				{X: 100, Y: 100},
				{X: 0, Y: 100},
			},
		},
	}

	placement, found, err := placeOnePatternNFP(
		context.Background(),
		engine,
		nfpBuilder,
		pattern,
		surface,
		nil,
		[]float64{0},
	)

	require.NoError(t, err)
	require.True(t, found)

	assert.InDelta(
		t,
		10.0,
		placement.Candidate.X,
		config.Epsilon,
	)

	assert.InDelta(
		t,
		10.0,
		placement.Candidate.Y,
		config.Epsilon,
	)

	assert.InDelta(
		t,
		0.0,
		placement.Candidate.Rotation,
		config.Epsilon,
	)

	bounds, err := engine.Bounds(placement.Geometry)
	require.NoError(t, err)

	assert.InDelta(
		t,
		0.0,
		bounds.MinX,
		config.Epsilon,
	)

	assert.InDelta(
		t,
		0.0,
		bounds.MinY,
		config.Epsilon,
	)

	assert.InDelta(
		t,
		20.0,
		bounds.MaxX,
		config.Epsilon,
	)

	assert.InDelta(
		t,
		20.0,
		bounds.MaxY,
		config.Epsilon,
	)
}

func TestPlaceOnePatternNFPTouchesOccupied(t *testing.T) {
	engine := simplefeatures.NewEngine()
	nfpBuilder := nfp.NewBuilder(engine)

	pattern := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 0, Y: 0},
				{X: 20, Y: 0},
				{X: 20, Y: 20},
				{X: 0, Y: 20},
			},
		},
	}

	surface := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 0, Y: 0},
				{X: 100, Y: 0},
				{X: 100, Y: 100},
				{X: 0, Y: 100},
			},
		},
	}

	occupied := []domaingeometry.Polygon{
		{
			Exterior: domaingeometry.Ring{
				Points: []domaingeometry.Point{
					{X: 0, Y: 0},
					{X: 20, Y: 0},
					{X: 20, Y: 20},
					{X: 0, Y: 20},
				},
			},
		},
	}

	placement, found, err := placeOnePatternNFP(
		context.Background(),
		engine,
		nfpBuilder,
		pattern,
		surface,
		occupied,
		[]float64{0},
	)

	require.NoError(t, err)
	require.True(t, found)

	assert.InDelta(
		t,
		30.0,
		placement.Candidate.X,
		config.Epsilon,
	)

	assert.InDelta(
		t,
		10.0,
		placement.Candidate.Y,
		config.Epsilon,
	)

	intersects, err := engine.InteriorsIntersect(
		occupied[0],
		placement.Geometry,
	)
	require.NoError(t, err)

	assert.False(t, intersects)
}

func TestPlaceOnePatternNFPRotationNeeded(t *testing.T) {
	engine := simplefeatures.NewEngine()
	nfpBuilder := nfp.NewBuilder(engine)

	pattern := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 0, Y: 0},
				{X: 40, Y: 0},
				{X: 40, Y: 20},
				{X: 0, Y: 20},
			},
		},
	}

	surface := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 0, Y: 0},
				{X: 30, Y: 0},
				{X: 30, Y: 50},
				{X: 0, Y: 50},
			},
		},
	}

	placement, found, err := placeOnePatternNFP(
		context.Background(),
		engine,
		nfpBuilder,
		pattern,
		surface,
		nil,
		[]float64{0, 90},
	)

	require.NoError(t, err)
	require.True(t, found)

	assert.InDelta(
		t,
		90.0,
		placement.Candidate.Rotation,
		config.Epsilon,
	)
}

func TestPlaceOnePatternNFPNoMoreSpace(t *testing.T) {
	engine := simplefeatures.NewEngine()
	nfpBuilder := nfp.NewBuilder(engine)

	surface := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 0, Y: 0},
				{X: 20, Y: 0},
				{X: 20, Y: 20},
				{X: 0, Y: 20},
			},
		},
	}

	occupied := []domaingeometry.Polygon{
		surface,
	}

	pattern := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 0, Y: 0},
				{X: 10, Y: 0},
				{X: 10, Y: 10},
				{X: 0, Y: 10},
			},
		},
	}

	placement, found, err := placeOnePatternNFP(
		context.Background(),
		engine,
		nfpBuilder,
		pattern,
		surface,
		occupied,
		[]float64{0},
	)

	require.NoError(t, err)
	assert.False(t, found)
	assert.Zero(t, placement)
}

func TestPlaceOnePatternNFPCanceledContext(t *testing.T) {
	engine := simplefeatures.NewEngine()
	nfpBuilder := nfp.NewBuilder(engine)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	placement, found, err := placeOnePatternNFP(
		ctx,
		engine,
		nfpBuilder,
		domaingeometry.Polygon{},
		domaingeometry.Polygon{},
		nil,
		[]float64{0},
	)

	assert.Zero(t, placement)
	assert.False(t, found)
	assert.ErrorIs(t, err, context.Canceled)
}
