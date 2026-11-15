package projectpattern

import (
	"context"
	"errors"
	domainprojectpattern "server_nesting_optimizer/internal/domain/project_pattern"
	projecttransaction "server_nesting_optimizer/internal/usecase/project_transaction"
	"server_nesting_optimizer/pkg/apperror"
)

type DeleteProjectPatternUseCase struct {
	Transactions projecttransaction.Executor
	repo         ProjectPatternRepository
}

func NewDeleteProjectPatternUseCase(
	repo ProjectPatternRepository,
) *DeleteProjectPatternUseCase {
	return &DeleteProjectPatternUseCase{
		repo: repo,
	}
}

func (uc *DeleteProjectPatternUseCase) Execute(
	ctx context.Context,
	input DeleteProjectPatternInput,
) error {
	_, err := projecttransaction.Run(ctx, uc.Transactions, input.ProjectID, input.UserID, func(ctx context.Context) (struct{}, error) { return struct{}{}, uc.execute(ctx, input) })
	return err
}

func (uc *DeleteProjectPatternUseCase) execute(
	ctx context.Context,
	input DeleteProjectPatternInput,
) error {
	details := input.Validate()
	if len(details) > 0 {
		return apperror.Validation(
			"validation failed",
			details...,
		)
	}

	hasPlacements, checkErr := uc.repo.HasActivePlacements(ctx, input.ProjectPatternID, input.ProjectID, input.UserID)
	if checkErr != nil {
		return apperror.Internal("failed to check active placements", checkErr)
	}
	if hasPlacements {
		return apperror.Conflict("project pattern is used by active placements")
	}

	err := uc.repo.SoftDelete(
		ctx,
		input.ProjectPatternID,
		input.ProjectID,
		input.UserID,
	)
	if err != nil {
		if errors.Is(err, domainprojectpattern.ErrNotFound) {
			return apperror.NotFound(
				"project pattern not found",
			)
		}

		return apperror.Internal(
			"failed to soft delete project pattern",
			err,
		)
	}

	return nil
}
