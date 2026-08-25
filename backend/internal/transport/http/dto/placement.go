package dto

import "time"

type CreatePlacementRequest struct {
	ProjectPatternID int64   `json:"project_pattern_id"`
	X                float64 `json:"x"`
	Y                float64 `json:"y"`
	Rotation         float64 `json:"rotation"`
}

type CreatePlacementResponse struct {
	ID               int64     `json:"id"`
	ProjectSurfaceID int64     `json:"project_surface_id"`
	ProjectPatternID int64     `json:"project_pattern_id"`
	X                float64   `json:"x"`
	Y                float64   `json:"y"`
	Rotation         float64   `json:"rotation"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type GetPlacementByIDResponse struct {
	ID               int64           `json:"id"`
	ProjectSurfaceID int64           `json:"project_surface_id"`
	ProjectPatternID int64           `json:"project_pattern_id"`
	X                float64         `json:"x"`
	Y                float64         `json:"y"`
	Rotation         float64         `json:"rotation"`
	Geometry         PolygonGeometry `json:"geometry"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type ListPlacementsResponse struct {
	Items []ListPlacementsItemResponse `json:"items"`
}

type ListPlacementsItemResponse struct {
	ID               int64           `json:"id"`
	ProjectSurfaceID int64           `json:"project_surface_id"`
	ProjectPatternID int64           `json:"project_pattern_id"`
	X                float64         `json:"x"`
	Y                float64         `json:"y"`
	Rotation         float64         `json:"rotation"`
	Geometry         PolygonGeometry `json:"geometry"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}
