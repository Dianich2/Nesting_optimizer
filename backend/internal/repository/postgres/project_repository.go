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

func (r *ProjectRepository) GetByID(
	ctx context.Context,
	projectID int64,
	userID int64,
) (domainproject.Project, error) {
	var project domainproject.Project
	if err := r.db.GetContext(
		ctx,
		&project,
		getProjectByIDQuery,
		projectID,
		userID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainproject.Project{}, fmt.Errorf(
				"get project by id: %w",
				domainproject.ErrNotFound,
			)
		}

		return domainproject.Project{}, fmt.Errorf(
			"get project by id: %w",
			err,
		)
	}

	return project, nil
}

type ProjectListRow struct {
	ID          sql.NullInt64  `db:"id"`
	UserID      sql.NullInt64  `db:"user_id"`
	Name        sql.NullString `db:"name"`
	Description sql.NullString `db:"description"`
	CreatedAt   sql.NullTime   `db:"created_at"`
	UpdatedAt   sql.NullTime   `db:"updated_at"`
	Total       int64          `db:"total"`
}

func (r *ProjectRepository) ListByUserID(
	ctx context.Context,
	userID int64,
	limit int,
	offset int,
) (projectusecase.ProjectListResult, error) {
	rows := make([]ProjectListRow, 0)

	if err := r.db.SelectContext(
		ctx,
		&rows,
		listProjectsQuery,
		userID,
		limit,
		offset,
	); err != nil {
		return projectusecase.ProjectListResult{}, fmt.Errorf(
			"list projects by user id: %w",
			err,
		)
	}

	listOfProjects := projectusecase.ProjectListResult{
		Projects: make([]domainproject.Project, 0),
		Total:    0,
	}

	for _, row := range rows {
		listOfProjects.Total = row.Total

		if !row.ID.Valid {
			continue
		}

		project := r.projectListRowToDomain(row)

		listOfProjects.Projects = append(
			listOfProjects.Projects,
			project,
		)
	}

	return listOfProjects, nil
}

func (r *ProjectRepository) Update(
	ctx context.Context,
	projectID int64,
	userID int64,
	name *string,
	description *string,
) (domainproject.Project, error) {
	var project domainproject.Project
	if err := r.db.GetContext(
		ctx,
		&project,
		updateProjectQuery,
		name,
		description,
		projectID,
		userID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainproject.Project{}, fmt.Errorf(
				"update project: %w",
				domainproject.ErrNotFound,
			)
		}

		return domainproject.Project{}, fmt.Errorf(
			"update project: %w",
			err,
		)
	}

	return project, nil
}

func (r *ProjectRepository) SoftDelete(
	ctx context.Context,
	projectID int64,
	userID int64,
) error {
	affected, err := execAffected(
		ctx,
		r.db,
		softDeleteProjectQuery,
		projectID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("soft delete project: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf(
			"soft delete project: %w",
			domainproject.ErrNotFound,
		)
	}

	return nil
}
