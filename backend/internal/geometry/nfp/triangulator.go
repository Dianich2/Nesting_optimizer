package nfp

import (
	"errors"
	"fmt"
	"math"
	"server_nesting_optimizer/internal/config"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"

	"github.com/rclancey/earcut"
)

type triangulatorInput struct {
	Vertices    []float64
	Points      []domaingeometry.Point
	HoleIndices []int
}

func prepareTriangulationInput(
	polygon domaingeometry.Polygon,
) triangulatorInput {
	preparedTriangulatorInput := triangulatorInput{
		Points:      make([]domaingeometry.Point, 0, len(polygon.Exterior.Points)),
		HoleIndices: make([]int, 0),
		Vertices:    make([]float64, 0),
	}

	for _, exteriorPoint := range polygon.Exterior.Points {
		preparedTriangulatorInput.Points = append(
			preparedTriangulatorInput.Points,
			exteriorPoint,
		)

		preparedTriangulatorInput.Vertices = append(
			preparedTriangulatorInput.Vertices,
			exteriorPoint.X,
			exteriorPoint.Y,
		)
	}

	for _, hole := range polygon.Holes {
		preparedTriangulatorInput.HoleIndices = append(
			preparedTriangulatorInput.HoleIndices,
			len(preparedTriangulatorInput.Points),
		)

		for _, holePoint := range hole.Points {
			preparedTriangulatorInput.Points = append(
				preparedTriangulatorInput.Points,
				holePoint,
			)

			preparedTriangulatorInput.Vertices = append(
				preparedTriangulatorInput.Vertices,
				holePoint.X,
				holePoint.Y,
			)
		}
	}

	return preparedTriangulatorInput
}

func triangulate(
	polygon domaingeometry.Polygon,
) ([]domaingeometry.Polygon, error) {
	preparedTriangulationInput := prepareTriangulationInput(polygon)

	indices, err := earcut.Earcut(
		preparedTriangulationInput.Vertices,
		preparedTriangulationInput.HoleIndices,
		2,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"triangulate polygon: %w",
			err,
		)
	}

	if len(indices) == 0 {
		return nil, fmt.Errorf(
			"triangulate polygon: %w",
			errors.New("triangulation returned no triangles"),
		)
	}

	if len(indices)%3 != 0 {
		return nil, fmt.Errorf(
			"triangulate polygon: %w",
			errors.New("len indices must divide by 3"),
		)
	}

	for _, index := range indices {
		if index < 0 || index >= len(preparedTriangulationInput.Points) {
			return nil, fmt.Errorf(
				"triangulate polygon: %w",
				errors.New("invalid index"),
			)
		}
	}

	deviation := earcut.Deviation(
		preparedTriangulationInput.Vertices,
		preparedTriangulationInput.HoleIndices,
		2,
		indices,
	)

	if math.Abs(deviation) > config.TriangulationDeviationTolerance ||
		math.IsInf(deviation, 0) ||
		math.IsNaN(deviation) {
		return nil, fmt.Errorf(
			"triangulate polygon: %w",
			errors.New("invalid deviation"),
		)
	}

	l := len(indices)

	newPolygons := make([]domaingeometry.Polygon, 0, len(indices)/3)

	for i := 0; i < l; i += 3 {
		newPolygon := domaingeometry.Polygon{
			Exterior: domaingeometry.Ring{
				Points: []domaingeometry.Point{
					preparedTriangulationInput.Points[indices[i]],
					preparedTriangulationInput.Points[indices[i+1]],
					preparedTriangulationInput.Points[indices[i+2]],
				},
			},
		}

		newPolygons = append(newPolygons, newPolygon)
	}

	return newPolygons, nil
}
