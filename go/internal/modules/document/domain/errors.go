package domain

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNotFound         = errors.New("page document not found")
	ErrInvalidInput     = errors.New("invalid document input")
	ErrInvalidDocument  = errors.New("invalid design document")
	ErrInvalidPatch     = errors.New("invalid json patch")
	ErrRevisionConflict = errors.New("document revision conflict")
	ErrDocumentTooLarge = errors.New("document exceeds size limit")
)

// ConflictError carries the server revision so clients can rebase.
type ConflictError struct {
	CurrentRevision int64
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("document revision conflict (current revision %d)", e.CurrentRevision)
}

func (e *ConflictError) Unwrap() error { return ErrRevisionConflict }

// Validation issue codes (ADR-05 invariants).
const (
	IssueUnsupportedVersion = "UNSUPPORTED_SCHEMA_VERSION"
	IssueMissingNode        = "MISSING_NODE"
	IssueIDMismatch         = "NODE_ID_MISMATCH"
	IssueParentMismatch     = "PARENT_MISMATCH"
	IssueMultipleParents    = "MULTIPLE_PARENTS"
	IssueCycle              = "CYCLE"
	IssueOrphan             = "ORPHAN_NODE"
	IssueInvalidKind        = "INVALID_KIND"
	IssueTooManyNodes       = "TOO_MANY_NODES"
	IssueDuplicateChild     = "DUPLICATE_CHILD"
)

type ValidationIssue struct {
	Code    string `json:"code"`
	NodeID  string `json:"node_id,omitempty"`
	Message string `json:"message"`
}

// ValidationError aggregates every invariant violation found in a document.
type ValidationError struct {
	Issues []ValidationIssue
}

func (e *ValidationError) Error() string {
	parts := make([]string, 0, len(e.Issues))
	for i, is := range e.Issues {
		if i == 5 {
			parts = append(parts, fmt.Sprintf("... and %d more", len(e.Issues)-5))
			break
		}
		parts = append(parts, is.Code+": "+is.Message)
	}
	return "invalid design document: " + strings.Join(parts, "; ")
}

func (e *ValidationError) Unwrap() error { return ErrInvalidDocument }
