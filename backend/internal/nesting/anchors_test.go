package nesting

import (
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollectAnchors(t *testing.T) {
	tests := []struct {
		name        string
		surface     domaingeometry.Polygon
		occupied    []domaingeometry.Polygon
		wantAnchors []domaingeometry.Point
	}{
		{
			name: "surface only",
			surface: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 20, Y: 0},
						{X: 20, Y: 20},
						{X: 0, Y: 20},
					},
				},
			},
			occupied: nil,
			wantAnchors: []domaingeometry.Point{
				{
					X: 0,
					Y: 0,
				},
				{
					X: 20,
					Y: 0,
				},
				{
					X: 20,
					Y: 20,
				},
				{
					X: 0,
					Y: 20,
				},
			},
		},
		{
			name: "surface + hole",
			surface: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 20, Y: 0},
						{X: 20, Y: 20},
						{X: 0, Y: 20},
						{X: 0, Y: 0},
					},
				},
				Holes: []domaingeometry.Ring{
					{
						Points: []domaingeometry.Point{
							{
								X: 5,
								Y: 5,
							},
							{
								X: 15,
								Y: 5,
							},
							{
								X: 15,
								Y: 15,
							},
							{
								X: 5,
								Y: 15,
							},
						},
					},
				},
			},
			occupied: nil,
			wantAnchors: []domaingeometry.Point{
				{
					X: 0,
					Y: 0,
				},
				{
					X: 20,
					Y: 0,
				},
				{
					X: 20,
					Y: 20,
				},
				{
					X: 0,
					Y: 20,
				},
				{
					X: 5,
					Y: 5,
				},
				{
					X: 15,
					Y: 5,
				},
				{
					X: 15,
					Y: 15,
				},
				{
					X: 5,
					Y: 15,
				},
			},
		},
		{
			name: "surface + one occupied polygon",
			surface: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 20, Y: 0},
						{X: 20, Y: 20},
						{X: 0, Y: 20},
						{X: 0, Y: 0},
					},
				},
			},
			occupied: []domaingeometry.Polygon{
				{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{
								X: 5,
								Y: 5,
							},
							{
								X: 15,
								Y: 5,
							},
							{
								X: 15,
								Y: 15,
							},
							{
								X: 5,
								Y: 15,
							},
						},
					},
				},
			},
			wantAnchors: []domaingeometry.Point{
				{
					X: 0,
					Y: 0,
				},
				{
					X: 20,
					Y: 0,
				},
				{
					X: 20,
					Y: 20,
				},
				{
					X: 0,
					Y: 20,
				},
				{
					X: 5,
					Y: 5,
				},
				{
					X: 15,
					Y: 5,
				},
				{
					X: 15,
					Y: 15,
				},
				{
					X: 5,
					Y: 15,
				},
			},
		},
		{
			name: "surface + several occupied polygons",
			surface: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 20, Y: 0},
						{X: 20, Y: 20},
						{X: 0, Y: 20},
						{X: 0, Y: 0},
					},
				},
			},
			occupied: []domaingeometry.Polygon{
				{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{
								X: 5,
								Y: 5,
							},
							{
								X: 15,
								Y: 5,
							},
							{
								X: 15,
								Y: 15,
							},
							{
								X: 5,
								Y: 15,
							},
						},
					},
				},
				{
					Exterior: domaingeometry.Ring{
						Points: []domaingeometry.Point{
							{
								X: 16,
								Y: 16,
							},
							{
								X: 19,
								Y: 16,
							},
							{
								X: 19,
								Y: 19,
							},
						},
					},
				},
			},
			wantAnchors: []domaingeometry.Point{
				{
					X: 0,
					Y: 0,
				},
				{
					X: 20,
					Y: 0,
				},
				{
					X: 20,
					Y: 20,
				},
				{
					X: 0,
					Y: 20,
				},
				{
					X: 5,
					Y: 5,
				},
				{
					X: 15,
					Y: 5,
				},
				{
					X: 15,
					Y: 15,
				},
				{
					X: 5,
					Y: 15,
				},
				{
					X: 16,
					Y: 16,
				},
				{
					X: 19,
					Y: 16,
				},
				{
					X: 19,
					Y: 19,
				},
			},
		},
		{
			name: "empty geometry",
			surface: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: nil,
				},
			},
			occupied:    nil,
			wantAnchors: nil,
		},
		{
			name: "closed rings don't duplicate closing points",
			surface: domaingeometry.Polygon{
				Exterior: domaingeometry.Ring{
					Points: []domaingeometry.Point{
						{X: 0, Y: 0},
						{X: 20, Y: 0},
						{X: 20, Y: 20},
						{X: 0, Y: 20},
						{X: 0, Y: 0},
					},
				},
			},
			occupied: nil,
			wantAnchors: []domaingeometry.Point{
				{
					X: 0,
					Y: 0,
				},
				{
					X: 20,
					Y: 0,
				},
				{
					X: 20,
					Y: 20,
				},
				{
					X: 0,
					Y: 20,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			assert.Equal(
				t,
				tt.wantAnchors,
				collectAnchors(
					tt.surface,
					tt.occupied,
				),
			)

		})
	}
}
