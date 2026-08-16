package surface

import (
	"context"
	"errors"
	domainsurface "server_nesting_optimizer/internal/domain/surface"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/pkg/apperror"
)

type CreateSurfaceUseCase struct {
	repo           SurfaceRepository
	geometryEngine geometry.Engine
}

func NewCreateSurfaceUseCase(
	repo SurfaceRepository,
	geometryEngine geometry.Engine,
) *CreateSurfaceUseCase {
	return &CreateSurfaceUseCase{
		repo:           repo,
		geometryEngine: geometryEngine,
	}
}

func (uc *CreateSurfaceUseCase) Execute(
	ctx context.Context,
	input CreateSurfaceInput,
) (CreateSurfaceOutput, error) {
	input = normalizeCreateSurfaceInput(input)
	details := input.Validate()
	if len(details) > 0 {
		return CreateSurfaceOutput{}, apperror.Validation(
			"validation failed",
			details...,
		)
	}

	normalizedGeometry, err := uc.geometryEngine.Normalize(input.Geometry)
	if err != nil {
		if errors.Is(err, geometry.ErrInvalidPolygon) {
			return CreateSurfaceOutput{}, apperror.Validation(
				"validation failed",
				apperror.NewFieldError(
					"geometry",
					apperror.FieldCodeInvalid,
					"invalid surface geometry",
				),
			)
		}

		return CreateSurfaceOutput{}, apperror.Internal(
			"failed to normalize surface geometry",
			err,
		)
	}

	input.Geometry = normalizedGeometry
	domainSurface := toSurface(input)

	createdSurface, err := uc.repo.Create(
		ctx,
		domainSurface,
	)
	if err != nil {
		if errors.Is(err, domainsurface.ErrOwnerNotFound) {
			return CreateSurfaceOutput{}, apperror.Unauthorized(
				"user is not active",
			)
		}

		return CreateSurfaceOutput{}, apperror.Internal(
			"failed to create surface",
			err,
		)
	}

	return toCreateSurfaceOutput(createdSurface), nil
}
