package placement

import (
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/domain/placement"
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

type PlacementWithPatternGeometry struct {
	Placement       placement.Placement
	PatternGeometry domaingeometry.Polygon
}

type GetPlacementByIDInput struct {
	UserID      int64
	ProjectID   int64
	PlacementID int64
}

type GetPlacementByIDOutput struct {
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

type ListPlacementsInput struct {
	UserID           int64
	ProjectID        int64
	ProjectSurfaceID int64
}

type ListPlacementsOutput struct {
	Items []ListPlacementsItem
}

type ListPlacementsItem struct {
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
