package application

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/klnp-studio/backend/internal/modules/workspace/domain"
	appErrors "github.com/klnp-studio/backend/internal/shared/errors"
)

var (
	slugRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,48}[a-z0-9]$`)
)

type WorkspaceUseCase struct {
	repo domain.WorkspaceRepository
}

func NewWorkspaceUseCase(repo domain.WorkspaceRepository) *WorkspaceUseCase {
	return &WorkspaceUseCase{repo: repo}
}

type CreateWorkspaceInput struct {
	Name string
	Slug string
}

func (uc *WorkspaceUseCase) CreateWorkspace(ctx context.Context, userID uuid.UUID, input CreateWorkspaceInput) (*domain.WorkspaceWithRole, error) {
	name := strings.TrimSpace(input.Name)
	slug := strings.ToLower(strings.TrimSpace(input.Slug))

	if name == "" || len(name) > 80 {
		return nil, errors.New("name must be between 1 and 80 characters")
	}

	if !slugRegex.MatchString(slug) {
		return nil, errors.New("slug must be 3-50 lowercase alphanumeric characters and hyphens, starting and ending with an alphanumeric character")
	}

	taken, err := uc.repo.IsSlugTaken(ctx, slug)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, appErrors.ErrConflict
	}

	ws := &domain.Workspace{
		ID:        uuid.New(),
		Name:      name,
		Slug:      slug,
		CreatedBy: userID,
	}

	member := &domain.WorkspaceMember{
		WorkspaceID: ws.ID,
		UserID:      userID,
		Role:        domain.RoleOwner,
	}

	if err := uc.repo.CreateWithMember(ctx, ws, member); err != nil {
		return nil, err
	}

	return &domain.WorkspaceWithRole{
		ID:        ws.ID,
		Name:      ws.Name,
		Slug:      ws.Slug,
		Role:      domain.RoleOwner,
		CreatedAt: ws.CreatedAt,
		UpdatedAt: ws.UpdatedAt,
	}, nil
}

func (uc *WorkspaceUseCase) ListWorkspaces(ctx context.Context, userID uuid.UUID) ([]domain.WorkspaceWithRole, error) {
	return uc.repo.ListByUserID(ctx, userID)
}

func (uc *WorkspaceUseCase) GetWorkspace(ctx context.Context, userID, workspaceID uuid.UUID) (*domain.WorkspaceWithRole, error) {
	member, err := uc.repo.GetMember(ctx, workspaceID, userID)
	if err != nil {
		if errors.Is(err, appErrors.ErrNotFound) {
			return nil, appErrors.ErrNotFound
		}
		return nil, err
	}

	ws, err := uc.repo.GetByID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	return &domain.WorkspaceWithRole{
		ID:        ws.ID,
		Name:      ws.Name,
		Slug:      ws.Slug,
		Role:      member.Role,
		CreatedAt: ws.CreatedAt,
		UpdatedAt: ws.UpdatedAt,
	}, nil
}

func (uc *WorkspaceUseCase) UpdateWorkspace(ctx context.Context, userID, workspaceID uuid.UUID, newName string) (*domain.WorkspaceWithRole, error) {
	member, err := uc.repo.GetMember(ctx, workspaceID, userID)
	if err != nil {
		return nil, err
	}

	if member.Role != domain.RoleOwner {
		return nil, appErrors.ErrForbidden
	}

	name := strings.TrimSpace(newName)
	if name == "" || len(name) > 80 {
		return nil, errors.New("name must be between 1 and 80 characters")
	}

	ws, err := uc.repo.GetByID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	ws.Name = name
	if err := uc.repo.Update(ctx, ws); err != nil {
		return nil, err
	}

	return &domain.WorkspaceWithRole{
		ID:        ws.ID,
		Name:      ws.Name,
		Slug:      ws.Slug,
		Role:      member.Role,
		CreatedAt: ws.CreatedAt,
		UpdatedAt: ws.UpdatedAt,
	}, nil
}
