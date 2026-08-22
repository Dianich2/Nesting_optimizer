package surface

import (
	"context"
	"server_nesting_optimizer/pkg/apperror"
)

type ListSurfacesUseCase struct {
	repo SurfaceRepository
}

func NewListSurfacesUseCase(
	repo SurfaceRepository,
) *ListSurfacesUseCase {
	return &ListSurfacesUseCase{
		repo: repo,
	}
}

func (uc *ListSurfacesUseCase) Execute(
	ctx context.Context,
	input ListSurfacesInput,
) (ListSurfacesOutput, error) {
	details := input.Validate()
	if len(details) > 0 {
		return ListSurfacesOutput{}, apperror.Validation(
			"validation failed",
			details...,
		)
	}

	listOfSurfaces := ListSurfacesOutput{
		Items:    make([]ListSurfacesItem, 0),
		Page:     input.Page,
		PageSize: input.PageSize,
	}

	repoListOfSurfaces, err := uc.repo.ListByUserID(
		ctx,
		input.UserID,
		input.PageSize,
		(input.Page-1)*input.PageSize,
	)
	if err != nil {
		return ListSurfacesOutput{}, apperror.Internal(
			"failed to get surfaces by user id",
			err,
		)
	}

	for _, surface := range repoListOfSurfaces.Surfaces {
		curSurface := ListSurfacesItem{
			ID:        surface.ID,
			UserID:    surface.UserID,
			Name:      surface.Name,
			Geometry:  surface.Geometry,
			CreatedAt: surface.CreatedAt,
			UpdatedAt: surface.UpdatedAt,
		}

		listOfSurfaces.Items = append(
			listOfSurfaces.Items,
			curSurface,
		)
	}

	listOfSurfaces.Total = repoListOfSurfaces.Total

	if repoListOfSurfaces.Total == 0 {
		listOfSurfaces.TotalPages = 0
	} else {
		listOfSurfaces.TotalPages = (listOfSurfaces.Total + int64(listOfSurfaces.PageSize) - 1) / int64(listOfSurfaces.PageSize)
	}

	return listOfSurfaces, nil
}
