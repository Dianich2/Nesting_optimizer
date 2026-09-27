package nesting

type Fitness struct {
	UnplacedCount int
	UsedArea      float64
	UsedHeight    float64
}

func isFitnessBetter(
	candidate Fitness,
	best Fitness,
) bool {
	if candidate.UnplacedCount < best.UnplacedCount {
		return true
	}

	if candidate.UnplacedCount > best.UnplacedCount {
		return false
	}

	if !floatsEqual(candidate.UsedArea, best.UsedArea) {
		return candidate.UsedArea < best.UsedArea
	}

	if !floatsEqual(candidate.UsedHeight, best.UsedHeight) {
		return candidate.UsedHeight < best.UsedHeight
	}

	return false
}
