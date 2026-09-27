package nesting

import (
	"server_nesting_optimizer/internal/config"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRandom struct {
	intnValues []int
	intnIndex  int
}

func newFakeRandom(intnValues ...int) *fakeRandom {
	return &fakeRandom{
		intnValues: intnValues,
		intnIndex:  0,
	}
}

func (f *fakeRandom) Shuffle(
	n int,
	swap func(i int, j int),
) {
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}

func (f *fakeRandom) Intn(n int) int {
	curInt := f.intnValues[f.intnIndex]
	f.intnIndex++
	return curInt
}

func assertChromosome(
	t *testing.T,
	want Chromosome,
	got Chromosome,
) {
	t.Helper()

	require.Len(t, got.Genes, len(want.Genes))

	for i := range want.Genes {
		assertGenes(
			t,
			want.Genes[i],
			got.Genes[i],
			config.Epsilon,
		)
	}
}

func TestGenerateInitialPopulationBasicScenario(t *testing.T) {
	// arrange
	genes := []Gene{
		{
			InstanceID: 0,
			PatternID:  10,
		},
		{
			InstanceID: 1,
			PatternID:  20,
		},
		{
			InstanceID: 2,
			PatternID:  30,
		},
	}

	allowedRotations := []float64{0, 90, 180}
	populationSize := 1

	wantPopulation := []Chromosome{
		{
			Genes: []Gene{
				{
					InstanceID: 2,
					PatternID:  30,
					Rotation:   180,
				},
				{
					InstanceID: 1,
					PatternID:  20,
					Rotation:   0,
				},
				{
					InstanceID: 0,
					PatternID:  10,
					Rotation:   90,
				},
			},
		},
	}

	rng := newFakeRandom(2, 0, 1)

	// act
	initialPopulation := generateInitialPopulation(
		genes,
		allowedRotations,
		populationSize,
		rng,
	)

	// assert
	assert.Len(
		t,
		initialPopulation,
		len(wantPopulation),
	)

	for i := range wantPopulation {
		assertChromosome(
			t,
			wantPopulation[i],
			initialPopulation[i],
		)
	}
}

func TestGenerateInitialPopulationCreatesIndependentChromosomes(t *testing.T) {
	genes := []Gene{
		{
			InstanceID: 0,
			PatternID:  10,
		},
		{
			InstanceID: 1,
			PatternID:  20,
		},
		{
			InstanceID: 2,
			PatternID:  30,
		},
	}

	allowedRotations := []float64{0, 90, 180}

	rng := newFakeRandom(
		2, 0, 1,
		1, 2, 0,
	)

	populationSize := 2

	// act
	population := generateInitialPopulation(
		genes,
		allowedRotations,
		populationSize,
		rng,
	)

	require.Len(
		t,
		population,
		populationSize,
	)

	population[0].Genes[0].Rotation = 999

	assert.NotEqual(
		t,
		float64(999),
		population[1].Genes[0].Rotation,
	)
}

func TestGenerateInitialPopulationDoesNotMutateInputGenes(t *testing.T) {
	genes := []Gene{
		{
			InstanceID: 0,
			PatternID:  10,
		},
		{
			InstanceID: 1,
			PatternID:  20,
		},
		{
			InstanceID: 2,
			PatternID:  30,
		},
	}

	originalGenes := make([]Gene, len(genes))
	copy(originalGenes, genes)

	allowedRotations := []float64{0, 90, 180}

	rng := newFakeRandom(
		2, 0, 1,
	)

	populationSize := 1

	// act
	generateInitialPopulation(
		genes,
		allowedRotations,
		populationSize,
		rng,
	)

	assert.Equal(
		t,
		originalGenes,
		genes,
	)
}
