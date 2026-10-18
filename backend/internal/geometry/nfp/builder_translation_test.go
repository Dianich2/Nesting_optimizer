package nfp

import (
	"math"
	"testing"

	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry/simplefeatures"

	"github.com/stretchr/testify/require"
)

func translatePolygon(
	polygon domaingeometry.Polygon,
	dx, dy float64,
) domaingeometry.Polygon {
	result := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: make(
				[]domaingeometry.Point,
				len(polygon.Exterior.Points),
			),
		},
		Holes: make(
			[]domaingeometry.Ring,
			len(polygon.Holes),
		),
	}

	for i, point := range polygon.Exterior.Points {
		result.Exterior.Points[i] = domaingeometry.Point{
			X: point.X + dx,
			Y: point.Y + dy,
		}
	}

	for i, hole := range polygon.Holes {
		result.Holes[i].Points = make(
			[]domaingeometry.Point,
			len(hole.Points),
		)

		for j, point := range hole.Points {
			result.Holes[i].Points[j] = domaingeometry.Point{
				X: point.X + dx,
				Y: point.Y + dy,
			}
		}
	}

	return result
}

func TestBuildExternalStationaryTranslationRectangles(t *testing.T) {
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

	const dx = 25.0
	const dy = -17.0

	original, err := builder.BuildExternal(
		stationary,
		moving,
	)
	require.NoError(t, err)
	require.Len(t, original.Polygons, 1)

	shiftedStationary := translatePolygon(
		stationary,
		dx,
		dy,
	)

	shifted, err := builder.BuildExternal(
		shiftedStationary,
		moving,
	)
	require.NoError(t, err)
	require.Len(t, shifted.Polygons, 1)

	originalBounds, err := engine.Bounds(
		original.Polygons[0],
	)
	require.NoError(t, err)

	shiftedBounds, err := engine.Bounds(
		shifted.Polygons[0],
	)
	require.NoError(t, err)

	require.InDelta(
		t,
		originalBounds.MinX+dx,
		shiftedBounds.MinX,
		1e-8,
	)
	require.InDelta(
		t,
		originalBounds.MaxX+dx,
		shiftedBounds.MaxX,
		1e-8,
	)
	require.InDelta(
		t,
		originalBounds.MinY+dy,
		shiftedBounds.MinY,
		1e-8,
	)
	require.InDelta(
		t,
		originalBounds.MaxY+dy,
		shiftedBounds.MaxY,
		1e-8,
	)

	originalArea, err := engine.Area(original.Polygons[0])
	require.NoError(t, err)

	shiftedArea, err := engine.Area(shifted.Polygons[0])
	require.NoError(t, err)

	require.InDelta(
		t,
		originalArea,
		shiftedArea,
		1e-8,
	)
}

func requireTranslatedRingVertices(
	t *testing.T,
	original domaingeometry.Ring,
	shifted domaingeometry.Ring,
	dx, dy float64,
) {
	t.Helper()

	require.Len(t, shifted.Points, len(original.Points))

	used := make([]bool, len(shifted.Points))

	for _, originalPoint := range original.Points {
		expectedX := originalPoint.X + dx
		expectedY := originalPoint.Y + dy

		found := false

		for i, shiftedPoint := range shifted.Points {
			if used[i] {
				continue
			}

			if math.Abs(shiftedPoint.X-expectedX) <= 1e-7 &&
				math.Abs(shiftedPoint.Y-expectedY) <= 1e-7 {
				used[i] = true
				found = true
				break
			}
		}

		require.True(
			t,
			found,
			"translated vertex (%.10f, %.10f) not found",
			expectedX,
			expectedY,
		)
	}
}

func TestBuildExternalStationaryTranslationComplexPolygons(t *testing.T) {
	engine := simplefeatures.NewEngine()
	builder := NewBuilder(engine)

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

	tests := []struct {
		name       string
		stationary domaingeometry.Polygon
		dx         float64
		dy         float64
	}{
		{
			name: "concave polygon",
			stationary: domaingeometry.Polygon{
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
			},
			dx: 37,
			dy: -23,
		},
		{
			name: "polygon with hole",
			stationary: domaingeometry.Polygon{
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
			},
			dx: -31,
			dy: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original, err := builder.BuildExternal(
				tt.stationary,
				moving,
			)
			require.NoError(t, err)

			shiftedStationary := translatePolygon(
				tt.stationary,
				tt.dx,
				tt.dy,
			)

			shifted, err := builder.BuildExternal(
				shiftedStationary,
				moving,
			)
			require.NoError(t, err)

			require.Len(
				t,
				shifted.Polygons,
				len(original.Polygons),
			)

			// These fixtures currently produce one NFP polygon.
			// Verify that assumption explicitly.
			require.Len(t, original.Polygons, 1)

			originalPolygon := original.Polygons[0]
			shiftedPolygon := shifted.Polygons[0]

			require.NoError(
				t,
				engine.ValidatePolygon(originalPolygon),
			)
			require.NoError(
				t,
				engine.ValidatePolygon(shiftedPolygon),
			)

			require.Len(
				t,
				shiftedPolygon.Holes,
				len(originalPolygon.Holes),
			)

			requireTranslatedRingVertices(
				t,
				originalPolygon.Exterior,
				shiftedPolygon.Exterior,
				tt.dx,
				tt.dy,
			)

			for i := range originalPolygon.Holes {
				requireTranslatedRingVertices(
					t,
					originalPolygon.Holes[i],
					shiftedPolygon.Holes[i],
					tt.dx,
					tt.dy,
				)
			}

			originalArea, err := engine.Area(originalPolygon)
			require.NoError(t, err)

			shiftedArea, err := engine.Area(shiftedPolygon)
			require.NoError(t, err)

			require.InDelta(
				t,
				originalArea,
				shiftedArea,
				1e-7,
			)
		})
	}
}
