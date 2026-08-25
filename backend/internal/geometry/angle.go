package geometry

import "math"

func NormalizeDegrees(angle float64) float64 {
	normalizedAngle := math.Mod(angle, 360)
	if normalizedAngle < 0 {
		normalizedAngle += 360
	}

	return normalizedAngle
}
