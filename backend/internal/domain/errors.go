package domain

import "errors"

var (
	// ErrInvalidFilter dikembalikan saat filter query metrik tidak valid.
	ErrInvalidFilter = errors.New("filter tidak valid")
	// ErrInvalidEvent dikembalikan saat event delivery tidak valid.
	ErrInvalidEvent = errors.New("event tidak valid")
)
