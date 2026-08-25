package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	domainplacement "server_nesting_optimizer/internal/domain/placement"
	"server_nesting_optimizer/internal/geometry"
	placementusecase "server_nesting_optimizer/internal/usecase/placement"
	"time"
)

type PlacementRepository struct {
	db            DBTX
	geometryCodec geometry.Codec
}

func NewPlacementRepository(
	db DBTX,
	geometryCodec geometry.Codec,
) *PlacementRepository {
	return &PlacementRepository{
		db:            db,
		geometryCodec: geometryCodec,
	}
}

type PlacementRow struct {
	ID               int64     `db:"id"`
	ProjectSurfaceID int64     `db:"project_surface_id"`
	ProjectPatternID int64     `db:"project_pattern_id"`
	X                float64   `db:"x"`
	Y                float64   `db:"y"`
	Rotation         float64   `db:"rotation"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

func (r *PlacementRepository) Create(
	ctx context.Context,
	input domainplacement.Placement,
	projectID int64,
	userID int64,
) (domainplacement.Placement, error) {
	var placement PlacementRow
	if err := r.db.GetContext(
		ctx,
		&placement,
		createPlacementQuery,
		projectID,
		input.ProjectSurfaceID,
		input.ProjectPatternID,
		input.X,
		input.Y,
		input.Rotation,
		userID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainplacement.Placement{}, domainplacement.ErrNotFound
		}

		return domainplacement.Placement{}, fmt.Errorf(
			"create placement: %w",
			err,
		)
	}

	return placementRowToDomain(placement), nil
}

type CollisionPlacementRow struct {
	ID              int64   `db:"id"`
	PatternGeometry []byte  `db:"pattern_geometry"`
	X               float64 `db:"x"`
	Y               float64 `db:"y"`
	Rotation        float64 `db:"rotation"`
}

func (r *PlacementRepository) ListForCollisionCheck(
	ctx context.Context,
	projectSurfaceID int64,
	projectID int64,
	userID int64,
) ([]placementusecase.CollisionPlacement, error) {
	var rows []CollisionPlacementRow

	if err := r.db.SelectContext(
		ctx,
		&rows,
		listPlacementsForCollisionCheckQuery,
		projectSurfaceID,
		projectID,
		userID,
	); err != nil {
		return nil, fmt.Errorf(
			"list placements for collision check: %w",
			err,
		)
	}

	placements := make([]placementusecase.CollisionPlacement, 0, len(rows))

	for _, row := range rows {
		placement, err := r.toCollisionPlacement(row)
		if err != nil {
			return nil, err
		}

		placements = append(placements, placement)
	}

	return placements, nil
}

type PlacementWithPatternGeometryDB struct {
	ID               int64     `db:"id"`
	ProjectSurfaceID int64     `db:"project_surface_id"`
	ProjectPatternID int64     `db:"project_pattern_id"`
	X                float64   `db:"x"`
	Y                float64   `db:"y"`
	Rotation         float64   `db:"rotation"`
	PatternGeometry  []byte    `db:"pattern_geometry"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

func (r *PlacementRepository) GetByIDWithPatternGeometry(
	ctx context.Context,
	userID int64,
	projectID int64,
	placementID int64,
) (placementusecase.PlacementWithPatternGeometry, error) {
	var placementWithPattern PlacementWithPatternGeometryDB
	if err := r.db.GetContext(
		ctx,
		&placementWithPattern,
		getPlacementByIDQuery,
		placementID,
		projectID,
		userID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return placementusecase.PlacementWithPatternGeometry{},
				domainplacement.ErrNotFound
		}

		return placementusecase.PlacementWithPatternGeometry{}, fmt.Errorf(
			"get placement by id: %w",
			err,
		)
	}

	polygon, err := r.geometryCodec.DecodeWKB(placementWithPattern.PatternGeometry)
	if err != nil {
		return placementusecase.PlacementWithPatternGeometry{}, fmt.Errorf(
			"decode placement pattern geometry: %w",
			err,
		)
	}

	return placementusecase.PlacementWithPatternGeometry{
		Placement: domainplacement.Placement{
			ID:               placementWithPattern.ID,
			ProjectSurfaceID: placementWithPattern.ProjectSurfaceID,
			ProjectPatternID: placementWithPattern.ProjectPatternID,
			X:                placementWithPattern.X,
			Y:                placementWithPattern.Y,
			Rotation:         placementWithPattern.Rotation,
			CreatedAt:        placementWithPattern.CreatedAt,
			UpdatedAt:        placementWithPattern.UpdatedAt,
		},
		PatternGeometry: polygon,
	}, nil
}
