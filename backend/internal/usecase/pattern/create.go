package pattern

import (
	"context"
	"errors"
	domainpattern "server_nesting_optimizer/internal/domain/pattern"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/pkg/apperror"
)

type CreatePatternUseCase struct {
	repo           PatternRepository
	geometryEngine geometry.Engine
}

func NewCreatePatternUseCase(
	repo PatternRepository,
	geometryEngine geometry.Engine,
) *CreatePatternUseCase {
	return &CreatePatternUseCase{
		repo:           repo,
		geometryEngine: geometryEngine,
	}
}

func (uc *CreatePatternUseCase) Execute(
	ctx context.Context,
	input CreatePatternInput,
) (CreatePatternOutput, error) {
	input = normalizeCreatePatternInput(input)
	details := input.Validate()
	if len(details) > 0 {
		return CreatePatternOutput{}, apperror.Validation(
			"validation failed",
			details...,
		)
	}

	normalizedGeometry, err := uc.geometryEngine.Normalize(input.Geometry)
	if err != nil {
		if errors.Is(err, geometry.ErrInvalidPolygon) {
			return CreatePatternOutput{}, apperror.Validation(
				"validation failed",
				apperror.NewFieldError(
					"geometry",
					apperror.FieldCodeInvalid,
					"invalid pattern geometry",
				),
			)
		}

		return CreatePatternOutput{}, apperror.Internal(
			"failed to normalize pattern geometry",
			err,
		)
	}

	input.Geometry = normalizedGeometry
	domainPattern := toPattern(input)

	createdPattern, err := uc.repo.Create(
		ctx,
		domainPattern,
	)
	if err != nil {
		if errors.Is(err, domainpattern.ErrOwnerNotFound) {
			return CreatePatternOutput{}, apperror.Unauthorized(
				"user is not active",
			)
		}

		return CreatePatternOutput{}, apperror.Internal(
			"failed to create pattern",
			err,
		)
	}

	return toCreatePatternOutput(createdPattern), nil
}
