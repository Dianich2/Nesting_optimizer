package projectpattern

import (
	"context"
	"errors"
	domainprojectpattern "server_nesting_optimizer/internal/domain/project_pattern"
	"server_nesting_optimizer/pkg/apperror"
)

type GetProjectPatternByIDUseCase struct {
	repo ProjectPatternRepository
}

func NewGetProjectPatternByIDUseCase(
	repo ProjectPatternRepository,
) *GetProjectPatternByIDUseCase {
	return &GetProjectPatternByIDUseCase{
		repo: repo,
	}
}

func (uc *GetProjectPatternByIDUseCase) Execute(
	ctx context.Context,
	input GetProjectPatternByIDInput,
) (GetProjectPatternByIDOutput, error) {
	details := input.Validate()
	if len(details) > 0 {
		return GetProjectPatternByIDOutput{}, apperror.Validation(
			"validation failed",
			details...,
		)
	}

	projectPattern, err := uc.repo.GetByID(
		ctx,
		input.UserID,
		input.ProjectID,
		input.ProjectPatternID,
	)
	if err != nil {
		if errors.Is(err, domainprojectpattern.ErrNotFound) {
			return GetProjectPatternByIDOutput{}, apperror.NotFound(
				"project pattern not found",
			)
		}

		return GetProjectPatternByIDOutput{}, apperror.Internal(
			"failed to get project pattern by id",
			err,
		)
	}

	return toGetProjectPatternByIDOutput(projectPattern), nil
}
