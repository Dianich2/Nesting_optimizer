package nesting

import (
	"server_nesting_optimizer/internal/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertFitness(
	t *testing.T,
	want Fitness,
	got Fitness,
	delta float64,
) {
	t.Helper()

	assert.Equal(t, want.UnplacedCount, got.UnplacedCount)
	assert.InDelta(t, want.UsedArea, got.UsedArea, delta)
	assert.InDelta(t, want.UsedHeight, got.UsedHeight, delta)
}

func TestTournamentSelection(t *testing.T) {
	tests := []struct {
		name           string
		population     []EvaluatedChromosome
		tournamentSize int
		randomIndexes  []int
		want           EvaluatedChromosome
	}{
		{
			name: "best by unplaced_count not first",
			population: []EvaluatedChromosome{
				{
					Chromosome: Chromosome{
						Genes: []Gene{
							{
								InstanceID: 0,
							},
						},
					},
					Fitness: Fitness{
						UnplacedCount: 2,
						UsedArea:      2000,
						UsedHeight:    100,
					},
				},
				{
					Chromosome: Chromosome{
						Genes: []Gene{
							{
								InstanceID: 1,
							},
						},
					},
					Fitness: Fitness{
						UnplacedCount: 4,
						UsedArea:      1600,
						UsedHeight:    80,
					},
				},
				{
					Chromosome: Chromosome{
						Genes: []Gene{
							{
								InstanceID: 2,
							},
						},
					},
					Fitness: Fitness{
						UnplacedCount: 1,
						UsedArea:      2500,
						UsedHeight:    120,
					},
				},
			},
			tournamentSize: 3,
			randomIndexes:  []int{1, 0, 2},
			want: EvaluatedChromosome{
				Chromosome: Chromosome{
					Genes: []Gene{
						{
							InstanceID: 2,
						},
					},
				},
				Fitness: Fitness{
					UnplacedCount: 1,
					UsedArea:      2500,
					UsedHeight:    120,
				},
			},
		},
		{
			name: "best by used_area",
			population: []EvaluatedChromosome{
				{
					Chromosome: Chromosome{
						Genes: []Gene{
							{
								InstanceID: 0,
							},
						},
					},
					Fitness: Fitness{
						UnplacedCount: 2,
						UsedArea:      2000,
						UsedHeight:    100,
					},
				},
				{
					Chromosome: Chromosome{
						Genes: []Gene{
							{
								InstanceID: 1,
							},
						},
					},
					Fitness: Fitness{
						UnplacedCount: 2,
						UsedArea:      2200,
						UsedHeight:    180,
					},
				},
				{
					Chromosome: Chromosome{
						Genes: []Gene{
							{
								InstanceID: 2,
							},
						},
					},
					Fitness: Fitness{
						UnplacedCount: 2,
						UsedArea:      1700,
						UsedHeight:    120,
					},
				},
			},
			tournamentSize: 3,
			randomIndexes:  []int{1, 0, 2},
			want: EvaluatedChromosome{
				Chromosome: Chromosome{
					Genes: []Gene{
						{
							InstanceID: 2,
						},
					},
				},
				Fitness: Fitness{
					UnplacedCount: 2,
					UsedArea:      1700,
					UsedHeight:    120,
				},
			},
		},
		{
			name: "best by used_height",
			population: []EvaluatedChromosome{
				{
					Chromosome: Chromosome{
						Genes: []Gene{
							{
								InstanceID: 0,
							},
						},
					},
					Fitness: Fitness{
						UnplacedCount: 2,
						UsedArea:      2000,
						UsedHeight:    100,
					},
				},
				{
					Chromosome: Chromosome{
						Genes: []Gene{
							{
								InstanceID: 1,
							},
						},
					},
					Fitness: Fitness{
						UnplacedCount: 2,
						UsedArea:      2000,
						UsedHeight:    180,
					},
				},
				{
					Chromosome: Chromosome{
						Genes: []Gene{
							{
								InstanceID: 2,
							},
						},
					},
					Fitness: Fitness{
						UnplacedCount: 2,
						UsedArea:      2000,
						UsedHeight:    120,
					},
				},
			},
			tournamentSize: 3,
			randomIndexes:  []int{2, 0, 1},
			want: EvaluatedChromosome{
				Chromosome: Chromosome{
					Genes: []Gene{
						{
							InstanceID: 0,
						},
					},
				},
				Fitness: Fitness{
					UnplacedCount: 2,
					UsedArea:      2000,
					UsedHeight:    100,
				},
			},
		},
		{
			name: "best already first",
			population: []EvaluatedChromosome{
				{
					Chromosome: Chromosome{
						Genes: []Gene{
							{
								InstanceID: 0,
							},
						},
					},
					Fitness: Fitness{
						UnplacedCount: 2,
						UsedArea:      2000,
						UsedHeight:    100,
					},
				},
				{
					Chromosome: Chromosome{
						Genes: []Gene{
							{
								InstanceID: 1,
							},
						},
					},
					Fitness: Fitness{
						UnplacedCount: 4,
						UsedArea:      1600,
						UsedHeight:    80,
					},
				},
				{
					Chromosome: Chromosome{
						Genes: []Gene{
							{
								InstanceID: 2,
							},
						},
					},
					Fitness: Fitness{
						UnplacedCount: 1,
						UsedArea:      1700,
						UsedHeight:    120,
					},
				},
			},
			tournamentSize: 3,
			randomIndexes:  []int{2, 0, 1},
			want: EvaluatedChromosome{
				Chromosome: Chromosome{
					Genes: []Gene{
						{
							InstanceID: 2,
						},
					},
				},
				Fitness: Fitness{
					UnplacedCount: 1,
					UsedArea:      1700,
					UsedHeight:    120,
				},
			},
		},
		{
			name: "tournamentSize = 1",
			population: []EvaluatedChromosome{
				{
					Chromosome: Chromosome{
						Genes: []Gene{
							{
								InstanceID: 0,
							},
						},
					},
					Fitness: Fitness{
						UnplacedCount: 2,
						UsedArea:      2000,
						UsedHeight:    100,
					},
				},
				{
					Chromosome: Chromosome{
						Genes: []Gene{
							{
								InstanceID: 1,
							},
						},
					},
					Fitness: Fitness{
						UnplacedCount: 4,
						UsedArea:      1600,
						UsedHeight:    80,
					},
				},
				{
					Chromosome: Chromosome{
						Genes: []Gene{
							{
								InstanceID: 2,
							},
						},
					},
					Fitness: Fitness{
						UnplacedCount: 1,
						UsedArea:      1700,
						UsedHeight:    120,
					},
				},
			},
			tournamentSize: 1,
			randomIndexes:  []int{1, 0, 2},
			want: EvaluatedChromosome{
				Chromosome: Chromosome{
					Genes: []Gene{
						{
							InstanceID: 1,
						},
					},
				},
				Fitness: Fitness{
					UnplacedCount: 4,
					UsedArea:      1600,
					UsedHeight:    80,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rng := newFakeRandom(tt.randomIndexes...)

			got := tournamentSelection(
				tt.population,
				tt.tournamentSize,
				rng,
			)

			assertChromosome(
				t,
				tt.want.Chromosome,
				got.Chromosome,
			)

			assertFitness(
				t,
				tt.want.Fitness,
				got.Fitness,
				config.Epsilon,
			)

		})
	}
}
