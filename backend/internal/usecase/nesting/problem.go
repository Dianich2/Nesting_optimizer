package nesting

import (
	"fmt"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	domainplacement "server_nesting_optimizer/internal/domain/placement"
	domainprojectpattern "server_nesting_optimizer/internal/domain/project_pattern"
	"server_nesting_optimizer/internal/geometry"
	"server_nesting_optimizer/internal/nesting"
)

func buildProblem(
	engine geometry.Engine,
	input RunNestingInput,
	projectSurface domaingeometry.Polygon,
	projectPatterns []domainprojectpattern.ProjectPattern,
	collisionPlacements []domainplacement.CollisionPlacement,
) (nesting.Problem, error) {
	problem := nesting.Problem{
		Surface: projectSurface,
	}

	patternsByID := make(map[int64]domainprojectpattern.ProjectPattern, len(projectPatterns))

	for _, projectPattern := range projectPatterns {
		patternsByID[projectPattern.ID] = projectPattern
	}

	nestingPatterns := make([]nesting.PatternItem, 0, len(input.Patterns))

	for _, reqPattern := range input.Patterns {
		projectPattern, ok := patternsByID[reqPattern.ProjectPatternID]

		if !ok {
			return nesting.Problem{}, fmt.Errorf(
				"build problem: project pattern %d not found",
				reqPattern.ProjectPatternID,
			)
		}

		nestingPatterns = append(nestingPatterns, nesting.PatternItem{
			PatternID: projectPattern.ID,
			Geometry:  projectPattern.Geometry,
			Quantity:  reqPattern.Quantity,
		})
	}

	problem.Patterns = nestingPatterns
	problem.AllowedRotations = NormalizeAllowedRotations(input.AllowedRotations)

	obstacles := make([]domaingeometry.Polygon, 0, len(collisionPlacements))
	for _, placement := range collisionPlacements {
		transformedGeometry, err := engine.Transform(
			placement.PatternGeometry,
			placement.X,
			placement.Y,
			placement.Rotation,
		)
		if err != nil {
			return nesting.Problem{}, fmt.Errorf(
				"build problem: %w",
				err,
			)
		}

		obstacles = append(obstacles, transformedGeometry)
	}

	problem.Obstacles = obstacles

	return problem, nil
}
