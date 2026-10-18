package nesting

import (
	"context"
	"fmt"

	"server_nesting_optimizer/internal/geometry"
)

type HybridOptimizer struct {
	engine             geometry.Engine
	geneticOptimizer   Optimizer
	annealingOptimizer Optimizer
	compactionConfig   CompactionConfig
}

var _ Optimizer = (*HybridOptimizer)(nil)

func NewHybridOptimizer(
	engine geometry.Engine,
	geneticOptimizer Optimizer,
	annealingOptimizer Optimizer,
	compactionConfig CompactionConfig,
) (*HybridOptimizer, error) {
	if geneticOptimizer == nil {
		return nil, fmt.Errorf(
			"new hybrid optimizer: genetic optimizer is nil",
		)
	}

	if annealingOptimizer == nil {
		return nil, fmt.Errorf(
			"new hybrid optimizer: annealing optimizer is nil",
		)
	}

	if err := validateCompactionConfig(compactionConfig); err != nil {
		return nil, fmt.Errorf(
			"new hybrid optimizer: %w",
			err,
		)
	}

	return &HybridOptimizer{
		engine:             engine,
		geneticOptimizer:   geneticOptimizer,
		annealingOptimizer: annealingOptimizer,
		compactionConfig:   compactionConfig,
	}, nil
}

func (o *HybridOptimizer) Optimize(
	ctx context.Context,
	problem Problem,
) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, fmt.Errorf(
			"hybrid optimizer optimize: %w",
			err,
		)
	}

	if err := validateProblem(
		o.engine,
		problem,
	); err != nil {
		return Result{}, fmt.Errorf(
			"hybrid optimizer optimize: %w",
			err,
		)
	}

	geneticResult, err := o.geneticOptimizer.Optimize(
		ctx,
		problem,
	)
	if err != nil {
		return Result{}, fmt.Errorf(
			"hybrid optimizer optimize genetic: %w",
			err,
		)
	}

	geneticCandidate, err := o.compactCandidate(
		ctx,
		problem,
		geneticResult,
	)
	if err != nil {
		return Result{}, fmt.Errorf(
			"hybrid optimizer compact genetic: %w",
			err,
		)
	}

	if err := ctx.Err(); err != nil {
		return Result{}, fmt.Errorf(
			"hybrid optimizer optimize: %w",
			err,
		)
	}

	annealingResult, err := o.annealingOptimizer.Optimize(
		ctx,
		problem,
	)
	if err != nil {
		return Result{}, fmt.Errorf(
			"hybrid optimizer optimize annealing: %w",
			err,
		)
	}

	annealingCandidate, err := o.compactCandidate(
		ctx,
		problem,
		annealingResult,
	)
	if err != nil {
		return Result{}, fmt.Errorf(
			"hybrid optimizer compact annealing: %w",
			err,
		)
	}

	if isFitnessBetter(
		annealingCandidate.fitness,
		geneticCandidate.fitness,
	) {
		return annealingCandidate.result, nil
	}

	return geneticCandidate.result, nil
}

type hybridCandidate struct {
	result  Result
	fitness Fitness
}

func (o *HybridOptimizer) compactCandidate(
	ctx context.Context,
	problem Problem,
	result Result,
) (hybridCandidate, error) {
	compactedPlacements, err := compactPlacements(
		ctx,
		o.engine,
		problem,
		result.Placements,
		o.compactionConfig,
	)
	if err != nil {
		return hybridCandidate{}, fmt.Errorf(
			"compact hybrid candidate: %w",
			err,
		)
	}

	fitness, err := calculatePlacementsFitness(
		o.engine,
		problem,
		compactedPlacements,
	)
	if err != nil {
		return hybridCandidate{}, fmt.Errorf(
			"compact hybrid candidate: %w",
			err,
		)
	}

	compactedResult, err := buildResultFromDecoded(
		o.engine,
		problem,
		DecodedChromosome{
			Placements: compactedPlacements,
			Fitness:    fitness,
		},
	)
	if err != nil {
		return hybridCandidate{}, fmt.Errorf(
			"compact hybrid candidate: %w",
			err,
		)
	}

	return hybridCandidate{
		result:  compactedResult,
		fitness: fitness,
	}, nil
}
