package nesting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMutateSwap(t *testing.T) {
	tests := []struct {
		name       string
		chromosome Chromosome
		firstIdx   int
		secondIdx  int
		want       Chromosome
	}{
		{
			name: "basic swap",
			chromosome: Chromosome{
				Genes: []Gene{
					{
						InstanceID: 0,
						PatternID:  1,
						Rotation:   0,
					},
					{
						InstanceID: 1,
						PatternID:  1,
						Rotation:   90,
					},
					{
						InstanceID: 2,
						PatternID:  2,
						Rotation:   0,
					},
					{
						InstanceID: 3,
						PatternID:  2,
						Rotation:   180,
					},
				},
			},
			firstIdx:  0,
			secondIdx: 2,
			want: Chromosome{
				Genes: []Gene{
					{
						InstanceID: 2,
						PatternID:  2,
						Rotation:   0,
					},
					{
						InstanceID: 1,
						PatternID:  1,
						Rotation:   90,
					},
					{
						InstanceID: 0,
						PatternID:  1,
						Rotation:   0,
					},

					{
						InstanceID: 3,
						PatternID:  2,
						Rotation:   180,
					},
				},
			},
		},
		{
			name: "firstIdx = secondIdx",
			chromosome: Chromosome{
				Genes: []Gene{
					{
						InstanceID: 0,
						PatternID:  1,
						Rotation:   0,
					},
					{
						InstanceID: 1,
						PatternID:  1,
						Rotation:   90,
					},
					{
						InstanceID: 2,
						PatternID:  2,
						Rotation:   0,
					},
					{
						InstanceID: 3,
						PatternID:  2,
						Rotation:   180,
					},
				},
			},
			firstIdx:  2,
			secondIdx: 2,
			want: Chromosome{
				Genes: []Gene{
					{
						InstanceID: 0,
						PatternID:  1,
						Rotation:   0,
					},
					{
						InstanceID: 1,
						PatternID:  1,
						Rotation:   90,
					},
					{
						InstanceID: 2,
						PatternID:  2,
						Rotation:   0,
					},
					{
						InstanceID: 3,
						PatternID:  2,
						Rotation:   180,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mutateSwap(
				tt.chromosome,
				tt.firstIdx,
				tt.secondIdx,
			)

			assertChromosome(
				t,
				tt.want,
				got,
			)
		})
	}
}

func TestMutateSwapNotMutateInputChromosome(t *testing.T) {
	// arrange
	chromosome := Chromosome{
		Genes: []Gene{
			{
				InstanceID: 0,
				PatternID:  1,
				Rotation:   0,
			},
			{
				InstanceID: 1,
				PatternID:  1,
				Rotation:   90,
			},
			{
				InstanceID: 2,
				PatternID:  2,
				Rotation:   0,
			},
			{
				InstanceID: 3,
				PatternID:  2,
				Rotation:   180,
			},
		},
	}

	originalChromosome := Chromosome{
		Genes: make([]Gene, len(chromosome.Genes)),
	}
	copy(originalChromosome.Genes, chromosome.Genes)

	// act
	mutateSwap(
		chromosome,
		0,
		2,
	)

	// assert
	assert.Equal(
		t,
		originalChromosome,
		chromosome,
	)
}

func TestMutateRotation(t *testing.T) {
	tests := []struct {
		name       string
		chromosome Chromosome
		geneIdx    int
		rotation   float64
		want       Chromosome
	}{
		{
			name: "basic rotation",
			chromosome: Chromosome{
				Genes: []Gene{
					{
						InstanceID: 0,
						PatternID:  1,
						Rotation:   0,
					},
					{
						InstanceID: 1,
						PatternID:  1,
						Rotation:   90,
					},
					{
						InstanceID: 2,
						PatternID:  2,
						Rotation:   0,
					},
					{
						InstanceID: 3,
						PatternID:  2,
						Rotation:   180,
					},
				},
			},
			geneIdx:  2,
			rotation: 125,
			want: Chromosome{
				Genes: []Gene{
					{
						InstanceID: 0,
						PatternID:  1,
						Rotation:   0,
					},
					{
						InstanceID: 1,
						PatternID:  1,
						Rotation:   90,
					},
					{
						InstanceID: 2,
						PatternID:  2,
						Rotation:   125,
					},
					{
						InstanceID: 3,
						PatternID:  2,
						Rotation:   180,
					},
				},
			},
		},
		{
			name: "rotation = previous rotation",
			chromosome: Chromosome{
				Genes: []Gene{
					{
						InstanceID: 0,
						PatternID:  1,
						Rotation:   0,
					},
					{
						InstanceID: 1,
						PatternID:  1,
						Rotation:   90,
					},
					{
						InstanceID: 2,
						PatternID:  2,
						Rotation:   0,
					},
					{
						InstanceID: 3,
						PatternID:  2,
						Rotation:   180,
					},
				},
			},
			geneIdx:  2,
			rotation: 0,
			want: Chromosome{
				Genes: []Gene{
					{
						InstanceID: 0,
						PatternID:  1,
						Rotation:   0,
					},
					{
						InstanceID: 1,
						PatternID:  1,
						Rotation:   90,
					},
					{
						InstanceID: 2,
						PatternID:  2,
						Rotation:   0,
					},
					{
						InstanceID: 3,
						PatternID:  2,
						Rotation:   180,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mutateRotation(
				tt.chromosome,
				tt.geneIdx,
				tt.rotation,
			)

			assertChromosome(
				t,
				tt.want,
				got,
			)
		})
	}
}

func TestMutateRotationNotMutateInputChromosome(t *testing.T) {
	// arrange
	chromosome := Chromosome{
		Genes: []Gene{
			{
				InstanceID: 0,
				PatternID:  1,
				Rotation:   0,
			},
			{
				InstanceID: 1,
				PatternID:  1,
				Rotation:   90,
			},
			{
				InstanceID: 2,
				PatternID:  2,
				Rotation:   0,
			},
			{
				InstanceID: 3,
				PatternID:  2,
				Rotation:   180,
			},
		},
	}

	originalChromosome := Chromosome{
		Genes: make([]Gene, len(chromosome.Genes)),
	}
	copy(originalChromosome.Genes, chromosome.Genes)

	// act
	mutateRotation(
		chromosome,
		0,
		90,
	)

	// assert
	assert.Equal(
		t,
		originalChromosome,
		chromosome,
	)
}
