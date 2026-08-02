package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	domainproject "server_nesting_optimizer/internal/domain/project"
	projectusecase "server_nesting_optimizer/internal/usecase/project"
)

type ProjectRepository struct {
	db DBTX
}

func NewProjectRepository(db DBTX) *ProjectRepository {
	return &ProjectRepository{
		db: db,
	}
}

var _ projectusecase.ProjectRepository = (*ProjectRepository)(nil)

func (r *ProjectRepository) Create(
	ctx context.Context,
	project domainproject.Project,
) (domainproject.Project, error) {
	var createdProject domainproject.Project
	if err := r.db.GetContext(
		ctx,
		&createdProject,
		createProjectQuery,
		project.UserID,
		project.Name,
		project.Description,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainproject.Project{}, fmt.Errorf(
				"create project: %w",
				domainproject.ErrOwnerNotFound,
			)
		}

		return domainproject.Project{}, fmt.Errorf(
			"create project: %w",
			err,
		)
	}

	return createdProject, nil
}
