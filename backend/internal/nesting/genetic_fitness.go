package nesting

type Fitness struct {
	UnplacedCount int
	PlacedArea    float64
	UsedArea      float64
	UsedHeight    float64
}

func isFitnessBetter(
	candidate Fitness,
	best Fitness,
) bool {
	if candidate.UnplacedCount != best.UnplacedCount {
		return candidate.UnplacedCount < best.UnplacedCount
	}

	if !floatsEqual(
		candidate.PlacedArea,
		best.PlacedArea,
	) {
		return candidate.PlacedArea > best.PlacedArea
	}

	if !floatsEqual(
		candidate.UsedArea,
		best.UsedArea,
	) {
		return candidate.UsedArea < best.UsedArea
	}

	if !floatsEqual(
		candidate.UsedHeight,
		best.UsedHeight,
	) {
		return candidate.UsedHeight < best.UsedHeight
	}

	return false
}
