package errors

import "errors"

var (
	ErrNotFound     = errors.New("resource not found")
	ErrForbidden    = errors.New("permission denied")
	ErrUnauthorized = errors.New("unauthorized")
	ErrConflict     = errors.New("resource conflict")
	ErrValidation   = errors.New("validation failed")
)

