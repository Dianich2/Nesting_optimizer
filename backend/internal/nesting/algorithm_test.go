package nesting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlgorithmIsValid(t *testing.T) {
	tests := []struct {
		name      string
		algorithm Algorithm
		want      bool
	}{
		{
			name:      "baseline",
			algorithm: BaselineAlgorithm,
			want:      true,
		},
		{
			name:      "nfp greedy",
			algorithm: NFPGreedyAlgorithm,
			want:      true,
		},
		{
			name:      "genetic",
			algorithm: GeneticAlgorithm,
			want:      true,
		},
		{
			name:      "simulated annealing",
			algorithm: SimulatedAnnealingAlgorithm,
			want:      true,
		},
		{
			name:      "unknown",
			algorithm: Algorithm("unknown"),
			want:      false,
		},
		{
			name:      "empty",
			algorithm: Algorithm(""),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.algorithm.IsValid()

			assert.Equal(
				t,
				tt.want,
				got,
			)
		})
	}
}
