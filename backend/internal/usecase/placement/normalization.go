package placement

import "math"

func NormalizeRotation(rotation float64) float64 {
	normalizedRotation := math.Mod(rotation, 360)
	if normalizedRotation < 0 {
		normalizedRotation += 360
	}

	return normalizedRotation
}
