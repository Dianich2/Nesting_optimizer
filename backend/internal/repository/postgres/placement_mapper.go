package postgres

import (
	"fmt"
	domainplacement "server_nesting_optimizer/internal/domain/placement"
	placementusecase "server_nesting_optimizer/internal/usecase/placement"
)

func (r *PlacementRepository) toCollisionPlacement(
	row CollisionPlacementRow,
) (placementusecase.CollisionPlacement, error) {
	polygon, err := r.geometryCodec.DecodeWKB(row.PatternGeometry)
	if err != nil {
		return placementusecase.CollisionPlacement{}, fmt.Errorf(
			"decode collision placement geometry: %w",
			err,
		)
	}

	return placementusecase.CollisionPlacement{
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
