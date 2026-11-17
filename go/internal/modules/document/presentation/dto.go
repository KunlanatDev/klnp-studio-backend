package presentation

import (
	"encoding/json"
	"time"

	"github.com/klnp-studio/backend/internal/modules/document/domain"
)

type ApplyPatchRequest struct {
	BaseRevision int64           `json:"base_revision"`
	Patch        json.RawMessage `json:"patch"`
}

type DocumentResponse struct {
	PageID    string          `json:"page_id"`
	Revision  int64           `json:"revision"`
	Document  json.RawMessage `json:"document"`
	UpdatedBy *string         `json:"updated_by,omitempty"`
	UpdatedAt string          `json:"updated_at"`
}

// PatchResultResponse is intentionally small: the client already has the
// document and only needs the new revision.
type PatchResultResponse struct {
	PageID    string `json:"page_id"`
	Revision  int64  `json:"revision"`
	UpdatedAt string `json:"updated_at"`
}

type OperationResponse struct {
	ID           string          `json:"id"`
	BaseRevision int64           `json:"base_revision"`
	Revision     int64           `json:"revision"`
	Patch        json.RawMessage `json:"patch"`
	CreatedBy    *string         `json:"created_by,omitempty"`
	CreatedAt    string          `json:"created_at"`
}

func ToDocumentResponse(d domain.PageDocument) DocumentResponse {
	return DocumentResponse{
		PageID: d.PageID, Revision: d.Revision, Document: d.Content,
		UpdatedBy: d.UpdatedBy, UpdatedAt: d.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func ToPatchResultResponse(d domain.PageDocument) PatchResultResponse {
	return PatchResultResponse{PageID: d.PageID, Revision: d.Revision, UpdatedAt: d.UpdatedAt.UTC().Format(time.RFC3339)}
}

func ToOperationResponse(o domain.Operation) OperationResponse {
	return OperationResponse{
		ID: o.ID, BaseRevision: o.BaseRevision, Revision: o.Revision, Patch: o.Patch,
		CreatedBy: o.CreatedBy, CreatedAt: o.CreatedAt.UTC().Format(time.RFC3339),
	}
}
