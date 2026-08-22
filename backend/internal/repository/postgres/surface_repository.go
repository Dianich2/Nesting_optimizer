package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	domainsurface "server_nesting_optimizer/internal/domain/surface"
	"server_nesting_optimizer/internal/geometry"
	surfaceusecase "server_nesting_optimizer/internal/usecase/surface"
	"time"
)

type SurfaceRepository struct {
	db            DBTX
	geometryCodec geometry.Codec
}

func NewSurfaceRepository(
	db DBTX,
	geometryCodec geometry.Codec,
) *SurfaceRepository {
	return &SurfaceRepository{
		db:            db,
		geometryCodec: geometryCodec,
	}
}

var _ surfaceusecase.SurfaceRepository = (*SurfaceRepository)(nil)

type SurfaceRow struct {
	ID        int64        `db:"id"`
	UserID    int64        `db:"user_id"`
	Name      string       `db:"name"`
	Geometry  []byte       `db:"geometry"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

func (r *SurfaceRepository) Create(
	ctx context.Context,
	input domainsurface.Surface,
) (domainsurface.Surface, error) {
	encoded, err := r.geometryCodec.EncodeWKB(input.Geometry)
	if err != nil {
		return domainsurface.Surface{}, fmt.Errorf(
			"create surface: encode geometry: %w",
			err,
		)
	}

	var surfaceRow SurfaceRow
	if err := r.db.GetContext(
		ctx,
		&surfaceRow,
		createSurfaceQuery,
		input.UserID,
		input.Name,
		encoded,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainsurface.Surface{}, fmt.Errorf(
				"create surface: %w",
				domainsurface.ErrOwnerNotFound,
			)
		}

		return domainsurface.Surface{}, fmt.Errorf(
			"create surface: %w",
			err,
		)
	}

	surface, err := r.surfaceRowToDomain(surfaceRow)
	if err != nil {
		return domainsurface.Surface{}, fmt.Errorf(
			"create surface: map row: %w",
			err,
		)
	}

	return surface, nil
}

func (r *SurfaceRepository) GetByID(
	ctx context.Context,
	surfaceID int64,
	userID int64,
) (domainsurface.Surface, error) {
	var surface SurfaceRow
	if err := r.db.GetContext(
		ctx,
		&surface,
		getSurfaceByIDQuery,
		surfaceID,
		userID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainsurface.Surface{}, fmt.Errorf(
				"get surface: %w",
				domainsurface.ErrNotFound,
			)
		}

		return domainsurface.Surface{}, fmt.Errorf(
			"get surface: %w",
			err,
		)
	}

	surfaceD, err := r.surfaceRowToDomain(surface)
	if err != nil {
		return domainsurface.Surface{}, fmt.Errorf(
			"get surface: map row: %w",
			err,
		)
	}

	return surfaceD, nil
}

type SurfaceListRow struct {
	ID        sql.NullInt64  `db:"id"`
	UserID    sql.NullInt64  `db:"user_id"`
	Name      sql.NullString `db:"name"`
	Geometry  []byte         `db:"geometry"`
	CreatedAt sql.NullTime   `db:"created_at"`
	UpdatedAt sql.NullTime   `db:"updated_at"`
	Total     int64          `db:"total"`
}

func (r *SurfaceRepository) ListByUserID(
	ctx context.Context,
	userID int64,
	limit int,
	offset int,
) (surfaceusecase.SurfaceListResult, error) {
	rows := make([]SurfaceListRow, 0)

	if err := r.db.SelectContext(
		ctx,
		&rows,
		listSurfacesQuery,
		userID,
		limit,
		offset,
	); err != nil {
		return surfaceusecase.SurfaceListResult{}, fmt.Errorf(
			"list surfaces by user id: %w",
			err,
		)
	}

	listOfSurfaces := surfaceusecase.SurfaceListResult{
		Surfaces: make([]domainsurface.Surface, 0),
		Total:    0,
	}

	for _, row := range rows {
		listOfSurfaces.Total = row.Total

		if !row.ID.Valid {
			continue
		}

		polygon, err := r.geometryCodec.DecodeWKB(row.Geometry)
		if err != nil {
			return surfaceusecase.SurfaceListResult{}, fmt.Errorf(
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

		listOfSurfaces.Surfaces = append(
			listOfSurfaces.Surfaces,
			surface,
		)
	}

	return listOfSurfaces, nil
}

func (r *SurfaceRepository) Update(
	ctx context.Context,
	surfaceID int64,
	userID int64,
	name *string,
	geometry *domaingeometry.Polygon,
) (domainsurface.Surface, error) {
	var encodedGeometry []byte

	if geometry != nil {
		encoded, err := r.geometryCodec.EncodeWKB(*geometry)
		if err != nil {
			return domainsurface.Surface{}, fmt.Errorf(
				"update surface: encode geometry: %w",
				err,
			)
		}

		encodedGeometry = encoded
	}

	var surfaceRow SurfaceRow
	if err := r.db.GetContext(
		ctx,
		&surfaceRow,
		updateSurfaceQuery,
		surfaceID,
		userID,
		name,
		encodedGeometry,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainsurface.Surface{}, fmt.Errorf(
				"update surface: %w",
				domainsurface.ErrNotFound,
			)
		}

		return domainsurface.Surface{}, fmt.Errorf(
			"update surface: %w",
			err,
		)
	}

	surface, err := r.surfaceRowToDomain(surfaceRow)
	if err != nil {
		return domainsurface.Surface{}, fmt.Errorf(
			"update surface: map row: %w",
			err,
		)
	}

	return surface, nil
}

func (r *SurfaceRepository) SoftDelete(
	ctx context.Context,
	surfaceID int64,
	userID int64,
) error {
	res, err := r.db.ExecContext(
		ctx,
		softDeleteSurfaceQuery,
		surfaceID,
		userID,
	)
	if err != nil {
		return fmt.Errorf(
			"soft delete surface: %w",
			err,
		)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"soft delete surface: %w",
			err,
		)
	}

	if count == 0 {
		return fmt.Errorf(
			"soft delete surface: %w",
			domainsurface.ErrNotFound,
		)
	}

	return nil
}
