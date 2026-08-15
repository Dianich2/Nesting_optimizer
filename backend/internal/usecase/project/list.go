package project

import (
	"context"
	"server_nesting_optimizer/pkg/apperror"
)

type ListProjectsUseCase struct {
	repo ProjectRepository
}

func NewListProjectsUseCase(
	repo ProjectRepository,
) *ListProjectsUseCase {
	return &ListProjectsUseCase{
		repo: repo,
	}
}

func (uc *ListProjectsUseCase) Execute(
	ctx context.Context,
	input ListProjectsInput,
) (ListProjectsOutput, error) {
	details := input.Validate()
	if len(details) > 0 {
		return ListProjectsOutput{}, apperror.Validation(
			"validation failed",
			details...,
		)
	}

	listOfProjects := ListProjectsOutput{
		Items:    make([]ListProjectsItem, 0),
		Page:     input.Page,
		PageSize: input.PageSize,
	}

	repoListOfProjects, err := uc.repo.ListByUserID(
		ctx,
		input.UserID,
		input.PageSize,
		(input.Page-1)*input.PageSize,
	)
	if err != nil {
		return ListProjectsOutput{}, apperror.Internal(
			"failed to get projects by user id",
			err,
		)
	}

	for _, project := range repoListOfProjects.Projects {
		curProject := ListProjectsItem{
			ID:          project.ID,
			UserID:      project.UserID,
			Name:        project.Name,
			Description: project.Description,
			CreatedAt:   project.CreatedAt,
			UpdatedAt:   project.UpdatedAt,
		}

		listOfProjects.Items = append(
			listOfProjects.Items,
			curProject,
		)
	}

	listOfProjects.Total = repoListOfProjects.Total

	if repoListOfProjects.Total == 0 {
		listOfProjects.TotalPages = 0
	} else {
		listOfProjects.TotalPages = (listOfProjects.Total + int64(listOfProjects.PageSize) - 1) / int64(listOfProjects.PageSize)
	}

	return listOfProjects, nil
}
