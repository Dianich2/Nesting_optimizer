package user

import "time"

type CreateUserInput struct {
	Login     string
	Email     string
	Password  string
	FirstName string
	LastName  string
}

type CreateUserOutput struct {
	ID        int64
	Login     string
	Email     string
	FirstName string
	LastName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
