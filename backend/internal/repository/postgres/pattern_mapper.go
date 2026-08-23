package postgres

import (
	"fmt"
	domainpattern "server_nesting_optimizer/internal/domain/pattern"
	"time"
)

func (r *PatternRepository) patternRowToDomain(
	row PatternRow,
) (domainpattern.Pattern, error) {
	polygon, err := r.geometryCodec.DecodeWKB(row.Geometry)
	if err != nil {
		return domainpattern.Pattern{}, fmt.Errorf(
			"decode pattern geometry: %w",
			err,
		)
	}

	var deletedAt *time.Time
	if row.DeletedAt.Valid {
		t := row.DeletedAt.Time
		deletedAt = &t
	}

	return domainpattern.Pattern{
		ID:        row.ID,
		UserID:    row.UserID,
		Name:      row.Name,
		Geometry:  polygon,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		DeletedAt: deletedAt,
	}, nil
}
