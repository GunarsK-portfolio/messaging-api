package repository

import "errors"

// ErrNotFound is returned when a resource is not found
var ErrNotFound = errors.New("resource not found")
