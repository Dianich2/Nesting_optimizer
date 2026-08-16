package surface

import (
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
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
