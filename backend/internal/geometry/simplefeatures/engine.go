package simplefeatures

import (
	"errors"
	"fmt"
	"math"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"
)

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

var _ geometry.Engine = (*Engine)(nil)

func (e *Engine) ValidatePolygon(
	polygon domaingeometry.Polygon,
) error {
	if len(polygon.Exterior.Points) == 0 {
		return fmt.Errorf(
			"validate polygon: %w",
			geometry.ErrInvalidPolygon,
		)
	}

	sfPolygon := toSimpleFeaturesPolygon(polygon)

	if sfPolygon.IsEmpty() {
		return fmt.Errorf(
			"validate polygon: %w",
			geometry.ErrInvalidPolygon,
		)
	}

	if err := sfPolygon.Validate(); err != nil {
		return fmt.Errorf(
			"validate polygon: %w",
			errors.Join(
				geometry.ErrInvalidPolygon,
				err,
			),
		)
	}

	return nil
}

func (e *Engine) Area(
	polygon domaingeometry.Polygon,
) (float64, error) {
	err := e.ValidatePolygon(polygon)

	if err != nil {
		return 0, fmt.Errorf(
			"calculate polygon area: %w",
			err,
		)
	}

	sfPolygon := toSimpleFeaturesPolygon(polygon)

	return sfPolygon.Area(), nil
}

func (e *Engine) Normalize(
	polygon domaingeometry.Polygon,
) (domaingeometry.Polygon, error) {
	if err := e.ValidatePolygon(polygon); err != nil {
		return domaingeometry.Polygon{}, fmt.Errorf(
			"normalize polygon: %w",
			err,
		)
	}

	var minX = math.Inf(1)
	var minY = math.Inf(1)

	for _, point := range polygon.Exterior.Points {
		if point.X < minX {
			minX = point.X
		}
		if point.Y < minY {
			minY = point.Y
		}
	}

	newPolygon := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: make([]domaingeometry.Point, 0, len(polygon.Exterior.Points)),
		},
	}

	for _, point := range polygon.Exterior.Points {
		newPolygon.Exterior.Points = append(
			newPolygon.Exterior.Points,
			domaingeometry.Point{
				X: point.X - minX,
				Y: point.Y - minY,
			},
		)
	}

	var holes []domaingeometry.Ring

	if len(polygon.Holes) > 0 {
		holes = make([]domaingeometry.Ring, 0, len(polygon.Holes))
		for _, hole := range polygon.Holes {
			newHole := domaingeometry.Ring{
				Points: make([]domaingeometry.Point, 0, len(hole.Points)),
			}

			for _, point := range hole.Points {
				newHole.Points = append(
					newHole.Points,
					domaingeometry.Point{
						X: point.X - minX,
						Y: point.Y - minY,
					},
				)
			}

			holes = append(holes, newHole)
		}
	}

	newPolygon.Holes = holes

	return newPolygon, nil
}

func (e *Engine) Scale(
	polygon domaingeometry.Polygon,
	factor float64,
) (domaingeometry.Polygon, error) {
	if factor <= 0 ||
		math.IsNaN(factor) ||
		math.IsInf(factor, 0) {
		return domaingeometry.Polygon{}, fmt.Errorf(
			"scale polygon: %w",
			geometry.ErrInvalidScale,
		)
	}

	normalizedPolygon, err := e.Normalize(polygon)
	if err != nil {
		return domaingeometry.Polygon{}, fmt.Errorf(
			"scale polygon: %w",
			err,
		)
	}

	newPolygon := domaingeometry.Polygon{
		Exterior: domaingeometry.Ring{
			Points: make([]domaingeometry.Point, 0, len(normalizedPolygon.Exterior.Points)),
		},
	}

	for _, point := range normalizedPolygon.Exterior.Points {
		newPolygon.Exterior.Points = append(
			newPolygon.Exterior.Points,
			domaingeometry.Point{
				X: point.X * factor,
				Y: point.Y * factor,
			},
		)
	}

	var holes []domaingeometry.Ring

	if len(normalizedPolygon.Holes) > 0 {
		holes = make([]domaingeometry.Ring, 0, len(normalizedPolygon.Holes))
		for _, hole := range normalizedPolygon.Holes {
			newHole := domaingeometry.Ring{
				Points: make([]domaingeometry.Point, 0, len(hole.Points)),
			}

			for _, point := range hole.Points {
				newHole.Points = append(
					newHole.Points,
					domaingeometry.Point{
						X: point.X * factor,
						Y: point.Y * factor,
					},
				)
			}

			holes = append(holes, newHole)
		}
	}

	newPolygon.Holes = holes

	return newPolygon, nil
}
