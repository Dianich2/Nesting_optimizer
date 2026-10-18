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
	surfaceBounds, err := engine.Bounds(surface)
	if err != nil {
		return 0, fmt.Errorf(
			"calculate annealing energy: %w",
			err,
		)
	}

	surfaceArea, err := engine.Area(surface)
	if err != nil {
		return 0, fmt.Errorf(
			"calculate annealing energy: %w",
			err,
		)
	}

	if surfaceArea <= 0 {
		return 0, fmt.Errorf(
			"calculate annealing energy: invalid surface area",
		)
	}

	surfaceWidth :=
		surfaceBounds.MaxX - surfaceBounds.MinX

	surfaceHeight :=
		surfaceBounds.MaxY - surfaceBounds.MinY

	if surfaceWidth <= 0 || surfaceHeight <= 0 {
		return 0, fmt.Errorf(
			"calculate annealing energy: invalid surface bounds",
		)
	}

	boundingArea :=
		surfaceWidth * surfaceHeight

	normalizedUnplaced :=
		float64(fitness.UnplacedCount)

	normalizedMissingPlacedArea :=
		1 - fitness.PlacedArea/surfaceArea

	if normalizedMissingPlacedArea < 0 {
		normalizedMissingPlacedArea = 0
	}

	normalizedUsedArea :=
		fitness.UsedArea / boundingArea

	normalizedHeight :=
		fitness.UsedHeight / surfaceHeight

	const (
		unplacedPenalty   = 4.0
		placedAreaPenalty = 2.0
	)

	energy :=
		normalizedUnplaced*unplacedPenalty +
			normalizedMissingPlacedArea*placedAreaPenalty +
			normalizedUsedArea +
			normalizedHeight

	return energy, nil
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
