package projectpattern

import (
	"context"
	"errors"
	domainprojectpattern "server_nesting_optimizer/internal/domain/project_pattern"
	"server_nesting_optimizer/pkg/apperror"
)

type DeleteProjectPatternUseCase struct {
	repo ProjectPatternRepository
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
	details := input.Validate()
	if len(details) > 0 {
		return apperror.Validation(
			"validation failed",
			details...,
		)
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
