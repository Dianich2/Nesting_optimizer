package projectsurface

import (
	"context"
	"errors"
	domainprojectsurface "server_nesting_optimizer/internal/domain/project_surface"
	projecttransaction "server_nesting_optimizer/internal/usecase/project_transaction"
	"server_nesting_optimizer/pkg/apperror"
)

type DeleteProjectSurfaceUseCase struct {
	Transactions projecttransaction.Executor
	repo         ProjectSurfaceRepository
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
	_, err := projecttransaction.Run(ctx, uc.Transactions, input.ProjectID, input.UserID, func(ctx context.Context) (struct{}, error) { return struct{}{}, uc.execute(ctx, input) })
	return err
}

func (uc *DeleteProjectSurfaceUseCase) execute(
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

	hasPlacements, checkErr := uc.repo.HasActivePlacements(ctx, input.ProjectSurfaceID, input.ProjectID, input.UserID)
	if checkErr != nil {
		return apperror.Internal("failed to check active placements", checkErr)
	}
	if hasPlacements {
		return apperror.Conflict("project surface is used by active placements")
	}

	err := uc.repo.Delete(
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
			"failed to delete project surface",
			err,
		)
	}

	return nil
}
