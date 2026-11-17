package persistence

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/klnp-studio/backend/internal/modules/workspace/domain"
	appErrors "github.com/klnp-studio/backend/internal/shared/errors"
	"gorm.io/gorm"
)

type PostgresWorkspaceRepository struct {
	db *gorm.DB
}

func NewPostgresWorkspaceRepository(db *gorm.DB) *PostgresWorkspaceRepository {
	return &PostgresWorkspaceRepository{db: db}
}

func (r *PostgresWorkspaceRepository) CreateWithMember(ctx context.Context, ws *domain.Workspace, member *domain.WorkspaceMember) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(ws).Error; err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return appErrors.ErrConflict
			}
			return err
		}

		if err := tx.Create(member).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *PostgresWorkspaceRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.WorkspaceWithRole, error) {
	results := make([]domain.WorkspaceWithRole, 0)

	err := r.db.WithContext(ctx).
		Table("workspaces w").
		Select("w.id, w.name, w.slug, wm.role, w.created_at, w.updated_at").
		Joins("JOIN workspace_members wm ON wm.workspace_id = w.id").
		Where("wm.user_id = ?", userID).
		Order("w.created_at DESC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (r *PostgresWorkspaceRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Workspace, error) {
	var ws domain.Workspace
	err := r.db.WithContext(ctx).First(&ws, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrNotFound
		}
		return nil, err
	}
	return &ws, nil
}

func (r *PostgresWorkspaceRepository) GetMember(ctx context.Context, workspaceID, userID uuid.UUID) (*domain.WorkspaceMember, error) {
	var member domain.WorkspaceMember
	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrNotFound
		}
		return nil, err
	}
	return &member, nil
}

func (r *PostgresWorkspaceRepository) Update(ctx context.Context, ws *domain.Workspace) error {
	result := r.db.WithContext(ctx).Model(ws).Updates(map[string]any{
		"name": ws.Name,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return appErrors.ErrNotFound
	}
	return nil
}

func (r *PostgresWorkspaceRepository) IsSlugTaken(ctx context.Context, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Workspace{}).
		Where("slug = ?", slug).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
