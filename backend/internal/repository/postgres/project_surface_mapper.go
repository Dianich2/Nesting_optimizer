package postgres

import (
	"fmt"
	domainprojectsurface "server_nesting_optimizer/internal/domain/project_surface"
)

func (r *ProjectSurfaceRepository) projectSurfaceRowToDomain(
	row ProjectSurfaceRow,
) (domainprojectsurface.ProjectSurface, error) {
	polygon, err := r.geometryCodec.DecodeWKB(row.Geometry)
	if err != nil {
		return domainprojectsurface.ProjectSurface{}, fmt.Errorf(
			"decode project surface geometry: %w",
			err,
		)
	}

	var sourceSurfaceID *int64
	if row.SourceSurfaceID.Valid {
		id := row.SourceSurfaceID.Int64
		sourceSurfaceID = &id
	}

	return domainprojectsurface.ProjectSurface{
		ID:              row.ID,
		ProjectID:       row.ProjectID,
		SourceSurfaceID: sourceSurfaceID,
		Name:            row.Name,
		Geometry:        polygon,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
		DeletedAt:       nil,
	}, nil
}
