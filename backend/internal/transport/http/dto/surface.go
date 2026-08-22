package dto

import "time"

type CreateSurfaceRequest struct {
	Name     string          `json:"name"`
	Geometry PolygonGeometry `json:"geometry"`
}

type CreateSurfaceResponse struct {
	ID        int64           `json:"id"`
	UserID    int64           `json:"user_id"`
	Name      string          `json:"name"`
	Geometry  PolygonGeometry `json:"geometry"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type GetSurfaceResponse struct {
	ID        int64           `json:"id"`
	Name      string          `json:"name"`
	Geometry  PolygonGeometry `json:"geometry"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type ListSurfacesItemResponse struct {
	ID        int64           `json:"id"`
	Name      string          `json:"name"`
	Geometry  PolygonGeometry `json:"geometry"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
type ListSurfacesResponse struct {
	Items      []ListSurfacesItemResponse `json:"items"`
	Page       int                        `json:"page"`
	PageSize   int                        `json:"page_size"`
	Total      int64                      `json:"total"`
	TotalPages int64                      `json:"total_pages"`
}
