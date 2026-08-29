package nesting

import domaingeometry "server_nesting_optimizer/internal/domain/geometry"

type Candidate struct {
	X        float64
	Y        float64
	Rotation float64
}

type CandidateScore struct {
	UsedArea   float64
	UsedHeight float64
	UsedWidth  float64
}

type CandidatePlacement struct {
	Candidate Candidate
	Geometry  domaingeometry.Polygon
}
