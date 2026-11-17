package persistence

import "time"

type ProjectModel struct {
	ID          string     `gorm:"column:id;primaryKey"`
	WorkspaceID string     `gorm:"column:workspace_id"`
	Name        string     `gorm:"column:name"`
	Slug        string     `gorm:"column:slug"`
	Description *string    `gorm:"column:description"`
	Icon        *string    `gorm:"column:icon"`
	Platform    string     `gorm:"column:platform"`
	Status      string     `gorm:"column:status"`
	CreatedBy   string     `gorm:"column:created_by"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
	ArchivedAt  *time.Time `gorm:"column:archived_at"`
}

func (ProjectModel) TableName() string {
	return "projects"
}

