package pattern

import "errors"

var (
	ErrNotFound      = errors.New("pattern not found")
	ErrOwnerNotFound = errors.New("pattern owner not found")
)
