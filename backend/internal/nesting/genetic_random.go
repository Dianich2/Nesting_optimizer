package nesting

import "math/rand"

type Random interface {
	Intn(n int) int
	Shuffle(n int, swap func(i int, j int))
}

type defaultRandom struct{}

func newDefaultRandom() Random {
	return defaultRandom{}
}

func (defaultRandom) Intn(
	n int,
) int {
	return rand.Intn(n)
}

func (defaultRandom) Shuffle(
	n int,
	swap func(i int, j int),
) {
	rand.Shuffle(n, swap)
}
