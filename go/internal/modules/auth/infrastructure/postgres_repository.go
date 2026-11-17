package infrastructure

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/klnp-studio/backend/internal/modules/auth/domain"
	appErrors "github.com/klnp-studio/backend/internal/shared/errors"
	"gorm.io/gorm"
)

type PostgresProfileRepository struct {
	db *gorm.DB
}

func NewPostgresProfileRepository(db *gorm.DB) *PostgresProfileRepository {
	return &PostgresProfileRepository{db: db}
}

func (r *PostgresProfileRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Profile, error) {
	var profile domain.Profile
	err := r.db.WithContext(ctx).First(&profile, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrNotFound
		}
		return nil, err
	}
	return &profile, nil
}

func (r *PostgresProfileRepository) Update(ctx context.Context, profile *domain.Profile) error {
	result := r.db.WithContext(ctx).Model(profile).Updates(map[string]any{
		"display_name": profile.DisplayName,
		"avatar_url":   profile.AvatarURL,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return appErrors.ErrNotFound
	}
	return nil
}
