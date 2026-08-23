package pattern

import (
	"context"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	domainpattern "server_nesting_optimizer/internal/domain/pattern"
)

type PatternRepository interface {
	Create(
		ctx context.Context,
		pattern domainpattern.Pattern,
	) (domainpattern.Pattern, error)

	GetByID(
		ctx context.Context,
		patternID int64,
		userID int64,
	) (domainpattern.Pattern, error)

	ListByUserID(
		ctx context.Context,
		userID int64,
		limit int,
		offset int,
	) (PatternListResult, error)

	Update(
		ctx context.Context,
		patternID int64,
		userID int64,
		name *string,
		geometry *domaingeometry.Polygon,
	) (domainpattern.Pattern, error)

	SoftDelete(
		ctx context.Context,
		patternID int64,
		userID int64,
	) error
}
