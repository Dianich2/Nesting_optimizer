package nesting

import "math/rand"

type Random interface {
	Intn(n int) int
	Shuffle(n int, swap func(i, j int))
}

type AnnealingRandom interface {
	Random
	Float64() float64
}

type defaultRandom struct {
	source *rand.Rand
}

func newDefaultRandom() Random {
	return &defaultRandom{
		source: rand.New(
			rand.NewSource(rand.Int63()),
		),
	}
}

func newDefaultAnnealingRandom() AnnealingRandom {
	return &defaultRandom{
		source: rand.New(
			rand.NewSource(rand.Int63()),
		),
	}
}

func (r *defaultRandom) Intn(
	n int,
) int {
	return r.source.Intn(n)
}

func (r *defaultRandom) Shuffle(
	n int,
	swap func(i, j int),
) {
	r.source.Shuffle(n, swap)
}

func (r *defaultRandom) Float64() float64 {
	return r.source.Float64()
}
