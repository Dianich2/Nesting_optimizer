package placement

import (
	"context"
	"errors"
	domainprojectsurface "server_nesting_optimizer/internal/domain/project_surface"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/pkg/apperror"
)

type ListPlacementsUseCase struct {
	placementRepo      PlacementRepository
	projectSurfaceRepo ProjectSurfaceRepository
	geometryEngine     geometry.Engine
}

func NewListPlacementsUseCase(
	placementRepo PlacementRepository,
	projectSurfaceRepo ProjectSurfaceRepository,
	geometryEngine geometry.Engine,
) *ListPlacementsUseCase {
	return &ListPlacementsUseCase{
		placementRepo:      placementRepo,
		projectSurfaceRepo: projectSurfaceRepo,
		geometryEngine:     geometryEngine,
	}
}

func (uc *ListPlacementsUseCase) Execute(
	ctx context.Context,
	input ListPlacementsInput,
) (ListPlacementsOutput, error) {
	details := input.Validate()
	if len(details) > 0 {
		return ListPlacementsOutput{}, apperror.Validation(
			"validation failed",
			details...,
		)
	}

	_, err := uc.projectSurfaceRepo.GetByID(
		ctx,
		input.UserID,
		input.ProjectID,
		input.ProjectSurfaceID,
	)
	if err != nil {
		if errors.Is(err, domainprojectsurface.ErrNotFound) {
			return ListPlacementsOutput{}, apperror.NotFound(
				"project surface not found",
			)
		}

		return ListPlacementsOutput{}, apperror.Internal(
			"failed to get project surface by id",
			err,
		)
	}

	listOfPlacements := ListPlacementsOutput{
		Items: make([]ListPlacementsItem, 0),
	}

	repoListOfPlacements, err := uc.placementRepo.ListPlacements(
		ctx,
		input.UserID,
		input.ProjectID,
		input.ProjectSurfaceID,
	)
	if err != nil {
		return ListPlacementsOutput{}, apperror.Internal(
			"failed to get placements",
			err,
		)
	}

	for _, placement := range repoListOfPlacements {
		curPlacement := ListPlacementsItem{
			ID:               placement.Placement.ID,
			ProjectSurfaceID: placement.Placement.ProjectSurfaceID,
			ProjectPatternID: placement.Placement.ProjectPatternID,
			X:                placement.Placement.X,
			Y:                placement.Placement.Y,
			Rotation:         placement.Placement.Rotation,
			CreatedAt:        placement.Placement.CreatedAt,
			UpdatedAt:        placement.Placement.UpdatedAt,
		}

		transformedGeometry, err := uc.geometryEngine.Transform(
			placement.PatternGeometry,
			placement.Placement.X,
			placement.Placement.Y,
			placement.Placement.Rotation,
		)
		if err != nil {
			return ListPlacementsOutput{}, apperror.Internal(
				"failed to transform placement geometry",
				err,
			)
		}
		curPlacement.Geometry = transformedGeometry

		listOfPlacements.Items = append(
			listOfPlacements.Items,
			curPlacement,
		)
	}

	return listOfPlacements, nil
}
