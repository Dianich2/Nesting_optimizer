package project

import (
	"context"
	"errors"
	domainproject "server_nesting_optimizer/internal/domain/project"
	"server_nesting_optimizer/pkg/apperror"
)

type GetProjectByIDUseCase struct {
	repo ProjectRepository
}

func NewGetProjectByIDUseCase(
	repo ProjectRepository,
) *GetProjectByIDUseCase {
	return &GetProjectByIDUseCase{
		repo: repo,
	}
}

func (uc *GetProjectByIDUseCase) Execute(
	ctx context.Context,
	input GetProjectByIDInput,
) (GetProjectByIDOutput, error) {
	details := input.Validate()
	if len(details) > 0 {
		return GetProjectByIDOutput{}, apperror.Validation(
			"validation failed",
			details...,
		)
	}

	project, err := uc.repo.GetByID(
		ctx,
		input.ProjectID,
		input.UserID,
	)
	if err != nil {
		if errors.Is(err, domainproject.ErrNotFound) {
			return GetProjectByIDOutput{}, apperror.NotFound(
				"project not found",
			)
		}

		return GetProjectByIDOutput{}, apperror.Internal(
			"failed to get project by id",
			err,
		)
	}

	return toGetProjectByIDOutput(project), nil
}
