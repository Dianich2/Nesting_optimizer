package nesting

import (
	"context"
	"errors"
	"fmt"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	domainplacement "server_nesting_optimizer/internal/domain/placement"
	projectpattern "server_nesting_optimizer/internal/domain/project_pattern"
	domainprojectsurface "server_nesting_optimizer/internal/domain/project_surface"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/internal/nesting"
	"server_nesting_optimizer/pkg/apperror"
)

type RunNestingUseCase struct {
	projectSurfaceRepository ProjectSurfaceRepository
	projectPatternRepository ProjectPatternRepository
	placementRepository      PlacementRepository
	geometryEngine           geometry.Engine
	optimizer                nesting.Optimizer
	unitOfWork               UnitOfWork
}

func NewRunNestingUseCase(
	projectSurfaceRepository ProjectSurfaceRepository,
	projectPatternRepository ProjectPatternRepository,
	placementRepository PlacementRepository,
	geometryEngine geometry.Engine,
	optimizer nesting.Optimizer,
	unitOfWork UnitOfWork,
) *RunNestingUseCase {
	return &RunNestingUseCase{
		projectSurfaceRepository: projectSurfaceRepository,
		projectPatternRepository: projectPatternRepository,
		placementRepository:      placementRepository,
		geometryEngine:           geometryEngine,
		optimizer:                optimizer,
		unitOfWork:               unitOfWork,
	}
}

type preparedPlacement struct {
	Placement domainplacement.Placement
	Geometry  domaingeometry.Polygon
}

func (uc *RunNestingUseCase) Execute(
	ctx context.Context,
	input RunNestingInput,
) (RunNestingOutput, error) {
	if err := ctx.Err(); err != nil {
		return RunNestingOutput{}, fmt.Errorf(
			"run nesting: %w",
			err,
		)
	}

	details := input.Validate()
	if len(details) > 0 {
		return RunNestingOutput{}, apperror.Validation(
			"validation failed",
			details...,
		)
	}

	projectSurface, err := uc.projectSurfaceRepository.GetByID(
		ctx,
		input.UserID,
		input.ProjectID,
		input.ProjectSurfaceID,
	)
	if err != nil {
		if errors.Is(err, domainprojectsurface.ErrNotFound) {
			return RunNestingOutput{}, apperror.NotFound(
				"project surface not found",
			)
		}

		if errors.Is(err, context.Canceled) {
			return RunNestingOutput{}, fmt.Errorf(
				"run nesting: %w",
				err,
			)
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return RunNestingOutput{}, fmt.Errorf(
				"run nesting: %w",
				err,
			)
		}

		return RunNestingOutput{}, apperror.Internal(
			"failed to get project surface by id",
			err,
		)
	}

	projectPatternIDs := make([]int64, 0, len(input.Patterns))
	for _, projectPattern := range input.Patterns {
		projectPatternIDs = append(projectPatternIDs, projectPattern.ProjectPatternID)
	}

	projectPatternsFromRepository, err := uc.projectPatternRepository.GetByIDs(
		ctx,
		input.UserID,
		input.ProjectID,
		projectPatternIDs,
	)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return RunNestingOutput{}, fmt.Errorf(
				"run nesting: %w",
				err,
			)
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return RunNestingOutput{}, fmt.Errorf(
				"run nesting: %w",
				err,
			)
		}

		return RunNestingOutput{}, apperror.Internal(
			"failed to get project patterns",
			err,
		)
	}

	foundIDs := make(map[int64]projectpattern.ProjectPattern)

	for _, projectPatternRepo := range projectPatternsFromRepository {
		foundIDs[projectPatternRepo.ID] = projectPatternRepo
	}

	for _, reqProjectPattern := range input.Patterns {
		_, ok := foundIDs[reqProjectPattern.ProjectPatternID]

		if !ok {
			return RunNestingOutput{}, apperror.NotFound(
				"project pattern not found",
			)
		}
	}

	var collisionPlacements []domainplacement.CollisionPlacement
	if input.KeepExisting {
		collisionPlacements, err = uc.placementRepository.ListForCollisionCheck(
			ctx,
			input.ProjectSurfaceID,
			input.ProjectID,
			input.UserID,
		)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return RunNestingOutput{}, fmt.Errorf(
					"run nesting: %w",
					err,
				)
			}

			if errors.Is(err, context.DeadlineExceeded) {
				return RunNestingOutput{}, fmt.Errorf(
					"run nesting: %w",
					err,
				)
			}

			return RunNestingOutput{}, apperror.Internal(
				"failed to get collision placements",
				err,
			)
		}
	}

	problem, err := buildProblem(
		uc.geometryEngine,
		input,
		projectSurface.Geometry,
		projectPatternsFromRepository,
		collisionPlacements,
	)
	if err != nil {
		return RunNestingOutput{}, apperror.Internal(
			"failed to build nesting problem",
			err,
		)
	}

	result, err := uc.optimizer.Optimize(
		ctx,
		problem,
	)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return RunNestingOutput{}, fmt.Errorf(
				"run nesting: %w",
				err,
			)
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return RunNestingOutput{}, fmt.Errorf(
				"run nesting: %w",
				err,
			)
		}

		return RunNestingOutput{}, apperror.Internal(
			"failed to optimize nesting problem",
			err,
		)
	}

	preparedPlacements := make([]preparedPlacement, 0, len(result.Placements))
	for _, resPlacement := range result.Placements {
		projectPattern, ok := foundIDs[resPlacement.PatternID]
		if !ok {
			return RunNestingOutput{}, apperror.Internal(
				"optimizer returned unknown project pattern",
				fmt.Errorf(
					"optimizer returned unknown project pattern: %d",
					resPlacement.PatternID,
				),
			)
		}

		transformedGeometry, err := uc.geometryEngine.Transform(
			projectPattern.Geometry,
			resPlacement.X,
			resPlacement.Y,
			resPlacement.Rotation,
		)
		if err != nil {
			return RunNestingOutput{}, apperror.Internal(
				"failed to transform pattern geometry",
				err,
			)
		}

		prepPlacement := domainplacement.Placement{
			ProjectSurfaceID: input.ProjectSurfaceID,
			ProjectPatternID: resPlacement.PatternID,
			X:                resPlacement.X,
			Y:                resPlacement.Y,
			Rotation:         resPlacement.Rotation,
		}

		preparedPlacements = append(preparedPlacements, preparedPlacement{
			Placement: prepPlacement,
			Geometry:  transformedGeometry,
		})
	}

	savedPlacements := make([]domainplacement.Placement, 0, len(result.Placements))

	err = uc.unitOfWork.WithinTransaction(
		ctx,
		func(repositories TransactionRepositories) error {
			if !input.KeepExisting {
				err := repositories.Placements.DeleteByProjectSurface(
					ctx,
					input.UserID,
					input.ProjectID,
					input.ProjectSurfaceID,
				)
				if err != nil {
					return fmt.Errorf("delete existing placements: %w", err)
				}
			}

			for _, placementToCreate := range preparedPlacements {
				savedPlacement, err := repositories.Placements.Create(
					ctx,
					placementToCreate.Placement,
					input.ProjectID,
					input.UserID,
				)
				if err != nil {
					return fmt.Errorf("create nesting placement: %w", err)
				}

				savedPlacements = append(savedPlacements, savedPlacement)
			}

			return nil
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return RunNestingOutput{}, fmt.Errorf(
				"run nesting: %w",
				err,
			)

		case errors.Is(err, context.DeadlineExceeded):
			return RunNestingOutput{}, fmt.Errorf(
				"run nesting: %w",
				err,
			)

		default:
			return RunNestingOutput{}, apperror.Internal(
				"failed to persist nesting result",
				err,
			)
		}
	}

	runNestingOutput := RunNestingOutput{
		Placements: make([]PlacementItem, 0, len(savedPlacements)),
		Unplaced:   result.Unplaced,
		Metrics:    result.Metrics,
	}

	for i, savedPlacement := range savedPlacements {
		preparedPlacement := preparedPlacements[i]

		runNestingOutput.Placements = append(
			runNestingOutput.Placements,
			PlacementItem{
				ID:               savedPlacement.ID,
				ProjectSurfaceID: savedPlacement.ProjectSurfaceID,
				ProjectPatternID: savedPlacement.ProjectPatternID,
				X:                savedPlacement.X,
				Y:                savedPlacement.Y,
				Rotation:         savedPlacement.Rotation,
				Geometry:         preparedPlacement.Geometry,
				CreatedAt:        savedPlacement.CreatedAt,
				UpdatedAt:        savedPlacement.UpdatedAt,
			},
		)
	}

	return runNestingOutput, nil
}
