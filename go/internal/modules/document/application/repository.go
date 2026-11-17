package application

import (
	"context"

	"github.com/klnp-studio/backend/internal/modules/document/domain"
)

// MutateFunc computes the next document content (and the patch that records it)
// from the locked current state. Returning an error aborts the transaction.
type MutateFunc func(current *domain.PageDocument) (newContent []byte, patch []byte, err error)

type Repository interface {
	// Get returns the current document. A page that was never edited yields the
	// empty document at revision 1 without writing to the database.
	Get(ctx context.Context, workspaceID, projectID, pageID string) (*domain.PageDocument, error)

	// Mutate atomically locks the page document, verifies baseRevision, runs fn,
	// stores the new content with revision+1, appends the operation and creates a
	// checkpoint version every domain.SnapshotEvery revisions.
	Mutate(ctx context.Context, workspaceID, projectID, pageID string, baseRevision int64, actorID string, fn MutateFunc) (*domain.PageDocument, error)

	ListOperations(ctx context.Context, workspaceID, projectID, pageID string, afterRevision int64, limit int) ([]domain.Operation, error)
}
