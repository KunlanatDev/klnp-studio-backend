package presentation

import "github.com/klnp-studio/backend/internal/modules/page/domain"

type CreatePageRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug,omitempty"`
}

type UpdatePageRequest struct {
	Name *string `json:"name,omitempty"`
	Slug *string `json:"slug,omitempty"`
}

type ReorderPagesRequest struct {
	PageIDs []string `json:"page_ids"`
}

type PageResponse struct {
	ID              string `json:"id"`
	ProjectID       string `json:"project_id"`
	WorkspaceID     string `json:"workspace_id"`
	Name            string `json:"name"`
	Slug            string `json:"slug"`
	Position        int    `json:"position"`
	IsHome          bool   `json:"is_home"`
	Status          string `json:"status"`
	DocumentVersion int    `json:"document_version"`
	CreatedBy       string `json:"created_by,omitempty"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
	ArchivedAt      *string `json:"archived_at,omitempty"`
}

func ToPageResponse(p domain.Page) PageResponse {
	resp := PageResponse{
		ID:              p.ID,
		ProjectID:       p.ProjectID,
		WorkspaceID:     p.WorkspaceID,
		Name:            p.Name,
		Slug:            p.Slug,
		Position:        p.Position,
		IsHome:          p.IsHome,
		Status:          string(p.Status),
		DocumentVersion: p.DocumentVersion,
		CreatedBy:       p.CreatedBy,
		CreatedAt:       p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if p.ArchivedAt != nil {
		archived := p.ArchivedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.ArchivedAt = &archived
	}
	return resp
}
