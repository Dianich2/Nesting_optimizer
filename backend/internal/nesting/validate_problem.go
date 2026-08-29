package nesting

import (
	"fmt"
	"math"
	"server_nesting_optimizer/internal/geometry"
)

func validateProblem(
	engine geometry.Engine,
	problem Problem,
) error {
	if err := engine.ValidatePolygon(problem.Surface); err != nil {
		return fmt.Errorf(
			"validate problem surface: %w",
			err,
		)
	}

	if len(problem.AllowedRotations) == 0 {
		return fmt.Errorf(
			"validate problem rotations: %w",
			ErrEmptyRotations,
		)
	}

	for _, rotation := range problem.AllowedRotations {
		if math.IsNaN(rotation) || math.IsInf(rotation, 0) {
			return fmt.Errorf(
				"validate problem rotation: %w",
				ErrInvalidRotation,
			)
		}
	}

	for _, obstacle := range problem.Obstacles {
		if err := engine.ValidatePolygon(obstacle); err != nil {
			return fmt.Errorf(
				"validate problem obstacle: %w",
				err,
			)
		}
	}

	for _, patternItem := range problem.Patterns {
		if err := engine.ValidatePolygon(patternItem.Geometry); err != nil {
			return fmt.Errorf(
				"validate problem pattern: %w",
				err,
			)
		}

		if patternItem.Quantity <= 0 {
			return fmt.Errorf(
				"validate problem quantity: %w",
				ErrInvalidQuantity,
			)
		}
	}

	return nil
}
