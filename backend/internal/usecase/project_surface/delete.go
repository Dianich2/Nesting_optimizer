package projectsurface

import (
	"context"
	"errors"
	domainprojectsurface "server_nesting_optimizer/internal/domain/project_surface"
	"server_nesting_optimizer/pkg/apperror"
)

type DeleteProjectSurfaceUseCase struct {
	repo ProjectSurfaceRepository
}

func NewDeleteProjectSurfaceUseCase(
	repo ProjectSurfaceRepository,
) *DeleteProjectSurfaceUseCase {
	return &DeleteProjectSurfaceUseCase{
		repo: repo,
	}
}

func (uc *DeleteProjectSurfaceUseCase) Execute(
	ctx context.Context,
	input DeleteProjectSurfaceInput,
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
		input.ProjectSurfaceID,
		input.ProjectID,
		input.UserID,
	)
	if err != nil {
		if errors.Is(err, domainprojectsurface.ErrNotFound) {
			return apperror.NotFound(
				"project surface not found",
			)
		}

		return apperror.Internal(
			"failed to soft delete project surface",
			err,
		)
	}

	return nil
}
