package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
