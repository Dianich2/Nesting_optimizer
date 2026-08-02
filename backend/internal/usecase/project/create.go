package project

import (
	"context"
	"errors"
	domainproject "server_nesting_optimizer/internal/domain/project"
	"server_nesting_optimizer/pkg/apperror"
)

type CreateProjectUseCase struct {
	repo ProjectRepository
}

func NewCreateProjectUseCase(
	repo ProjectRepository,
) *CreateProjectUseCase {
	return &CreateProjectUseCase{
		repo: repo,
	}
}

func (uc *CreateProjectUseCase) Execute(
	ctx context.Context,
	input CreateProjectInput,
) (CreateProjectOutput, error) {
	input = normalizeCreateProjectInput(input)
	details := input.Validate()
	if len(details) > 0 {
		return CreateProjectOutput{}, apperror.Validation(
			"validation failed",
			details...,
		)
	}

	domainProject := toProject(
		input,
	)

	createdProject, err := uc.repo.Create(
		ctx,
		domainProject,
	)
	if err != nil {
		if errors.Is(err, domainproject.ErrOwnerNotFound) {
			return CreateProjectOutput{}, apperror.Unauthorized(
				"user is not active",
			)
		}

		return CreateProjectOutput{}, apperror.Internal(
			"failed to create project",
			err,
		)
	}

	return toCreateProjectOutput(createdProject), nil
}
