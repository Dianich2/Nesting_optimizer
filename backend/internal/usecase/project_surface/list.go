package projectsurface

import (
	"context"
	"errors"
	domainprojectsurface "server_nesting_optimizer/internal/domain/project_surface"
	"server_nesting_optimizer/pkg/apperror"
)

type ListProjectSurfacesUseCase struct {
	repo ProjectSurfaceRepository
}

func NewListProjectSurfacesUseCase(
	repo ProjectSurfaceRepository,
) *ListProjectSurfacesUseCase {
	return &ListProjectSurfacesUseCase{
		repo: repo,
	}
}

func (uc *ListProjectSurfacesUseCase) Execute(
	ctx context.Context,
	input ListProjectSurfacesInput,
) (ListProjectSurfacesOutput, error) {
	details := input.Validate()
	if len(details) > 0 {
		return ListProjectSurfacesOutput{}, apperror.Validation(
			"validation failed",
			details...,
		)
	}

	listOfProjectSurfaces := ListProjectSurfacesOutput{
		Items:    make([]ListProjectSurfacesItem, 0),
		Page:     input.Page,
		PageSize: input.PageSize,
	}

	repoListOfProjectSurfaces, err := uc.repo.ListByProjectID(
		ctx,
		input.UserID,
		input.ProjectID,
		input.PageSize,
		(input.Page-1)*input.PageSize,
	)
	if err != nil {
		if errors.Is(err, domainprojectsurface.ErrNotFound) {
			return ListProjectSurfacesOutput{}, apperror.NotFound(
				"project not found",
			)
		}

		return ListProjectSurfacesOutput{}, apperror.Internal(
			"failed to get project surfaces by project id",
			err,
		)
	}

	for _, projectSurface := range repoListOfProjectSurfaces.ProjectSurfaces {
		curProjectSurface := ListProjectSurfacesItem{
			ID:              projectSurface.ID,
			ProjectID:       projectSurface.ProjectID,
			SourceSurfaceID: projectSurface.SourceSurfaceID,
			Name:            projectSurface.Name,
			Geometry:        projectSurface.Geometry,
			CreatedAt:       projectSurface.CreatedAt,
			UpdatedAt:       projectSurface.UpdatedAt,
		}

		listOfProjectSurfaces.Items = append(
			listOfProjectSurfaces.Items,
			curProjectSurface,
		)
	}

	listOfProjectSurfaces.Total = repoListOfProjectSurfaces.Total

	if repoListOfProjectSurfaces.Total == 0 {
		listOfProjectSurfaces.TotalPages = 0
	} else {
		listOfProjectSurfaces.TotalPages = (listOfProjectSurfaces.Total + int64(listOfProjectSurfaces.PageSize) - 1) / int64(listOfProjectSurfaces.PageSize)
	}

	return listOfProjectSurfaces, nil
}
