package projectsurface

import (
	"context"
	"errors"
	domainprojectsurface "server_nesting_optimizer/internal/domain/project_surface"
	"server_nesting_optimizer/pkg/apperror"
)

type GetProjectSurfaceByIDUseCase struct {
	repo ProjectSurfaceRepository
}

func NewGetProjectSurfaceByIDUseCase(
	repo ProjectSurfaceRepository,
) *GetProjectSurfaceByIDUseCase {
	return &GetProjectSurfaceByIDUseCase{
		repo: repo,
	}
}

func (uc *GetProjectSurfaceByIDUseCase) Execute(
	ctx context.Context,
	input GetProjectSurfaceByIDInput,
) (GetProjectSurfaceByIDOutput, error) {
	details := input.Validate()
	if len(details) > 0 {
		return GetProjectSurfaceByIDOutput{}, apperror.Validation(
			"validation failed",
			details...,
		)
	}

	projectSurface, err := uc.repo.GetByID(
		ctx,
		input.UserID,
		input.ProjectID,
		input.ProjectSurfaceID,
	)
	if err != nil {
		if errors.Is(err, domainprojectsurface.ErrNotFound) {
			return GetProjectSurfaceByIDOutput{}, apperror.NotFound(
				"project surface not found",
			)
		}

		return GetProjectSurfaceByIDOutput{}, apperror.Internal(
			"failed to get project surface by id",
			err,
		)
	}

	return toGetProjectSurfaceByIDOutput(projectSurface), nil
}
