package surface

import (
	"context"
	"errors"
	domainsurface "server_nesting_optimizer/internal/domain/surface"
	"server_nesting_optimizer/pkg/apperror"
)

type DeleteSurfaceUseCase struct {
	repo SurfaceRepository
}

func NewDeleteSurfaceUseCase(
	repo SurfaceRepository,
) *DeleteSurfaceUseCase {
	return &DeleteSurfaceUseCase{
		repo: repo,
	}
}

func (uc *DeleteSurfaceUseCase) Execute(
	ctx context.Context,
	input DeleteSurfaceInput,
) error {
	details := input.Validate()
	if len(details) > 0 {
		return apperror.Validation(
			"validation failed",
			details...,
		)
	}

	err := uc.repo.SoftDelete(
		ctx,
		input.SurfaceID,
		input.UserID,
	)
	if err != nil {
		if errors.Is(err, domainsurface.ErrNotFound) {
			return apperror.NotFound(
				"surface not found",
			)
		}

		return apperror.Internal(
			"failed to soft delete surface",
			err,
		)
	}

	return nil
}
