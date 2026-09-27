package nesting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsFitnessBetter(t *testing.T) {
	tests := []struct {
		name      string
		candidate Fitness
		best      Fitness
		want      bool
	}{
		{
			name: "candidate has less unplaced_count than best",
			candidate: Fitness{
				UnplacedCount: 2,
				UsedArea:      2000,
				UsedHeight:    100,
			},
			best: Fitness{
				UnplacedCount: 3,
				UsedArea:      1800,
				UsedHeight:    120,
			},
			want: true,
		},
		{
			name: "candidate has more unplaced_count than best",
			candidate: Fitness{
				UnplacedCount: 3,
				UsedArea:      1800,
				UsedHeight:    120,
			},
			best: Fitness{
				UnplacedCount: 2,
				UsedArea:      2000,
				UsedHeight:    100,
			},
			want: false,
		},
		{
			name: "candidate has less used_area than best",
			candidate: Fitness{
				UnplacedCount: 3,
				UsedArea:      1800,
				UsedHeight:    120,
			},
			best: Fitness{
				UnplacedCount: 3,
				UsedArea:      2000,
				UsedHeight:    100,
			},
			want: true,
		},
		{
			name: "candidate has more used_area than best",
			candidate: Fitness{
				UnplacedCount: 3,
				UsedArea:      2100,
				UsedHeight:    120,
			},
			best: Fitness{
				UnplacedCount: 3,
				UsedArea:      2000,
				UsedHeight:    100,
			},
			want: false,
		},
		{
			name: "candidate has less used_height than best",
			candidate: Fitness{
				UnplacedCount: 3,
				UsedArea:      1800,
				UsedHeight:    80,
			},
			best: Fitness{
				UnplacedCount: 3,
				UsedArea:      1800,
				UsedHeight:    100,
			},
			want: true,
		},
		{
			name: "candidate has more used_height than best",
			candidate: Fitness{
				UnplacedCount: 3,
				UsedArea:      1800,
				UsedHeight:    120,
			},
			best: Fitness{
				UnplacedCount: 3,
				UsedArea:      1800,
				UsedHeight:    100,
			},
			want: false,
		},
		{
			name: "equals",
			candidate: Fitness{
				UnplacedCount: 3,
				UsedArea:      1800,
				UsedHeight:    100,
			},
			best: Fitness{
				UnplacedCount: 3,
				UsedArea:      1800,
				UsedHeight:    100,
			},
			want: false,
		},
		{
			name: "almost equals by used_area",
			candidate: Fitness{
				UnplacedCount: 3,
				UsedArea:      1800.0000000890,
				UsedHeight:    100,
			},
			best: Fitness{
				UnplacedCount: 3,
				UsedArea:      1800.0000000899,
				UsedHeight:    120,
			},
			want: true,
		},
		{
			name: "almost equals by used_height",
			candidate: Fitness{
				UnplacedCount: 3,
				UsedArea:      1800.0000000890,
				UsedHeight:    100.00000000089,
			},
			best: Fitness{
				UnplacedCount: 3,
				UsedArea:      1800.0000000899,
				UsedHeight:    100.00000000088,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isFitnessBetter(
				tt.candidate,
				tt.best,
			)

			assert.Equal(
				t,
				tt.want,
				got,
			)
		})
	}
}
