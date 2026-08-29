package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	domainprojectpattern "server_nesting_optimizer/internal/domain/project_pattern"
	"server_nesting_optimizer/internal/geometry"
	nestingusecase "server_nesting_optimizer/internal/usecase/nesting"
	projectpatternusecase "server_nesting_optimizer/internal/usecase/project_pattern"
	"time"

	"github.com/lib/pq"
)

type ProjectPatternRepository struct {
	db            DBTX
	geometryCodec geometry.Codec
}

func NewProjectPatternRepository(
	db DBTX,
	geometryCodec geometry.Codec,
) *ProjectPatternRepository {
	return &ProjectPatternRepository{
		db:            db,
		geometryCodec: geometryCodec,
	}
}

var _ projectpatternusecase.ProjectPatternRepository = (*ProjectPatternRepository)(nil)
var _ nestingusecase.ProjectPatternRepository = (*ProjectPatternRepository)(nil)

type ProjectPatternRow struct {
	ID              int64         `db:"id"`
	ProjectID       int64         `db:"project_id"`
	SourcePatternID sql.NullInt64 `db:"source_pattern_id"`
	Name            string        `db:"name"`
	Geometry        []byte        `db:"geometry"`
	CreatedAt       time.Time     `db:"created_at"`
	UpdatedAt       time.Time     `db:"updated_at"`
}

func (r *ProjectPatternRepository) Create(
	ctx context.Context,
	input domainprojectpattern.ProjectPattern,
	userID int64,
) (domainprojectpattern.ProjectPattern, error) {
	encoded, err := r.geometryCodec.EncodeWKB(input.Geometry)
	if err != nil {
		return domainprojectpattern.ProjectPattern{}, fmt.Errorf(
			"create project pattern: encode geometry: %w",
			err,
		)
	}

	var projectPatternRow ProjectPatternRow
	if err := r.db.GetContext(
		ctx,
		&projectPatternRow,
		createProjectPatternQuery,
		userID,
		input.ProjectID,
		input.SourcePatternID,
		input.Name,
		encoded,
	); err != nil {
		return domainprojectpattern.ProjectPattern{}, fmt.Errorf(
			"create project pattern: %w",
			err,
		)
	}

	projectPattern, err := r.projectPatternRowToDomain(projectPatternRow)
	if err != nil {
		return domainprojectpattern.ProjectPattern{}, fmt.Errorf(
			"create project pattern: map row: %w",
			err,
		)
	}

	return projectPattern, nil
}

func (r *ProjectPatternRepository) GetByID(
	ctx context.Context,
	userID int64,
	projectID int64,
	projectPatternID int64,
) (domainprojectpattern.ProjectPattern, error) {
	var projectPatternRow ProjectPatternRow
	if err := r.db.GetContext(
		ctx,
		&projectPatternRow,
		getProjectPatternByIDQuery,
		projectPatternID,
		projectID,
		userID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainprojectpattern.ProjectPattern{}, fmt.Errorf(
				"get project pattern by ID: %w",
				domainprojectpattern.ErrNotFound,
			)
		}

		return domainprojectpattern.ProjectPattern{}, fmt.Errorf(
			"get project pattern by ID: %w",
			err,
		)
	}

	projectPattern, err := r.projectPatternRowToDomain(projectPatternRow)
	if err != nil {
		return domainprojectpattern.ProjectPattern{}, fmt.Errorf(
			"get project pattern by ID: map row: %w",
			err,
		)
	}

	return projectPattern, nil
}

type ProjectPatternListRow struct {
	ID              sql.NullInt64  `db:"id"`
	ProjectID       sql.NullInt64  `db:"project_id"`
	SourcePatternID sql.NullInt64  `db:"source_pattern_id"`
	Name            sql.NullString `db:"name"`
	Geometry        []byte         `db:"geometry"`
	CreatedAt       sql.NullTime   `db:"created_at"`
	UpdatedAt       sql.NullTime   `db:"updated_at"`
	Total           int64          `db:"total"`
	ProjectExists   bool           `db:"project_exists"`
}

func (r *ProjectPatternRepository) ListByProjectID(
	ctx context.Context,
	userID int64,
	projectID int64,
	limit int,
	offset int,
) (projectpatternusecase.ProjectPatternListResult, error) {
	rows := make([]ProjectPatternListRow, 0)

	if err := r.db.SelectContext(
		ctx,
		&rows,
		listProjectPatternsQuery,
		projectID,
		userID,
		limit,
		offset,
	); err != nil {
		return projectpatternusecase.ProjectPatternListResult{}, fmt.Errorf(
			"list project patterns by project id: %w",
			err,
		)
	}

	listOfPatterns := projectpatternusecase.ProjectPatternListResult{
		ProjectPatterns: make([]domainprojectpattern.ProjectPattern, 0),
		Total:           0,
	}
	for _, row := range rows {
		if !row.ProjectExists {
			return projectpatternusecase.ProjectPatternListResult{},
				fmt.Errorf(
					"list project patterns by project ID: %w",
					domainprojectpattern.ErrNotFound,
				)
		}

		listOfPatterns.Total = row.Total

		if !row.ID.Valid {
			continue
		}

		projectPattern, err := r.projectPatternListRowToDomain(row)
		if err != nil {
			return projectpatternusecase.ProjectPatternListResult{}, fmt.Errorf(
				"list project patterns: map row: %w",
				err,
			)
		}

		listOfPatterns.ProjectPatterns = append(
			listOfPatterns.ProjectPatterns,
			projectPattern,
		)
	}

	return listOfPatterns, nil
}

func (r *ProjectPatternRepository) Update(
	ctx context.Context,
	projectPatternID int64,
	projectID int64,
	userID int64,
	name *string,
	geometry *domaingeometry.Polygon,
) (domainprojectpattern.ProjectPattern, error) {
	var encodedGeometry []byte
	if geometry != nil {
		encoded, err := r.geometryCodec.EncodeWKB(*geometry)
		if err != nil {
			return domainprojectpattern.ProjectPattern{}, fmt.Errorf(
				"update project pattern: encode geometry: %w",
				err,
			)
		}

		encodedGeometry = encoded
	}

	var projectPatternRow ProjectPatternRow
	if err := r.db.GetContext(
		ctx,
		&projectPatternRow,
		updateProjectPatternQuery,
		projectPatternID,
		projectID,
		userID,
		name,
		encodedGeometry,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainprojectpattern.ProjectPattern{}, fmt.Errorf(
				"update project pattern: %w",
				domainprojectpattern.ErrNotFound,
			)
		}

		return domainprojectpattern.ProjectPattern{}, fmt.Errorf(
			"update project pattern: %w",
			err,
		)
	}

	projectPattern, err := r.projectPatternRowToDomain(projectPatternRow)
	if err != nil {
		return domainprojectpattern.ProjectPattern{}, fmt.Errorf(
			"update project pattern: map row: %w",
			err,
		)
	}

	return projectPattern, nil
}

func (r *ProjectPatternRepository) SoftDelete(
	ctx context.Context,
	projectPatternID int64,
	projectID int64,
	userID int64,
) error {
	affected, err := execAffected(
		ctx,
		r.db,
		softDeleteProjectPatternQuery,
		projectPatternID,
		projectID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("soft delete project pattern: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf(
			"soft delete project pattern: %w",
			domainprojectpattern.ErrNotFound,
		)
	}

	return nil
}

func (r *ProjectPatternRepository) HasActivePlacements(
	ctx context.Context,
	projectPatternID int64,
	projectID int64,
	userID int64,
) (bool, error) {
	var hasActivePlacements bool
	if err := r.db.GetContext(
		ctx,
		&hasActivePlacements,
		hasActivePlacementsByProjectPatternIDQuery,
		projectPatternID,
		projectID,
		userID,
	); err != nil {
		return false, fmt.Errorf(
			"check active placements for project pattern: %w",
			err,
		)
	}

	return hasActivePlacements, nil
}

func (r *ProjectPatternRepository) GetByIDs(
	ctx context.Context,
	userID int64,
	projectID int64,
	ids []int64,
) ([]domainprojectpattern.ProjectPattern, error) {
	var projectPatternRows []ProjectPatternRow
	if err := r.db.SelectContext(
		ctx,
		&projectPatternRows,
		getProjectPatternsByIDsQuery,
		pq.Array(ids),
		projectID,
		userID,
	); err != nil {
		return nil, fmt.Errorf(
			"get project patterns by ids: %w",
			err,
		)
	}

	res := make([]domainprojectpattern.ProjectPattern, 0, len(projectPatternRows))
	for _, projectPatternRow := range projectPatternRows {
		projectPattern, err := r.projectPatternRowToDomain(projectPatternRow)
		if err != nil {
			return nil, fmt.Errorf(
				"get project patterns by ids: map row: %w",
				err,
			)
		}

		res = append(res, projectPattern)
	}

	return res, nil
}
