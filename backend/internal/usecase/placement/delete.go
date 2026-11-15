package placement

import (
	"context"
	"errors"
	domainplacement "server_nesting_optimizer/internal/domain/placement"
	projecttransaction "server_nesting_optimizer/internal/usecase/project_transaction"
	"server_nesting_optimizer/pkg/apperror"
)

type DeletePlacementUseCase struct {
	Transactions projecttransaction.Executor
	repo         PlacementRepository
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
	_, err := projecttransaction.Run(ctx, uc.Transactions, input.ProjectID, input.UserID, func(ctx context.Context) (struct{}, error) { return struct{}{}, uc.execute(ctx, input) })
	return err
}

func (uc *DeletePlacementUseCase) execute(
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
