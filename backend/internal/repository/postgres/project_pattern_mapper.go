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

	return domainprojectpattern.ProjectPattern{
		ID:              row.ID,
		ProjectID:       row.ProjectID,
		SourcePatternID: nullInt64ToPtr(row.SourcePatternID),
		Name:            row.Name,
		Geometry:        polygon,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
		DeletedAt:       nil,
	}, nil
}

func (r *ProjectPatternRepository) projectPatternListRowToDomain(
	row ProjectPatternListRow,
) (domainprojectpattern.ProjectPattern, error) {
	polygon, err := r.geometryCodec.DecodeWKB(row.Geometry)
	if err != nil {
		return domainprojectpattern.ProjectPattern{}, fmt.Errorf(
			"decode project pattern geometry: %w",
			err,
		)
	}

	projectPattern := domainprojectpattern.ProjectPattern{
		ID:              row.ID.Int64,
		ProjectID:       row.ProjectID.Int64,
		SourcePatternID: nullInt64ToPtr(row.SourcePatternID),
		Name:            row.Name.String,
		Geometry:        polygon,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
		DeletedAt:       nil,
	}

	return projectPattern, nil
}
