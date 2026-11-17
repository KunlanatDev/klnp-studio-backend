package application_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/klnp-studio/backend/internal/modules/page/application"
	"github.com/klnp-studio/backend/internal/modules/page/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPageRepository struct {
	mock.Mock
}

func (m *MockPageRepository) Create(ctx context.Context, input domain.CreatePageInput) (*domain.Page, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Page), args.Error(1)
}

func (m *MockPageRepository) GetByID(ctx context.Context, workspaceID, projectID, pageID string) (*domain.Page, error) {
	args := m.Called(ctx, workspaceID, projectID, pageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Page), args.Error(1)
}

func (m *MockPageRepository) List(ctx context.Context, workspaceID, projectID string) ([]domain.Page, error) {
	args := m.Called(ctx, workspaceID, projectID)
	return args.Get(0).([]domain.Page), args.Error(1)
}

func (m *MockPageRepository) Update(ctx context.Context, input domain.UpdatePageInput) (*domain.Page, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Page), args.Error(1)
}

func (m *MockPageRepository) Delete(ctx context.Context, workspaceID, projectID, pageID string) error {
	return m.Called(ctx, workspaceID, projectID, pageID).Error(0)
}

func (m *MockPageRepository) Duplicate(ctx context.Context, workspaceID, projectID, pageID, actorID string) (*domain.Page, error) {
	args := m.Called(ctx, workspaceID, projectID, pageID, actorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Page), args.Error(1)
}

func (m *MockPageRepository) SetHome(ctx context.Context, workspaceID, projectID, pageID string) error {
	return m.Called(ctx, workspaceID, projectID, pageID).Error(0)
}

func (m *MockPageRepository) Reorder(ctx context.Context, input domain.ReorderPagesInput) error {
	return m.Called(ctx, input).Error(0)
}

func (m *MockPageRepository) Archive(ctx context.Context, workspaceID, projectID, pageID string) error {
	return m.Called(ctx, workspaceID, projectID, pageID).Error(0)
}

func (m *MockPageRepository) Restore(ctx context.Context, workspaceID, projectID, pageID string) error {
	return m.Called(ctx, workspaceID, projectID, pageID).Error(0)
}

func TestPageService_Lifecycle(t *testing.T) {
	ctx := context.Background()
	repo := new(MockPageRepository)
	svc := application.NewService(repo)

	t.Run("PAG-001: Create First Page", func(t *testing.T) {
		input := domain.CreatePageInput{
			WorkspaceID: "ws-1",
			ProjectID:   "p-1",
			Name:        "Home",
			CreatedBy:   "user-1",
		}
		expectedInput := domain.CreatePageInput{
			WorkspaceID: "ws-1",
			ProjectID:   "p-1",
			Name:        "Home",
			Slug:        "home",
			CreatedBy:   "user-1",
		}
		expected := &domain.Page{
			ID:          "page-1",
			ProjectID:   "p-1",
			WorkspaceID: "ws-1",
			Name:        "Home",
			Slug:        "home",
			Position:    0,
			IsHome:      true,
			Status:      domain.PageActive,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		repo.On("Create", ctx, expectedInput).Return(expected, nil).Once()

		page, err := svc.Create(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, "Home", page.Name)
		assert.True(t, page.IsHome)
	})

	t.Run("PAG-002: Create Second Page", func(t *testing.T) {
		input := domain.CreatePageInput{
			WorkspaceID: "ws-1",
			ProjectID:   "p-1",
			Name:        "Profile",
			CreatedBy:   "user-1",
		}
		expectedInput := domain.CreatePageInput{
			WorkspaceID: "ws-1",
			ProjectID:   "p-1",
			Name:        "Profile",
			Slug:        "profile",
			CreatedBy:   "user-1",
		}
		expected := &domain.Page{
			ID:          "page-2",
			ProjectID:   "p-1",
			WorkspaceID: "ws-1",
			Name:        "Profile",
			Slug:        "profile",
			Position:    1,
			IsHome:      false,
			Status:      domain.PageActive,
		}

		repo.On("Create", ctx, expectedInput).Return(expected, nil).Once()

		page, err := svc.Create(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, "Profile", page.Name)
		assert.False(t, page.IsHome)
	})

	t.Run("PAG-003: Rename Page", func(t *testing.T) {
		newName := "User Profile"
		input := domain.UpdatePageInput{
			WorkspaceID: "ws-1",
			ProjectID:   "p-1",
			PageID:      "page-2",
			Name:        &newName,
		}
		expected := &domain.Page{ID: "page-2", Name: "User Profile"}
		repo.On("Update", ctx, input).Return(expected, nil).Once()

		page, err := svc.Update(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, "User Profile", page.Name)
	})

	t.Run("PAG-004: Delete Page", func(t *testing.T) {
		repo.On("Delete", ctx, "ws-1", "p-1", "page-2").Return(nil).Once()
		err := svc.Delete(ctx, "ws-1", "p-1", "page-2")
		assert.NoError(t, err)
	})

	t.Run("PAG-005: Duplicate Page", func(t *testing.T) {
		expected := &domain.Page{ID: "page-dup", Name: "Profile (Copy)"}
		repo.On("Duplicate", ctx, "ws-1", "p-1", "page-2", "user-1").Return(expected, nil).Once()

		page, err := svc.Duplicate(ctx, "ws-1", "p-1", "page-2", "user-1")
		assert.NoError(t, err)
		assert.Equal(t, "Profile (Copy)", page.Name)
	})

	t.Run("PAG-006: Reorder Pages", func(t *testing.T) {
		reorderInput := domain.ReorderPagesInput{
			WorkspaceID: "ws-1",
			ProjectID:   "p-1",
			PageIDs:     []string{"page-2", "page-1"},
		}
		repo.On("Reorder", ctx, reorderInput).Return(nil).Once()

		err := svc.Reorder(ctx, reorderInput)
		assert.NoError(t, err)
	})

	t.Run("PAG-007: Set Home Page", func(t *testing.T) {
		repo.On("SetHome", ctx, "ws-1", "p-1", "page-2").Return(nil).Once()
		err := svc.SetHome(ctx, "ws-1", "p-1", "page-2")
		assert.NoError(t, err)
	})

	t.Run("PAG-008: Delete Last Home Denied", func(t *testing.T) {
		repo.On("Delete", ctx, "ws-1", "p-1", "page-1").Return(domain.ErrCannotDeleteHome).Once()
		err := svc.Delete(ctx, "ws-1", "p-1", "page-1")
		assert.ErrorIs(t, err, domain.ErrCannotDeleteHome)
	})

	t.Run("PAG-009: Archive Page", func(t *testing.T) {
		repo.On("Archive", ctx, "ws-1", "p-1", "page-2").Return(nil).Once()
		err := svc.Archive(ctx, "ws-1", "p-1", "page-2")
		assert.NoError(t, err)
	})

	t.Run("PAG-010: Restore Page", func(t *testing.T) {
		repo.On("Restore", ctx, "ws-1", "p-1", "page-2").Return(nil).Once()
		err := svc.Restore(ctx, "ws-1", "p-1", "page-2")
		assert.NoError(t, err)
	})
}

func TestPageService_Validation(t *testing.T) {
	ctx := context.Background()
	repo := new(MockPageRepository)
	svc := application.NewService(repo)

	t.Run("Empty Page Name", func(t *testing.T) {
		_, err := svc.Create(ctx, domain.CreatePageInput{
			WorkspaceID: "ws-1",
			ProjectID:   "p-1",
			Name:        "  ",
			CreatedBy:   "user-1",
		})
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("Page Name Exceeds Limit", func(t *testing.T) {
		_, err := svc.Create(ctx, domain.CreatePageInput{
			WorkspaceID: "ws-1",
			ProjectID:   "p-1",
			Name:        strings.Repeat("p", 101),
			CreatedBy:   "user-1",
		})
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})
}
