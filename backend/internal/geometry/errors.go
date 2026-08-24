package geometry

import "errors"

var (
	ErrInvalidPolygon   = errors.New("invalid polygon")
	ErrInvalidScale     = errors.New("invalid scale factor")
	ErrInvalidWKB       = errors.New("invalid geometry WKB")
	ErrInvalidTransform = errors.New("invalid geometry transform")
)
