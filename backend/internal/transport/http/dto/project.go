package dto

import "time"

type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateProjectResponse struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type GetProjectByIDResponse struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ListProjectsItemResponse struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
type ListProjectsResponse struct {
	Items      []ListProjectsItemResponse `json:"items"`
	Page       int                        `json:"page"`
	PageSize   int                        `json:"page_size"`
	Total      int64                      `json:"total"`
	TotalPages int64                      `json:"total_pages"`
}
