package nesting

import (
	"context"
	"math"

	nestingrun "server_nesting_optimizer/internal/domain/nesting_run"
	"server_nesting_optimizer/internal/validation"
	"server_nesting_optimizer/pkg/apperror"
)

type HistoryRepository interface {
	List(context.Context, int64, int64, int64, int, int) ([]nestingrun.NestingRun, error)
}

func (uc *RunNestingUseCase) ListHistory(ctx context.Context, userID, projectID, surfaceID int64, page, pageSize int) ([]nestingrun.NestingRun, bool, error) {
	if details := validation.ValidatePagination(page, pageSize); len(details) > 0 {
		return nil, false, apperror.Validation("validation failed", details...)
	}
	if page-1 > math.MaxInt/pageSize {
		return nil, false, apperror.Validation("invalid history page", apperror.NewFieldError("page", apperror.FieldCodeInvalid, "page number is too large"))
	}

	if _, err := uc.projectSurfaceRepository.GetByID(ctx, userID, projectID, surfaceID); err != nil {
		return nil, false, operationSurfaceError(err)
	}
	rows, err := uc.History.List(ctx, userID, projectID, surfaceID, pageSize+1, (page-1)*pageSize)
	if err != nil {
		return nil, false, apperror.Internal("list nesting history", err)
	}
	if len(rows) > pageSize {
		return rows[:pageSize], true, nil
	}
	return rows, false, nil
}
