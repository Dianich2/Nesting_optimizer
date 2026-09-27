package nesting

import (
	"fmt"
	"math"

	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
)

func calculateAnnealingEnergy(
	engine geometry.Engine,
	surface domaingeometry.Polygon,
	fitness Fitness,
) (float64, error) {
	bounds, err := engine.Bounds(surface)
	if err != nil {
		return 0, fmt.Errorf(
			"calculate annealing energy: %w",
			err,
		)
	}

	surfaceWidth := bounds.MaxX - bounds.MinX
	surfaceHeight := bounds.MaxY - bounds.MinY

	if surfaceWidth <= 0 || surfaceHeight <= 0 {
		return 0, fmt.Errorf(
			"calculate annealing energy: invalid surface bounds",
		)
	}

	boundingArea := surfaceWidth * surfaceHeight

	normalizedArea := fitness.UsedArea / boundingArea
	normalizedHeight := fitness.UsedHeight / surfaceHeight

	const unplacedPenalty = 3.0

	return float64(fitness.UnplacedCount)*unplacedPenalty +
		normalizedArea +
		normalizedHeight, nil
}

func shouldAcceptWorse(
	currentEnergy float64,
	candidateEnergy float64,
	temperature float64,
	rng AnnealingRandom,
) bool {
	if candidateEnergy <= currentEnergy {
		return true
	}

	if temperature <= 0 {
		return false
	}

	delta := candidateEnergy - currentEnergy

	probability := math.Exp(
		-delta / temperature,
	)

	return rng.Float64() < probability
}
