package project

import "time"

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
