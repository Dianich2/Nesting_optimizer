package geometry

import (
	"server_nesting_optimizer/internal/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeDegrees(t *testing.T) {
	tests := []struct {
		name               string
		angle              float64
		wantNormalizeAngle float64
	}{
		{
			name:               "positive angle already normalized",
			angle:              20,
			wantNormalizeAngle: 20,
		},
		{
			name:               "negative angle within one turn",
			angle:              -90,
			wantNormalizeAngle: 270,
		},
		{
			name:               "exact positive full turn",
			angle:              360,
			wantNormalizeAngle: 0,
		},
		{
			name:               "positive angle over one turn",
			angle:              450,
			wantNormalizeAngle: 90,
		},
		{
			name:               "fractional angle already normalized",
			angle:              100.22,
			wantNormalizeAngle: 100.22,
		},
		{
			name:               "fractional angle over one turn",
			angle:              450.22,
			wantNormalizeAngle: 90.22,
		},
		{
			name:               "fractional negative angle",
			angle:              -90.22,
			wantNormalizeAngle: 269.78,
		},
		{
			name:               "zero angle",
			angle:              0,
			wantNormalizeAngle: 0,
		},
		{
			name:               "exact negative full turn",
			angle:              -360,
			wantNormalizeAngle: 0,
		},
		{
			name:               "positive angle over two turns",
			angle:              810,
			wantNormalizeAngle: 90,
		},
		{
			name:               "negative angle below one turn",
			angle:              -450,
			wantNormalizeAngle: 270,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			assert.InDelta(
				t,
				tt.wantNormalizeAngle,
				NormalizeDegrees(tt.angle),
				config.Epsilon,
			)

		})
	}
}
