package bike

import "errors"

var (
	ErrNotFound     = errors.New("bike not found")
	ErrInvalidInput = errors.New("invalid input")
)
