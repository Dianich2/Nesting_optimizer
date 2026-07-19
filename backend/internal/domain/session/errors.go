package session

import "errors"

var ErrNotFound = errors.New("session not found")
var ErrSessionChanged = errors.New("session changed")
