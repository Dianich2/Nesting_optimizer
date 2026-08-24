package placement

import "errors"

var (
	ErrNotFound       = errors.New("placement not found")
	ErrOutsideSurface = errors.New("placement outside surface")
	ErrOverlap        = errors.New("placement overlaps another placement")
)
