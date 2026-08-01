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

type GetCurrentUserInput struct {
	ID int64
}

type GetCurrentUserOutput struct {
	ID        int64
	Login     string
	Email     string
	FirstName string
	LastName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UpdateProfileInput struct {
	FirstName *string
	LastName  *string
}

type UpdateProfileOutput struct {
	ID        int64
	Login     string
	Email     string
	FirstName string
	LastName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ChangePasswordInput struct {
	OldPassword       string
	NewPassword       string
	RepeatNewPassword string
}
