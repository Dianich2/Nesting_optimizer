package pattern

import (
	"context"
	"server_nesting_optimizer/pkg/apperror"
)

type ListPatternsUseCase struct {
	repo PatternRepository
}

func NewListPatternsUseCase(
	repo PatternRepository,
) *ListPatternsUseCase {
	return &ListPatternsUseCase{
		repo: repo,
	}
}

func (uc *ListPatternsUseCase) Execute(
	ctx context.Context,
	input ListPatternsInput,
) (ListPatternsOutput, error) {
	details := input.Validate()
	if len(details) > 0 {
		return ListPatternsOutput{}, apperror.Validation(
			"validation failed",
			details...,
		)
	}

	listOfPatterns := ListPatternsOutput{
		Items:    make([]ListPatternsItem, 0),
		Page:     input.Page,
		PageSize: input.PageSize,
	}

	repoListOfPatterns, err := uc.repo.ListByUserID(
		ctx,
		input.UserID,
		input.PageSize,
		(input.Page-1)*input.PageSize,
	)
	if err != nil {
		return ListPatternsOutput{}, apperror.Internal(
			"failed to get patterns by user id",
			err,
		)
	}

	for _, pattern := range repoListOfPatterns.Patterns {
		curPattern := ListPatternsItem{
			ID:        pattern.ID,
			Name:      pattern.Name,
			Geometry:  pattern.Geometry,
			CreatedAt: pattern.CreatedAt,
			UpdatedAt: pattern.UpdatedAt,
		}

		listOfPatterns.Items = append(
			listOfPatterns.Items,
			curPattern,
		)
	}

	listOfPatterns.Total = repoListOfPatterns.Total

	if repoListOfPatterns.Total == 0 {
		listOfPatterns.TotalPages = 0
	} else {
		listOfPatterns.TotalPages = (listOfPatterns.Total + int64(listOfPatterns.PageSize) - 1) / int64(listOfPatterns.PageSize)
	}

	return listOfPatterns, nil
}
