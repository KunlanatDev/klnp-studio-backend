package domain

import "errors"

var (
	ErrNotFound         = errors.New("page not found")
	ErrInvalidInput     = errors.New("invalid page input")
	ErrPathConflict     = errors.New("page path conflict in project")
	ErrQuotaExceeded    = errors.New("workspace page quota exceeded")
	ErrCannotDeleteHome = errors.New("cannot delete home page")
)
