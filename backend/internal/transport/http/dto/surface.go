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
