package postgres

import (
	"fmt"
	domainplacement "server_nesting_optimizer/internal/domain/placement"
)

func (r *PlacementRepository) toCollisionPlacement(
	row CollisionPlacementRow,
) (domainplacement.CollisionPlacement, error) {
	polygon, err := r.geometryCodec.DecodeWKB(row.PatternGeometry)
	if err != nil {
		return domainplacement.CollisionPlacement{}, fmt.Errorf(
			"decode collision placement geometry: %w",
			err,
		)
	}

	return domainplacement.CollisionPlacement{
		ID:              row.ID,
		PatternGeometry: polygon,
		X:               row.X,
		Y:               row.Y,
		Rotation:        row.Rotation,
	}, nil
}

func placementRowToDomain(
	row PlacementRow,
) domainplacement.Placement {
	return domainplacement.Placement{
		ID:               row.ID,
		ProjectSurfaceID: row.ProjectSurfaceID,
		ProjectPatternID: row.ProjectPatternID,
		X:                row.X,
		Y:                row.Y,
		Rotation:         row.Rotation,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
		DeletedAt:        nil,
	}
}
