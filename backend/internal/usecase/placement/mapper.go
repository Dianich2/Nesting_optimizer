package placement

import domainplacement "server_nesting_optimizer/internal/domain/placement"

func toCreatePlacementOutput(
	placement domainplacement.Placement,
) CreatePlacementOutput {
	return CreatePlacementOutput{
		ID:               placement.ID,
		ProjectSurfaceID: placement.ProjectSurfaceID,
		ProjectPatternID: placement.ProjectPatternID,
		X:                placement.X,
		Y:                placement.Y,
		Rotation:         placement.Rotation,
		CreatedAt:        placement.CreatedAt,
		UpdatedAt:        placement.UpdatedAt,
	}
}
