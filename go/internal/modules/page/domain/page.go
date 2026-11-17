package domain

import "time"

type PageStatus string

const (
	PageActive   PageStatus = "active"
	PageArchived PageStatus = "archived"
)

type Page struct {
	ID              string     `json:"id"`
	ProjectID       string     `json:"project_id"`
	WorkspaceID     string     `json:"workspace_id"`
	Name            string     `json:"name"`
	Slug            string     `json:"slug"`
	Position        int        `json:"position"`
	IsHome          bool       `json:"is_home"`
	Status          PageStatus `json:"status"`
	DocumentVersion int        `json:"document_version"`
	CreatedBy       string     `json:"created_by"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	ArchivedAt      *time.Time `json:"archived_at,omitempty"`
}

type CreatePageInput struct {
	WorkspaceID string `json:"workspace_id"`
	ProjectID   string `json:"project_id"`
	Name        string `json:"name"`
	Slug        string `json:"slug,omitempty"`
	CreatedBy   string `json:"created_by"`
}

type UpdatePageInput struct {
	WorkspaceID string  `json:"workspace_id"`
	ProjectID   string  `json:"project_id"`
	PageID      string  `json:"page_id"`
	Name        *string `json:"name,omitempty"`
	Slug        *string `json:"slug,omitempty"`
}

type ReorderPagesInput struct {
	WorkspaceID string   `json:"workspace_id"`
	ProjectID   string   `json:"project_id"`
	PageIDs     []string `json:"page_ids"`
}
