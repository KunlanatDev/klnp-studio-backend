package domain

import "errors"

var (
	ErrNotFound      = errors.New("project not found")
	ErrInvalidInput  = errors.New("invalid project input")
	ErrSlugConflict  = errors.New("project slug conflict")
	ErrQuotaExceeded = errors.New("workspace page quota exceeded")
)

