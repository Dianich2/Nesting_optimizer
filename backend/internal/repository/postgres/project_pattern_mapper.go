package postgres

import (
	"fmt"
	domainprojectpattern "server_nesting_optimizer/internal/domain/project_pattern"
)

func (r *ProjectPatternRepository) projectPatternRowToDomain(
	row ProjectPatternRow,
) (domainprojectpattern.ProjectPattern, error) {
	polygon, err := r.geometryCodec.DecodeWKB(row.Geometry)
	if err != nil {
		return domainprojectpattern.ProjectPattern{}, fmt.Errorf(
			"decode project pattern geometry: %w",
			err,
		)
	}

	var sourcePatternID *int64
	if row.SourcePatternID.Valid {
		id := row.SourcePatternID.Int64
		sourcePatternID = &id
	}

	return domainprojectpattern.ProjectPattern{
		ID:              row.ID,
		ProjectID:       row.ProjectID,
		SourcePatternID: sourcePatternID,
		Name:            row.Name,
		Geometry:        polygon,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
		DeletedAt:       nil,
	}, nil
}
