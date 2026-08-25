package placement

import (
	"context"
	"errors"
	domainplacement "server_nesting_optimizer/internal/domain/placement"
	domainprojectpattern "server_nesting_optimizer/internal/domain/project_pattern"
	domainprojectsurface "server_nesting_optimizer/internal/domain/project_surface"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/pkg/apperror"
)

type CreatePlacementUseCase struct {
	placementRepo      PlacementRepository
	projectSurfaceRepo ProjectSurfaceRepository
	projectPatternRepo ProjectPatternRepository
	geometryEngine     geometry.Engine
}

func NewCreatePlacementUseCase(
	placementRepo PlacementRepository,
	projectSurfaceRepo ProjectSurfaceRepository,
	projectPatternRepo ProjectPatternRepository,
	geometryEngine geometry.Engine,
) *CreatePlacementUseCase {
	return &CreatePlacementUseCase{
		placementRepo:      placementRepo,
		projectSurfaceRepo: projectSurfaceRepo,
		projectPatternRepo: projectPatternRepo,
		geometryEngine:     geometryEngine,
	}
}

func (uc *CreatePlacementUseCase) Execute(
	ctx context.Context,
	input CreatePlacementInput,
) (CreatePlacementOutput, error) {
	details := input.Validate()
	if len(details) > 0 {
		return CreatePlacementOutput{}, apperror.Validation(
			"validation failed",
			details...,
		)
	}

	input.Rotation = NormalizeRotation(input.Rotation)

	projectSurface, err := uc.projectSurfaceRepo.GetByID(
		ctx,
		input.UserID,
		input.ProjectID,
		input.ProjectSurfaceID,
	)
	if err != nil {
		if errors.Is(err, domainprojectsurface.ErrNotFound) {
			return CreatePlacementOutput{}, apperror.NotFound(
				"project surface not found",
			)
		}

		return CreatePlacementOutput{}, apperror.Internal(
			"failed to get project surface by id",
			err,
		)
	}

	projectPattern, err := uc.projectPatternRepo.GetByID(
		ctx,
		input.UserID,
		input.ProjectID,
		input.ProjectPatternID,
	)
	if err != nil {
		if errors.Is(err, domainprojectpattern.ErrNotFound) {
			return CreatePlacementOutput{}, apperror.NotFound(
				"project pattern not found",
			)
		}

		return CreatePlacementOutput{}, apperror.Internal(
			"failed to get project pattern by id",
			err,
		)
	}

	transformedGeometry, err := uc.geometryEngine.Transform(
		projectPattern.Geometry,
		input.X,
		input.Y,
		input.Rotation,
	)
	if err != nil {
		if errors.Is(err, geometry.ErrInvalidTransform) {
			return CreatePlacementOutput{}, apperror.Validation(
				"validation failed",
				apperror.NewFieldError(
					"transform",
					apperror.FieldCodeInvalid,
					"invalid pattern transform",
				),
			)
		}

		return CreatePlacementOutput{}, apperror.Internal(
			"failed to transform pattern geometry",
			err,
		)
	}

	isCoveredBy, err := uc.geometryEngine.CoveredBy(
		transformedGeometry,
		projectSurface.Geometry,
	)
	if err != nil {
		return CreatePlacementOutput{}, apperror.Internal(
			"failed to check whether pattern is covered by surface",
			err,
		)
	}

	if !isCoveredBy {
		return CreatePlacementOutput{}, apperror.Conflict(
			"placement is outside project surface",
		)
	}

	existingPlacements, err := uc.placementRepo.ListForCollisionCheck(
		ctx,
		input.ProjectSurfaceID,
		input.ProjectID,
		input.UserID,
	)
	if err != nil {
		return CreatePlacementOutput{}, apperror.Internal(
			"failed to list placements for collision check",
			err,
		)
	}

	for _, existing := range existingPlacements {
		existingGeometry, err := uc.geometryEngine.Transform(
			existing.PatternGeometry,
			existing.X,
			existing.Y,
			existing.Rotation,
		)
		if err != nil {
			return CreatePlacementOutput{}, apperror.Internal(
				"failed to transform existing placement geometry",
				err,
			)
		}

		intersects, err := uc.geometryEngine.InteriorsIntersect(
			transformedGeometry,
			existingGeometry,
		)
		if err != nil {
			return CreatePlacementOutput{}, apperror.Internal(
				"failed to check placement intersection",
				err,
			)
		}

		if intersects {
			return CreatePlacementOutput{}, apperror.Conflict(
				"placement overlaps another placement",
			)
		}
	}

	placement := domainplacement.Placement{
		ProjectSurfaceID: input.ProjectSurfaceID,
		ProjectPatternID: input.ProjectPatternID,
		X:                input.X,
		Y:                input.Y,
		Rotation:         input.Rotation,
	}

	createdPlacement, err := uc.placementRepo.Create(
		ctx,
		placement,
		input.ProjectID,
		input.UserID,
	)
	if err != nil {
		if errors.Is(err, domainplacement.ErrNotFound) {
			return CreatePlacementOutput{}, apperror.NotFound(
				"project surface or project pattern not found",
			)
		}

		return CreatePlacementOutput{}, apperror.Internal(
			"failed to create placement",
			err,
		)
	}

	return toCreatePlacementOutput(createdPlacement, transformedGeometry), nil
}
