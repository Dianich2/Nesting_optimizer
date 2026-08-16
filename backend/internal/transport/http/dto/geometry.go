package dto

type GeometryPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type GeometryRing struct {
	Points []GeometryPoint `json:"points"`
}

type PolygonGeometry struct {
	Exterior GeometryRing   `json:"exterior"`
	Holes    []GeometryRing `json:"holes"`
}
