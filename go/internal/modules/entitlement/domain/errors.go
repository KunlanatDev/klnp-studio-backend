package domain

import "errors"

var (
	ErrNotFound             = errors.New("entitlement not found")
	ErrFeatureDisabled      = errors.New("feature disabled")
	ErrFeatureNotConfigured = errors.New("feature not configured")
	ErrQuotaExceeded        = errors.New("quota exceeded")
	ErrInvalidLimit         = errors.New("invalid limit")
	ErrPlanInactive         = errors.New("plan inactive")
	ErrAssignmentConflict   = errors.New("assignment conflict")
	ErrPermissionDenied     = errors.New("permission denied")
)

