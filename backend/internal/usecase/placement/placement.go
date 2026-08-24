package placement

import (
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"time"
)

type CreatePlacementInput struct {
	UserID           int64
	ProjectID        int64
	ProjectSurfaceID int64
	ProjectPatternID int64
	X                float64
	Y                float64
	Rotation         float64
}

type CreatePlacementOutput struct {
	ID               int64
	ProjectSurfaceID int64
	ProjectPatternID int64
	X                float64
	Y                float64
	Rotation         float64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type CollisionPlacement struct {
	ID              int64
	PatternGeometry domaingeometry.Polygon
	X               float64
	Y               float64
	Rotation        float64
}
