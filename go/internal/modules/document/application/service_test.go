package application_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/klnp-studio/backend/internal/modules/document/application"
	"github.com/klnp-studio/backend/internal/modules/document/domain"
)

// memRepo is an in-memory Repository with the same revision semantics as Postgres.
type memRepo struct {
	doc *domain.PageDocument
	ops []domain.Operation
}

func newMemRepo() *memRepo {
	return &memRepo{doc: &domain.PageDocument{PageID: "p", Revision: 1, Content: json.RawMessage(domain.EmptyDocument)}}
}

func (m *memRepo) Get(_ context.Context, _, _, _ string) (*domain.PageDocument, error) {
	return m.doc, nil
}

func (m *memRepo) Mutate(_ context.Context, _, _, _ string, base int64, actor string, fn application.MutateFunc) (*domain.PageDocument, error) {
	if m.doc.Revision != base {
		return nil, &domain.ConflictError{CurrentRevision: m.doc.Revision}
	}
	content, patch, err := fn(m.doc)
	if err != nil {
		return nil, err
	}
	next := *m.doc
	next.Content = content
	next.Revision++
	m.ops = append(m.ops, domain.Operation{BaseRevision: base, Revision: next.Revision, Patch: patch})
	m.doc = &next
	return m.doc, nil
}

func (m *memRepo) ListOperations(_ context.Context, _, _, _ string, after int64, limit int) ([]domain.Operation, error) {
	var out []domain.Operation
	for _, o := range m.ops {
		if o.Revision > after && len(out) < limit {
			out = append(out, o)
		}
	}
	return out, nil
}

var ref = application.PageRef{WorkspaceID: "w", ProjectID: "pr", PageID: "p"}

func apply(s *application.Service, base int64, patch string) (*domain.PageDocument, error) {
	return s.ApplyPatch(context.Background(), application.ApplyPatchInput{
		Page: ref, ActorID: "u", BaseRevision: base, Patch: json.RawMessage(patch),
	})
}

const addFrame = `[
 {"op":"add","path":"/nodes/f","value":{"id":"f","kind":"frame","children":[]}},
 {"op":"add","path":"/rootNodeIds/-","value":"f"}]`

func TestApplyPatch_AddsNodeAndBumpsRevision(t *testing.T) {
	repo := newMemRepo()
	s := application.NewService(repo)

	doc, err := apply(s, 1, addFrame)
	require.NoError(t, err)
	assert.Equal(t, int64(2), doc.Revision)
	assert.Contains(t, string(doc.Content), `"f"`)
	require.Len(t, repo.ops, 1)
	assert.Equal(t, int64(1), repo.ops[0].BaseRevision)
}

func TestApplyPatch_StaleRevisionConflicts(t *testing.T) {
	repo := newMemRepo()
	s := application.NewService(repo)
	_, err := apply(s, 1, addFrame)
	require.NoError(t, err)

	_, err = apply(s, 1, `[{"op":"replace","path":"/schemaVersion","value":"1.1.0"}]`)
	assert.ErrorIs(t, err, domain.ErrRevisionConflict)
	var ce *domain.ConflictError
	require.ErrorAs(t, err, &ce)
	assert.Equal(t, int64(2), ce.CurrentRevision)
	assert.Equal(t, int64(2), repo.doc.Revision, "failed patch must not change state")
}

func TestApplyPatch_RejectsGraphViolations(t *testing.T) {
	repo := newMemRepo()
	s := application.NewService(repo)
	// child referenced but never created
	_, err := apply(s, 1, `[{"op":"add","path":"/rootNodeIds/-","value":"ghost"}]`)
	assert.ErrorIs(t, err, domain.ErrInvalidDocument)
	assert.Equal(t, int64(1), repo.doc.Revision)
	assert.Empty(t, repo.ops)
}

func TestApplyPatch_RejectsBadPatches(t *testing.T) {
	s := application.NewService(newMemRepo())
	for name, patch := range map[string]string{
		"not json":       `nope`,
		"empty array":    `[]`,
		"unknown op":     `[{"op":"explode","path":"/x"}]`,
		"missing path":   `[{"op":"remove","path":"/nodes/zzz"}]`,
		"root add":       `[{"op":"add","path":"","value":{}}]`,
		"root w/ others": `[{"op":"replace","path":"","value":{}},{"op":"add","path":"/a","value":1}]`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := apply(s, 1, patch)
			assert.ErrorIs(t, err, domain.ErrInvalidPatch)
		})
	}
}

func TestApplyPatch_RootReplaceSavesWholeDocument(t *testing.T) {
	repo := newMemRepo()
	s := application.NewService(repo)
	full := `{"schemaVersion":"1.1.0","rootNodeIds":["a"],"nodes":{"a":{"id":"a","kind":"rect","children":[]}}}`
	doc, err := apply(s, 1, `[{"op":"replace","path":"","value":`+full+`}]`)
	require.NoError(t, err)
	assert.JSONEq(t, full, string(doc.Content))
}

func TestApplyPatch_RootReplaceStillValidated(t *testing.T) {
	s := application.NewService(newMemRepo())
	_, err := apply(s, 1, `[{"op":"replace","path":"","value":{"schemaVersion":"1.1.0","rootNodeIds":["x"],"nodes":{}}}]`)
	assert.ErrorIs(t, err, domain.ErrInvalidDocument)
}

func TestApplyPatch_InputValidation(t *testing.T) {
	s := application.NewService(newMemRepo())
	_, err := s.ApplyPatch(context.Background(), application.ApplyPatchInput{Page: ref, BaseRevision: 1, Patch: json.RawMessage(addFrame)})
	assert.ErrorIs(t, err, domain.ErrInvalidInput, "missing actor")
	_, err = apply(s, 0, addFrame)
	assert.ErrorIs(t, err, domain.ErrInvalidInput, "revision must be >= 1")
	_, err = s.Get(context.Background(), application.PageRef{})
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
}

func TestListOperations_ClampsLimit(t *testing.T) {
	repo := newMemRepo()
	s := application.NewService(repo)
	_, _ = apply(s, 1, addFrame)
	ops, err := s.ListOperations(context.Background(), ref, 0, 100000)
	require.NoError(t, err)
	assert.Len(t, ops, 1)
	_, err = s.ListOperations(context.Background(), ref, -1, 10)
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
}
