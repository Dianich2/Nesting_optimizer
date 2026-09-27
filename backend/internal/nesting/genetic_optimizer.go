package nesting

import (
	"context"
	"fmt"
	"sort"

	"server_nesting_optimizer/internal/geometry"
)

type GeneticOptimizer struct {
	engine     geometry.Engine
	nfpBuilder geometry.NFPBuilder
	config     GeneticConfig
	rng        Random
}

var _ Optimizer = (*GeneticOptimizer)(nil)

func NewGeneticOptimizer(
	engine geometry.Engine,
	nfpBuilder geometry.NFPBuilder,
	config GeneticConfig,
) (*GeneticOptimizer, error) {
	if err := validateGeneticConfig(config); err != nil {
		return nil, fmt.Errorf(
			"new genetic optimizer: %w",
			err,
		)
	}

	return &GeneticOptimizer{
		engine:     engine,
		nfpBuilder: nfpBuilder,
		config:     config,
		rng:        newDefaultRandom(),
	}, nil
}

func (o *GeneticOptimizer) Optimize(
	ctx context.Context,
	problem Problem,
) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, fmt.Errorf(
			"genetic optimizer optimize: %w",
			err,
		)
	}

	if err := validateProblem(
		o.engine,
		problem,
	); err != nil {
		return Result{}, fmt.Errorf(
			"genetic optimizer optimize: %w",
			err,
		)
	}

	genes := expandPatternInstances(
		problem.Patterns,
	)

	patternLookup := buildPatternLookup(
		problem.Patterns,
	)

	population := generateInitialPopulation(
		genes,
		problem.AllowedRotations,
		o.config.PopulationSize,
		o.rng,
	)

	evaluatedPopulation, err := o.evaluatePopulation(
		ctx,
		problem,
		patternLookup,
		population,
	)
	if err != nil {
		return Result{}, fmt.Errorf(
			"genetic optimizer optimize: %w",
			err,
		)
	}

	for generation := 1; generation < o.config.Generations; generation++ {
		if err := ctx.Err(); err != nil {
			return Result{}, fmt.Errorf(
				"genetic optimizer optimize: %w",
				err,
			)
		}

		population = o.createNextGeneration(
			evaluatedPopulation,
			problem.AllowedRotations,
		)

		evaluatedPopulation, err = o.evaluatePopulation(
			ctx,
			problem,
			patternLookup,
			population,
		)
		if err != nil {
			return Result{}, fmt.Errorf(
				"genetic optimizer optimize: %w",
				err,
			)
		}
	}

	best := findBestEvaluatedChromosome(
		evaluatedPopulation,
	)

	decoded, err := decodeChromosome(
		ctx,
		o.engine,
		o.nfpBuilder,
		problem,
		best.Chromosome,
		patternLookup,
	)
	if err != nil {
		return Result{}, fmt.Errorf(
			"genetic optimizer optimize: %w",
			err,
		)
	}

	result, err := o.buildResult(
		problem,
		decoded,
	)
	if err != nil {
		return Result{}, fmt.Errorf(
			"genetic optimizer optimize: %w",
			err,
		)
	}

	return result, nil
}

func (o *GeneticOptimizer) evaluatePopulation(
	ctx context.Context,
	problem Problem,
	patternLookup map[int64]PatternItem,
	population []Chromosome,
) ([]EvaluatedChromosome, error) {
	evaluated := make(
		[]EvaluatedChromosome,
		0,
		len(population),
	)

	for _, chromosome := range population {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf(
				"evaluate population: %w",
				err,
			)
		}

		decoded, err := decodeChromosome(
			ctx,
			o.engine,
			o.nfpBuilder,
			problem,
			chromosome,
			patternLookup,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"evaluate population: %w",
				err,
			)
		}

		evaluated = append(
			evaluated,
			EvaluatedChromosome{
				Chromosome: chromosome,
				Fitness:    decoded.Fitness,
			},
		)
	}

	return evaluated, nil
}

func (o *GeneticOptimizer) createNextGeneration(
	population []EvaluatedChromosome,
	allowedRotations []float64,
) []Chromosome {
	sortedPopulation := append(
		[]EvaluatedChromosome(nil),
		population...,
	)

	sort.SliceStable(
		sortedPopulation,
		func(i int, j int) bool {
			return isFitnessBetter(
				sortedPopulation[i].Fitness,
				sortedPopulation[j].Fitness,
			)
		},
	)

	nextPopulation := make(
		[]Chromosome,
		0,
		o.config.PopulationSize,
	)

	for i := 0; i < o.config.EliteCount; i++ {
		nextPopulation = append(
			nextPopulation,
			cloneChromosome(
				sortedPopulation[i].Chromosome,
			),
		)
	}

	for len(nextPopulation) < o.config.PopulationSize {
		parentA := tournamentSelection(
			population,
			o.config.TournamentSize,
			o.rng,
		)

		parentB := tournamentSelection(
			population,
			o.config.TournamentSize,
			o.rng,
		)

		child := o.crossover(
			parentA.Chromosome,
			parentB.Chromosome,
		)

		child = o.mutate(
			child,
			allowedRotations,
		)

		nextPopulation = append(
			nextPopulation,
			child,
		)
	}

	return nextPopulation
}

func (o *GeneticOptimizer) crossover(
	parentA Chromosome,
	parentB Chromosome,
) Chromosome {
	geneCount := len(parentA.Genes)

	if geneCount < 2 {
		return cloneChromosome(parentA)
	}

	start := o.rng.Intn(geneCount)

	end := start + 1 + o.rng.Intn(
		geneCount-start,
	)

	return orderCrossover(
		parentA,
		parentB,
		start,
		end,
	)
}

func (o *GeneticOptimizer) mutate(
	chromosome Chromosome,
	allowedRotations []float64,
) Chromosome {
	if len(chromosome.Genes) == 0 {
		return cloneChromosome(chromosome)
	}

	if !shouldMutate(
		o.config.MutationRate,
		o.rng,
	) {
		return cloneChromosome(chromosome)
	}

	if len(chromosome.Genes) == 1 {
		return o.mutateChromosomeRotation(
			chromosome,
			allowedRotations,
		)
	}

	switch o.rng.Intn(2) {
	case 0:
		return o.mutateChromosomeSwap(
			chromosome,
		)
	default:
		return o.mutateChromosomeRotation(
			chromosome,
			allowedRotations,
		)
	}
}

func (o *GeneticOptimizer) mutateChromosomeSwap(
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

func (o *GeneticOptimizer) mutateChromosomeRotation(
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

func (o *GeneticOptimizer) buildResult(
	problem Problem,
	decoded DecodedChromosome,
) (Result, error) {
	surfaceArea, err := o.engine.Area(
		problem.Surface,
	)
	if err != nil {
		return Result{}, fmt.Errorf(
			"build genetic result: %w",
			err,
		)
	}

	placedCountByPattern := make(
		map[int64]int,
	)

	for _, placement := range decoded.Placements {
		placedCountByPattern[placement.PatternID]++
	}

	unplaced := make(
		[]UnplacedPattern,
		0,
	)

	requestedCount := 0
	placedArea := 0.0

	for _, patternItem := range problem.Patterns {
		requestedCount += patternItem.Quantity

		placedCount := placedCountByPattern[patternItem.PatternID]

		remaining := patternItem.Quantity - placedCount

		if remaining > 0 {
			unplaced = append(
				unplaced,
				UnplacedPattern{
					PatternID: patternItem.PatternID,
					Quantity:  remaining,
				},
			)
		}

		patternArea, err := o.engine.Area(
			patternItem.Geometry,
		)
		if err != nil {
			return Result{}, fmt.Errorf(
				"build genetic result: %w",
				err,
			)
		}

		placedArea += patternArea * float64(
			placedCount,
		)
	}

	result := Result{
		Placements: decoded.Placements,
		Unplaced:   unplaced,
		Metrics: Metrics{
			RequestedCount: requestedCount,
			PlacedCount:    len(decoded.Placements),
			SurfaceArea:    surfaceArea,
			PlacedArea:     placedArea,
		},
	}

	if surfaceArea > 0 {
		result.Metrics.Utilization =
			placedArea / surfaceArea
	}

	return result, nil
}

func cloneChromosome(
	chromosome Chromosome,
) Chromosome {
	genes := make(
		[]Gene,
		len(chromosome.Genes),
	)

	copy(
		genes,
		chromosome.Genes,
	)

	return Chromosome{
		Genes: genes,
	}
}

func shouldMutate(
	mutationRate float64,
	rng Random,
) bool {
	if mutationRate <= 0 {
		return false
	}

	if mutationRate >= 1 {
		return true
	}

	const precision = 1_000_000

	threshold := int(
		mutationRate * precision,
	)

	return rng.Intn(precision) < threshold
}

func findBestEvaluatedChromosome(
	population []EvaluatedChromosome,
) EvaluatedChromosome {
	best := population[0]

	for i := 1; i < len(population); i++ {
		if isFitnessBetter(
			population[i].Fitness,
			best.Fitness,
		) {
			best = population[i]
		}
	}

	return best
}
