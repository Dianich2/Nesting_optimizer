package simplefeatures

import (
	"math"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const epsilon = 1e-9

func assertPolygonInDelta(
	t *testing.T,
	want domaingeometry.Polygon,
	got domaingeometry.Polygon,
	delta float64,
) {
	t.Helper()

	assert.Len(t, got.Exterior.Points, len(want.Exterior.Points))
	for i := range want.Exterior.Points {
		assert.InDelta(t, want.Exterior.Points[i].X, got.Exterior.Points[i].X, delta)
		assert.InDelta(t, want.Exterior.Points[i].Y, got.Exterior.Points[i].Y, delta)
	}

	assert.Len(t, got.Holes, len(want.Holes))
	for i := range want.Holes {
		assert.Len(t, got.Holes[i].Points, len(want.Holes[i].Points))

		for j := range want.Holes[i].Points {
			assert.InDelta(t, want.Holes[i].Points[j].X, got.Holes[i].Points[j].X, delta)
			assert.InDelta(t, want.Holes[i].Points[j].Y, got.Holes[i].Points[j].Y, delta)
		}
	}
}

func TestValidatePolygon(t *testing.T) {
	sfEngine := NewEngine()

	tests := []struct {
		name    string
		polygon domaingeometry.Polygon
		wantErr error
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
			wantErr: nil,
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
			wantErr: nil,
		},
		{
			name:    "empty polygon",
			polygon: domaingeometry.Polygon{},
			wantErr: geometry.ErrInvalidPolygon,
		},
		{
			name: "self-intersecting exterior",
			polygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 100, Y: 0},
						{X: 100, Y: 50},
						{X: 50, Y: -10},
					},
				},
			},
			wantErr: geometry.ErrInvalidPolygon,
		},
		{
			name: "hole intersect exterior",
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
							{X: 30, Y: 120},
							{X: 60, Y: 120},
							{X: 60, Y: 30},
						},
					},
				},
			},
			wantErr: geometry.ErrInvalidPolygon,
		},
		{
			name: "hole outside exterior",
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
							{X: 300, Y: 300},
							{X: 300, Y: 120},
							{X: 120, Y: 120},
							{X: 120, Y: 300},
						},
					},
				},
			},
			wantErr: geometry.ErrInvalidPolygon,
		},
		{
			name: "collinear exterior points",
			polygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 1, Y: 1},
						{X: 2, Y: 2},
						{X: 3, Y: 3},
					},
				},
			},
			wantErr: geometry.ErrInvalidPolygon,
		},
		{
			name: "NaN coordinate",
			polygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: math.NaN(), Y: 1},
						{X: 2, Y: math.NaN()},
						{X: 3, Y: 3},
					},
				},
			},
			wantErr: geometry.ErrInvalidPolygon,
		},
		{
			name: "Inf coordinate",
			polygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: math.Inf(1), Y: 1},
						{X: 2, Y: math.Inf(1)},
						{X: 3, Y: 3},
					},
				},
			},
			wantErr: geometry.ErrInvalidPolygon,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sfEngine.ValidatePolygon(tt.polygon)

			if tt.wantErr == nil {
				assert.NoError(t, err)
				return
			}

			assert.ErrorIs(
				t,
				err,
				tt.wantErr,
			)
		})
	}
}

func TestArea(t *testing.T) {
	sfEngine := NewEngine()

	tests := []struct {
		name     string
		polygon  domaingeometry.Polygon
		wantArea float64
		wantErr  error
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
			wantArea: 5000,
			wantErr:  nil,
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
			wantArea: 9100,
			wantErr:  nil,
		},
		{
			name: "valid polygon with two holes",
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
					domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 90, Y: 90},
							{X: 95, Y: 90},
							{X: 95, Y: 95},
							{X: 90, Y: 95},
						},
					},
				},
			},
			wantArea: 9075,
			wantErr:  nil,
		},
		{
			name: "valid polygon with fractional coordinates",
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
			wantArea: 53.8125,
			wantErr:  nil,
		},
		{
			name: "invalid polygon",
			polygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 100, Y: 0},
						{X: 100, Y: 50},
						{X: 50, Y: -10},
					},
				},
			},
			wantErr: geometry.ErrInvalidPolygon,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			area, err := sfEngine.Area(tt.polygon)

			if tt.wantErr == nil {
				assert.NoError(t, err)

				assert.InDelta(
					t,
					tt.wantArea,
					area,
					epsilon,
				)

				return
			}

			assert.Zero(t, area)

			assert.ErrorIs(
				t,
				err,
				tt.wantErr,
			)
		})
	}
}

func TestNormalize(t *testing.T) {
	sfEngine := NewEngine()

	tests := []struct {
		name        string
		polygon     domaingeometry.Polygon
		wantPolygon domaingeometry.Polygon
		wantErr     error
	}{
		{
			name: "normalized polygon",
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
			wantPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 100, Y: 0},
						{X: 100, Y: 50},
						{X: 0, Y: 50},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "positive translated polygon",
			polygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 10, Y: 10},
						{X: 100, Y: 10},
						{X: 100, Y: 50},
						{X: 10, Y: 50},
					},
				},
			},
			wantPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 90, Y: 0},
						{X: 90, Y: 40},
						{X: 0, Y: 40},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "negative translated polygon",
			polygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: -10, Y: -10},
						{X: 100, Y: -10},
						{X: 100, Y: 50},
						{X: -10, Y: 50},
					},
				},
			},
			wantPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 110, Y: 0},
						{X: 110, Y: 60},
						{X: 0, Y: 60},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "polygon with fractional coordinates",
			polygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 2.55, Y: 2.55},
						{X: 10.5, Y: 2.55},
						{X: 10.5, Y: 9.45},
						{X: 2.55, Y: 9.45},
					},
				},
			},
			wantPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 7.95, Y: 0},
						{X: 7.95, Y: 6.9},
						{X: 0, Y: 6.9},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "polygon with two holes",
			polygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 5, Y: 5},
						{X: 100, Y: 5},
						{X: 100, Y: 100},
						{X: 5, Y: 100},
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
					domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 90, Y: 90},
							{X: 95, Y: 90},
							{X: 95, Y: 95},
							{X: 90, Y: 95},
						},
					},
				},
			},
			wantPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 95, Y: 0},
						{X: 95, Y: 95},
						{X: 0, Y: 95},
					},
				},
				Holes: []domaingeometry.Ring{
					domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 25, Y: 25},
							{X: 25, Y: 55},
							{X: 55, Y: 55},
							{X: 55, Y: 25},
						},
					},
					domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 85, Y: 85},
							{X: 90, Y: 85},
							{X: 90, Y: 90},
							{X: 85, Y: 90},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid polygon",
			polygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 100, Y: 0},
						{X: 100, Y: 50},
						{X: 50, Y: -10},
					},
				},
			},
			wantErr: geometry.ErrInvalidPolygon,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizedPolygon, err := sfEngine.Normalize(tt.polygon)

			if tt.wantErr == nil {
				assert.NoError(t, err)

				assertPolygonInDelta(
					t,
					tt.wantPolygon,
					normalizedPolygon,
					epsilon,
				)

				return
			}

			assert.Zero(t, normalizedPolygon)

			assert.ErrorIs(
				t,
				err,
				tt.wantErr,
			)
		})
	}
}

func TestArea_BeforeAndAfterNormalize(t *testing.T) {
	// arrange
	sfEngine := NewEngine()

	polygon := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: []domaingeometry.Point{
				{X: 5, Y: 5},
				{X: 100, Y: 5},
				{X: 100, Y: 100},
				{X: 5, Y: 100},
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
			domaingeometry.Ring{
				Points: []domaingeometry.Point{
					{X: 90, Y: 90},
					{X: 95, Y: 90},
					{X: 95, Y: 95},
					{X: 90, Y: 95},
				},
			},
		},
	}

	// act
	normalizePolygon, err := sfEngine.Normalize(polygon)
	require.NoError(t, err)

	areaBefore, err := sfEngine.Area(polygon)
	require.NoError(t, err)

	areaAfter, err := sfEngine.Area(normalizePolygon)
	require.NoError(t, err)

	// assert
	assert.InDelta(
		t,
		areaAfter,
		areaBefore,
		epsilon,
	)
}

func TestScale(t *testing.T) {
	sfEngine := NewEngine()

	tests := []struct {
		name        string
		polygon     domaingeometry.Polygon
		factor      float64
		wantPolygon domaingeometry.Polygon
		wantErr     error
	}{
		{
			name: "factor 2",
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
			factor: 2,
			wantPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 200, Y: 0},
						{X: 200, Y: 100},
						{X: 0, Y: 100},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "factor 0.5",
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
			factor: 0.5,
			wantPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 50, Y: 0},
						{X: 50, Y: 25},
						{X: 0, Y: 25},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "factor 1",
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
			factor: 1,
			wantPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 100, Y: 0},
						{X: 100, Y: 50},
						{X: 0, Y: 50},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "factor causes coordinate overflow",
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
			factor:  math.MaxFloat64,
			wantErr: geometry.ErrInvalidScale,
		},
		{
			name: "non normalized polygon factor 2",
			polygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 5, Y: 5},
						{X: 100, Y: 5},
						{X: 100, Y: 50},
						{X: 5, Y: 50},
					},
				},
			},
			factor: 2,
			wantPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 190, Y: 0},
						{X: 190, Y: 90},
						{X: 0, Y: 90},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "polygon with hole factor 2",
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
			factor: 2,
			wantPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 200, Y: 0},
						{X: 200, Y: 200},
						{X: 0, Y: 200},
					},
				},
				Holes: []domaingeometry.Ring{
					domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 60, Y: 60},
							{X: 60, Y: 120},
							{X: 120, Y: 120},
							{X: 120, Y: 60},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid polygon",
			polygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 100, Y: 0},
						{X: 100, Y: 50},
						{X: 50, Y: -10},
					},
				},
			},
			factor:  2,
			wantErr: geometry.ErrInvalidPolygon,
		},
		{
			name: "factor 0",
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
			factor:  0,
			wantErr: geometry.ErrInvalidScale,
		},
		{
			name: "negative factor",
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
			factor:  -2,
			wantErr: geometry.ErrInvalidScale,
		},
		{
			name: "NaN factor",
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
			factor:  math.NaN(),
			wantErr: geometry.ErrInvalidScale,
		},
		{
			name: "Inf factor",
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
			factor:  math.Inf(1),
			wantErr: geometry.ErrInvalidScale,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scaledPolygon, err := sfEngine.Scale(tt.polygon, tt.factor)

			if tt.wantErr == nil {
				assert.NoError(t, err)

				assertPolygonInDelta(
					t,
					tt.wantPolygon,
					scaledPolygon,
					epsilon,
				)

				return
			}

			assert.Zero(t, scaledPolygon)

			assert.ErrorIs(
				t,
				err,
				tt.wantErr,
			)
		})
	}
}

func TestArea_BeforeAndAfterScale(t *testing.T) {
	// arrange
	sfEngine := NewEngine()

	polygon := domaingeometry.Polygon{
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
			domaingeometry.Ring{
				Points: []domaingeometry.Point{
					{X: 90, Y: 90},
					{X: 95, Y: 90},
					{X: 95, Y: 95},
					{X: 90, Y: 95},
				},
			},
		},
	}

	factor := 2.0

	scaledPolygon, err := sfEngine.Scale(polygon, factor)
	require.NoError(t, err)

	areaBefore, err := sfEngine.Area(polygon)
	require.NoError(t, err)

	areaAfter, err := sfEngine.Area(scaledPolygon)
	require.NoError(t, err)

	assert.InDelta(
		t,
		areaBefore*factor*factor,
		areaAfter,
		epsilon,
	)
}
