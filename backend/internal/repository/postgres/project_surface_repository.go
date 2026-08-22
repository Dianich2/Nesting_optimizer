package postgres

import (
	"context"
	"database/sql"
	"fmt"
	domainprojectsurface "server_nesting_optimizer/internal/domain/project_surface"
	"server_nesting_optimizer/internal/geometry"
	projectsurfaceusecase "server_nesting_optimizer/internal/usecase/project_surface"
	"time"
)

type ProjectSurfaceRepository struct {
	db            DBTX
	geometryCodec geometry.Codec
}

func NewProjectSurfaceRepository(
	db DBTX,
	geometryCodec geometry.Codec,
) *ProjectSurfaceRepository {
	return &ProjectSurfaceRepository{
		db:            db,
		geometryCodec: geometryCodec,
	}
}

var _ projectsurfaceusecase.ProjectSurfaceRepository = (*ProjectSurfaceRepository)(nil)

type ProjectSurfaceRow struct {
	ID              int64         `db:"id"`
	ProjectID       int64         `db:"project_id"`
	SourceSurfaceID sql.NullInt64 `db:"source_surface_id"`
	Name            string        `db:"name"`
	Geometry        []byte        `db:"geometry"`
	CreatedAt       time.Time     `db:"created_at"`
	UpdatedAt       time.Time     `db:"updated_at"`
}

func (r *ProjectSurfaceRepository) Create(
	ctx context.Context,
	input domainprojectsurface.ProjectSurface,
	userID int64,
) (domainprojectsurface.ProjectSurface, error) {
	encoded, err := r.geometryCodec.EncodeWKB(input.Geometry)
	if err != nil {
		return domainprojectsurface.ProjectSurface{}, fmt.Errorf(
			"create project surface: encode geometry: %w",
			err,
		)
	}

	var projectSurfaceRow ProjectSurfaceRow
	if err := r.db.GetContext(
		ctx,
		&projectSurfaceRow,
		createProjectSurfaceQuery,
		userID,
		input.ProjectID,
		input.SourceSurfaceID,
		input.Name,
		encoded,
	); err != nil {
		return domainprojectsurface.ProjectSurface{}, fmt.Errorf(
			"create project surface: %w",
			err,
		)
	}

	projectSurface, err := r.projectSurfaceRowToDomain(projectSurfaceRow)
	if err != nil {
		return domainprojectsurface.ProjectSurface{}, fmt.Errorf(
			"create project surface: map row: %w",
			err,
		)
	}

	return projectSurface, nil
}
