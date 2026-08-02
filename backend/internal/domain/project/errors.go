package project

import "errors"

var (
	ErrNotFound      = errors.New("project not found")
	ErrOwnerNotFound = errors.New("project owner not found")
)
