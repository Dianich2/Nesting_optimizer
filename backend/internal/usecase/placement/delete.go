package placement

import (
	"context"
	"errors"
	domainplacement "server_nesting_optimizer/internal/domain/placement"
	"server_nesting_optimizer/pkg/apperror"
)

type DeletePlacementUseCase struct {
	repo PlacementRepository
}

func NewDeletePlacementUseCase(
	repo PlacementRepository,
) *DeletePlacementUseCase {
	return &DeletePlacementUseCase{
		repo: repo,
	}
}

func (uc *DeletePlacementUseCase) Execute(
	ctx context.Context,
	input DeletePlacementInput,
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
		input.PlacementID,
		input.ProjectID,
		input.UserID,
	)
	if err != nil {
		if errors.Is(err, domainplacement.ErrNotFound) {
			return apperror.NotFound(
				"placement not found",
			)
		}

		return apperror.Internal(
			"failed to soft delete placement",
			err,
		)
	}

	return nil
}
