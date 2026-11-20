package nesting

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	nestingrun "server_nesting_optimizer/internal/domain/nesting_run"
	domainplacement "server_nesting_optimizer/internal/domain/placement"
	projectpattern "server_nesting_optimizer/internal/domain/project_pattern"
	domainprojectsurface "server_nesting_optimizer/internal/domain/project_surface"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/internal/nesting"
	projecttransaction "server_nesting_optimizer/internal/usecase/project_transaction"
	"server_nesting_optimizer/pkg/apperror"
	"sort"
	"time"
)

type RunNestingUseCase struct {
	Budgets                  InputBudget
	Previews                 PreviewRepository
	Limiter                  *ComputeLimiter
	History                  HistoryRepository
	Operations               OperationRepository
	Transactions             projecttransaction.Executor
	projectSurfaceRepository ProjectSurfaceRepository
	projectPatternRepository ProjectPatternRepository
	placementRepository      PlacementRepository
	geometryEngine           geometry.Engine
	optimizerRegistry        *nesting.OptimizerRegistry
	unitOfWork               UnitOfWork
}

func NewRunNestingUseCase(
	projectSurfaceRepository ProjectSurfaceRepository,
	projectPatternRepository ProjectPatternRepository,
	placementRepository PlacementRepository,
	geometryEngine geometry.Engine,
	optimizerRegistry *nesting.OptimizerRegistry,
	unitOfWork UnitOfWork,
) *RunNestingUseCase {
	return &RunNestingUseCase{
		Limiter:                  NewComputeLimiter(2, 4),
		projectSurfaceRepository: projectSurfaceRepository,
		projectPatternRepository: projectPatternRepository,
		placementRepository:      placementRepository,
		geometryEngine:           geometryEngine,
		optimizerRegistry:        optimizerRegistry,
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
	return uc.ExecuteWithKey(ctx, input, "")
}

func (uc *RunNestingUseCase) ExecuteWithKey(ctx context.Context, input RunNestingInput, rawKey string) (RunNestingOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, ExecutionTimeout)
	defer cancel()
	if rawKey != "" {
		key, err := parseOperationKey(rawKey)
		if err != nil {
			return RunNestingOutput{}, err
		}
		return uc.runOperation(ctx, input, key)
	}
	release, err := uc.Limiter.Acquire(ctx)
	if err != nil {
		return RunNestingOutput{}, err
	}
	defer release()
	return uc.execute(ctx, input)
}

type calculatedLayout struct {
	previewID  string
	result     nesting.Result
	placements []preparedPlacement
	duration   time.Duration
	snapshot   preparedSnapshot
}

type preparedSnapshot struct {
	Placements []domainplacement.CollisionPlacement
	Problem    nesting.Problem
	Hash       [32]byte
}

func (uc *RunNestingUseCase) prepare(ctx context.Context, input RunNestingInput) (preparedSnapshot, error) {
	return projecttransaction.Run(ctx, uc.Transactions, input.ProjectID, input.UserID, func(ctx context.Context) (preparedSnapshot, error) {
		return uc.loadSnapshot(ctx, input)
	})
}

func (uc *RunNestingUseCase) loadSnapshot(ctx context.Context, input RunNestingInput) (preparedSnapshot, error) {
	if err := ctx.Err(); err != nil {
		return preparedSnapshot{}, err
	}
	if details := input.Validate(); len(details) > 0 {
		return preparedSnapshot{}, apperror.Validation("validation failed", details...)
	}

	if uc.Budgets != nil {
		if err := uc.Budgets.CheckNestingBudget(ctx, input); err != nil {
			return preparedSnapshot{}, err
		}
	}

	projectSurface, err := uc.projectSurfaceRepository.GetByID(
		ctx,
		input.UserID,
		input.ProjectID,
		input.ProjectSurfaceID,
	)
	if err != nil {
		if errors.Is(err, domainprojectsurface.ErrNotFound) {
			return preparedSnapshot{}, apperror.NotFound(
				"project surface not found",
			)
		}

		if errors.Is(err, context.Canceled) {
			return preparedSnapshot{}, fmt.Errorf(
				"run nesting: %w",
				err,
			)
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return preparedSnapshot{}, fmt.Errorf(
				"run nesting: %w",
				err,
			)
		}

		return preparedSnapshot{}, apperror.Internal(
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
			return preparedSnapshot{}, fmt.Errorf(
				"run nesting: %w",
				err,
			)
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return preparedSnapshot{}, fmt.Errorf(
				"run nesting: %w",
				err,
			)
		}

		return preparedSnapshot{}, apperror.Internal(
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
			return preparedSnapshot{}, apperror.NotFound(
				"project pattern not found",
			)
		}
	}

	var collisionPlacements []domainplacement.CollisionPlacement
	{
		collisionPlacements, err = uc.placementRepository.ListForCollisionCheck(
			ctx,
			input.ProjectSurfaceID,
			input.ProjectID,
			input.UserID,
		)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return preparedSnapshot{}, fmt.Errorf(
					"run nesting: %w",
					err,
				)
			}

			if errors.Is(err, context.DeadlineExceeded) {
				return preparedSnapshot{}, fmt.Errorf(
					"run nesting: %w",
					err,
				)
			}

			return preparedSnapshot{}, apperror.Internal(
				"failed to get collision placements",
				err,
			)
		}
	}

	sort.Slice(collisionPlacements, func(i, j int) bool { return collisionPlacements[i].ID < collisionPlacements[j].ID })
	problem, err := buildProblem(
		uc.geometryEngine,
		input,
		projectSurface.Geometry,
		projectPatternsFromRepository,
		nil,
	)
	if err != nil {
		return preparedSnapshot{}, apperror.Internal(
			"failed to build nesting problem",
			err,
		)
	}
	encoded, err := json.Marshal(struct {
		Problem    nesting.Problem
		Placements []domainplacement.CollisionPlacement
	}{problem, collisionPlacements})
	if err != nil {
		return preparedSnapshot{}, apperror.Internal("encode snapshot", err)
	}
	return preparedSnapshot{Problem: problem, Hash: sha256.Sum256(encoded), Placements: collisionPlacements}, nil
}

func (uc *RunNestingUseCase) calculate(ctx context.Context, input RunNestingInput) (layout calculatedLayout, calculationErr error) {
	defer func() {
		if failure := recover(); failure != nil {
			layout = calculatedLayout{}
			calculationErr = apperror.Internal("nesting calculation panicked", fmt.Errorf("%v", failure))
		}
	}()
	snapshot, err := uc.prepare(ctx, input)
	if err != nil {
		return calculatedLayout{}, err
	}
	optimizer, err := uc.optimizerRegistry.Get(input.Algorithm)
	if err != nil {
		return calculatedLayout{}, apperror.Validation("unsupported algorithm")
	}
	problem := snapshot.Problem
	if input.KeepExisting {
		for _, placement := range snapshot.Placements {
			if err := ctx.Err(); err != nil {
				return calculatedLayout{}, err
			}
			transformed, err := uc.geometryEngine.Transform(placement.PatternGeometry, placement.X, placement.Y, placement.Rotation)
			if err != nil {
				return calculatedLayout{}, apperror.Internal("prepare obstacle geometry", err)
			}
			problem.Obstacles = append(problem.Obstacles, transformed)
		}
	}
	snapshot.Problem = problem

	startAt := time.Now()

	result, err := optimizer.Optimize(
		ctx,
		problem,
	)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return calculatedLayout{}, fmt.Errorf(
				"run nesting: %w",
				err,
			)
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return calculatedLayout{}, fmt.Errorf(
				"run nesting: %w",
				err,
			)
		}

		if errors.Is(err, geometry.ErrComplexity) {
			return calculatedLayout{}, apperror.Validation(err.Error(), apperror.NewFieldError("geometry", apperror.FieldCodeTooLong, "nesting complexity budget exceeded"))
		}
		return calculatedLayout{}, apperror.Internal(
			"failed to optimize nesting problem",
			err,
		)
	}

	duration := time.Since(startAt)
	preparedPlacements, err := validateResult(ctx, uc.geometryEngine, input.ProjectSurfaceID, problem, result)
	if err != nil {
		return calculatedLayout{}, apperror.Internal("optimizer returned an invalid layout", err)
	}

	return calculatedLayout{result: result, placements: preparedPlacements, duration: duration, snapshot: snapshot}, nil
}

func (uc *RunNestingUseCase) Preview(ctx context.Context, input RunNestingInput) (RunNestingOutput, error) {
	return uc.PreviewWithKey(ctx, input, "")
}

func (uc *RunNestingUseCase) PreviewWithKey(ctx context.Context, input RunNestingInput, operationKey string) (RunNestingOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, ExecutionTimeout)
	defer cancel()
	if input.PreviewID != "" {
		return RunNestingOutput{}, apperror.Validation("preview_id is only accepted when saving")
	}
	if operationKey != "" {
		if _, err := parseOperationKey(operationKey); err != nil {
			return RunNestingOutput{}, err
		}
	}
	release, err := uc.Limiter.Acquire(ctx)
	if err != nil {
		return RunNestingOutput{}, err
	}
	defer release()
	if operationKey != "" {
		if uc.Previews == nil {
			return RunNestingOutput{}, apperror.Internal("preview storage unavailable", nil)
		}
		record, found, err := uc.Previews.FindByKey(ctx, input, operationKey)
		if err != nil {
			return RunNestingOutput{}, err
		}
		if found {
			actual, err := Fingerprint(input)
			if err != nil {
				return RunNestingOutput{}, err
			}
			expected, err := Fingerprint(record.Data.Input)
			if err != nil {
				return RunNestingOutput{}, err
			}
			if actual != expected {
				return RunNestingOutput{}, apperror.Conflict("idempotency key already used for different preview parameters")
			}
			placements, err := validateResult(ctx, uc.geometryEngine, input.ProjectSurfaceID, record.Data.Problem, record.Data.Result)
			if err != nil {
				return RunNestingOutput{}, apperror.Internal("validate stored preview", err)
			}
			result := previewOutput(calculatedLayout{result: record.Data.Result, placements: placements})
			result.PreviewID = record.ID
			result.PreviewExpiresAt = record.ExpiresAt
			return result, nil
		}
	}
	return uc.preview(ctx, input, operationKey)
}

func (uc *RunNestingUseCase) preview(ctx context.Context, input RunNestingInput, operationKey string) (RunNestingOutput, error) {
	computed, err := uc.calculate(ctx, input)
	if err != nil {
		return RunNestingOutput{}, err
	}
	if err := ctx.Err(); err != nil {
		return RunNestingOutput{}, err
	}
	output := previewOutput(computed)
	if err := uc.retainPreview(ctx, input, computed, &output, operationKey); err != nil {
		return RunNestingOutput{}, err
	}
	return output, nil
}

func previewOutput(computed calculatedLayout) RunNestingOutput {
	output := RunNestingOutput{
		Placements: make([]PlacementItem, 0, len(computed.placements)),
		Unplaced:   computed.result.Unplaced, Metrics: computed.result.Metrics,
	}
	for _, item := range computed.placements {
		output.Placements = append(output.Placements, PlacementItem{
			ProjectSurfaceID: item.Placement.ProjectSurfaceID, ProjectPatternID: item.Placement.ProjectPatternID,
			X: item.Placement.X, Y: item.Placement.Y, Rotation: item.Placement.Rotation, Geometry: item.Geometry,
		})
	}
	return output
}

func (uc *RunNestingUseCase) execute(ctx context.Context, input RunNestingInput) (RunNestingOutput, error) {
	return uc.executeAndCommit(ctx, input, nil)
}

func (uc *RunNestingUseCase) executeAndCommit(ctx context.Context, input RunNestingInput, finish func(context.Context, RunNestingOutput) error) (RunNestingOutput, error) {
	if input.PreviewID != "" {
		preview, err := uc.loadPreview(ctx, input)
		if err != nil {
			return RunNestingOutput{}, err
		}
		computed := calculatedLayout{previewID: preview.ID, snapshot: preparedSnapshot{Hash: preview.Data.SnapshotHash, Problem: preview.Data.Problem}, result: preview.Data.Result, duration: preview.Data.Duration}
		return uc.commit(ctx, input, computed, finish)
	}
	computed, err := uc.calculate(ctx, input)
	if err != nil {
		return RunNestingOutput{}, err
	}
	return uc.commit(ctx, input, computed, finish)
}

func (uc *RunNestingUseCase) commit(ctx context.Context, input RunNestingInput, computed calculatedLayout, finish func(context.Context, RunNestingOutput) error) (RunNestingOutput, error) {
	// Geometry is checked outside the write transaction. Under the final lock,
	// matching the complete source hash proves these same inputs are still valid.
	placements, err := validateResult(ctx, uc.geometryEngine, input.ProjectSurfaceID, computed.snapshot.Problem, computed.result)
	if err != nil {
		return RunNestingOutput{}, apperror.Internal("revalidate nesting result", err)
	}
	computed.placements = placements
	return projecttransaction.Run(ctx, uc.Transactions, input.ProjectID, input.UserID, func(ctx context.Context) (RunNestingOutput, error) {
		if computed.previewID != "" {
			preview, err := uc.loadPreview(ctx, input)
			if err != nil {
				return RunNestingOutput{}, err
			}
			if preview.Saved != nil {
				if finish != nil {
					if err := finish(ctx, *preview.Saved); err != nil {
						return RunNestingOutput{}, err
					}
				}
				return *preview.Saved, nil
			}
		}
		current, err := uc.loadSnapshot(ctx, input)
		if err != nil {
			return RunNestingOutput{}, err
		}
		if !bytes.Equal(current.Hash[:], computed.snapshot.Hash[:]) {
			return RunNestingOutput{}, &apperror.AppError{Code: apperror.AppCodeStale, Message: "nesting source data changed; recalculate"}
		}

		output, err := uc.persist(ctx, input, computed)
		if err != nil {
			return RunNestingOutput{}, err
		}
		if computed.previewID != "" {
			if err := uc.Previews.MarkSaved(ctx, computed.previewID, output); err != nil {
				return RunNestingOutput{}, err
			}
		}
		if finish != nil {
			if err := finish(ctx, output); err != nil {
				return RunNestingOutput{}, err
			}
		}
		return output, nil
	})
}

func (uc *RunNestingUseCase) persist(ctx context.Context, input RunNestingInput, computed calculatedLayout) (RunNestingOutput, error) {
	var err error

	result := computed.result
	preparedPlacements := computed.placements

	parameters, err := json.Marshal(input)
	if err != nil {
		return RunNestingOutput{}, apperror.Internal("encode nesting history parameters", err)
	}
	nestingRun := nestingrun.NestingRun{
		Parameters:       parameters,
		ProjectSurfaceID: input.ProjectSurfaceID,
		Algorithm:        input.Algorithm,
		KeepExisting:     input.KeepExisting,

		RequestedCount: result.Metrics.RequestedCount,
		PlacedCount:    result.Metrics.PlacedCount,
		SurfaceArea:    result.Metrics.SurfaceArea,
		PlacedArea:     result.Metrics.PlacedArea,
		Utilization:    result.Metrics.Utilization,

		Duration: computed.duration,
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

			_, err := repositories.NestingRuns.Create(
				ctx,
				nestingRun,
			)
			if err != nil {
				return fmt.Errorf("create nesting run: %w", err)
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
