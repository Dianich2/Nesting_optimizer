package pattern

import (
	"context"
	"errors"
	domainpattern "server_nesting_optimizer/internal/domain/pattern"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/pkg/apperror"
)

type UpdatePatternUseCase struct {
	repo           PatternRepository
	geometryEngine geometry.Engine
}

func NewUpdatePatternUseCase(
	repo PatternRepository,
	geometryEngine geometry.Engine,
) *UpdatePatternUseCase {
	return &UpdatePatternUseCase{
		repo:           repo,
		geometryEngine: geometryEngine,
	}
}

func (uc *UpdatePatternUseCase) Execute(
	ctx context.Context,
	input UpdatePatternInput,
) (UpdatePatternOutput, error) {
	input = normalizeUpdatePatternInput(input)
	details := input.Validate()
	if len(details) > 0 {
		return UpdatePatternOutput{}, apperror.Validation(
			"validation failed",
			details...,
		)
	}

	if input.Geometry != nil {
		normalizedGeometry, err := uc.geometryEngine.Normalize(*input.Geometry)
		if err != nil {
			if errors.Is(err, geometry.ErrInvalidPolygon) {
				return UpdatePatternOutput{}, apperror.Validation(
					"validation failed",
					apperror.NewFieldError(
						"geometry",
						apperror.FieldCodeInvalid,
						"invalid pattern geometry",
					),
				)
			}

			return UpdatePatternOutput{}, apperror.Internal(
				"failed to normalize pattern geometry",
				err,
			)
		}

		input.Geometry = &normalizedGeometry
	}

	updatedPattern, err := uc.repo.Update(
		ctx,
		input.PatternID,
		input.UserID,
		input.Name,
		input.Geometry,
	)
	if err != nil {
		if errors.Is(err, domainpattern.ErrNotFound) {
			return UpdatePatternOutput{}, apperror.NotFound(
				"pattern not found",
			)
		}

		return UpdatePatternOutput{}, apperror.Internal(
			"failed to update pattern",
			err,
		)
	}

	return toUpdatePatternOutput(updatedPattern), nil
}
