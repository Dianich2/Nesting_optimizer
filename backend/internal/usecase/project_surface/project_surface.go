package projectsurface

import (
	"server_nesting_optimizer/internal/domain/geometry"
	"time"
)

type CreateProjectSurfaceInput struct {
	UserID          int64
	ProjectID       int64
	SourceSurfaceID int64
	Scale           float64
}

type CreateProjectSurfaceOutput struct {
	ID              int64
	ProjectID       int64
	SourceSurfaceID int64
	Name            string
	Geometry        geometry.Polygon
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type GetProjectSurfaceByIDInput struct {
	UserID           int64
	ProjectID        int64
	ProjectSurfaceID int64
}

type GetProjectSurfaceByIDOutput struct {
	ID              int64
	ProjectID       int64
	SourceSurfaceID *int64
	Name            string
	Geometry        geometry.Polygon
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
