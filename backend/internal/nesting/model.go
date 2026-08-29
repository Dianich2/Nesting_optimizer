package nesting

import "server_nesting_optimizer/internal/domain/geometry"

type PatternItem struct {
	PatternID int64
	Geometry  geometry.Polygon
	Quantity  int
}

type Problem struct {
	Surface          geometry.Polygon
	Patterns         []PatternItem
	Obstacles        []geometry.Polygon
	AllowedRotations []float64
}

type Result struct {
	Placements []Placement
	Unplaced   []UnplacedPattern
	Metrics    Metrics
}

type Placement struct {
	PatternID int64
	X         float64
	Y         float64
	Rotation  float64
}

type UnplacedPattern struct {
	PatternID int64
	Quantity  int
}

type Metrics struct {
	RequestedCount int
	PlacedCount    int

	SurfaceArea float64
	PlacedArea  float64

	Utilization float64
}
