package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Profile) TableName() string {
	return "profiles"
}

type ProfileRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Profile, error)
	Update(ctx context.Context, profile *Profile) error
}
