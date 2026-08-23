package pattern

import (
	"server_nesting_optimizer/internal/domain/geometry"
	"time"
)

type Pattern struct {
	ID        int64
	UserID    int64
	Name      string
	Geometry  geometry.Polygon
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
