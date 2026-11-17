package persistence

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/klnp-studio/backend/internal/modules/project/domain"
	"gorm.io/gorm"
)

var slugSanitizer = regexp.MustCompile(`[^a-z0-9]+`)

func toSlug(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = slugSanitizer.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "project"
	}
	if len(s) > 40 {
		s = s[:40]
	}
	return s
}

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(
	ctx context.Context,
	input domain.CreateProjectInput,
) (*domain.Project, error) {
	projID := uuid.New().String()
	baseSlug := toSlug(input.Name)
	slug := fmt.Sprintf("%s-%s", baseSlug, projID[:8])

	proj := &ProjectModel{
		ID:          projID,
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		Slug:        slug,
		Description: input.Description,
		Icon:        input.Icon,
		Platform:    "mobile",
		Status:      "active",
		CreatedBy:   input.CreatedBy,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	homePageID := uuid.New().String()

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Create project
		if err := tx.Create(proj).Error; err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return domain.ErrSlugConflict
			}
			return err
		}

		// 2. Create initial Home Page (Quota trigger will verify free page entitlement)
		homePage := map[string]any{
			"id":           homePageID,
			"workspace_id": input.WorkspaceID,
			"project_id":   projID,
			"name":         "Home",
			"slug":         "home",
			"position":     0,
			"is_home":      true,
			"status":       "active",
			"created_by":   input.CreatedBy,
			"created_at":   time.Now().UTC(),
			"updated_at":   time.Now().UTC(),
		}

		if err := tx.Table("pages").Create(homePage).Error; err != nil {
			if strings.Contains(err.Error(), "WORKSPACE_PAGE_QUOTA_EXCEEDED") {
				return domain.ErrQuotaExceeded
			}
			return fmt.Errorf("create initial home page: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &domain.Project{
		ID:          proj.ID,
		WorkspaceID: proj.WorkspaceID,
		Name:        proj.Name,
		Slug:        proj.Slug,
		Description: proj.Description,
		Icon:        proj.Icon,
		Platform:    proj.Platform,
		Status:      domain.ProjectStatus(proj.Status),
		PageCount:   1,
		HomePageID:  homePageID,
		CreatedBy:   proj.CreatedBy,
		CreatedAt:   proj.CreatedAt,
		UpdatedAt:   proj.UpdatedAt,
	}, nil
}

func (r *ProjectRepository) GetByID(
	ctx context.Context,
	workspaceID string,
	projectID string,
) (*domain.Project, error) {
	var model ProjectModel
	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND id = ?", workspaceID, projectID).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	var pageCount int64
	r.db.WithContext(ctx).Table("pages").
		Where("project_id = ? AND workspace_id = ? AND status = 'active'", projectID, workspaceID).
		Count(&pageCount)

	var homePage struct {
		ID string
	}
	r.db.WithContext(ctx).Table("pages").
		Select("id").
		Where("project_id = ? AND workspace_id = ? AND is_home = TRUE AND status = 'active'", projectID, workspaceID).
		Take(&homePage)

	return &domain.Project{
		ID:          model.ID,
		WorkspaceID: model.WorkspaceID,
		Name:        model.Name,
		Slug:        model.Slug,
		Description: model.Description,
		Icon:        model.Icon,
		Platform:    model.Platform,
		Status:      domain.ProjectStatus(model.Status),
		PageCount:   pageCount,
		HomePageID:  homePage.ID,
		CreatedBy:   model.CreatedBy,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		ArchivedAt:  model.ArchivedAt,
	}, nil
}

func (r *ProjectRepository) List(
	ctx context.Context,
	workspaceID string,
	keyword string,
	limit int,
	offset int,
) ([]domain.Project, int64, error) {
	query := r.db.WithContext(ctx).Model(&ProjectModel{}).Where("workspace_id = ?", workspaceID)

	if keyword != "" {
		query = query.Where("name ILIKE ?", "%"+keyword+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var models []ProjectModel
	err := query.Order("updated_at DESC").Limit(limit).Offset(offset).Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	results := make([]domain.Project, len(models))
	for i, m := range models {
		var pageCount int64
		r.db.WithContext(ctx).Table("pages").
			Where("project_id = ? AND workspace_id = ? AND status = 'active'", m.ID, workspaceID).
			Count(&pageCount)

		results[i] = domain.Project{
			ID:          m.ID,
			WorkspaceID: m.WorkspaceID,
			Name:        m.Name,
			Slug:        m.Slug,
			Description: m.Description,
			Icon:        m.Icon,
			Platform:    m.Platform,
			Status:      domain.ProjectStatus(m.Status),
			PageCount:   pageCount,
			CreatedBy:   m.CreatedBy,
			CreatedAt:   m.CreatedAt,
			UpdatedAt:   m.UpdatedAt,
			ArchivedAt:  m.ArchivedAt,
		}
	}

	return results, total, nil
}

func (r *ProjectRepository) Update(
	ctx context.Context,
	workspaceID string,
	input domain.UpdateProjectInput,
) (*domain.Project, error) {
	updates := map[string]any{
		"updated_at": time.Now().UTC(),
	}

	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.Icon != nil {
		updates["icon"] = *input.Icon
	}

	res := r.db.WithContext(ctx).Model(&ProjectModel{}).
		Where("workspace_id = ? AND id = ?", workspaceID, input.ProjectID).
		Updates(updates)

	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, domain.ErrNotFound
	}

	return r.GetByID(ctx, workspaceID, input.ProjectID)
}

func (r *ProjectRepository) Archive(
	ctx context.Context,
	workspaceID string,
	projectID string,
) error {
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).Model(&ProjectModel{}).
		Where("workspace_id = ? AND id = ?", workspaceID, projectID).
		Updates(map[string]any{
			"status":      "archived",
			"archived_at": now,
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

func (r *ProjectRepository) Restore(
	ctx context.Context,
	workspaceID string,
	projectID string,
) error {
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).Model(&ProjectModel{}).
		Where("workspace_id = ? AND id = ?", workspaceID, projectID).
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

func (r *ProjectRepository) Delete(
	ctx context.Context,
	workspaceID string,
	projectID string,
) error {
	res := r.db.WithContext(ctx).
		Where("workspace_id = ? AND id = ?", workspaceID, projectID).
		Delete(&ProjectModel{})

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

