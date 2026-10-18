package nesting

import (
	"testing"

	domaingeometry "server_nesting_optimizer/internal/domain/geometry"

	"github.com/stretchr/testify/require"
)

func TestAABBInteriorsMayOverlap(t *testing.T) {
	first := domaingeometry.Bounds{
		MinX: 10,
		MinY: 10,
		MaxX: 30,
		MaxY: 30,
	}

	tests := []struct {
		name   string
		second domaingeometry.Bounds
		want   bool
	}{
		{
			name: "separated horizontally",
			second: domaingeometry.Bounds{
				MinX: 40, MinY: 10,
				MaxX: 60, MaxY: 30,
			},
			want: false,
		},
		{
			name: "separated vertically",
			second: domaingeometry.Bounds{
				MinX: 10, MinY: 40,
				MaxX: 30, MaxY: 60,
			},
			want: false,
		},
		{
			name: "touching edge",
			second: domaingeometry.Bounds{
				MinX: 30, MinY: 10,
				MaxX: 50, MaxY: 30,
			},
			want: false,
		},
		{
			name: "touching corner",
			second: domaingeometry.Bounds{
				MinX: 30, MinY: 30,
				MaxX: 50, MaxY: 50,
			},
			want: false,
		},
		{
			name: "partially overlapping",
			second: domaingeometry.Bounds{
				MinX: 20, MinY: 20,
				MaxX: 40, MaxY: 40,
			},
			want: true,
		},
		{
			name: "fully contained",
			second: domaingeometry.Bounds{
				MinX: 15, MinY: 15,
				MaxX: 25, MaxY: 25,
			},
			want: true,
		},
		{
			name: "small positive overlap",
			second: domaingeometry.Bounds{
				MinX: 29.999999,
				MinY: 10,
				MaxX: 50,
				MaxY: 30,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(
				t,
				tt.want,
				aabbInteriorsMayOverlap(first, tt.second),
			)
		})
	}
}
