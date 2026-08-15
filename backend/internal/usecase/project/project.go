package project

import (
	domainproject "server_nesting_optimizer/internal/domain/project"
	"time"
)

type CreateProjectInput struct {
	UserID      int64
	Name        string
	Description string
}

type CreateProjectOutput struct {
	ID          int64
	UserID      int64
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type GetProjectByIDInput struct {
	ProjectID int64
	UserID    int64
}

type GetProjectByIDOutput struct {
	ID          int64
	UserID      int64
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ListProjectsInput struct {
	UserID   int64
	Page     int
	PageSize int
}

type ProjectListResult struct {
	Projects []domainproject.Project
	Total    int64
}

type ListProjectsItem struct {
	ID          int64
	UserID      int64
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ListProjectsOutput struct {
	Items      []ListProjectsItem
	Page       int
	PageSize   int
	Total      int64
	TotalPages int64
}
