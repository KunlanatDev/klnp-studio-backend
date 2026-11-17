package application

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/klnp-studio/backend/internal/modules/project/domain"
)

type Service struct {
	repo ProjectRepository
}

func NewService(repo ProjectRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(
	ctx context.Context,
	input domain.CreateProjectInput,
) (*domain.Project, error) {
	input.Name = strings.TrimSpace(input.Name)

	if input.WorkspaceID == "" ||
		input.CreatedBy == "" ||
		input.Name == "" ||
		utf8.RuneCountInString(input.Name) > 100 {
		return nil, domain.ErrInvalidInput
	}

	return s.repo.Create(ctx, input)
}

func (s *Service) Get(
	ctx context.Context,
	workspaceID string,
	projectID string,
) (*domain.Project, error) {
	if workspaceID == "" || projectID == "" {
		return nil, domain.ErrNotFound
	}
	return s.repo.GetByID(ctx, workspaceID, projectID)
}

func (s *Service) List(
	ctx context.Context,
	workspaceID string,
	keyword string,
	limit int,
	offset int,
) ([]domain.Project, int64, error) {
	if workspaceID == "" {
		return nil, 0, domain.ErrNotFound
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, workspaceID, keyword, limit, offset)
}

func (s *Service) Update(
	ctx context.Context,
	workspaceID string,
	input domain.UpdateProjectInput,
) (*domain.Project, error) {
	if workspaceID == "" || input.ProjectID == "" {
		return nil, domain.ErrNotFound
	}
	if input.Name != nil {
		trimmed := strings.TrimSpace(*input.Name)
		if trimmed == "" || utf8.RuneCountInString(trimmed) > 100 {
			return nil, domain.ErrInvalidInput
		}
		input.Name = &trimmed
	}
	return s.repo.Update(ctx, workspaceID, input)
}

func (s *Service) Archive(
	ctx context.Context,
	workspaceID string,
	projectID string,
) error {
	if workspaceID == "" || projectID == "" {
		return domain.ErrNotFound
	}
	return s.repo.Archive(ctx, workspaceID, projectID)
}

func (s *Service) Restore(
	ctx context.Context,
	workspaceID string,
	projectID string,
) error {
	if workspaceID == "" || projectID == "" {
		return domain.ErrNotFound
	}
	return s.repo.Restore(ctx, workspaceID, projectID)
}

func (s *Service) Delete(
	ctx context.Context,
	workspaceID string,
	projectID string,
) error {
	if workspaceID == "" || projectID == "" {
		return domain.ErrNotFound
	}
	return s.repo.Delete(ctx, workspaceID, projectID)
}

