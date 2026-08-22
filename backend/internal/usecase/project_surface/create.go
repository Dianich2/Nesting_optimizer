package projectsurface

import (
	"context"
	"errors"
	domainproject "server_nesting_optimizer/internal/domain/project"
	domainprojectsurface "server_nesting_optimizer/internal/domain/project_surface"
	domainsurface "server_nesting_optimizer/internal/domain/surface"
	"server_nesting_optimizer/internal/geometry"
	projectusecase "server_nesting_optimizer/internal/usecase/project"
	surfaceusecase "server_nesting_optimizer/internal/usecase/surface"
	"server_nesting_optimizer/pkg/apperror"
)

type CreateProjectSurfaceUseCase struct {
	projectSurfaceRepo ProjectSurfaceRepository
	projectRepo        projectusecase.ProjectRepository
	surfaceRepo        surfaceusecase.SurfaceRepository
	geometryEngine     geometry.Engine
}

func NewCreateProjectSurfaceUseCase(
	projectSurfaceRepo ProjectSurfaceRepository,
	projectRepo projectusecase.ProjectRepository,
	surfaceRepo surfaceusecase.SurfaceRepository,
	geometryEngine geometry.Engine,
) *CreateProjectSurfaceUseCase {
	return &CreateProjectSurfaceUseCase{
		projectSurfaceRepo: projectSurfaceRepo,
		projectRepo:        projectRepo,
		surfaceRepo:        surfaceRepo,
		geometryEngine:     geometryEngine,
	}
}

func (uc *CreateProjectSurfaceUseCase) Execute(
	ctx context.Context,
	input CreateProjectSurfaceInput,
) (CreateProjectSurfaceOutput, error) {
	details := input.Validate()
	if len(details) > 0 {
		return CreateProjectSurfaceOutput{}, apperror.Validation(
			"validation failed",
			details...,
		)
	}

	_, err := uc.projectRepo.GetByID(
		ctx,
		input.ProjectID,
		input.UserID,
	)
	if err != nil {
		if errors.Is(err, domainproject.ErrNotFound) {
			return CreateProjectSurfaceOutput{}, apperror.NotFound(
				"project not found",
			)
		}

		return CreateProjectSurfaceOutput{}, apperror.Internal(
			"failed to get project by id",
			err,
		)
	}

	surface, err := uc.surfaceRepo.GetByID(
		ctx,
		input.SourceSurfaceID,
		input.UserID,
	)
	if err != nil {
		if errors.Is(err, domainsurface.ErrNotFound) {
			return CreateProjectSurfaceOutput{}, apperror.NotFound(
				"surface not found",
			)
		}

		return CreateProjectSurfaceOutput{}, apperror.Internal(
			"failed to get surface by id",
			err,
		)
	}

	scaledGeometry, err := uc.geometryEngine.Scale(surface.Geometry, input.Scale)
	if err != nil {
		if errors.Is(err, geometry.ErrInvalidScale) {
			return CreateProjectSurfaceOutput{}, apperror.Validation(
				"validation failed",
				apperror.NewFieldError(
					"scale",
					apperror.FieldCodeInvalid,
					"invalid surface scale",
				),
			)
		}

		if errors.Is(err, geometry.ErrInvalidPolygon) {
			return CreateProjectSurfaceOutput{}, apperror.Validation(
				"validation failed",
				apperror.NewFieldError(
					"geometry",
					apperror.FieldCodeInvalid,
					"invalid surface geometry",
				),
			)
		}

		return CreateProjectSurfaceOutput{}, apperror.Internal(
			"failed to scale surface geometry",
			err,
		)
	}

	domainProjectSurface := domainprojectsurface.ProjectSurface{
		ProjectID:       input.ProjectID,
		SourceSurfaceID: &input.SourceSurfaceID,
		Name:            surface.Name,
		Geometry:        scaledGeometry,
	}

	createdProjectSurface, err := uc.projectSurfaceRepo.Create(
		ctx,
		domainProjectSurface,
		input.UserID,
	)
	if err != nil {
		return CreateProjectSurfaceOutput{}, apperror.Internal(
			"failed to create project surface",
			err,
		)
	}

	return toCreateProjectSurfaceOutput(createdProjectSurface), nil
}
