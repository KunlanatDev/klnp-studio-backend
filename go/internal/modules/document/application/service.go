package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch/v5"

	"github.com/klnp-studio/backend/internal/modules/document/domain"
)

const (
	DefaultOperationsLimit = 100
	MaxOperationsLimit     = 500
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

type PageRef struct {
	WorkspaceID string
	ProjectID   string
	PageID      string
}

func (r PageRef) valid() bool {
	return r.WorkspaceID != "" && r.ProjectID != "" && r.PageID != ""
}

func (s *Service) Get(ctx context.Context, ref PageRef) (*domain.PageDocument, error) {
	if !ref.valid() {
		return nil, domain.ErrInvalidInput
	}
	return s.repo.Get(ctx, ref.WorkspaceID, ref.ProjectID, ref.PageID)
}

type ApplyPatchInput struct {
	Page         PageRef
	ActorID      string
	BaseRevision int64
	Patch        json.RawMessage // RFC 6902 array
}

// ApplyPatch applies a JSON Patch on top of baseRevision. A stale baseRevision
// is rejected with *domain.ConflictError so no edit is ever silently lost.
// A single `replace` on the root path ("") replaces the whole document.
func (s *Service) ApplyPatch(ctx context.Context, in ApplyPatchInput) (*domain.PageDocument, error) {
	if !in.Page.valid() || in.ActorID == "" || in.BaseRevision < 1 {
		return nil, domain.ErrInvalidInput
	}
	if len(in.Patch) == 0 || len(in.Patch) > domain.MaxPatchBytes {
		return nil, fmt.Errorf("%w: patch must be 1..%d bytes", domain.ErrInvalidPatch, domain.MaxPatchBytes)
	}

	return s.repo.Mutate(ctx, in.Page.WorkspaceID, in.Page.ProjectID, in.Page.PageID, in.BaseRevision, in.ActorID,
		func(current *domain.PageDocument) ([]byte, []byte, error) {
			next, err := applyPatch(current.Content, in.Patch)
			if err != nil {
				return nil, nil, err
			}
			if err := validateContent(next); err != nil {
				return nil, nil, err
			}
			return next, in.Patch, nil
		})
}

func (s *Service) ListOperations(ctx context.Context, ref PageRef, after int64, limit int) ([]domain.Operation, error) {
	if !ref.valid() || after < 0 {
		return nil, domain.ErrInvalidInput
	}
	if limit <= 0 {
		limit = DefaultOperationsLimit
	}
	if limit > MaxOperationsLimit {
		limit = MaxOperationsLimit
	}
	return s.repo.ListOperations(ctx, ref.WorkspaceID, ref.ProjectID, ref.PageID, after, limit)
}

func validateContent(content []byte) error {
	doc, err := domain.ParseDocument(content)
	if err != nil {
		return err
	}
	return doc.Validate()
}

type patchOp struct {
	Op    string          `json:"op"`
	Path  string          `json:"path"`
	Value json.RawMessage `json:"value"`
}

func applyPatch(content, patch []byte) ([]byte, error) {
	var ops []patchOp
	if err := json.Unmarshal(patch, &ops); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidPatch, err)
	}
	if len(ops) == 0 {
		return nil, fmt.Errorf("%w: patch has no operations", domain.ErrInvalidPatch)
	}

	if len(ops) == 1 && ops[0].Op == "replace" && ops[0].Path == "" {
		if len(ops[0].Value) == 0 {
			return nil, fmt.Errorf("%w: root replace requires a value", domain.ErrInvalidPatch)
		}
		return ops[0].Value, nil
	}
	for _, o := range ops {
		if o.Path == "" {
			return nil, fmt.Errorf("%w: root path is only allowed as a single replace", domain.ErrInvalidPatch)
		}
	}

	decoded, err := jsonpatch.DecodePatch(patch)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidPatch, err)
	}
	out, err := decoded.Apply(content)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidPatch, err)
	}
	return out, nil
}

// IsClientError reports whether err was caused by the request (4xx).
func IsClientError(err error) bool {
	return errors.Is(err, domain.ErrInvalidInput) ||
		errors.Is(err, domain.ErrInvalidPatch) ||
		errors.Is(err, domain.ErrInvalidDocument) ||
		errors.Is(err, domain.ErrRevisionConflict) ||
		errors.Is(err, domain.ErrDocumentTooLarge) ||
		errors.Is(err, domain.ErrNotFound)
}
