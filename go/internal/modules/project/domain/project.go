package domain

import "time"

type ProjectStatus string

const (
	ProjectActive   ProjectStatus = "active"
	ProjectArchived ProjectStatus = "archived"
)

type Project struct {
	ID          string        `json:"id"`
	WorkspaceID string        `json:"workspace_id"`
	Name        string        `json:"name"`
	Slug        string        `json:"slug"`
	Description *string       `json:"description,omitempty"`
	Icon        *string       `json:"icon,omitempty"`
	Platform    string        `json:"platform"`
	Status      ProjectStatus `json:"status"`
	PageCount   int64         `json:"page_count"`
	HomePageID  string        `json:"home_page_id,omitempty"`
	CreatedBy   string        `json:"created_by"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	ArchivedAt  *time.Time    `json:"archived_at,omitempty"`
}

type CreateProjectInput struct {
	WorkspaceID string  `json:"workspace_id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	CreatedBy   string  `json:"created_by"`
}

type UpdateProjectInput struct {
	ProjectID   string  `json:"project_id"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty"`
}

