package application

import (
	"context"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/klnp-studio/backend/internal/modules/page/domain"
)

var slugSanitizer = regexp.MustCompile(`[^a-z0-9]+`)

func toSlug(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = slugSanitizer.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "page"
	}
	if len(s) > 40 {
		s = s[:40]
	}
	return s
}

type Service struct {
	repo PageRepository
}

func NewService(repo PageRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(
	ctx context.Context,
	input domain.CreatePageInput,
) (*domain.Page, error) {
	input.Name = strings.TrimSpace(input.Name)

	if input.WorkspaceID == "" ||
		input.ProjectID == "" ||
		input.CreatedBy == "" ||
		input.Name == "" ||
		utf8.RuneCountInString(input.Name) > 100 {
		return nil, domain.ErrInvalidInput
	}

	if input.Slug == "" {
		input.Slug = toSlug(input.Name)
	} else {
		input.Slug = toSlug(input.Slug)
	}

	return s.repo.Create(ctx, input)
}

func (s *Service) Get(
	ctx context.Context,
	workspaceID string,
	projectID string,
	pageID string,
) (*domain.Page, error) {
	if workspaceID == "" || projectID == "" || pageID == "" {
		return nil, domain.ErrNotFound
	}
	return s.repo.GetByID(ctx, workspaceID, projectID, pageID)
}

func (s *Service) List(
	ctx context.Context,
	workspaceID string,
	projectID string,
) ([]domain.Page, error) {
	if workspaceID == "" || projectID == "" {
		return nil, domain.ErrNotFound
	}
	return s.repo.List(ctx, workspaceID, projectID)
}

func (s *Service) Update(
	ctx context.Context,
	input domain.UpdatePageInput,
) (*domain.Page, error) {
	if input.WorkspaceID == "" || input.ProjectID == "" || input.PageID == "" {
		return nil, domain.ErrNotFound
	}

	if input.Name != nil {
		trimmed := strings.TrimSpace(*input.Name)
		if trimmed == "" || utf8.RuneCountInString(trimmed) > 100 {
			return nil, domain.ErrInvalidInput
		}
		input.Name = &trimmed
	}

	if input.Slug != nil {
		slug := toSlug(*input.Slug)
		input.Slug = &slug
	}

	return s.repo.Update(ctx, input)
}

func (s *Service) Delete(
	ctx context.Context,
	workspaceID string,
	projectID string,
	pageID string,
) error {
	if workspaceID == "" || projectID == "" || pageID == "" {
		return domain.ErrNotFound
	}
	return s.repo.Delete(ctx, workspaceID, projectID, pageID)
}

func (s *Service) Duplicate(
	ctx context.Context,
	workspaceID string,
	projectID string,
	pageID string,
	actorID string,
) (*domain.Page, error) {
	if workspaceID == "" || projectID == "" || pageID == "" || actorID == "" {
		return nil, domain.ErrInvalidInput
	}
	return s.repo.Duplicate(ctx, workspaceID, projectID, pageID, actorID)
}

func (s *Service) SetHome(
	ctx context.Context,
	workspaceID string,
	projectID string,
	pageID string,
) error {
	if workspaceID == "" || projectID == "" || pageID == "" {
		return domain.ErrNotFound
	}
	return s.repo.SetHome(ctx, workspaceID, projectID, pageID)
}

func (s *Service) Reorder(
	ctx context.Context,
	input domain.ReorderPagesInput,
) error {
	if input.WorkspaceID == "" || input.ProjectID == "" || len(input.PageIDs) == 0 {
		return domain.ErrInvalidInput
	}
	return s.repo.Reorder(ctx, input)
}

func (s *Service) Archive(
	ctx context.Context,
	workspaceID string,
	projectID string,
	pageID string,
) error {
	if workspaceID == "" || projectID == "" || pageID == "" {
		return domain.ErrNotFound
	}
	return s.repo.Archive(ctx, workspaceID, projectID, pageID)
}

func (s *Service) Restore(
	ctx context.Context,
	workspaceID string,
	projectID string,
	pageID string,
) error {
	if workspaceID == "" || projectID == "" || pageID == "" {
		return domain.ErrNotFound
	}
	return s.repo.Restore(ctx, workspaceID, projectID, pageID)
}
