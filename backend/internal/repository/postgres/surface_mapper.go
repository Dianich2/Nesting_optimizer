package postgres

import (
	"fmt"
	domainsurface "server_nesting_optimizer/internal/domain/surface"
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

	return domainsurface.Surface{
		ID:        row.ID,
		UserID:    row.UserID,
		Name:      row.Name,
		Geometry:  polygon,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		DeletedAt: nullTimeToPtr(row.DeletedAt),
	}, nil
}

func (r *SurfaceRepository) surfaceListRowToDomain(
	row SurfaceListRow,
) (domainsurface.Surface, error) {
	polygon, err := r.geometryCodec.DecodeWKB(row.Geometry)
	if err != nil {
		return domainsurface.Surface{}, fmt.Errorf(
			"decode surface geometry: %w",
			err,
		)
	}

	surface := domainsurface.Surface{
		ID:        row.ID.Int64,
		UserID:    row.UserID.Int64,
		Name:      row.Name.String,
		Geometry:  polygon,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
		DeletedAt: nil,
	}

	return surface, nil
}
