package dto

import "time"

type RunNestingPatternRequest struct {
	ProjectPatternID int64 `json:"project_pattern_id"`
	Quantity         int   `json:"quantity"`
}

type RunNestingUnplacedPatternResponse struct {
	ProjectPatternID int64 `json:"project_pattern_id"`
	Quantity         int   `json:"quantity"`
}

type RunNestingRequest struct {
	Patterns         []RunNestingPatternRequest `json:"patterns"`
	AllowedRotations []float64                  `json:"allowed_rotations"`
	KeepExisting     bool                       `json:"keep_existing"`
}

type RunNestingPlacementResponse struct {
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

type RunNestingMetricsResponse struct {
	RequestedCount int `json:"requested_count"`
	PlacedCount    int `json:"placed_count"`

	SurfaceArea float64 `json:"surface_area"`
	PlacedArea  float64 `json:"placed_area"`

	Utilization float64 `json:"utilization"`
}

type RunNestingResponse struct {
	Placements []RunNestingPlacementResponse       `json:"placements"`
	Unplaced   []RunNestingUnplacedPatternResponse `json:"unplaced"`
	Metrics    RunNestingMetricsResponse           `json:"metrics"`
}
