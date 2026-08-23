package pattern

import (
	"context"
	"errors"
	domainpattern "server_nesting_optimizer/internal/domain/pattern"
	"server_nesting_optimizer/pkg/apperror"
)

type DeletePatternUseCase struct {
	repo PatternRepository
}

func NewDeletePatternUseCase(
	repo PatternRepository,
) *DeletePatternUseCase {
	return &DeletePatternUseCase{
		repo: repo,
	}
}

func (uc *DeletePatternUseCase) Execute(
	ctx context.Context,
	input DeletePatternInput,
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
		input.PatternID,
		input.UserID,
	)
	if err != nil {
		if errors.Is(err, domainpattern.ErrNotFound) {
			return apperror.NotFound(
				"pattern not found",
			)
		}

		return apperror.Internal(
			"failed to soft delete pattern",
			err,
		)
	}

	return nil
}
