package nesting

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultRandomHasIndependentState(t *testing.T) {
	first, ok := newDefaultRandom().(*defaultRandom)
	require.True(t, ok)

	second, ok := newDefaultRandom().(*defaultRandom)
	require.True(t, ok)

	require.NotSame(t, first, second)
	require.NotSame(t, first.source, second.source)

	// Set fixed seeds to test isolation deterministically.
	first.source.Seed(42)
	second.source.Seed(42)

	expectedFirst := rand.New(rand.NewSource(42))
	expectedSecond := rand.New(rand.NewSource(42))

	// Interleave calls: each generator must follow
	// its own independent sequence.
	for i := 0; i < 100; i++ {
		require.Equal(
			t,
			expectedFirst.Intn(1000),
			first.Intn(1000),
		)

		require.Equal(
			t,
			expectedSecond.Intn(1000),
			second.Intn(1000),
		)
	}
}

func TestAnnealingRandomHasIndependentState(t *testing.T) {
	first, ok := newDefaultAnnealingRandom().(*defaultRandom)
	require.True(t, ok)

	second, ok := newDefaultAnnealingRandom().(*defaultRandom)
	require.True(t, ok)

	require.NotSame(t, first.source, second.source)

	first.source.Seed(123)
	second.source.Seed(123)

	expectedFirst := rand.New(rand.NewSource(123))
	expectedSecond := rand.New(rand.NewSource(123))

	for i := 0; i < 100; i++ {
		require.Equal(
			t,
			expectedFirst.Float64(),
			first.Float64(),
		)

		require.Equal(
			t,
			expectedSecond.Intn(1000),
			second.Intn(1000),
		)
	}
}
