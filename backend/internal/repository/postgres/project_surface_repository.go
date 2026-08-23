package postgres

import (
	"context"
	"database/sql"
	"errors"
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

func (r *ProjectSurfaceRepository) GetByID(
	ctx context.Context,
	userID int64,
	projectID int64,
	projectSurfaceID int64,
) (domainprojectsurface.ProjectSurface, error) {
	var projectSurfaceRow ProjectSurfaceRow
	if err := r.db.GetContext(
		ctx,
		&projectSurfaceRow,
		getProjectSurfaceByIDQuery,
		projectSurfaceID,
		projectID,
		userID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainprojectsurface.ProjectSurface{}, fmt.Errorf(
				"get project surface by ID: %w",
				domainprojectsurface.ErrNotFound,
			)
		}

		return domainprojectsurface.ProjectSurface{}, fmt.Errorf(
			"get project surface by ID: %w",
			err,
		)
	}

	projectSurface, err := r.projectSurfaceRowToDomain(projectSurfaceRow)
	if err != nil {
		return domainprojectsurface.ProjectSurface{}, fmt.Errorf(
			"get project surface by ID: map row: %w",
			err,
		)
	}

	return projectSurface, nil
}

type ProjectSurfaceListRow struct {
	ID              sql.NullInt64  `db:"id"`
	ProjectID       sql.NullInt64  `db:"project_id"`
	SourceSurfaceID sql.NullInt64  `db:"source_surface_id"`
	Name            sql.NullString `db:"name"`
	Geometry        []byte         `db:"geometry"`
	CreatedAt       sql.NullTime   `db:"created_at"`
	UpdatedAt       sql.NullTime   `db:"updated_at"`
	Total           int64          `db:"total"`
	ProjectExists   bool           `db:"project_exists"`
}

func (r *ProjectSurfaceRepository) ListByProjectID(
	ctx context.Context,
	userID int64,
	projectID int64,
	limit int,
	offset int,
) (projectsurfaceusecase.ProjectSurfaceListResult, error) {
	rows := make([]ProjectSurfaceListRow, 0)

	if err := r.db.SelectContext(
		ctx,
		&rows,
		listProjectSurfacesQuery,
		projectID,
		userID,
		limit,
		offset,
	); err != nil {
		return projectsurfaceusecase.ProjectSurfaceListResult{}, fmt.Errorf(
			"list project surfaces by project id: %w",
			err,
		)
	}

	listOfSurfaces := projectsurfaceusecase.ProjectSurfaceListResult{
		ProjectSurfaces: make([]domainprojectsurface.ProjectSurface, 0),
		Total:           0,
	}
	for _, row := range rows {
		if !row.ProjectExists {
			return projectsurfaceusecase.ProjectSurfaceListResult{},
				fmt.Errorf(
					"list project surfaces by project ID: %w",
					domainprojectsurface.ErrNotFound,
				)
		}

		listOfSurfaces.Total = row.Total

		if !row.ID.Valid {
			continue
		}

		polygon, err := r.geometryCodec.DecodeWKB(row.Geometry)
		if err != nil {
			return projectsurfaceusecase.ProjectSurfaceListResult{}, fmt.Errorf(
				"decode project surface geometry: %w",
				err,
			)
		}

		var sourceSurfaceID *int64
		if row.SourceSurfaceID.Valid {
			id := row.SourceSurfaceID.Int64
			sourceSurfaceID = &id
		}

		projectSurface := domainprojectsurface.ProjectSurface{
			ID:              row.ID.Int64,
			ProjectID:       row.ProjectID.Int64,
			SourceSurfaceID: sourceSurfaceID,
			Name:            row.Name.String,
			Geometry:        polygon,
			CreatedAt:       row.CreatedAt.Time,
			UpdatedAt:       row.UpdatedAt.Time,
			DeletedAt:       nil,
		}

		listOfSurfaces.ProjectSurfaces = append(
			listOfSurfaces.ProjectSurfaces,
			projectSurface,
		)
	}

	return listOfSurfaces, nil
}
