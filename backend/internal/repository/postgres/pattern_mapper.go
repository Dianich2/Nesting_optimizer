package postgres

import (
	"fmt"
	domainpattern "server_nesting_optimizer/internal/domain/pattern"
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

	return domainpattern.Pattern{
		ID:        row.ID,
		UserID:    row.UserID,
		Name:      row.Name,
		Geometry:  polygon,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		DeletedAt: nullTimeToPtr(row.DeletedAt),
	}, nil
}

func (r *PatternRepository) patternListRowToDomain(
	row PatternListRow,
) (domainpattern.Pattern, error) {
	polygon, err := r.geometryCodec.DecodeWKB(row.Geometry)
	if err != nil {
		return domainpattern.Pattern{}, fmt.Errorf(
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

	return pattern, nil
}
