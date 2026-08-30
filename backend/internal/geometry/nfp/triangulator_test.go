package nfp

import (
	"math"
	"server_nesting_optimizer/internal/config"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"testing"
)

func ringArea(
	points []domaingeometry.Point,
) float64 {
	var twiceArea float64

	for i := range points {
		next := (i + 1) % len(points)

		twiceArea +=
			points[i].X*points[next].Y -
				points[next].X*points[i].Y
	}

	return math.Abs(twiceArea) / 2
}

func TestTriangulator(t *testing.T) {
	tests := []struct {
		name              string
		polygon           domaingeometry.Polygon
		wantTriangleCount int
		wantArea          float64
	}{
		{
			name: "triangle",
			polygon: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 100, Y: 0},
						{X: 50, Y: 50},
					},
				},
			},
			wantTriangleCount: 1,
			wantArea:          2500,
		},
		{
			name: "rectangle",
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
			wantTriangleCount: 2,
			wantArea:          5000,
		},
		{
			name: "concave L shape",
			polygon: domaingeometry.Polygon{
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
			wantTriangleCount: 4,
			wantArea:          7500,
		},
		{
			name: "rectangle with rectangular hole",
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
			wantTriangleCount: 8,
			wantArea:          7500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			triangles, err := triangulate(tt.polygon)
			if err != nil {
				t.Fatalf(
					"triangulate() unexpected error: %v",
					err,
				)
			}

			if len(triangles) != tt.wantTriangleCount {
				t.Fatalf(
					"triangulate() returned %d triangles, want %d",
					len(triangles),
					tt.wantTriangleCount,
				)
			}

			var totalArea float64

			for i, triangle := range triangles {
				if len(triangle.Exterior.Points) != 3 {
					t.Fatalf(
						"triangle %d has %d points, want 3",
						i,
						len(triangle.Exterior.Points),
					)
				}

				if len(triangle.Holes) != 0 {
					t.Fatalf(
						"triangle %d contains holes",
						i,
					)
				}

				area := ringArea(
					triangle.Exterior.Points,
				)

				if area <= 0 {
					t.Fatalf(
						"triangle %d has invalid area: %v",
						i,
						area,
					)
				}

				totalArea += area
			}

			if math.Abs(totalArea-tt.wantArea) > config.TriangulationDeviationTolerance {
				t.Fatalf(
					"triangles area = %v, want %v",
					totalArea,
					tt.wantArea,
				)
			}
		})
	}
}

func TestPrepareTriangulationInputWithHoles(t *testing.T) {
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
			{
				Points: []domaingeometry.Point{
					{X: 10, Y: 10},
					{X: 20, Y: 10},
					{X: 20, Y: 20},
					{X: 10, Y: 20},
				},
			},
			{
				Points: []domaingeometry.Point{
					{X: 50, Y: 50},
					{X: 60, Y: 50},
					{X: 55, Y: 60},
				},
			},
		},
	}

	got := prepareTriangulationInput(polygon)

	if len(got.Points) != 11 {
		t.Fatalf(
			"points count = %d, want 11",
			len(got.Points),
		)
	}

	if len(got.Vertices) != 22 {
		t.Fatalf(
			"vertices count = %d, want 22",
			len(got.Vertices),
		)
	}

	if len(got.HoleIndices) != 2 {
		t.Fatalf(
			"hole indices count = %d, want 2",
			len(got.HoleIndices),
		)
	}

	if got.HoleIndices[0] != 4 {
		t.Errorf(
			"first hole index = %d, want 4",
			got.HoleIndices[0],
		)
	}

	if got.HoleIndices[1] != 8 {
		t.Errorf(
			"second hole index = %d, want 8",
			got.HoleIndices[1],
		)
	}
}
