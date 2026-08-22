package surface

import (
	"context"
	"errors"
	domainsurface "server_nesting_optimizer/internal/domain/surface"
	"server_nesting_optimizer/pkg/apperror"
)

type GetSurfaceByIDUseCase struct {
	repo SurfaceRepository
}

func NewGetSurfaceByIDUseCase(
	repo SurfaceRepository,
) *GetSurfaceByIDUseCase {
	return &GetSurfaceByIDUseCase{
		repo: repo,
	}
}

func (uc *GetSurfaceByIDUseCase) Execute(
	ctx context.Context,
	input GetSurfaceByIDInput,
) (GetSurfaceByIDOutput, error) {
	details := input.Validate()
	if len(details) > 0 {
		return GetSurfaceByIDOutput{}, apperror.Validation(
			"validation failed",
			details...,
		)
	}

	surface, err := uc.repo.GetByID(
		ctx,
		input.SurfaceID,
		input.UserID,
	)
	if err != nil {
		if errors.Is(err, domainsurface.ErrNotFound) {
			return GetSurfaceByIDOutput{}, apperror.NotFound(
				"surface not found",
			)
		}

		return GetSurfaceByIDOutput{}, apperror.Internal(
			"failed to get surface by id",
			err,
		)
	}

	return toGetSurfaceByIDOutput(surface), nil
}
