package geometry

type Polygon struct {
	Exterior Ring
	Holes    []Ring
}
