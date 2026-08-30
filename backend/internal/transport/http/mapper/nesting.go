package mapper

import (
	"server_nesting_optimizer/internal/nesting"
	"server_nesting_optimizer/internal/transport/http/dto"
	nestingusecase "server_nesting_optimizer/internal/usecase/nesting"
)

func ToRunNestingInput(
	req dto.RunNestingRequest,
	userID int64,
	projectID int64,
	projectSurfaceID int64,
) nestingusecase.RunNestingInput {
	var nestingPatterns []nestingusecase.RunNestingPatternInput
	for _, pattern := range req.Patterns {
		patternInput := nestingusecase.RunNestingPatternInput{
			ProjectPatternID: pattern.ProjectPatternID,
			Quantity:         pattern.Quantity,
		}

		nestingPatterns = append(nestingPatterns, patternInput)
	}

	return nestingusecase.RunNestingInput{
		Algorithm:        nesting.Algorithm(req.Algorithm),
		UserID:           userID,
		ProjectID:        projectID,
		ProjectSurfaceID: projectSurfaceID,
		Patterns:         nestingPatterns,
		AllowedRotations: req.AllowedRotations,
		KeepExisting:     req.KeepExisting,
	}
}

func ToRunNestingResponse(
	output nestingusecase.RunNestingOutput,
) dto.RunNestingResponse {
	placements := make([]dto.RunNestingPlacementResponse, 0, len(output.Placements))
	for _, placement := range output.Placements {
		placements = append(placements, dto.RunNestingPlacementResponse{
			ID:               placement.ID,
			ProjectSurfaceID: placement.ProjectSurfaceID,
			ProjectPatternID: placement.ProjectPatternID,
			X:                placement.X,
			Y:                placement.Y,
			Rotation:         placement.Rotation,
			Geometry:         toGeometryPolygon(placement.Geometry),
			CreatedAt:        placement.CreatedAt,
			UpdatedAt:        placement.UpdatedAt,
		})
	}

	unplaced := make([]dto.RunNestingUnplacedPatternResponse, 0, len(output.Unplaced))
	for _, unplacedPattern := range output.Unplaced {
		unplaced = append(unplaced, dto.RunNestingUnplacedPatternResponse{
			ProjectPatternID: unplacedPattern.PatternID,
			Quantity:         unplacedPattern.Quantity,
		})
	}

	return dto.RunNestingResponse{
		Placements: placements,
		Unplaced:   unplaced,
		Metrics: dto.RunNestingMetricsResponse{
			RequestedCount: output.Metrics.RequestedCount,
			PlacedCount:    output.Metrics.PlacedCount,
			SurfaceArea:    output.Metrics.SurfaceArea,
			PlacedArea:     output.Metrics.PlacedArea,
			Utilization:    output.Metrics.Utilization,
		},
	}
}
