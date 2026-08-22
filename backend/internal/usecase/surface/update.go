package surface

import (
	"context"
	"errors"
	domainsurface "server_nesting_optimizer/internal/domain/surface"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/pkg/apperror"
)

type UpdateSurfaceUseCase struct {
	repo           SurfaceRepository
	geometryEngine geometry.Engine
}

func NewUpdateSurfaceUseCase(
	repo SurfaceRepository,
	geometryEngine geometry.Engine,
) *UpdateSurfaceUseCase {
	return &UpdateSurfaceUseCase{
		repo:           repo,
		geometryEngine: geometryEngine,
	}
}

func (uc *UpdateSurfaceUseCase) Execute(
	ctx context.Context,
	input UpdateSurfaceInput,
) (UpdateSurfaceOutput, error) {
	input = normalizeUpdateSurfaceInput(input)
	details := input.Validate()
	if len(details) > 0 {
		return UpdateSurfaceOutput{}, apperror.Validation(
			"validation failed",
			details...,
		)
	}

	if input.Geometry != nil {
		normalizedGeometry, err := uc.geometryEngine.Normalize(*input.Geometry)
		if err != nil {
			if errors.Is(err, geometry.ErrInvalidPolygon) {
				return UpdateSurfaceOutput{}, apperror.Validation(
					"validation failed",
					apperror.NewFieldError(
						"geometry",
						apperror.FieldCodeInvalid,
						"invalid surface geometry",
					),
				)
			}

			return UpdateSurfaceOutput{}, apperror.Internal(
				"failed to normalize surface geometry",
				err,
			)
		}

		input.Geometry = &normalizedGeometry
	}

	updatedSurface, err := uc.repo.Update(
		ctx,
		input.SurfaceID,
		input.UserID,
		input.Name,
		input.Geometry,
	)
	if err != nil {
		if errors.Is(err, domainsurface.ErrNotFound) {
			return UpdateSurfaceOutput{}, apperror.NotFound(
				"surface not found",
			)
		}

		return UpdateSurfaceOutput{}, apperror.Internal(
			"failed to update surface",
			err,
		)
	}

	return toUpdateSurfaceOutput(updatedSurface), nil
}
