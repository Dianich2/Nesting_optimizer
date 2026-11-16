package project

import (
	"context"
	"errors"
	domainproject "server_nesting_optimizer/internal/domain/project"
	projecttransaction "server_nesting_optimizer/internal/usecase/project_transaction"
	"server_nesting_optimizer/pkg/apperror"
)

type DeleteProjectUseCase struct {
	Transactions projecttransaction.Executor
	repo         ProjectRepository
}

func NewDeleteProjectUseCase(
	repo ProjectRepository,
) *DeleteProjectUseCase {
	return &DeleteProjectUseCase{
		repo: repo,
	}
}

func (uc *DeleteProjectUseCase) Execute(
	ctx context.Context,
	input DeleteProjectInput,
) error {
	_, err := projecttransaction.Run(ctx, uc.Transactions, input.ProjectID, input.UserID, func(ctx context.Context) (struct{}, error) { return struct{}{}, uc.execute(ctx, input) })
	return err
}

func (uc *DeleteProjectUseCase) execute(
	ctx context.Context,
	input DeleteProjectInput,
) error {
	details := input.Validate()
	if len(details) > 0 {
		return apperror.Validation(
			"validation failed",
			details...,
		)
	}

	err := uc.repo.Delete(
		ctx,
		input.ProjectID,
		input.UserID,
	)
	if err != nil {
		if errors.Is(err, domainproject.ErrNotFound) {
			return apperror.NotFound(
				"project not found",
			)
		}

		return apperror.Internal(
			"failed to delete project",
			err,
		)
	}

	return nil
}
