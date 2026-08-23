package projectpattern

import (
	"server_nesting_optimizer/internal/domain/geometry"
	domainprojectpattern "server_nesting_optimizer/internal/domain/project_pattern"
	"time"
)

type CreateProjectPatternInput struct {
	UserID          int64
	ProjectID       int64
	SourcePatternID int64
	Scale           float64
}

type CreateProjectPatternOutput struct {
	ID              int64
	ProjectID       int64
	SourcePatternID int64
	Name            string
	Geometry        geometry.Polygon
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type GetProjectPatternByIDInput struct {
	UserID           int64
	ProjectID        int64
	ProjectPatternID int64
}

type GetProjectPatternByIDOutput struct {
	ID              int64
	ProjectID       int64
	SourcePatternID *int64
	Name            string
	Geometry        geometry.Polygon
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ListProjectPatternsInput struct {
	UserID    int64
	ProjectID int64
	Page      int
	PageSize  int
}

type ListProjectPatternsItem struct {
	ID              int64
	ProjectID       int64
	SourcePatternID *int64
	Name            string
	Geometry        geometry.Polygon
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ListProjectPatternsOutput struct {
	Items      []ListProjectPatternsItem
	Page       int
	PageSize   int
	Total      int64
	TotalPages int64
}

type ProjectPatternListResult struct {
	ProjectPatterns []domainprojectpattern.ProjectPattern
	Total           int64
}

type UpdateProjectPatternInput struct {
	UserID           int64
	ProjectPatternID int64
	ProjectID        int64
	Name             *string
	Scale            *float64
}

type UpdateProjectPatternOutput struct {
	ID              int64
	ProjectID       int64
	SourcePatternID *int64
	Name            string
	Geometry        geometry.Polygon
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type DeleteProjectPatternInput struct {
	UserID           int64
	ProjectPatternID int64
	ProjectID        int64
}
