package persistence

import (
	"context"
	"encoding/json"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/klnp-studio/backend/internal/modules/document/application"
	"github.com/klnp-studio/backend/internal/modules/document/domain"
)

type DocumentRepository struct {
	db *gorm.DB
}

var _ application.Repository = (*DocumentRepository)(nil)

func NewDocumentRepository(db *gorm.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

type documentRow struct {
	PageID    string
	Revision  int64
	Document  string
	UpdatedBy *string
	UpdatedAt time.Time
}

type operationRow struct {
	ID           string
	PageID       string
	BaseRevision int64
	Revision     int64
	Patch        string
	CreatedBy    *string
	CreatedAt    time.Time
}

func (r *DocumentRepository) pageExists(tx *gorm.DB, workspaceID, projectID, pageID string) error {
	var n int64
	err := tx.Table("pages").
		Where("id = ? AND project_id = ? AND workspace_id = ? AND status = 'active'", pageID, projectID, workspaceID).
		Count(&n).Error
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func toPageDocument(row documentRow, workspaceID, projectID string) *domain.PageDocument {
	return &domain.PageDocument{
		PageID:      row.PageID,
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		Revision:    row.Revision,
		Content:     json.RawMessage(row.Document),
		UpdatedBy:   row.UpdatedBy,
		UpdatedAt:   row.UpdatedAt,
	}
}

func (r *DocumentRepository) Get(ctx context.Context, workspaceID, projectID, pageID string) (*domain.PageDocument, error) {
	db := r.db.WithContext(ctx)
	if err := r.pageExists(db, workspaceID, projectID, pageID); err != nil {
		return nil, err
	}

	var row documentRow
	res := db.Raw(
		`SELECT page_id::text AS page_id, revision, document::text AS document,
		        updated_by::text AS updated_by, updated_at
		   FROM page_documents WHERE page_id = ?`, pageID).Scan(&row)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return &domain.PageDocument{
			PageID: pageID, WorkspaceID: workspaceID, ProjectID: projectID,
			Revision: 1, Content: json.RawMessage(domain.EmptyDocument), UpdatedAt: time.Now().UTC(),
		}, nil
	}
	return toPageDocument(row, workspaceID, projectID), nil
}

func (r *DocumentRepository) Mutate(
	ctx context.Context,
	workspaceID, projectID, pageID string,
	baseRevision int64,
	actorID string,
	fn application.MutateFunc,
) (*domain.PageDocument, error) {
	var result *domain.PageDocument

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := r.pageExists(tx, workspaceID, projectID, pageID); err != nil {
			return err
		}

		// Ensure the row exists (revision 1 = empty document), then lock it.
		if err := tx.Exec(
			`INSERT INTO page_documents (page_id, schema_version, revision, document)
			 VALUES (?, 1, 1, ?::jsonb) ON CONFLICT (page_id) DO NOTHING`,
			pageID, domain.EmptyDocument).Error; err != nil {
			return err
		}

		var cur documentRow
		q := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Raw(
			`SELECT page_id::text AS page_id, revision, document::text AS document,
			        updated_by::text AS updated_by, updated_at
			   FROM page_documents WHERE page_id = ?`, pageID).Scan(&cur)
		if q.Error != nil {
			return q.Error
		}
		if q.RowsAffected == 0 {
			return domain.ErrNotFound
		}

		if cur.Revision != baseRevision {
			return &domain.ConflictError{CurrentRevision: cur.Revision}
		}

		newContent, patch, err := fn(toPageDocument(cur, workspaceID, projectID))
		if err != nil {
			return err
		}
		if len(newContent) > domain.MaxDocumentBytes {
			return domain.ErrDocumentTooLarge
		}

		next := cur.Revision + 1
		var updated documentRow
		if err := tx.Raw(
			`UPDATE page_documents
			    SET document = ?::jsonb, revision = ?, updated_by = ?::uuid, updated_at = now()
			  WHERE page_id = ?
			RETURNING page_id::text AS page_id, revision, document::text AS document,
			          updated_by::text AS updated_by, updated_at`,
			string(newContent), next, actorID, pageID).Scan(&updated).Error; err != nil {
			return err
		}

		if err := tx.Exec(
			`INSERT INTO document_operations (page_id, base_revision, revision, patch, created_by)
			 VALUES (?, ?, ?, ?::jsonb, ?::uuid)`,
			pageID, cur.Revision, next, string(patch), actorID).Error; err != nil {
			return err
		}

		if next%domain.SnapshotEvery == 0 {
			if err := tx.Exec(
				`INSERT INTO document_versions (page_id, revision, document, created_by, is_checkpoint)
				 VALUES (?, ?, ?::jsonb, ?::uuid, true) ON CONFLICT (page_id, revision) DO NOTHING`,
				pageID, next, string(newContent), actorID).Error; err != nil {
				return err
			}
		}

		if err := tx.Exec(`UPDATE pages SET document_version = ? WHERE id = ?`, next, pageID).Error; err != nil {
			return err
		}

		result = toPageDocument(updated, workspaceID, projectID)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *DocumentRepository) ListOperations(
	ctx context.Context,
	workspaceID, projectID, pageID string,
	afterRevision int64,
	limit int,
) ([]domain.Operation, error) {
	db := r.db.WithContext(ctx)
	if err := r.pageExists(db, workspaceID, projectID, pageID); err != nil {
		return nil, err
	}

	var rows []operationRow
	if err := db.Raw(
		`SELECT id::text AS id, page_id::text AS page_id, base_revision, revision,
		        patch::text AS patch, created_by::text AS created_by, created_at
		   FROM document_operations
		  WHERE page_id = ? AND revision > ?
		  ORDER BY revision ASC LIMIT ?`, pageID, afterRevision, limit).Scan(&rows).Error; err != nil {
		return nil, err
	}

	ops := make([]domain.Operation, len(rows))
	for i, row := range rows {
		ops[i] = domain.Operation{
			ID: row.ID, PageID: row.PageID, BaseRevision: row.BaseRevision, Revision: row.Revision,
			Patch: json.RawMessage(row.Patch), CreatedBy: row.CreatedBy, CreatedAt: row.CreatedAt,
		}
	}
	return ops, nil
}
