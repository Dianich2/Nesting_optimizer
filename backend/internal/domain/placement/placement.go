package placement

import "time"

type Placement struct {
	ID               int64
	ProjectSurfaceID int64
	ProjectPatternID int64
	X                float64
	Y                float64
	Rotation         float64
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}
