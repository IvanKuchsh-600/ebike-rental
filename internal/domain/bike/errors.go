package bike

import "errors"

var (
	ErrNotFound     = errors.New("bike not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrHasRentals   = errors.New("bike has rentals and cannot be deleted")
)
