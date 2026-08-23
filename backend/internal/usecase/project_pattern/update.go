package projectpattern

import (
	"context"
	"errors"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	domainprojectpattern "server_nesting_optimizer/internal/domain/project_pattern"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/pkg/apperror"
)

type UpdateProjectPatternUseCase struct {
	repo           ProjectPatternRepository
	geometryEngine geometry.Engine
}

func NewUpdateProjectPatternUseCase(
	repo ProjectPatternRepository,
	geometryEngine geometry.Engine,
) *UpdateProjectPatternUseCase {
	return &UpdateProjectPatternUseCase{
		repo:           repo,
		geometryEngine: geometryEngine,
	}
}

func (uc *UpdateProjectPatternUseCase) Execute(
	ctx context.Context,
	input UpdateProjectPatternInput,
) (UpdateProjectPatternOutput, error) {
	input = normalizeUpdateProjectPatternInput(input)
	details := input.Validate()
	if len(details) > 0 {
		return UpdateProjectPatternOutput{}, apperror.Validation(
			"validation failed",
			details...,
		)
	}

	projectPattern, err := uc.repo.GetByID(
		ctx,
		input.UserID,
		input.ProjectID,
		input.ProjectPatternID,
	)
	if err != nil {
		if errors.Is(err, domainprojectpattern.ErrNotFound) {
			return UpdateProjectPatternOutput{}, apperror.NotFound(
				"project pattern not found",
			)
		}

		return UpdateProjectPatternOutput{}, apperror.Internal(
			"failed to retrieve project pattern",
			err,
		)
	}

	var scaledGeometry *domaingeometry.Polygon

	if input.Scale != nil {
		scaled, err := uc.geometryEngine.Scale(projectPattern.Geometry, *input.Scale)
		if err != nil {
			if errors.Is(err, geometry.ErrInvalidScale) {
				return UpdateProjectPatternOutput{}, apperror.Validation(
					"validation failed",
					apperror.NewFieldError(
						"scale",
						apperror.FieldCodeInvalid,
						"invalid project pattern scale",
					),
				)
			}

			if errors.Is(err, geometry.ErrInvalidPolygon) {
				return UpdateProjectPatternOutput{}, apperror.Validation(
					"validation failed",
					apperror.NewFieldError(
						"geometry",
						apperror.FieldCodeInvalid,
						"invalid project pattern geometry",
					),
				)
			}

			return UpdateProjectPatternOutput{}, apperror.Internal(
				"failed to scale project pattern geometry",
				err,
			)
		}

		scaledGeometry = &scaled
	}

	updatedPattern, err := uc.repo.Update(
		ctx,
		input.ProjectPatternID,
		input.ProjectID,
		input.UserID,
		input.Name,
		scaledGeometry,
	)
	if err != nil {
		if errors.Is(err, domainprojectpattern.ErrNotFound) {
			return UpdateProjectPatternOutput{}, apperror.NotFound(
				"project pattern not found",
			)
		}

		return UpdateProjectPatternOutput{}, apperror.Internal(
			"failed to update project pattern",
			err,
		)
	}

	return toUpdateProjectPatternOutput(updatedPattern), nil
}
