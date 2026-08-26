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

	return domainprojectsurface.ProjectSurface{
		ID:              row.ID,
		ProjectID:       row.ProjectID,
		SourceSurfaceID: nullInt64ToPtr(row.SourceSurfaceID),
		Name:            row.Name,
		Geometry:        polygon,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
		DeletedAt:       nil,
	}, nil
}

func (r *ProjectSurfaceRepository) projectSurfaceListRowToDomain(
	row ProjectSurfaceListRow,
) (domainprojectsurface.ProjectSurface, error) {
	polygon, err := r.geometryCodec.DecodeWKB(row.Geometry)
	if err != nil {
		return domainprojectsurface.ProjectSurface{}, fmt.Errorf(
			"decode project surface geometry: %w",
			err,
		)
	}

	projectSurface := domainprojectsurface.ProjectSurface{
		ID:              row.ID.Int64,
		ProjectID:       row.ProjectID.Int64,
		SourceSurfaceID: nullInt64ToPtr(row.SourceSurfaceID),
		Name:            row.Name.String,
		Geometry:        polygon,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
		DeletedAt:       nil,
	}

	return projectSurface, nil
}
