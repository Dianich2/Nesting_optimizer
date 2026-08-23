package pattern

import (
	"context"
	"errors"
	domainpattern "server_nesting_optimizer/internal/domain/pattern"
	"server_nesting_optimizer/pkg/apperror"
)

type GetPatternByIDUseCase struct {
	repo PatternRepository
}

func NewGetPatternByIDUseCase(
	repo PatternRepository,
) *GetPatternByIDUseCase {
	return &GetPatternByIDUseCase{
		repo: repo,
	}
}

func (uc *GetPatternByIDUseCase) Execute(
	ctx context.Context,
	input GetPatternByIDInput,
) (GetPatternByIDOutput, error) {
	details := input.Validate()
	if len(details) > 0 {
		return GetPatternByIDOutput{}, apperror.Validation(
			"validation failed",
			details...,
		)
	}

	pattern, err := uc.repo.GetByID(
		ctx,
		input.PatternID,
		input.UserID,
	)
	if err != nil {
		if errors.Is(err, domainpattern.ErrNotFound) {
			return GetPatternByIDOutput{}, apperror.NotFound(
				"pattern not found",
			)
		}

		return GetPatternByIDOutput{}, apperror.Internal(
			"failed to get pattern by id",
			err,
		)
	}

	return toGetPatternByIDOutput(pattern), nil
}
