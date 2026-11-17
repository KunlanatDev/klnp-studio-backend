package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type WorkspaceRole string

const (
	RoleOwner  WorkspaceRole = "owner"
	RoleAdmin  WorkspaceRole = "admin"
	RoleEditor WorkspaceRole = "editor"
	RoleViewer WorkspaceRole = "viewer"
)

type Workspace struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Slug      string    `gorm:"not null;uniqueIndex" json:"slug"`
	CreatedBy uuid.UUID `gorm:"type:uuid;not null" json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Workspace) TableName() string {
	return "workspaces"
}

type WorkspaceMember struct {
	WorkspaceID uuid.UUID     `gorm:"type:uuid;primaryKey" json:"workspace_id"`
	UserID      uuid.UUID     `gorm:"type:uuid;primaryKey" json:"user_id"`
	Role        WorkspaceRole `gorm:"not null" json:"role"`
	CreatedAt   time.Time     `json:"created_at"`
}

func (WorkspaceMember) TableName() string {
	return "workspace_members"
}

type WorkspaceWithRole struct {
	ID        uuid.UUID     `json:"id"`
	Name      string        `json:"name"`
	Slug      string        `json:"slug"`
	Role      WorkspaceRole `json:"role"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type WorkspaceRepository interface {
	CreateWithMember(ctx context.Context, ws *Workspace, member *WorkspaceMember) error
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]WorkspaceWithRole, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Workspace, error)
	GetMember(ctx context.Context, workspaceID, userID uuid.UUID) (*WorkspaceMember, error)
	Update(ctx context.Context, ws *Workspace) error
	IsSlugTaken(ctx context.Context, slug string) (bool, error)
}
