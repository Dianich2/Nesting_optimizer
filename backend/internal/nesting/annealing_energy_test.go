package nesting

import (
	"server_nesting_optimizer/internal/geometry/simplefeatures"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAnnealingRandom struct {
	floatValue float64
}

func (f fakeAnnealingRandom) Intn(
	n int,
) int {
	return 0
}

func (f fakeAnnealingRandom) Shuffle(
	n int,
	swap func(i int, j int),
) {
}

func (f fakeAnnealingRandom) Float64() float64 {
	return f.floatValue
}

func TestShouldAcceptWorse(t *testing.T) {
	tests := []struct {
		name            string
		currentEnergy   float64
		candidateEnergy float64
		temperature     float64
		randomValue     float64
		want            bool
	}{
		{
			name:            "better candidate always accepted",
			currentEnergy:   2,
			candidateEnergy: 1,
			temperature:     1,
			randomValue:     0.99,
			want:            true,
		},
		{
			name:            "equal candidate accepted",
			currentEnergy:   1,
			candidateEnergy: 1,
			temperature:     1,
			randomValue:     0.99,
			want:            true,
		},
		{
			name:            "worse candidate accepted by probability",
			currentEnergy:   1,
			candidateEnergy: 2,
			temperature:     1,
			randomValue:     0.3,
			want:            true,
		},
		{
			name:            "worse candidate rejected by probability",
			currentEnergy:   1,
			candidateEnergy: 2,
			temperature:     1,
			randomValue:     0.5,
			want:            false,
		},
		{
			name:            "worse candidate rejected at zero temperature",
			currentEnergy:   1,
			candidateEnergy: 2,
			temperature:     0,
			randomValue:     0,
			want:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rng := fakeAnnealingRandom{
				floatValue: tt.randomValue,
			}

			got := shouldAcceptWorse(
				tt.currentEnergy,
				tt.candidateEnergy,
				tt.temperature,
				rng,
			)

			assert.Equal(
				t,
				tt.want,
				got,
			)
		})
	}
}

func TestCalculateAnnealingEnergyPrefersMorePlacedArea(
	t *testing.T,
) {
	engine := simplefeatures.NewEngine()
	surface := rectangle(100, 100)

	smallerPlacedArea := Fitness{
		UnplacedCount: 1,
		PlacedArea:    3000,
		UsedArea:      5000,
		UsedHeight:    50,
	}

	largerPlacedArea := Fitness{
		UnplacedCount: 1,
		PlacedArea:    4000,
		UsedArea:      5000,
		UsedHeight:    50,
	}

	smallerEnergy, err := calculateAnnealingEnergy(
		engine,
		surface,
		smallerPlacedArea,
	)
	require.NoError(t, err)

	largerEnergy, err := calculateAnnealingEnergy(
		engine,
		surface,
		largerPlacedArea,
	)
	require.NoError(t, err)

	require.Less(
		t,
		largerEnergy,
		smallerEnergy,
	)
}
