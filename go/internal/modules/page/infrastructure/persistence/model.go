package persistence

import (
	"time"

	"github.com/klnp-studio/backend/internal/modules/page/domain"
)

type PageModel struct {
	ID              string     `gorm:"column:id;primaryKey"`
	ProjectID       string     `gorm:"column:project_id"`
	WorkspaceID     string     `gorm:"column:workspace_id"`
	Name            string     `gorm:"column:name"`
	Route           string     `gorm:"column:route"`
	Slug            string     `gorm:"column:slug"`
	Position        int        `gorm:"column:position"`
	IsHome          bool       `gorm:"column:is_home"`
	Status          string     `gorm:"column:status"`
	DocumentVersion int        `gorm:"column:document_version"`
	CreatedBy       *string    `gorm:"column:created_by"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
	ArchivedAt      *time.Time `gorm:"column:archived_at"`
}

func (PageModel) TableName() string {
	return "pages"
}

func (m PageModel) ToDomain() domain.Page {
	p := domain.Page{
		ID:              m.ID,
		ProjectID:       m.ProjectID,
		WorkspaceID:     m.WorkspaceID,
		Name:            m.Name,
		Slug:            m.Slug,
		Position:        m.Position,
		IsHome:          m.IsHome,
		Status:          domain.PageStatus(m.Status),
		DocumentVersion: m.DocumentVersion,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
		ArchivedAt:      m.ArchivedAt,
	}
	if m.CreatedBy != nil {
		p.CreatedBy = *m.CreatedBy
	}
	return p
}
