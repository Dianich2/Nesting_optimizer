package projectsurface

import (
	"context"
	"errors"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	domainprojectsurface "server_nesting_optimizer/internal/domain/project_surface"
	"server_nesting_optimizer/internal/geometry"
	projecttransaction "server_nesting_optimizer/internal/usecase/project_transaction"
	"server_nesting_optimizer/pkg/apperror"
)

type UpdateProjectSurfaceUseCase struct {
	Transactions   projecttransaction.Executor
	repo           ProjectSurfaceRepository
	geometryEngine geometry.Engine
}

func NewUpdateProjectSurfaceUseCase(
	repo ProjectSurfaceRepository,
	geometryEngine geometry.Engine,
) *UpdateProjectSurfaceUseCase {
	return &UpdateProjectSurfaceUseCase{
		repo:           repo,
		geometryEngine: geometryEngine,
	}
}

func (uc *UpdateProjectSurfaceUseCase) Execute(
	ctx context.Context,
	input UpdateProjectSurfaceInput,
) (UpdateProjectSurfaceOutput, error) {
	return projecttransaction.Run(ctx, uc.Transactions, input.ProjectID, input.UserID, func(ctx context.Context) (UpdateProjectSurfaceOutput, error) {
		return uc.execute(ctx, input)
	})
}

func (uc *UpdateProjectSurfaceUseCase) execute(
	ctx context.Context,
	input UpdateProjectSurfaceInput,
) (UpdateProjectSurfaceOutput, error) {
	input = normalizeUpdateProjectSurfaceInput(input)
	details := input.Validate()
	if len(details) > 0 {
		return UpdateProjectSurfaceOutput{}, apperror.Validation(
			"validation failed",
			details...,
		)
	}

	projectSurface, err := uc.repo.GetByID(
		ctx,
		input.UserID,
		input.ProjectID,
		input.ProjectSurfaceID,
	)
	if err != nil {
		if errors.Is(err, domainprojectsurface.ErrNotFound) {
			return UpdateProjectSurfaceOutput{}, apperror.NotFound(
				"project surface not found",
			)
		}

		return UpdateProjectSurfaceOutput{}, apperror.Internal(
			"failed to retrieve project surface",
			err,
		)
	}

	var scaledGeometry *domaingeometry.Polygon

	if input.Scale != nil || input.Geometry != nil {
		hasPlacements, err := uc.repo.HasActivePlacements(
			ctx,
			input.ProjectSurfaceID,
			input.ProjectID,
			input.UserID,
		)
		if err != nil {
			return UpdateProjectSurfaceOutput{}, apperror.Internal(
				"failed to check project surface placements",
				err,
			)
		}

		if hasPlacements {
			return UpdateProjectSurfaceOutput{}, apperror.Conflict(
				"project surface is used by active placements",
			)
		}

		var scaled domaingeometry.Polygon
		if input.Geometry != nil {
			scaled, err = uc.geometryEngine.Normalize(*input.Geometry)
		} else {
			scaled, err = uc.geometryEngine.Scale(projectSurface.Geometry, *input.Scale)
		}

		if err != nil {
			if errors.Is(err, geometry.ErrInvalidScale) {
				return UpdateProjectSurfaceOutput{}, apperror.Validation(
					"validation failed",
					apperror.NewFieldError(
						"scale",
						apperror.FieldCodeInvalid,
						"invalid project surface scale",
					),
				)
			}

			if errors.Is(err, geometry.ErrInvalidPolygon) {
				return UpdateProjectSurfaceOutput{}, apperror.Validation(
					"validation failed",
					apperror.NewFieldError(
						"geometry",
						apperror.FieldCodeInvalid,
						"invalid project surface geometry",
					),
				)
			}

			return UpdateProjectSurfaceOutput{}, apperror.Internal(
				"failed to scale project surface geometry",
				err,
			)
		}

		scaledGeometry = &scaled
	}

	updatedSurface, err := uc.repo.Update(
		ctx,
		input.ProjectSurfaceID,
		input.ProjectID,
		input.UserID,
		input.Name,
		scaledGeometry,
	)
	if err != nil {
		if errors.Is(err, domainprojectsurface.ErrNotFound) {
			return UpdateProjectSurfaceOutput{}, apperror.NotFound(
				"project surface not found",
			)
		}

		return UpdateProjectSurfaceOutput{}, apperror.Internal(
			"failed to update project surface",
			err,
		)
	}

	return toUpdateProjectSurfaceOutput(updatedSurface), nil
}
