package placement

import (
	"server_nesting_optimizer/internal/domain/geometry"
	domainplacement "server_nesting_optimizer/internal/domain/placement"
)

func toCreatePlacementOutput(
	placement domainplacement.Placement,
	transformedGeometry geometry.Polygon,
) CreatePlacementOutput {
	return CreatePlacementOutput{
		ID:               placement.ID,
		ProjectSurfaceID: placement.ProjectSurfaceID,
		ProjectPatternID: placement.ProjectPatternID,
		X:                placement.X,
		Y:                placement.Y,
		Rotation:         placement.Rotation,
		Geometry:         transformedGeometry,
		CreatedAt:        placement.CreatedAt,
		UpdatedAt:        placement.UpdatedAt,
	}
}

func toGetPlacementByIDOutput(
	placement PlacementWithPatternGeometry,
	transformedGeometry geometry.Polygon,
) GetPlacementByIDOutput {
	return GetPlacementByIDOutput{
		ID:               placement.Placement.ID,
		ProjectSurfaceID: placement.Placement.ProjectSurfaceID,
		ProjectPatternID: placement.Placement.ProjectPatternID,
		X:                placement.Placement.X,
		Y:                placement.Placement.Y,
		Rotation:         placement.Placement.Rotation,
		Geometry:         transformedGeometry,
		CreatedAt:        placement.Placement.CreatedAt,
		UpdatedAt:        placement.Placement.UpdatedAt,
	}
}

func toUpdatePlacementOutput(
	placement domainplacement.Placement,
	transformedGeometry geometry.Polygon,
) UpdatePlacementOutput {
	return UpdatePlacementOutput{
		ID:               placement.ID,
		ProjectSurfaceID: placement.ProjectSurfaceID,
		ProjectPatternID: placement.ProjectPatternID,
		X:                placement.X,
		Y:                placement.Y,
		Rotation:         placement.Rotation,
		Geometry:         transformedGeometry,
		CreatedAt:        placement.CreatedAt,
		UpdatedAt:        placement.UpdatedAt,
	}
}
