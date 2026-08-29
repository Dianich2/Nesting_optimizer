package nesting

import "errors"

var (
	ErrInvalidQuantity = errors.New("invalid quantity")
	ErrEmptyRotations  = errors.New("allowed rotations list is empty")
	ErrInvalidRotation = errors.New("invalid rotation")
)
