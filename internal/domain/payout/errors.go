package payout

import "errors"

var (
	ErrNotFound      = errors.New("payout not found")
	ErrInvalidInput  = errors.New("invalid input")
	ErrAlreadyExists = errors.New("payout for this period already exists")
	ErrAlreadyPaid   = errors.New("payout already paid")
)
