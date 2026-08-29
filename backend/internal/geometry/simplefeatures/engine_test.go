package simplefeatures

import (
	"math"
	"server_nesting_optimizer/internal/config"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
				require.NoError(t, err)
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
					config.Epsilon,
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
				require.NoError(t, err)

				assertPolygonInDelta(
					t,
					tt.wantPolygon,
					normalizedPolygon,
					config.Epsilon,
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
		config.Epsilon,
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
				require.NoError(t, err)

				assertPolygonInDelta(
					t,
					tt.wantPolygon,
					scaledPolygon,
					config.Epsilon,
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

	// act
	scaledPolygon, err := sfEngine.Scale(polygon, factor)
	require.NoError(t, err)

	areaBefore, err := sfEngine.Area(polygon)
	require.NoError(t, err)

	areaAfter, err := sfEngine.Area(scaledPolygon)
	require.NoError(t, err)

	// assert
	assert.InDelta(
		t,
		areaBefore*factor*factor,
		areaAfter,
		config.Epsilon,
	)
}

func TestTransform(t *testing.T) {
	sfEngine := NewEngine()

	tests := []struct {
		name        string
		polygon     domaingeometry.Polygon
		rotation    float64
		x           float64
		y           float64
		wantPolygon domaingeometry.Polygon
		wantErr     error
	}{
		{
			name: "rotation 0 and new center",
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
			rotation: 0,
			x:        55,
			y:        30,
			wantPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 5, Y: 5},
						{X: 105, Y: 5},
						{X: 105, Y: 55},
						{X: 5, Y: 55},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "rotation 90",
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
			rotation: 90,
			x:        50,
			y:        25,
			wantPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 75, Y: -25},
						{X: 75, Y: 75},
						{X: 25, Y: 75},
						{X: 25, Y: -25},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "rotation 450",
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
			rotation: 450,
			x:        50,
			y:        25,
			wantPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 75, Y: -25},
						{X: 75, Y: 75},
						{X: 25, Y: 75},
						{X: 25, Y: -25},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "rotation -90",
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
			rotation: -90,
			x:        50,
			y:        25,
			wantPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 25, Y: 75},
						{X: 25, Y: -25},
						{X: 75, Y: -25},
						{X: 75, Y: 75},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "polygon with holes",
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
			rotation: 90,
			x:        50,
			y:        50,
			wantPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 100, Y: 0},
						{X: 100, Y: 100},
						{X: 0, Y: 100},
						{X: 0, Y: 0},
					},
				},
				Holes: []domaingeometry.Ring{
					domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 70, Y: 30},
							{X: 40, Y: 30},
							{X: 40, Y: 60},
							{X: 70, Y: 60},
						},
					},
					domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{X: 10, Y: 90},
							{X: 10, Y: 95},
							{X: 5, Y: 95},
							{X: 5, Y: 90},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "non normalized polygon new center",
			polygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 100, Y: 100},
						{X: 200, Y: 100},
						{X: 200, Y: 150},
						{X: 100, Y: 150},
					},
				},
			},
			rotation: 0,
			x:        50,
			y:        25,
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
		},
		{
			name: "nan x",
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
			rotation: -90,
			x:        math.NaN(),
			y:        25,
			wantErr:  geometry.ErrInvalidTransform,
		},
		{
			name: "inf x",
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
			rotation: -90,
			x:        math.Inf(1),
			y:        25,
			wantErr:  geometry.ErrInvalidTransform,
		},
		{
			name: "nan y",
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
			rotation: -90,
			x:        25,
			y:        math.NaN(),
			wantErr:  geometry.ErrInvalidTransform,
		},
		{
			name: "inf y",
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
			rotation: -90,
			x:        25,
			y:        math.Inf(1),
			wantErr:  geometry.ErrInvalidTransform,
		},
		{
			name: "nan rotation",
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
			rotation: math.NaN(),
			x:        25,
			y:        25,
			wantErr:  geometry.ErrInvalidTransform,
		},
		{
			name: "inf rotation",
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
			rotation: math.Inf(1),
			x:        25,
			y:        25,
			wantErr:  geometry.ErrInvalidTransform,
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
			rotation: -90,
			x:        50,
			y:        25,
			wantErr:  geometry.ErrInvalidPolygon,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformedPolygon, err := sfEngine.Transform(
				tt.polygon,
				tt.x,
				tt.y,
				tt.rotation,
			)

			if tt.wantErr == nil {
				require.NoError(t, err)

				assertPolygonInDelta(
					t,
					tt.wantPolygon,
					transformedPolygon,
					config.Epsilon,
				)

				return
			}

			assert.Zero(t, transformedPolygon)

			assert.ErrorIs(
				t,
				err,
				tt.wantErr,
			)
		})
	}
}

func TestCoveredBy(t *testing.T) {
	sfEngine := NewEngine()

	tests := []struct {
		name          string
		innerPolygon  domaingeometry.Polygon
		outerPolygon  domaingeometry.Polygon
		wantIsCovered bool
		wantErr       error
	}{
		{
			name: "polygon completely inside",
			innerPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 20, Y: 20},
						{X: 50, Y: 20},
						{X: 50, Y: 50},
						{X: 20, Y: 50},
					},
				},
			},
			outerPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 100, Y: 0},
						{X: 100, Y: 50},
						{X: 0, Y: 50},
					},
				},
			},
			wantIsCovered: true,
			wantErr:       nil,
		},
		{
			name: "inner touches the boundary of outer",
			innerPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 5, Y: 0},
						{X: 50, Y: 0},
						{X: 50, Y: 20},
						{X: 5, Y: 20},
					},
				},
			},
			outerPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 100, Y: 0},
						{X: 100, Y: 50},
						{X: 0, Y: 50},
					},
				},
			},
			wantIsCovered: true,
			wantErr:       nil,
		},
		{
			name: "inner touches the boundary of outer one dot",
			innerPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 25},
						{X: 50, Y: 75},
						{X: 75, Y: 25},
						{X: 50, Y: 5},
					},
				},
			},
			outerPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 100, Y: 0},
						{X: 100, Y: 100},
						{X: 0, Y: 100},
					},
				},
			},
			wantIsCovered: true,
			wantErr:       nil,
		},
		{
			name: "polygon partially protrudes outwards",
			innerPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 25},
						{X: 150, Y: 75},
						{X: 75, Y: 25},
						{X: 150, Y: 5},
					},
				},
			},
			outerPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 100, Y: 0},
						{X: 100, Y: 100},
						{X: 0, Y: 100},
					},
				},
			},
			wantIsCovered: false,
			wantErr:       nil,
		},
		{
			name: "polygon completely outside",
			innerPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 200, Y: 200},
						{X: 300, Y: 200},
						{X: 300, Y: 300},
						{X: 200, Y: 300},
					},
				},
			},
			outerPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 100, Y: 0},
						{X: 100, Y: 100},
						{X: 0, Y: 100},
					},
				},
			},
			wantIsCovered: false,
			wantErr:       nil,
		},
		{
			name: "equals polygons",
			innerPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 25},
						{X: 150, Y: 75},
						{X: 75, Y: 25},
						{X: 150, Y: 5},
					},
				},
			},
			outerPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 25},
						{X: 150, Y: 75},
						{X: 75, Y: 25},
						{X: 150, Y: 5},
					},
				},
			},
			wantIsCovered: true,
			wantErr:       nil,
		},
		{
			name: "pattern completely inside outer hole",
			innerPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 35, Y: 35},
						{X: 45, Y: 35},
						{X: 45, Y: 45},
						{X: 35, Y: 45},
					},
				},
			},
			outerPolygon: domaingeometry.Polygon{
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
			wantIsCovered: false,
			wantErr:       nil,
		},
		{
			name: "pattern crosses outer hole",
			innerPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 35, Y: 35},
						{X: 75, Y: 35},
						{X: 75, Y: 75},
						{X: 35, Y: 75},
					},
				},
			},
			outerPolygon: domaingeometry.Polygon{
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
			wantIsCovered: false,
			wantErr:       nil,
		},
		{
			name: "invalid inner polygon",
			innerPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 25},
						{X: 0, Y: 75},
						{X: 0, Y: 125},
						{X: 0, Y: 175},
					},
				},
			},
			outerPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 100, Y: 0},
						{X: 100, Y: 100},
						{X: 0, Y: 100},
					},
				},
			},
			wantErr: geometry.ErrInvalidPolygon,
		},
		{
			name: "invalid outer polygon",
			innerPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 25},
						{X: 50, Y: 75},
						{X: 75, Y: 25},
						{X: 50, Y: 5},
					},
				},
			},
			outerPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 0, Y: 0},
						{X: 0, Y: 0},
						{X: 0, Y: 0},
					},
				},
			},
			wantErr: geometry.ErrInvalidPolygon,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isCoveredBy, err := sfEngine.CoveredBy(
				tt.innerPolygon,
				tt.outerPolygon,
			)

			if tt.wantErr == nil {
				require.NoError(t, err)

				assert.Equal(
					t,
					tt.wantIsCovered,
					isCoveredBy,
				)

				return
			}

			assert.Zero(t, isCoveredBy)

			assert.ErrorIs(
				t,
				err,
				tt.wantErr,
			)
		})
	}
}

func TestInteriorsIntersect(t *testing.T) {
	sfEngine := NewEngine()

	tests := []struct {
		name            string
		firstPolygon    domaingeometry.Polygon
		secondPolygon   domaingeometry.Polygon
		wantIsIntersect bool
		wantErr         error
	}{
		{
			name: "polygons are far apart",
			firstPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 20, Y: 0},
						{X: 20, Y: 20},
						{X: 0, Y: 20},
					},
				},
			},
			secondPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 100, Y: 100},
						{X: 120, Y: 100},
						{X: 120, Y: 120},
						{X: 100, Y: 120},
					},
				},
			},
			wantIsIntersect: false,
			wantErr:         nil,
		},
		{
			name: "boundaries touch along a side",
			firstPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 20, Y: 0},
						{X: 20, Y: 20},
						{X: 0, Y: 20},
					},
				},
			},
			secondPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 20, Y: 10},
						{X: 30, Y: 10},
						{X: 40, Y: 20},
						{X: 20, Y: 20},
					},
				},
			},
			wantIsIntersect: false,
			wantErr:         nil,
		},
		{
			name: "boundaries touch along a dot",
			firstPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 20, Y: 0},
						{X: 20, Y: 20},
						{X: 0, Y: 20},
					},
				},
			},
			secondPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 20, Y: 10},
						{X: 30, Y: 10},
						{X: 40, Y: 20},
						{X: 30, Y: 20},
					},
				},
			},
			wantIsIntersect: false,
			wantErr:         nil,
		},
		{
			name: "partially overlap",
			firstPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 20, Y: 0},
						{X: 20, Y: 20},
						{X: 0, Y: 20},
					},
				},
			},
			secondPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 15, Y: 15},
						{X: 30, Y: 15},
						{X: 30, Y: 30},
						{X: 15, Y: 30},
					},
				},
			},
			wantIsIntersect: true,
			wantErr:         nil,
		},
		{
			name: "one inside the other",
			firstPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 20, Y: 0},
						{X: 20, Y: 20},
						{X: 0, Y: 20},
					},
				},
			},
			secondPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 10, Y: 10},
						{X: 15, Y: 10},
						{X: 15, Y: 15},
						{X: 10, Y: 15},
					},
				},
			},
			wantIsIntersect: true,
			wantErr:         nil,
		},
		{
			name: "equals polygons",
			firstPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 20, Y: 0},
						{X: 20, Y: 20},
						{X: 0, Y: 20},
					},
				},
			},
			secondPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 20, Y: 0},
						{X: 20, Y: 20},
						{X: 0, Y: 20},
					},
				},
			},
			wantIsIntersect: true,
			wantErr:         nil,
		},
		{
			name: "first polygon inside second polygon hole",
			firstPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 35, Y: 35},
						{X: 45, Y: 35},
						{X: 45, Y: 45},
						{X: 35, Y: 45},
					},
				},
			},
			secondPolygon: domaingeometry.Polygon{
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
			wantIsIntersect: false,
			wantErr:         nil,
		},
		{
			name: "invalid first polygon",
			firstPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 0, Y: 10},
						{X: 0, Y: 20},
						{X: 0, Y: 30},
					},
				},
			},
			secondPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 100, Y: 100},
						{X: 120, Y: 100},
						{X: 120, Y: 120},
						{X: 100, Y: 120},
					},
				},
			},
			wantErr: geometry.ErrInvalidPolygon,
		},
		{
			name: "invalid second polygon",
			firstPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 100, Y: 100},
						{X: 120, Y: 100},
						{X: 120, Y: 120},
						{X: 100, Y: 120},
					},
				},
			},
			secondPolygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 0, Y: 10},
						{X: 0, Y: 20},
						{X: 0, Y: 30},
					},
				},
			},
			wantErr: geometry.ErrInvalidPolygon,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doInteriorsIntersect, err := sfEngine.InteriorsIntersect(
				tt.firstPolygon,
				tt.secondPolygon,
			)

			if tt.wantErr == nil {
				require.NoError(t, err)

				assert.Equal(
					t,
					tt.wantIsIntersect,
					doInteriorsIntersect,
				)

				return
			}

			assert.Zero(t, doInteriorsIntersect)

			assert.ErrorIs(
				t,
				err,
				tt.wantErr,
			)
		})
	}
}

func TestArea_BeforeAndAfterTransform(t *testing.T) {
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

	x := 100.0
	y := 100.0
	rotation := 90.0

	transformedPolygon, err := sfEngine.Transform(
		polygon,
		x,
		y,
		rotation,
	)
	require.NoError(t, err)

	areaBefore, err := sfEngine.Area(polygon)
	require.NoError(t, err)

	areaAfter, err := sfEngine.Area(transformedPolygon)
	require.NoError(t, err)

	assert.InDelta(
		t,
		areaBefore,
		areaAfter,
		config.Epsilon,
	)
}

func TestBounds(t *testing.T) {
	sfEngine := NewEngine()

	tests := []struct {
		name       string
		polygon    domaingeometry.Polygon
		wantBounds domaingeometry.Bounds
		wantErr    error
	}{
		{
			name: "rectangle polygon",
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
			wantBounds: domaingeometry.Bounds{
				MinX: 0,
				MinY: 0,
				MaxX: 20,
				MaxY: 20,
			},
			wantErr: nil,
		},
		{
			name: "polygon with negative coordinates",
			polygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: -20, Y: 0},
						{X: -20, Y: -20},
						{X: 0, Y: -20},
					},
				},
			},
			wantBounds: domaingeometry.Bounds{
				MinX: -20,
				MinY: -20,
				MaxX: 0,
				MaxY: 0,
			},
			wantErr: nil,
		},
		{
			name: "polygon with fractional coordinates",
			polygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 2.5, Y: 2.5},
						{X: 7.5, Y: 2.5},
						{X: 7.5, Y: 7.5},
						{X: 2.5, Y: 7.5},
					},
				},
			},
			wantBounds: domaingeometry.Bounds{
				MinX: 2.5,
				MinY: 2.5,
				MaxX: 7.5,
				MaxY: 7.5,
			},
			wantErr: nil,
		},
		{
			name: "polygon with holes",
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
			wantBounds: domaingeometry.Bounds{
				MinX: 0,
				MinY: 0,
				MaxX: 100,
				MaxY: 100,
			},
			wantErr: nil,
		},
		{
			name: "non-normalized polygon",
			polygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 2.5, Y: 2.5},
						{X: 7.5, Y: 2.5},
						{X: 7.5, Y: 7.5},
						{X: 2.5, Y: 7.5},
					},
				},
			},
			wantBounds: domaingeometry.Bounds{
				MinX: 2.5,
				MinY: 2.5,
				MaxX: 7.5,
				MaxY: 7.5,
			},
			wantErr: nil,
		},
		{
			name: "invalid polygon",
			polygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 0, Y: 10},
						{X: 0, Y: 20},
						{X: 0, Y: 30},
					},
				},
			},
			wantErr: geometry.ErrInvalidPolygon,
		},
		{
			name: "empty polygon",
			polygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 0, Y: 0},
						{X: 0, Y: 0},
						{X: 0, Y: 0},
					},
				},
			},
			wantErr: geometry.ErrInvalidPolygon,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			polygonBounds, err := sfEngine.Bounds(
				tt.polygon,
			)

			if tt.wantErr == nil {
				require.NoError(t, err)

				assert.Equal(
					t,
					tt.wantBounds,
					polygonBounds,
				)

				return
			}

			assert.Zero(t, polygonBounds)

			assert.ErrorIs(
				t,
				err,
				tt.wantErr,
			)
		})
	}
}
