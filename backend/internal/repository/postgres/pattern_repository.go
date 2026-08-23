package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	domainpattern "server_nesting_optimizer/internal/domain/pattern"
	"server_nesting_optimizer/internal/geometry"
	patternusecase "server_nesting_optimizer/internal/usecase/pattern"
	"time"
)

type PatternRepository struct {
	db            DBTX
	geometryCodec geometry.Codec
}

func NewPatternRepository(
	db DBTX,
	geometryCodec geometry.Codec,
) *PatternRepository {
	return &PatternRepository{
		db:            db,
		geometryCodec: geometryCodec,
	}
}

var _ patternusecase.PatternRepository = (*PatternRepository)(nil)

type PatternRow struct {
	ID        int64        `db:"id"`
	UserID    int64        `db:"user_id"`
	Name      string       `db:"name"`
	Geometry  []byte       `db:"geometry"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

func (r *PatternRepository) Create(
	ctx context.Context,
	input domainpattern.Pattern,
) (domainpattern.Pattern, error) {
	encoded, err := r.geometryCodec.EncodeWKB(input.Geometry)
	if err != nil {
		return domainpattern.Pattern{}, fmt.Errorf(
			"create pattern: encode geometry: %w",
			err,
		)
	}

	var patternRow PatternRow
	if err := r.db.GetContext(
		ctx,
		&patternRow,
		createPatternQuery,
		input.UserID,
		input.Name,
		encoded,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainpattern.Pattern{}, fmt.Errorf(
				"create pattern: %w",
				domainpattern.ErrOwnerNotFound,
			)
		}

		return domainpattern.Pattern{}, fmt.Errorf(
			"create pattern: %w",
			err,
		)
	}

	pattern, err := r.patternRowToDomain(patternRow)
	if err != nil {
		return domainpattern.Pattern{}, fmt.Errorf(
			"create pattern: map row: %w",
			err,
		)
	}

	return pattern, nil
}

func (r *PatternRepository) GetByID(
	ctx context.Context,
	patternID int64,
	userID int64,
) (domainpattern.Pattern, error) {
	var pattern PatternRow
	if err := r.db.GetContext(
		ctx,
		&pattern,
		getPatternByIDQuery,
		patternID,
		userID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainpattern.Pattern{}, fmt.Errorf(
				"get pattern: %w",
				domainpattern.ErrNotFound,
			)
		}

		return domainpattern.Pattern{}, fmt.Errorf(
			"get pattern: %w",
			err,
		)
	}

	patternD, err := r.patternRowToDomain(pattern)
	if err != nil {
		return domainpattern.Pattern{}, fmt.Errorf(
			"get pattern: map row: %w",
			err,
		)
	}

	return patternD, nil
}

type PatternListRow struct {
	ID        sql.NullInt64  `db:"id"`
	UserID    sql.NullInt64  `db:"user_id"`
	Name      sql.NullString `db:"name"`
	Geometry  []byte         `db:"geometry"`
	CreatedAt sql.NullTime   `db:"created_at"`
	UpdatedAt sql.NullTime   `db:"updated_at"`
	Total     int64          `db:"total"`
}

func (r *PatternRepository) ListByUserID(
	ctx context.Context,
	userID int64,
	limit int,
	offset int,
) (patternusecase.PatternListResult, error) {
	rows := make([]PatternListRow, 0)

	if err := r.db.SelectContext(
		ctx,
		&rows,
		listPatternsQuery,
		userID,
		limit,
		offset,
	); err != nil {
		return patternusecase.PatternListResult{}, fmt.Errorf(
			"list patterns by user id: %w",
			err,
		)
	}

	listOfPatterns := patternusecase.PatternListResult{
		Patterns: make([]domainpattern.Pattern, 0),
		Total:    0,
	}

	for _, row := range rows {
		listOfPatterns.Total = row.Total

		if !row.ID.Valid {
			continue
		}

		polygon, err := r.geometryCodec.DecodeWKB(row.Geometry)
		if err != nil {
			return patternusecase.PatternListResult{}, fmt.Errorf(
				"decode pattern geometry: %w",
				err,
			)
		}

		pattern := domainpattern.Pattern{
			ID:        row.ID.Int64,
			UserID:    row.UserID.Int64,
			Name:      row.Name.String,
			Geometry:  polygon,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
			DeletedAt: nil,
		}

		listOfPatterns.Patterns = append(
			listOfPatterns.Patterns,
			pattern,
		)
	}

	return listOfPatterns, nil
}

func (r *PatternRepository) Update(
	ctx context.Context,
	patternID int64,
	userID int64,
	name *string,
	geometry *domaingeometry.Polygon,
) (domainpattern.Pattern, error) {
	var encodedGeometry []byte

	if geometry != nil {
		encoded, err := r.geometryCodec.EncodeWKB(*geometry)
		if err != nil {
			return domainpattern.Pattern{}, fmt.Errorf(
				"update pattern: encode geometry: %w",
				err,
			)
		}

		encodedGeometry = encoded
	}

	var patternRow PatternRow
	if err := r.db.GetContext(
		ctx,
		&patternRow,
		updatePatternQuery,
		patternID,
		userID,
		name,
		encodedGeometry,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainpattern.Pattern{}, fmt.Errorf(
				"update pattern: %w",
				domainpattern.ErrNotFound,
			)
		}

		return domainpattern.Pattern{}, fmt.Errorf(
			"update pattern: %w",
			err,
		)
	}

	pattern, err := r.patternRowToDomain(patternRow)
	if err != nil {
		return domainpattern.Pattern{}, fmt.Errorf(
			"update pattern: map row: %w",
			err,
		)
	}

	return pattern, nil
}

func (r *PatternRepository) SoftDelete(
	ctx context.Context,
	patternID int64,
	userID int64,
) error {
	res, err := r.db.ExecContext(
		ctx,
		softDeletePatternQuery,
		patternID,
		userID,
	)
	if err != nil {
		return fmt.Errorf(
			"soft delete pattern: %w",
			err,
		)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"soft delete pattern: %w",
			err,
		)
	}

	if count == 0 {
		return fmt.Errorf(
			"soft delete pattern: %w",
			domainpattern.ErrNotFound,
		)
	}

	return nil
}
