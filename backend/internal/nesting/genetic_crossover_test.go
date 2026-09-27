package nesting

import "testing"

func TestOrderCrossover(t *testing.T) {
	tests := []struct {
		name    string
		parentA Chromosome
		parentB Chromosome
		start   int
		end     int
		want    Chromosome
	}{
		{
			name: "middle segment",
			parentA: Chromosome{
				Genes: []Gene{
					{
						InstanceID: 0,
						Rotation:   0,
					},
					{
						InstanceID: 1,
						Rotation:   90,
					},
					{
						InstanceID: 2,
						Rotation:   180,
					},
					{
						InstanceID: 3,
						Rotation:   270,
					},
				},
			},
			parentB: Chromosome{
				Genes: []Gene{
					{
						InstanceID: 3,
						Rotation:   45,
					},
					{
						InstanceID: 2,
						Rotation:   135,
					},
					{
						InstanceID: 1,
						Rotation:   225,
					},
					{
						InstanceID: 0,
						Rotation:   315,
					},
				},
			},
			start: 1,
			end:   3,
			want: Chromosome{
				Genes: []Gene{
					{
						InstanceID: 3,
						Rotation:   45,
					},
					{
						InstanceID: 1,
						Rotation:   90,
					},
					{
						InstanceID: 2,
						Rotation:   180,
					},
					{
						InstanceID: 0,
						Rotation:   315,
					},
				},
			},
		},
		{
			name: "segment start at zero",
			parentA: Chromosome{
				Genes: []Gene{
					{
						InstanceID: 0,
						Rotation:   0,
					},
					{
						InstanceID: 1,
						Rotation:   90,
					},
					{
						InstanceID: 2,
						Rotation:   180,
					},
					{
						InstanceID: 3,
						Rotation:   270,
					},
				},
			},
			parentB: Chromosome{
				Genes: []Gene{
					{
						InstanceID: 3,
						Rotation:   45,
					},
					{
						InstanceID: 2,
						Rotation:   135,
					},
					{
						InstanceID: 1,
						Rotation:   225,
					},
					{
						InstanceID: 0,
						Rotation:   315,
					},
				},
			},
			start: 0,
			end:   2,
			want: Chromosome{
				Genes: []Gene{
					{
						InstanceID: 0,
						Rotation:   0,
					},
					{
						InstanceID: 1,
						Rotation:   90,
					},
					{
						InstanceID: 3,
						Rotation:   45,
					},
					{
						InstanceID: 2,
						Rotation:   135,
					},
				},
			},
		},
		{
			name: "segment ends at chromosome length",
			parentA: Chromosome{
				Genes: []Gene{
					{
						InstanceID: 0,
						Rotation:   0,
					},
					{
						InstanceID: 1,
						Rotation:   90,
					},
					{
						InstanceID: 2,
						Rotation:   180,
					},
					{
						InstanceID: 3,
						Rotation:   270,
					},
				},
			},
			parentB: Chromosome{
				Genes: []Gene{
					{
						InstanceID: 3,
						Rotation:   45,
					},
					{
						InstanceID: 2,
						Rotation:   135,
					},
					{
						InstanceID: 1,
						Rotation:   225,
					},
					{
						InstanceID: 0,
						Rotation:   315,
					},
				},
			},
			start: 2,
			end:   4,
			want: Chromosome{
				Genes: []Gene{
					{
						InstanceID: 1,
						Rotation:   225,
					},
					{
						InstanceID: 0,
						Rotation:   315,
					},
					{
						InstanceID: 2,
						Rotation:   180,
					},
					{
						InstanceID: 3,
						Rotation:   270,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := orderCrossover(
				tt.parentA,
				tt.parentB,
				tt.start,
				tt.end,
			)

			assertChromosome(
				t,
				tt.want,
				got,
			)
		})
	}
}

func TestOrderCrossoverNotChangeInputChromosome(t *testing.T) {
	// arrange
	parentA := Chromosome{
		Genes: []Gene{
			{
				InstanceID: 0,
				Rotation:   0,
			},
			{
				InstanceID: 1,
				Rotation:   90,
			},
			{
				InstanceID: 2,
				Rotation:   180,
			},
			{
				InstanceID: 3,
				Rotation:   270,
			},
		},
	}

	parentB := Chromosome{
		Genes: []Gene{
			{
				InstanceID: 3,
				Rotation:   45,
			},
			{
				InstanceID: 2,
				Rotation:   135,
			},
			{
				InstanceID: 1,
				Rotation:   225,
			},
			{
				InstanceID: 0,
				Rotation:   315,
			},
		},
	}

	originalParentA := Chromosome{
		Genes: make([]Gene, len(parentA.Genes)),
	}
	copy(originalParentA.Genes, parentA.Genes)

	originalParentB := Chromosome{
		Genes: make([]Gene, len(parentB.Genes)),
	}
	copy(originalParentB.Genes, parentB.Genes)

	// act
	orderCrossover(
		parentA,
		parentB,
		0,
		2,
	)

	// assert
	assertChromosome(
		t,
		originalParentA,
		parentA,
	)

	assertChromosome(
		t,
		originalParentB,
		parentB,
	)
}
