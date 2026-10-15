package breakdown

import "errors"

var (
	ErrNotFound          = errors.New("breakdown not found")
	ErrInvalidInput      = errors.New("invalid input")
	ErrCannotDeleteFixed = errors.New("cannot delete fixed breakdown")
)
