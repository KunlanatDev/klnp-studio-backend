package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/klnp-studio/backend/internal/modules/workspace/domain"
	appErrors "github.com/klnp-studio/backend/internal/shared/errors"
	"github.com/stretchr/testify/require"
)

type mockWorkspaceRepo struct {
	workspaces map[uuid.UUID]*domain.Workspace
	members    map[string]*domain.WorkspaceMember
	slugs      map[string]bool
}

func newMockWorkspaceRepo() *mockWorkspaceRepo {
	return &mockWorkspaceRepo{
		workspaces: make(map[uuid.UUID]*domain.Workspace),
		members:    make(map[string]*domain.WorkspaceMember),
		slugs:      make(map[string]bool),
	}
}

func (m *mockWorkspaceRepo) CreateWithMember(ctx context.Context, ws *domain.Workspace, member *domain.WorkspaceMember) error {
	m.workspaces[ws.ID] = ws
	key := ws.ID.String() + ":" + member.UserID.String()
	m.members[key] = member
	m.slugs[ws.Slug] = true
	return nil
}

func (m *mockWorkspaceRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.WorkspaceWithRole, error) {
	var list []domain.WorkspaceWithRole
	for _, ws := range m.workspaces {
		key := ws.ID.String() + ":" + userID.String()
		if mem, exists := m.members[key]; exists {
			list = append(list, domain.WorkspaceWithRole{
				ID:        ws.ID,
				Name:      ws.Name,
				Slug:      ws.Slug,
				Role:      mem.Role,
				CreatedAt: ws.CreatedAt,
				UpdatedAt: ws.UpdatedAt,
			})
		}
	}
	return list, nil
}

func (m *mockWorkspaceRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Workspace, error) {
	if ws, exists := m.workspaces[id]; exists {
		return ws, nil
	}
	return nil, appErrors.ErrNotFound
}

func (m *mockWorkspaceRepo) GetMember(ctx context.Context, workspaceID, userID uuid.UUID) (*domain.WorkspaceMember, error) {
	key := workspaceID.String() + ":" + userID.String()
	if mem, exists := m.members[key]; exists {
		return mem, nil
	}
	return nil, appErrors.ErrNotFound
}

func (m *mockWorkspaceRepo) Update(ctx context.Context, ws *domain.Workspace) error {
	if _, exists := m.workspaces[ws.ID]; !exists {
		return appErrors.ErrNotFound
	}
	ws.UpdatedAt = time.Now()
	m.workspaces[ws.ID] = ws
	return nil
}

func (m *mockWorkspaceRepo) IsSlugTaken(ctx context.Context, slug string) (bool, error) {
	return m.slugs[slug], nil
}

func TestWorkspaceUseCase(t *testing.T) {
	repo := newMockWorkspaceRepo()
	uc := NewWorkspaceUseCase(repo)
	ctx := context.Background()
	user1 := uuid.New()
	user2 := uuid.New()

	t.Run("Create Workspace Success", func(t *testing.T) {
		res, err := uc.CreateWorkspace(ctx, user1, CreateWorkspaceInput{
			Name: "My Workspace",
			Slug: "my-workspace",
		})
		require.NoError(t, err)
		require.Equal(t, "My Workspace", res.Name)
		require.Equal(t, "my-workspace", res.Slug)
		require.Equal(t, domain.RoleOwner, res.Role)
	})

	t.Run("Create Workspace Invalid Slug", func(t *testing.T) {
		_, err := uc.CreateWorkspace(ctx, user1, CreateWorkspaceInput{
			Name: "Bad Slug",
			Slug: "-invalid-",
		})
		require.Error(t, err)
	})

	t.Run("Create Workspace Duplicate Slug Conflict", func(t *testing.T) {
		_, err := uc.CreateWorkspace(ctx, user1, CreateWorkspaceInput{
			Name: "Duplicate",
			Slug: "my-workspace",
		})
		require.ErrorIs(t, err, appErrors.ErrConflict)
	})

	t.Run("List Workspaces For Member", func(t *testing.T) {
		list, err := uc.ListWorkspaces(ctx, user1)
		require.NoError(t, err)
		require.Len(t, list, 1)

		list2, err := uc.ListWorkspaces(ctx, user2)
		require.NoError(t, err)
		require.Len(t, list2, 0)
	})

	t.Run("Update Workspace By Owner", func(t *testing.T) {
		list, _ := uc.ListWorkspaces(ctx, user1)
		wsID := list[0].ID

		updated, err := uc.UpdateWorkspace(ctx, user1, wsID, "Renamed Workspace")
		require.NoError(t, err)
		require.Equal(t, "Renamed Workspace", updated.Name)
	})

	t.Run("Update Workspace By Non-member Forbidden/NotFound", func(t *testing.T) {
		list, _ := uc.ListWorkspaces(ctx, user1)
		wsID := list[0].ID

		_, err := uc.UpdateWorkspace(ctx, user2, wsID, "Hacked Workspace")
		require.Error(t, err)
	})
}
