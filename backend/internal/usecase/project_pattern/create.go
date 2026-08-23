package projectpattern

import (
	"context"
	"errors"
	domainpattern "server_nesting_optimizer/internal/domain/pattern"
	domainproject "server_nesting_optimizer/internal/domain/project"
	domainprojectpattern "server_nesting_optimizer/internal/domain/project_pattern"
	"server_nesting_optimizer/internal/geometry"
	patternusecase "server_nesting_optimizer/internal/usecase/pattern"
	projectusecase "server_nesting_optimizer/internal/usecase/project"
	"server_nesting_optimizer/pkg/apperror"
)

type CreateProjectPatternUseCase struct {
	projectPatternRepo ProjectPatternRepository
	projectRepo        projectusecase.ProjectRepository
	patternRepo        patternusecase.PatternRepository
	geometryEngine     geometry.Engine
}

func NewCreateProjectPatternUseCase(
	projectPatternRepo ProjectPatternRepository,
	projectRepo projectusecase.ProjectRepository,
	patternRepo patternusecase.PatternRepository,
	geometryEngine geometry.Engine,
) *CreateProjectPatternUseCase {
	return &CreateProjectPatternUseCase{
		projectPatternRepo: projectPatternRepo,
		projectRepo:        projectRepo,
		patternRepo:        patternRepo,
		geometryEngine:     geometryEngine,
	}
}

func (uc *CreateProjectPatternUseCase) Execute(
	ctx context.Context,
	input CreateProjectPatternInput,
) (CreateProjectPatternOutput, error) {
	details := input.Validate()
	if len(details) > 0 {
		return CreateProjectPatternOutput{}, apperror.Validation(
			"validation failed",
			details...,
		)
	}

	_, err := uc.projectRepo.GetByID(
		ctx,
		input.ProjectID,
		input.UserID,
	)
	if err != nil {
		if errors.Is(err, domainproject.ErrNotFound) {
			return CreateProjectPatternOutput{}, apperror.NotFound(
				"project not found",
			)
		}

		return CreateProjectPatternOutput{}, apperror.Internal(
			"failed to get project by id",
			err,
		)
	}

	pattern, err := uc.patternRepo.GetByID(
		ctx,
		input.SourcePatternID,
		input.UserID,
	)
	if err != nil {
		if errors.Is(err, domainpattern.ErrNotFound) {
			return CreateProjectPatternOutput{}, apperror.NotFound(
				"pattern not found",
			)
		}

		return CreateProjectPatternOutput{}, apperror.Internal(
			"failed to get pattern by id",
			err,
		)
	}

	scaledGeometry, err := uc.geometryEngine.Scale(pattern.Geometry, input.Scale)
	if err != nil {
		if errors.Is(err, geometry.ErrInvalidScale) {
			return CreateProjectPatternOutput{}, apperror.Validation(
				"validation failed",
				apperror.NewFieldError(
					"scale",
					apperror.FieldCodeInvalid,
					"invalid pattern scale",
				),
			)
		}

		if errors.Is(err, geometry.ErrInvalidPolygon) {
			return CreateProjectPatternOutput{}, apperror.Validation(
				"validation failed",
				apperror.NewFieldError(
					"geometry",
					apperror.FieldCodeInvalid,
					"invalid pattern geometry",
				),
			)
		}

		return CreateProjectPatternOutput{}, apperror.Internal(
			"failed to scale pattern geometry",
			err,
		)
	}

	domainProjectPattern := domainprojectpattern.ProjectPattern{
		ProjectID:       input.ProjectID,
		SourcePatternID: &input.SourcePatternID,
		Name:            pattern.Name,
		Geometry:        scaledGeometry,
	}

	createdProjectPattern, err := uc.projectPatternRepo.Create(
		ctx,
		domainProjectPattern,
		input.UserID,
	)
	if err != nil {
		return CreateProjectPatternOutput{}, apperror.Internal(
			"failed to create project pattern",
			err,
		)
	}

	return toCreateProjectPatternOutput(createdProjectPattern), nil
}
