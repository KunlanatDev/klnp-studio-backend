package presentation

import "github.com/klnp-studio/backend/internal/modules/project/domain"

type CreateProjectRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty"`
}

type UpdateProjectRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty"`
}

type ProjectResponse struct {
	ID          string               `json:"id"`
	WorkspaceID string               `json:"workspace_id"`
	Name        string               `json:"name"`
	Slug        string               `json:"slug"`
	Description *string              `json:"description,omitempty"`
	Icon        *string              `json:"icon,omitempty"`
	Platform    string               `json:"platform"`
	Status      domain.ProjectStatus `json:"status"`
	PageCount   int64                `json:"page_count"`
	HomePageID  string               `json:"home_page_id,omitempty"`
	CreatedBy   string               `json:"created_by"`
	CreatedAt   string               `json:"created_at"`
	UpdatedAt   string               `json:"updated_at"`
	ArchivedAt  *string              `json:"archived_at,omitempty"`
}

type PaginationMeta struct {
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

func ToProjectResponse(p domain.Project) ProjectResponse {
	resp := ProjectResponse{
		ID:          p.ID,
		WorkspaceID: p.WorkspaceID,
		Name:        p.Name,
		Slug:        p.Slug,
		Description: p.Description,
		Icon:        p.Icon,
		Platform:    p.Platform,
		Status:      p.Status,
		PageCount:   p.PageCount,
		HomePageID:  p.HomePageID,
		CreatedBy:   p.CreatedBy,
		CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if p.ArchivedAt != nil {
		archived := p.ArchivedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.ArchivedAt = &archived
	}
	return resp
}
