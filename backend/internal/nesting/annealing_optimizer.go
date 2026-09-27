package nesting

import (
	"context"
	"fmt"

	"server_nesting_optimizer/internal/geometry"
)

type SimulatedAnnealingOptimizer struct {
	engine     geometry.Engine
	nfpBuilder geometry.NFPBuilder
	config     AnnealingConfig
	rng        AnnealingRandom
}

var _ Optimizer = (*SimulatedAnnealingOptimizer)(nil)

func NewSimulatedAnnealingOptimizer(
	engine geometry.Engine,
	nfpBuilder geometry.NFPBuilder,
	config AnnealingConfig,
) (*SimulatedAnnealingOptimizer, error) {
	if err := validateAnnealingConfig(config); err != nil {
		return nil, fmt.Errorf(
			"new simulated annealing optimizer: %w",
			err,
		)
	}

	return &SimulatedAnnealingOptimizer{
		engine:     engine,
		nfpBuilder: nfpBuilder,
		config:     config,
		rng:        newDefaultAnnealingRandom(),
	}, nil
}

func (o *SimulatedAnnealingOptimizer) Optimize(
	ctx context.Context,
	problem Problem,
) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, fmt.Errorf(
			"simulated annealing optimizer optimize: %w",
			err,
		)
	}

	if err := validateProblem(
		o.engine,
		problem,
	); err != nil {
		return Result{}, fmt.Errorf(
			"simulated annealing optimizer optimize: %w",
			err,
		)
	}

	genes := expandPatternInstances(
		problem.Patterns,
	)

	patternLookup := buildPatternLookup(
		problem.Patterns,
	)

	initialPopulation := generateInitialPopulation(
		genes,
		problem.AllowedRotations,
		1,
		o.rng,
	)

	current := initialPopulation[0]

	currentDecoded, err := decodeChromosome(
		ctx,
		o.engine,
		o.nfpBuilder,
		problem,
		current,
		patternLookup,
	)
	if err != nil {
		return Result{}, fmt.Errorf(
			"simulated annealing optimizer optimize: %w",
			err,
		)
	}

	currentEnergy, err := calculateAnnealingEnergy(
		o.engine,
		problem.Surface,
		currentDecoded.Fitness,
	)
	if err != nil {
		return Result{}, fmt.Errorf(
			"simulated annealing optimizer optimize: %w",
			err,
		)
	}

	bestDecoded := currentDecoded

	if len(genes) == 0 {
		return buildResultFromDecoded(
			o.engine,
			problem,
			bestDecoded,
		)
	}

	temperature := o.config.InitialTemperature

	for iteration := 0; iteration < o.config.MaxIterations &&
		temperature > o.config.MinTemperature; iteration++ {

		if err := ctx.Err(); err != nil {
			return Result{}, fmt.Errorf(
				"simulated annealing optimizer optimize: %w",
				err,
			)
		}

		candidate := o.generateNeighbor(
			current,
			problem.AllowedRotations,
		)

		candidateDecoded, err := decodeChromosome(
			ctx,
			o.engine,
			o.nfpBuilder,
			problem,
			candidate,
			patternLookup,
		)
		if err != nil {
			return Result{}, fmt.Errorf(
				"simulated annealing optimizer optimize: %w",
				err,
			)
		}

		candidateEnergy, err := calculateAnnealingEnergy(
			o.engine,
			problem.Surface,
			candidateDecoded.Fitness,
		)
		if err != nil {
			return Result{}, fmt.Errorf(
				"simulated annealing optimizer optimize: %w",
				err,
			)
		}

		accept := isFitnessBetter(
			candidateDecoded.Fitness,
			currentDecoded.Fitness,
		)

		if !accept {
			accept = shouldAcceptWorse(
				currentEnergy,
				candidateEnergy,
				temperature,
				o.rng,
			)
		}

		if accept {
			current = candidate
			currentDecoded = candidateDecoded
			currentEnergy = candidateEnergy

			if isFitnessBetter(
				currentDecoded.Fitness,
				bestDecoded.Fitness,
			) {
				bestDecoded = currentDecoded
			}
		}

		temperature *= o.config.CoolingRate
	}

	result, err := buildResultFromDecoded(
		o.engine,
		problem,
		bestDecoded,
	)
	if err != nil {
		return Result{}, fmt.Errorf(
			"simulated annealing optimizer optimize: %w",
			err,
		)
	}

	return result, nil
}

func (o *SimulatedAnnealingOptimizer) generateNeighbor(
	chromosome Chromosome,
	allowedRotations []float64,
) Chromosome {
	geneCount := len(chromosome.Genes)

	if geneCount == 0 {
		return cloneChromosome(chromosome)
	}

	if geneCount == 1 {
		return o.generateRotationNeighbor(
			chromosome,
			allowedRotations,
		)
	}

	switch o.rng.Intn(2) {
	case 0:
		return o.generateSwapNeighbor(
			chromosome,
		)
	default:
		return o.generateRotationNeighbor(
			chromosome,
			allowedRotations,
		)
	}
}

func (o *SimulatedAnnealingOptimizer) generateSwapNeighbor(
	chromosome Chromosome,
) Chromosome {
	firstIdx := o.rng.Intn(
		len(chromosome.Genes),
	)

	secondIdx := o.rng.Intn(
		len(chromosome.Genes) - 1,
	)

	if secondIdx >= firstIdx {
		secondIdx++
	}

	return mutateSwap(
		chromosome,
		firstIdx,
		secondIdx,
	)
}

func (o *SimulatedAnnealingOptimizer) generateRotationNeighbor(
	chromosome Chromosome,
	allowedRotations []float64,
) Chromosome {
	geneIdx := o.rng.Intn(
		len(chromosome.Genes),
	)

	rotationIdx := o.rng.Intn(
		len(allowedRotations),
	)

	return mutateRotation(
		chromosome,
		geneIdx,
		allowedRotations[rotationIdx],
	)
}
