package projectsurface

import (
	"server_nesting_optimizer/internal/domain/geometry"
	domainprojectsurface "server_nesting_optimizer/internal/domain/project_surface"
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

type ListProjectSurfacesInput struct {
	UserID    int64
	ProjectID int64
	Page      int
	PageSize  int
}

type ListProjectSurfacesItem struct {
	ID              int64
	ProjectID       int64
	SourceSurfaceID *int64
	Name            string
	Geometry        geometry.Polygon
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ListProjectSurfacesOutput struct {
	Items      []ListProjectSurfacesItem
	Page       int
	PageSize   int
	Total      int64
	TotalPages int64
}

type ProjectSurfaceListResult struct {
	ProjectSurfaces []domainprojectsurface.ProjectSurface
	Total           int64
}

type UpdateProjectSurfaceInput struct {
	UserID           int64
	ProjectSurfaceID int64
	ProjectID        int64
	Name             *string
	Scale            *float64
}

type UpdateProjectSurfaceOutput struct {
	ID              int64
	ProjectID       int64
	SourceSurfaceID *int64
	Name            string
	Geometry        geometry.Polygon
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type DeleteProjectSurfaceInput struct {
	UserID           int64
	ProjectSurfaceID int64
	ProjectID        int64
}
