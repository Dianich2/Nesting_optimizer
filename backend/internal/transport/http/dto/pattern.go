package dto

import "time"

type CreatePatternRequest struct {
	Name     string          `json:"name"`
	Geometry PolygonGeometry `json:"geometry"`
}

type CreatePatternResponse struct {
	ID        int64           `json:"id"`
	Name      string          `json:"name"`
	Geometry  PolygonGeometry `json:"geometry"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type GetPatternResponse struct {
	ID        int64           `json:"id"`
	Name      string          `json:"name"`
	Geometry  PolygonGeometry `json:"geometry"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type ListPatternsItemResponse struct {
	ID        int64           `json:"id"`
	Name      string          `json:"name"`
	Geometry  PolygonGeometry `json:"geometry"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
type ListPatternsResponse struct {
	Items      []ListPatternsItemResponse `json:"items"`
	Page       int                        `json:"page"`
	PageSize   int                        `json:"page_size"`
	Total      int64                      `json:"total"`
	TotalPages int64                      `json:"total_pages"`
}

type UpdatePatternRequest struct {
	Name     *string          `json:"name"`
	Geometry *PolygonGeometry `json:"geometry"`
}

type UpdatePatternResponse struct {
	ID        int64           `json:"id"`
	Name      string          `json:"name"`
	Geometry  PolygonGeometry `json:"geometry"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
