package persistence

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/klnp-studio/backend/internal/modules/page/domain"
	"gorm.io/gorm"
)

type PageRepository struct {
	db *gorm.DB
}

func NewPageRepository(db *gorm.DB) *PageRepository {
	return &PageRepository{db: db}
}

func (r *PageRepository) Create(
	ctx context.Context,
	input domain.CreatePageInput,
) (*domain.Page, error) {
	pageID := uuid.New().String()
	route := "/" + input.Slug
	if input.Slug == "home" {
		route = "/"
	}

	model := &PageModel{
		ID:              pageID,
		ProjectID:       input.ProjectID,
		WorkspaceID:     input.WorkspaceID,
		Name:            input.Name,
		Route:           route,
		Slug:            input.Slug,
		Position:        0,
		IsHome:          false,
		Status:          "active",
		DocumentVersion: 1,
		CreatedBy:       &input.CreatedBy,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Verify project exists in workspace
		var projCount int64
		if err := tx.Table("projects").
			Where("id = ? AND workspace_id = ? AND status = 'active'", input.ProjectID, input.WorkspaceID).
			Count(&projCount).Error; err != nil {
			return err
		}
		if projCount == 0 {
			return domain.ErrNotFound
		}

		if err := tx.Create(model).Error; err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.Code == "23505" {
					return domain.ErrPathConflict
				}
			}
			if strings.Contains(err.Error(), "WORKSPACE_PAGE_QUOTA_EXCEEDED") {
				return domain.ErrQuotaExceeded
			}
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Refetch to get initialized position and is_home from trigger
	return r.GetByID(ctx, input.WorkspaceID, input.ProjectID, pageID)
}

func (r *PageRepository) GetByID(
	ctx context.Context,
	workspaceID string,
	projectID string,
	pageID string,
) (*domain.Page, error) {
	var model PageModel
	err := r.db.WithContext(ctx).
		Where("id = ? AND project_id = ? AND workspace_id = ? AND status = 'active'", pageID, projectID, workspaceID).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	res := model.ToDomain()
	return &res, nil
}

func (r *PageRepository) List(
	ctx context.Context,
	workspaceID string,
	projectID string,
) ([]domain.Page, error) {
	var models []PageModel
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND workspace_id = ? AND status = 'active'", projectID, workspaceID).
		Order("position ASC, id ASC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	res := make([]domain.Page, len(models))
	for i, m := range models {
		res[i] = m.ToDomain()
	}
	return res, nil
}

func (r *PageRepository) Update(
	ctx context.Context,
	input domain.UpdatePageInput,
) (*domain.Page, error) {
	updates := map[string]any{
		"updated_at": time.Now().UTC(),
	}
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Slug != nil {
		updates["slug"] = *input.Slug
		updates["route"] = "/" + *input.Slug
	}

	res := r.db.WithContext(ctx).
		Table("pages").
		Where("id = ? AND project_id = ? AND workspace_id = ? AND status = 'active'", input.PageID, input.ProjectID, input.WorkspaceID).
		Updates(updates)

	if res.Error != nil {
		var pgErr *pgconn.PgError
		if errors.As(res.Error, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrPathConflict
		}
		return nil, res.Error
	}

	if res.RowsAffected == 0 {
		return nil, domain.ErrNotFound
	}

	return r.GetByID(ctx, input.WorkspaceID, input.ProjectID, input.PageID)
}

func (r *PageRepository) Delete(
	ctx context.Context,
	workspaceID string,
	projectID string,
	pageID string,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var page PageModel
		if err := tx.Where("id = ? AND project_id = ? AND workspace_id = ? AND status = 'active'", pageID, projectID, workspaceID).
			First(&page).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrNotFound
			}
			return err
		}

		if page.IsHome {
			return domain.ErrCannotDeleteHome
		}

		res := tx.Where("id = ? AND project_id = ? AND workspace_id = ?", pageID, projectID, workspaceID).
			Delete(&PageModel{})

		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return domain.ErrNotFound
		}

		return nil
	})
}

func (r *PageRepository) Duplicate(
	ctx context.Context,
	workspaceID string,
	projectID string,
	pageID string,
	actorID string,
) (*domain.Page, error) {
	var newPage *domain.Page

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var src PageModel
		if err := tx.Where("id = ? AND project_id = ? AND workspace_id = ? AND status = 'active'", pageID, projectID, workspaceID).
			First(&src).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrNotFound
			}
			return err
		}

		newID := uuid.New().String()
		newSlug := fmt.Sprintf("%s-copy-%s", src.Slug, newID[:4])
		newName := fmt.Sprintf("%s (Copy)", src.Name)

		dupModel := &PageModel{
			ID:              newID,
			ProjectID:       projectID,
			WorkspaceID:     workspaceID,
			Name:            newName,
			Route:           "/" + newSlug,
			Slug:            newSlug,
			Position:        0,
			IsHome:          false,
			Status:          "active",
			DocumentVersion: 1,
			CreatedBy:       &actorID,
			CreatedAt:       time.Now().UTC(),
			UpdatedAt:       time.Now().UTC(),
		}

		if err := tx.Create(dupModel).Error; err != nil {
			if strings.Contains(err.Error(), "WORKSPACE_PAGE_QUOTA_EXCEEDED") {
				return domain.ErrQuotaExceeded
			}
			return err
		}

		var created PageModel
		if err := tx.Where("id = ?", newID).First(&created).Error; err != nil {
			return err
		}
		p := created.ToDomain()
		newPage = &p
		return nil
	})

	if err != nil {
		return nil, err
	}
	return newPage, nil
}

func (r *PageRepository) SetHome(
	ctx context.Context,
	workspaceID string,
	projectID string,
	pageID string,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Serialize by locking project row
		var projCount int64
		if err := tx.Table("projects").
			Where("id = ? AND workspace_id = ? AND status = 'active'", projectID, workspaceID).
			Clauses().
			Count(&projCount).Error; err != nil {
			return err
		}
		if projCount == 0 {
			return domain.ErrNotFound
		}

		// Clear previous home
		if err := tx.Table("pages").
			Where("project_id = ? AND workspace_id = ? AND is_home = true", projectID, workspaceID).
			Update("is_home", false).Error; err != nil {
			return err
		}

		// Set new home
		res := tx.Table("pages").
			Where("id = ? AND project_id = ? AND workspace_id = ? AND status = 'active'", pageID, projectID, workspaceID).
			Update("is_home", true)

		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return domain.ErrNotFound
		}

		return nil
	})
}

func (r *PageRepository) Reorder(
	ctx context.Context,
	input domain.ReorderPagesInput,
) error {
	idsJoined := strings.Join(input.PageIDs, ",")
	sql := `SELECT public.s2_02_07_reorder_pages(?::uuid, ?::uuid, string_to_array(?, ',')::uuid[])`

	err := r.db.WithContext(ctx).Exec(sql, input.WorkspaceID, input.ProjectID, idsJoined).Error
	if err != nil {
		if strings.Contains(err.Error(), "Project not found") {
			return domain.ErrNotFound
		}
		if strings.Contains(err.Error(), "does not match active pages count") ||
			strings.Contains(err.Error(), "invalid or not active") {
			return domain.ErrInvalidInput
		}
		return err
	}
	return nil
}

func (r *PageRepository) Archive(
	ctx context.Context,
	workspaceID string,
	projectID string,
	pageID string,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var page PageModel
		if err := tx.Where("id = ? AND project_id = ? AND workspace_id = ? AND status = 'active'", pageID, projectID, workspaceID).
			First(&page).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrNotFound
			}
			return err
		}

		if page.IsHome {
			return domain.ErrCannotDeleteHome
		}

		now := time.Now().UTC()
		res := tx.Table("pages").
			Where("id = ? AND project_id = ? AND workspace_id = ?", pageID, projectID, workspaceID).
			Updates(map[string]any{
				"status":      "archived",
				"is_home":     false,
				"archived_at": &now,
				"updated_at":  now,
			})

		if res.Error != nil {
			return res.Error
		}
		return nil
	})
}

func (r *PageRepository) Restore(
	ctx context.Context,
	workspaceID string,
	projectID string,
	pageID string,
) error {
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).
		Table("pages").
		Where("id = ? AND project_id = ? AND workspace_id = ? AND status = 'archived'", pageID, projectID, workspaceID).
		Updates(map[string]any{
			"status":      "active",
			"archived_at": nil,
			"updated_at":  now,
		})

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
