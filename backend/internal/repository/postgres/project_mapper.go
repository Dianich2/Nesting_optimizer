package postgres

import (
	domainproject "server_nesting_optimizer/internal/domain/project"
)

func (r *ProjectRepository) projectListRowToDomain(
	row ProjectListRow,
) domainproject.Project {
	project := domainproject.Project{
		ID:          row.ID.Int64,
		UserID:      row.UserID.Int64,
		Name:        row.Name.String,
		Description: row.Description.String,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
		DeletedAt:   nil,
	}

	return project
}
