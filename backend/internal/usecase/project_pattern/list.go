package projectpattern

import (
	"context"
	"errors"
	domainprojectpattern "server_nesting_optimizer/internal/domain/project_pattern"
	"server_nesting_optimizer/pkg/apperror"
)

type ListProjectPatternsUseCase struct {
	repo ProjectPatternRepository
}

func NewListProjectPatternsUseCase(
	repo ProjectPatternRepository,
) *ListProjectPatternsUseCase {
	return &ListProjectPatternsUseCase{
		repo: repo,
	}
}

func (uc *ListProjectPatternsUseCase) Execute(
	ctx context.Context,
	input ListProjectPatternsInput,
) (ListProjectPatternsOutput, error) {
	details := input.Validate()
	if len(details) > 0 {
		return ListProjectPatternsOutput{}, apperror.Validation(
			"validation failed",
			details...,
		)
	}

	listOfProjectPatterns := ListProjectPatternsOutput{
		Items:    make([]ListProjectPatternsItem, 0),
		Page:     input.Page,
		PageSize: input.PageSize,
	}

	repoListOfProjectPatterns, err := uc.repo.ListByProjectID(
		ctx,
		input.UserID,
		input.ProjectID,
		input.PageSize,
		(input.Page-1)*input.PageSize,
	)
	if err != nil {
		if errors.Is(err, domainprojectpattern.ErrNotFound) {
			return ListProjectPatternsOutput{}, apperror.NotFound(
				"project not found",
			)
		}

		return ListProjectPatternsOutput{}, apperror.Internal(
			"failed to get project patterns by project id",
			err,
		)
	}

	for _, projectPattern := range repoListOfProjectPatterns.ProjectPatterns {
		curProjectPattern := ListProjectPatternsItem{
			ID:              projectPattern.ID,
			ProjectID:       projectPattern.ProjectID,
			SourcePatternID: projectPattern.SourcePatternID,
			Name:            projectPattern.Name,
			Geometry:        projectPattern.Geometry,
			CreatedAt:       projectPattern.CreatedAt,
			UpdatedAt:       projectPattern.UpdatedAt,
		}

		listOfProjectPatterns.Items = append(
			listOfProjectPatterns.Items,
			curProjectPattern,
		)
	}

	listOfProjectPatterns.Total = repoListOfProjectPatterns.Total

	if repoListOfProjectPatterns.Total == 0 {
		listOfProjectPatterns.TotalPages = 0
	} else {
		listOfProjectPatterns.TotalPages = (listOfProjectPatterns.Total + int64(listOfProjectPatterns.PageSize) - 1) / int64(listOfProjectPatterns.PageSize)
	}

	return listOfProjectPatterns, nil
}
