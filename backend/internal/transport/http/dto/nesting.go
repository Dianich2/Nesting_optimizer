package dto

import (
	"encoding/json"
	"time"
)

type RunNestingPatternRequest struct {
	ProjectPatternID int64 `json:"project_pattern_id"`
	Quantity         int   `json:"quantity"`
}

type RunNestingUnplacedPatternResponse struct {
	ProjectPatternID int64 `json:"project_pattern_id"`
	Quantity         int   `json:"quantity"`
}

type RunNestingRequest struct {
	PreviewID        string                     `json:"preview_id,omitempty"`
	Algorithm        string                     `json:"algorithm"`
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
	PreviewID        string                              `json:"preview_id,omitempty"`
	PreviewExpiresAt *time.Time                          `json:"preview_expires_at,omitempty"`
	Placements       []RunNestingPlacementResponse       `json:"placements"`
	Unplaced         []RunNestingUnplacedPatternResponse `json:"unplaced"`
	Metrics          RunNestingMetricsResponse           `json:"metrics"`
}

type NestingHistoryItem struct {
	Parameters     json.RawMessage `json:"parameters,omitempty" swaggertype:"object"`
	ID             int64           `json:"id"`
	Algorithm      string          `json:"algorithm"`
	KeepExisting   bool            `json:"keep_existing"`
	RequestedCount int             `json:"requested_count"`
	PlacedCount    int             `json:"placed_count"`
	SurfaceArea    float64         `json:"surface_area"`
	PlacedArea     float64         `json:"placed_area"`
	Utilization    float64         `json:"utilization"`
	DurationMS     int64           `json:"duration_ms"`
	CreatedAt      time.Time       `json:"created_at"`
}

type NestingHistoryResponse struct {
	Items   []NestingHistoryItem `json:"items"`
	HasNext bool                 `json:"has_next"`
}

type NestingOperationResponse struct {
	Failure *NestingOperationFailure `json:"failure,omitempty"`
	Status  string                   `json:"status"`
	Result  *RunNestingResponse      `json:"result,omitempty"`
}

type NestingOperationFailure struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
