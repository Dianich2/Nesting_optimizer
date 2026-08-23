package projectpattern

import (
	"server_nesting_optimizer/internal/domain/geometry"
	"time"
)

type ProjectPattern struct {
	ID              int64
	ProjectID       int64
	SourcePatternID *int64
	Name            string
	Geometry        geometry.Polygon
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}
