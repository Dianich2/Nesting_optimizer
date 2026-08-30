package nestingrun

import (
	"server_nesting_optimizer/internal/nesting"
	"time"
)

type NestingRun struct {
	ID               int64
	ProjectSurfaceID int64
	Algorithm        nesting.Algorithm
	KeepExisting     bool

	RequestedCount int
	PlacedCount    int
	SurfaceArea    float64
	PlacedArea     float64
	Utilization    float64

	Duration time.Duration

	CreatedAt time.Time
}
