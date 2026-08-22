package surface

import (
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	domainsurface "server_nesting_optimizer/internal/domain/surface"
	"time"
)

type CreateSurfaceInput struct {
	UserID   int64
	Name     string
	Geometry domaingeometry.Polygon
}

type CreateSurfaceOutput struct {
	ID        int64
	UserID    int64
	Name      string
	Geometry  domaingeometry.Polygon
	CreatedAt time.Time
	UpdatedAt time.Time
}

type GetSurfaceByIDInput struct {
	SurfaceID int64
	UserID    int64
}

type GetSurfaceByIDOutput struct {
	ID        int64
	Name      string
	Geometry  domaingeometry.Polygon
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ListSurfacesInput struct {
	UserID   int64
	Page     int
	PageSize int
}

type ListSurfacesItem struct {
	ID        int64
	UserID    int64
	Name      string
	Geometry  domaingeometry.Polygon
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ListSurfacesOutput struct {
	Items      []ListSurfacesItem
	Page       int
	PageSize   int
	Total      int64
	TotalPages int64
}

type SurfaceListResult struct {
	Surfaces []domainsurface.Surface
	Total    int64
}
