package project

import (
	"context"
	"errors"
	domainproject "server_nesting_optimizer/internal/domain/project"
	"server_nesting_optimizer/pkg/apperror"
)

type UpdateProjectUseCase struct {
	repo ProjectRepository
}

func NewUpdateProjectUseCase(
	repo ProjectRepository,
) *UpdateProjectUseCase {
	return &UpdateProjectUseCase{
		repo: repo,
	}
}

func (uc *UpdateProjectUseCase) Execute(
	ctx context.Context,
	input UpdateProjectInput,
) (UpdateProjectOutput, error) {
	input = normalizeUpdateProjectInput(input)
	details := input.Validate()
	if len(details) > 0 {
		return UpdateProjectOutput{}, apperror.Validation(
			"validation failed",
			details...,
		)
	}

	updatedProject, err := uc.repo.Update(
		ctx,
		input.ProjectID,
		input.UserID,
		input.Name,
		input.Description,
	)
	if err != nil {
		if errors.Is(err, domainproject.ErrNotFound) {
			return UpdateProjectOutput{}, apperror.NotFound(
				"project not found",
			)
		}

		return UpdateProjectOutput{}, apperror.Internal(
			"failed to update project",
			err,
		)
	}

	return toUpdateProjectOutput(updatedProject), nil
}
