package dto

import "time"

type CreateProjectPatternRequest struct {
	SourcePatternID int64   `json:"source_pattern_id"`
	Scale           float64 `json:"scale"`
}

type CreateProjectPatternResponse struct {
	ID              int64           `json:"id"`
	ProjectID       int64           `json:"project_id"`
	SourcePatternID int64           `json:"source_pattern_id"`
	Name            string          `json:"name"`
	Geometry        PolygonGeometry `json:"geometry"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type GetProjectPatternByIDResponse struct {
	ID              int64           `json:"id"`
	ProjectID       int64           `json:"project_id"`
	SourcePatternID *int64          `json:"source_pattern_id"`
	Name            string          `json:"name"`
	Geometry        PolygonGeometry `json:"geometry"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type ListProjectPatternsItemResponse struct {
	ID              int64           `json:"id"`
	ProjectID       int64           `json:"project_id"`
	SourcePatternID *int64          `json:"source_pattern_id"`
	Name            string          `json:"name"`
	Geometry        PolygonGeometry `json:"geometry"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}
type ListProjectPatternsResponse struct {
	Items      []ListProjectPatternsItemResponse `json:"items"`
	Page       int                               `json:"page"`
	PageSize   int                               `json:"page_size"`
	Total      int64                             `json:"total"`
	TotalPages int64                             `json:"total_pages"`
}

type UpdateProjectPatternRequest struct {
	Name  *string  `json:"name"`
	Scale *float64 `json:"scale"`
}

type UpdateProjectPatternResponse struct {
	ID              int64           `json:"id"`
	ProjectID       int64           `json:"project_id"`
	SourcePatternID *int64          `json:"source_pattern_id"`
	Name            string          `json:"name"`
	Geometry        PolygonGeometry `json:"geometry"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}
