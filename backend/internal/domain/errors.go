package domain

import "errors"

var (
	// ErrInvalidFilter dikembalikan saat filter query metrik tidak valid.
	ErrInvalidFilter = errors.New("filter tidak valid")
	// ErrInvalidEvent dikembalikan saat event delivery tidak valid.
	ErrInvalidEvent = errors.New("event tidak valid")
	// ErrTargetNotFound dikembalikan saat target KPI cabang untuk bulan yang diminta tidak ada.
	ErrTargetNotFound = errors.New("target tidak ditemukan")
)
