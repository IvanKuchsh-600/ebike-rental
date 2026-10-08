package rental

import "errors"

var (
	ErrNotFound         = errors.New("rental not found")
	ErrInvalidInput     = errors.New("invalid input")
	ErrBikeNotAvailable = errors.New("bike is not available")
	ErrBikeNotFound     = errors.New("bike not found")
	ErrAlreadyReturned  = errors.New("rental already returned")
)
