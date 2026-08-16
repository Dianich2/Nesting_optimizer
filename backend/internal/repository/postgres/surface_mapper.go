package postgres

import (
	"fmt"
	domainsurface "server_nesting_optimizer/internal/domain/surface"
	"time"
)

func (r *SurfaceRepository) surfaceRowToDomain(
	row SurfaceRow,
) (domainsurface.Surface, error) {
	polygon, err := r.geometryCodec.DecodeWKB(row.Geometry)
	if err != nil {
		return domainsurface.Surface{}, fmt.Errorf(
			"decode surface geometry: %w",
			err,
		)
	}

	var deletedAt *time.Time
	if row.DeletedAt.Valid {
		t := row.DeletedAt.Time
		deletedAt = &t
	}

	return domainsurface.Surface{
		ID:        row.ID,
		UserID:    row.UserID,
		Name:      row.Name,
		Geometry:  polygon,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		DeletedAt: deletedAt,
	}, nil
}
