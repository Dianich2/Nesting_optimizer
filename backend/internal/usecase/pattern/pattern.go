package pattern

import (
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	domainpattern "server_nesting_optimizer/internal/domain/pattern"
	"time"
)

type CreatePatternInput struct {
	UserID   int64
	Name     string
	Geometry domaingeometry.Polygon
}

type CreatePatternOutput struct {
	ID        int64
	Name      string
	Geometry  domaingeometry.Polygon
	CreatedAt time.Time
	UpdatedAt time.Time
}

type GetPatternByIDInput struct {
	PatternID int64
	UserID    int64
}

type GetPatternByIDOutput struct {
	ID        int64
	Name      string
	Geometry  domaingeometry.Polygon
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ListPatternsInput struct {
	UserID   int64
	Page     int
	PageSize int
}

type ListPatternsItem struct {
	ID        int64
	Name      string
	Geometry  domaingeometry.Polygon
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ListPatternsOutput struct {
	Items      []ListPatternsItem
	Page       int
	PageSize   int
	Total      int64
	TotalPages int64
}

type PatternListResult struct {
	Patterns []domainpattern.Pattern
	Total    int64
}

type UpdatePatternInput struct {
	PatternID int64
	UserID    int64
	Name      *string
	Geometry  *domaingeometry.Polygon
}

type UpdatePatternOutput struct {
	ID        int64
	Name      string
	Geometry  domaingeometry.Polygon
	CreatedAt time.Time
	UpdatedAt time.Time
}

type DeletePatternInput struct {
	PatternID int64
	UserID    int64
}
