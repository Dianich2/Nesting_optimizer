package dto

import "time"

type CreateProjectSurfaceRequest struct {
	SourceSurfaceID int64   `json:"source_surface_id"`
	Scale           float64 `json:"scale"`
}

type CreateProjectSurfaceResponse struct {
	ID              int64           `json:"id"`
	ProjectID       int64           `json:"project_id"`
	SourceSurfaceID int64           `json:"source_surface_id"`
	Name            string          `json:"name"`
	Geometry        PolygonGeometry `json:"geometry"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type GetProjectSurfaceByIDResponse struct {
	ID              int64           `json:"id"`
	ProjectID       int64           `json:"project_id"`
	SourceSurfaceID *int64          `json:"source_surface_id"`
	Name            string          `json:"name"`
	Geometry        PolygonGeometry `json:"geometry"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}
