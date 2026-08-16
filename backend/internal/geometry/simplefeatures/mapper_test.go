package simplefeatures

import (
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/stretchr/testify/assert"
)

func TestToSimpleFeaturesPolygon(t *testing.T) {
	tests := []struct {
		name    string
		polygon domaingeometry.Polygon
		want    string
	}{
		{
			name: "simple rectangle",
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
			want: "POLYGON((0 0,100 0,100 50,0 50,0 0))",
		},
		{
			name: "polygon with hole",
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
			want: "POLYGON((0 0,100 0,100 100,0 100,0 0),(30 30,30 60,60 60,60 30,30 30))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toSimpleFeaturesPolygon(tt.polygon)

			assert.Equal(
				t,
				tt.want,
				got.AsText(),
			)
		})
	}
}

func TestPolygonRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		polygon domaingeometry.Polygon
	}{
		{
			name: "simple rectangle",
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
			name: "polygon with hole",
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
			sfPolygon := toSimpleFeaturesPolygon(tt.polygon)
			got := fromSimpleFeaturesPolygon(sfPolygon)

			assert.Equal(
				t,
				tt.polygon,
				got,
			)
		})
	}
}

func TestFromSimpleFeaturesRing_UnclosedRing(t *testing.T) {
	// arrange
	ring := geom.NewLineStringXY(
		0, 0,
		100, 0,
		100, 50,
		0, 50,
	)

	wantRing := domaingeometry.Ring{
		Points: []domaingeometry.Point{
			{
				X: 0,
				Y: 0,
			},
			{
				X: 100,
				Y: 0,
			},
			{
				X: 100,
				Y: 50,
			},
			{
				X: 0,
				Y: 50,
			},
		},
	}

	// act
	got := fromSimpleFeaturesRing(ring)

	// assert
	assert.Equal(
		t,
		wantRing,
		got,
	)
}
