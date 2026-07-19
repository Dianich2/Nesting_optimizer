package user

import "errors"

var (
	ErrNotFound           = errors.New("user not found")
	ErrLoginAlreadyExists = errors.New("user login already exists")
	ErrEmailAlreadyExists = errors.New("user email already exists")
)
