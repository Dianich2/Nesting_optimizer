package placement

import (
	"context"
	"errors"
	domainplacement "server_nesting_optimizer/internal/domain/placement"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/pkg/apperror"
)

type GetPlacementByIDUseCase struct {
	repo           PlacementRepository
	geometryEngine geometry.Engine
}

func NewGetPlacementByIDUseCase(
	repo PlacementRepository,
	geometryEngine geometry.Engine,
) *GetPlacementByIDUseCase {
	return &GetPlacementByIDUseCase{
		repo:           repo,
		geometryEngine: geometryEngine,
	}
}

func (uc *GetPlacementByIDUseCase) Execute(
	ctx context.Context,
	input GetPlacementByIDInput,
) (GetPlacementByIDOutput, error) {
	details := input.Validate()
	if len(details) > 0 {
		return GetPlacementByIDOutput{}, apperror.Validation(
			"validation failed",
			details...,
		)
	}

	placement, err := uc.repo.GetByIDWithPatternGeometry(
		ctx,
		input.UserID,
		input.ProjectID,
		input.PlacementID,
	)
	if err != nil {
		if errors.Is(err, domainplacement.ErrNotFound) {
			return GetPlacementByIDOutput{}, apperror.NotFound(
				"placement not found",
			)
		}

		return GetPlacementByIDOutput{}, apperror.Internal(
			"failed to get placement by id",
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
		return GetPlacementByIDOutput{}, apperror.Internal(
			"failed to transform placement geometry",
			err,
		)
	}

	return toGetPlacementByIDOutput(placement, transformedGeometry), nil
}
