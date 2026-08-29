package nesting

import "server_nesting_optimizer/internal/geometry"

func NormalizeAllowedRotations(
	input []float64,
) []float64 {
	seenRotations := make(map[float64]struct{})
	res := make([]float64, 0, len(input))

	for _, rotation := range input {
		normalizedRotation := geometry.NormalizeDegrees(rotation)

		_, ok := seenRotations[normalizedRotation]
		if ok {
			continue
		}

		res = append(res, normalizedRotation)
		seenRotations[normalizedRotation] = struct{}{}
	}

	return res
}
