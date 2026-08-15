package project

import (
	"context"
	"errors"
	domainproject "server_nesting_optimizer/internal/domain/project"
	"server_nesting_optimizer/pkg/apperror"
)

type DeleteProjectUseCase struct {
	repo ProjectRepository
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
	details := input.Validate()
	if len(details) > 0 {
		return apperror.Validation(
			"validation failed",
			details...,
		)
	}

	err := uc.repo.SoftDelete(
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
			"failed to soft delete project",
			err,
		)
	}

	return nil
}
