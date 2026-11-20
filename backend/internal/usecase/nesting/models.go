package nesting

import (
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/nesting"
	"time"
)

const ExecutionTimeout = 2 * time.Minute

type RunNestingPatternInput struct {
	ProjectPatternID int64
	Quantity         int
}

type RunNestingInput struct {
	PreviewID        string `json:",omitempty"`
	Algorithm        nesting.Algorithm
	UserID           int64
	ProjectID        int64
	ProjectSurfaceID int64
	Patterns         []RunNestingPatternInput
	AllowedRotations []float64
	KeepExisting     bool
}

type PlacementItem struct {
	ID               int64
	ProjectSurfaceID int64
	ProjectPatternID int64
	X                float64
	Y                float64
	Rotation         float64
	Geometry         domaingeometry.Polygon
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type RunNestingOutput struct {
	PreviewID        string
	PreviewExpiresAt time.Time
	Placements       []PlacementItem
	Unplaced         []nesting.UnplacedPattern
	Metrics          nesting.Metrics
}
