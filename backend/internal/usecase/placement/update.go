package placement

import (
	"context"
	"errors"
	domainplacement "server_nesting_optimizer/internal/domain/placement"
	domainprojectsurface "server_nesting_optimizer/internal/domain/project_surface"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/pkg/apperror"
)

type UpdatePlacementUseCase struct {
	placementRepo      PlacementRepository
	projectSurfaceRepo ProjectSurfaceRepository
	geometryEngine     geometry.Engine
}

func NewUpdatePlacementUseCase(
	placementRepo PlacementRepository,
	projectSurfaceRepo ProjectSurfaceRepository,
	geometryEngine geometry.Engine,
) *UpdatePlacementUseCase {
	return &UpdatePlacementUseCase{
		placementRepo:      placementRepo,
		projectSurfaceRepo: projectSurfaceRepo,
		geometryEngine:     geometryEngine,
	}
}

func (uc *UpdatePlacementUseCase) Execute(
	ctx context.Context,
	input UpdatePlacementInput,
) (UpdatePlacementOutput, error) {
	details := input.Validate()
	if len(details) > 0 {
		return UpdatePlacementOutput{}, apperror.Validation(
			"validation failed",
			details...,
		)
	}

	if input.Rotation != nil {
		*input.Rotation = geometry.NormalizeDegrees(*input.Rotation)
	}

	placement, err := uc.placementRepo.GetByIDWithPatternGeometry(
		ctx,
		input.UserID,
		input.ProjectID,
		input.PlacementID,
	)
	if err != nil {
		if errors.Is(err, domainplacement.ErrNotFound) {
			return UpdatePlacementOutput{}, apperror.NotFound(
				"placement not found",
			)
		}

		return UpdatePlacementOutput{}, apperror.Internal(
			"failed to get placement by id",
			err,
		)
	}

	if input.X != nil {
		placement.Placement.X = *input.X
	}

	if input.Y != nil {
		placement.Placement.Y = *input.Y
	}

	if input.Rotation != nil {
		placement.Placement.Rotation = *input.Rotation
	}

	projectSurface, err := uc.projectSurfaceRepo.GetByID(
		ctx,
		input.UserID,
		input.ProjectID,
		placement.Placement.ProjectSurfaceID,
	)
	if err != nil {
		if errors.Is(err, domainprojectsurface.ErrNotFound) {
			return UpdatePlacementOutput{}, apperror.NotFound(
				"project surface not found",
			)
		}

		return UpdatePlacementOutput{}, apperror.Internal(
			"failed to get project surface by id",
			err,
		)
	}

	transformedGeometry, err := uc.geometryEngine.Transform(
		placement.PatternGeometry,
		placement.Placement.X,
		placement.Placement.Y,
		placement.Placement.Rotation,
	)
	if err != nil {
		return UpdatePlacementOutput{}, apperror.Internal(
			"failed to transform placement geometry",
			err,
		)
	}

	isCoveredBy, err := uc.geometryEngine.CoveredBy(
		transformedGeometry,
		projectSurface.Geometry,
	)
	if err != nil {
		return UpdatePlacementOutput{}, apperror.Internal(
			"failed to check whether pattern is covered by surface",
			err,
		)
	}

	if !isCoveredBy {
		return UpdatePlacementOutput{}, apperror.Conflict(
			"placement is outside project surface",
		)
	}

	existingPlacements, err := uc.placementRepo.ListForCollisionCheckExcluding(
		ctx,
		projectSurface.ID,
		input.ProjectID,
		input.UserID,
		input.PlacementID,
	)
	if err != nil {
		return UpdatePlacementOutput{}, apperror.Internal(
			"failed to list placements for collision check",
			err,
		)
	}

	for _, existing := range existingPlacements {
		existingGeometry, err := uc.geometryEngine.Transform(
			existing.PatternGeometry,
			existing.Placement.X,
			existing.Placement.Y,
			existing.Placement.Rotation,
		)
		if err != nil {
			return UpdatePlacementOutput{}, apperror.Internal(
				"failed to transform existing placement geometry",
				err,
			)
		}

		intersects, err := uc.geometryEngine.InteriorsIntersect(
			transformedGeometry,
			existingGeometry,
		)
		if err != nil {
			return UpdatePlacementOutput{}, apperror.Internal(
				"failed to check placement intersection",
				err,
			)
		}

		if intersects {
			return UpdatePlacementOutput{}, apperror.Conflict(
				"placement overlaps another placement",
			)
		}
	}

	placementNew := domainplacement.Placement{
		ID:               input.PlacementID,
		ProjectSurfaceID: projectSurface.ID,
		ProjectPatternID: placement.Placement.ProjectPatternID,
		X:                placement.Placement.X,
		Y:                placement.Placement.Y,
		Rotation:         placement.Placement.Rotation,
	}

	updatedPlacement, err := uc.placementRepo.Update(
		ctx,
		placementNew,
		input.ProjectID,
		input.UserID,
	)
	if err != nil {
		if errors.Is(err, domainplacement.ErrNotFound) {
			return UpdatePlacementOutput{}, apperror.NotFound(
				"project surface or project pattern not found",
			)
		}

		return UpdatePlacementOutput{}, apperror.Internal(
			"failed to update placement",
			err,
		)
	}

	return toUpdatePlacementOutput(updatedPlacement, transformedGeometry), nil
}
