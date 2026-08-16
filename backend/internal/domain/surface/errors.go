package surface

import "errors"

var (
	ErrNotFound      = errors.New("surface not found")
	ErrOwnerNotFound = errors.New("surface owner not found")
)
