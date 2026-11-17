package persistence_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/klnp-studio/backend/internal/modules/document/application"
	"github.com/klnp-studio/backend/internal/modules/document/domain"
	"github.com/klnp-studio/backend/internal/modules/document/infrastructure/persistence"
	"github.com/klnp-studio/backend/internal/platform/database"
)

// Integration tests run only when KLNP_TEST_DATABASE_URL points at a migrated
// local Supabase database (needs at least one auth.users row).
func setup(t *testing.T) (*gorm.DB, application.PageRef, string, *application.Service) {
	t.Helper()
	url := os.Getenv("KLNP_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("KLNP_TEST_DATABASE_URL not set")
	}
	db, err := database.NewPostgres(url)
	require.NoError(t, err)

	var userID string
	require.NoError(t, db.Raw(`SELECT id::text FROM auth.users LIMIT 1`).Scan(&userID).Error)
	require.NotEmpty(t, userID, "needs an auth.users row")

	ws, proj, page := uuid.NewString(), uuid.NewString(), uuid.NewString()
	require.NoError(t, db.Exec(`INSERT INTO workspaces (id, name, slug, created_by) VALUES (?, 'doc-it', ?, ?)`, ws, "doc-it-"+ws[:8], userID).Error)
	require.NoError(t, db.Exec(`INSERT INTO projects (id, workspace_id, name, slug, platform, status, created_by) VALUES (?, ?, 'P', ?, 'mobile', 'active', ?)`, proj, ws, "p-"+proj[:8], userID).Error)
	require.NoError(t, db.Exec(`INSERT INTO pages (id, workspace_id, project_id, name, slug, status, created_by) VALUES (?, ?, ?, 'Home', 'home', 'active', ?)`, page, ws, proj, userID).Error)
	t.Cleanup(func() { db.Exec(`DELETE FROM workspaces WHERE id = ?`, ws) })

	svc := application.NewService(persistence.NewDocumentRepository(db))
	return db, application.PageRef{WorkspaceID: ws, ProjectID: proj, PageID: page}, userID, svc
}

func addNode(id string) json.RawMessage {
	return json.RawMessage(fmt.Sprintf(`[
	 {"op":"add","path":"/nodes/%[1]s","value":{"id":"%[1]s","kind":"rect","children":[]}},
	 {"op":"add","path":"/rootNodeIds/-","value":"%[1]s"}]`, id))
}

func TestDocumentRepository_EndToEnd(t *testing.T) {
	db, ref, user, svc := setup(t)
	ctx := context.Background()

	// never-edited page: virtual empty document at revision 1, nothing written
	doc, err := svc.Get(ctx, ref)
	require.NoError(t, err)
	assert.Equal(t, int64(1), doc.Revision)
	var rows int64
	db.Raw(`SELECT count(*) FROM page_documents WHERE page_id = ?`, ref.PageID).Scan(&rows)
	assert.Zero(t, rows)

	// first save
	doc, err = svc.ApplyPatch(ctx, application.ApplyPatchInput{Page: ref, ActorID: user, BaseRevision: 1, Patch: addNode("a")})
	require.NoError(t, err)
	assert.Equal(t, int64(2), doc.Revision)

	// reload keeps content
	doc, err = svc.Get(ctx, ref)
	require.NoError(t, err)
	assert.Equal(t, int64(2), doc.Revision)
	parsed, err := domain.ParseDocument(doc.Content)
	require.NoError(t, err)
	assert.Contains(t, parsed.Nodes, "a")

	// pages.document_version follows the revision
	var dv int
	db.Raw(`SELECT document_version FROM pages WHERE id = ?`, ref.PageID).Scan(&dv)
	assert.Equal(t, 2, dv)

	// stale revision conflicts without writing
	_, err = svc.ApplyPatch(ctx, application.ApplyPatchInput{Page: ref, ActorID: user, BaseRevision: 1, Patch: addNode("b")})
	assert.ErrorIs(t, err, domain.ErrRevisionConflict)

	// invalid graph is rejected and rolled back
	_, err = svc.ApplyPatch(ctx, application.ApplyPatchInput{Page: ref, ActorID: user, BaseRevision: 2,
		Patch: json.RawMessage(`[{"op":"add","path":"/rootNodeIds/-","value":"ghost"}]`)})
	assert.ErrorIs(t, err, domain.ErrInvalidDocument)
	doc, _ = svc.Get(ctx, ref)
	assert.Equal(t, int64(2), doc.Revision)

	// operations are recorded in order
	ops, err := svc.ListOperations(ctx, ref, 0, 10)
	require.NoError(t, err)
	require.Len(t, ops, 1)
	assert.Equal(t, int64(2), ops[0].Revision)
	assert.Equal(t, int64(1), ops[0].BaseRevision)

	// page from another workspace/project is not addressable
	other := ref
	other.WorkspaceID = uuid.NewString()
	_, err = svc.Get(ctx, other)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	_, err = svc.ApplyPatch(ctx, application.ApplyPatchInput{Page: other, ActorID: user, BaseRevision: 2, Patch: addNode("z")})
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestDocumentRepository_ConcurrentWritersNeverLoseUpdates(t *testing.T) {
	_, ref, user, svc := setup(t)
	ctx := context.Background()

	const writers = 12
	var wins, conflicts int32
	var wg sync.WaitGroup
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := svc.ApplyPatch(ctx, application.ApplyPatchInput{
				Page: ref, ActorID: user, BaseRevision: 1, Patch: addNode(fmt.Sprintf("n%d", i)),
			})
			switch {
			case err == nil:
				atomic.AddInt32(&wins, 1)
			case assert.ErrorIs(t, err, domain.ErrRevisionConflict):
				atomic.AddInt32(&conflicts, 1)
			}
		}(i)
	}
	wg.Wait()

	assert.Equal(t, int32(1), wins, "exactly one writer wins a given base revision")
	assert.Equal(t, int32(writers-1), conflicts)
	doc, err := svc.Get(ctx, ref)
	require.NoError(t, err)
	assert.Equal(t, int64(2), doc.Revision)
}

func TestDocumentRepository_CheckpointEvery50Revisions(t *testing.T) {
	db, ref, user, svc := setup(t)
	ctx := context.Background()

	for rev := int64(1); rev < domain.SnapshotEvery; rev++ {
		_, err := svc.ApplyPatch(ctx, application.ApplyPatchInput{
			Page: ref, ActorID: user, BaseRevision: rev, Patch: addNode(fmt.Sprintf("n%d", rev)),
		})
		require.NoError(t, err)
	}

	var revs []int64
	require.NoError(t, db.Raw(`SELECT revision FROM document_versions WHERE page_id = ? AND is_checkpoint ORDER BY revision`, ref.PageID).Scan(&revs).Error)
	assert.Equal(t, []int64{domain.SnapshotEvery}, revs)

	var nOps int64
	db.Raw(`SELECT count(*) FROM document_operations WHERE page_id = ?`, ref.PageID).Scan(&nOps)
	assert.Equal(t, int64(domain.SnapshotEvery-1), nOps)
}
