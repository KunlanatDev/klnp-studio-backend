package application_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/klnp-studio/backend/internal/modules/project/application"
	"github.com/klnp-studio/backend/internal/modules/project/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockProjectRepository struct {
	mock.Mock
}

func (m *MockProjectRepository) Create(ctx context.Context, input domain.CreateProjectInput) (*domain.Project, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Project), args.Error(1)
}

func (m *MockProjectRepository) GetByID(ctx context.Context, workspaceID, projectID string) (*domain.Project, error) {
	args := m.Called(ctx, workspaceID, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Project), args.Error(1)
}

func (m *MockProjectRepository) List(ctx context.Context, workspaceID, keyword string, limit, offset int) ([]domain.Project, int64, error) {
	args := m.Called(ctx, workspaceID, keyword, limit, offset)
	return args.Get(0).([]domain.Project), args.Get(1).(int64), args.Error(2)
}

func (m *MockProjectRepository) Update(ctx context.Context, workspaceID string, input domain.UpdateProjectInput) (*domain.Project, error) {
	args := m.Called(ctx, workspaceID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Project), args.Error(1)
}

func (m *MockProjectRepository) Archive(ctx context.Context, workspaceID, projectID string) error {
	return m.Called(ctx, workspaceID, projectID).Error(0)
}

func (m *MockProjectRepository) Restore(ctx context.Context, workspaceID, projectID string) error {
	return m.Called(ctx, workspaceID, projectID).Error(0)
}

func (m *MockProjectRepository) Delete(ctx context.Context, workspaceID, projectID string) error {
	return m.Called(ctx, workspaceID, projectID).Error(0)
}

func TestProjectService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("PRJ-001: Create Project Success", func(t *testing.T) {
		repo := new(MockProjectRepository)
		svc := application.NewService(repo)

		input := domain.CreateProjectInput{
			WorkspaceID: "ws-1",
			Name:        "My First App",
			CreatedBy:   "user-1",
		}

		expected := &domain.Project{
			ID:          "p-1",
			WorkspaceID: "ws-1",
			Name:        "My First App",
			Status:      domain.ProjectActive,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		repo.On("Create", ctx, input).Return(expected, nil)

		p, err := svc.Create(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, "My First App", p.Name)
		repo.AssertExpectations(t)
	})

	t.Run("PRJ-002: Empty Project Name Fails", func(t *testing.T) {
		repo := new(MockProjectRepository)
		svc := application.NewService(repo)

		_, err := svc.Create(ctx, domain.CreateProjectInput{
			WorkspaceID: "ws-1",
			Name:        "   ",
			CreatedBy:   "user-1",
		})
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})

	t.Run("PRJ-003: Name Exceeds Limit Fails", func(t *testing.T) {
		repo := new(MockProjectRepository)
		svc := application.NewService(repo)

		_, err := svc.Create(ctx, domain.CreateProjectInput{
			WorkspaceID: "ws-1",
			Name:        strings.Repeat("a", 101),
			CreatedBy:   "user-1",
		})
		assert.ErrorIs(t, err, domain.ErrInvalidInput)
	})
}

func TestProjectService_Operations(t *testing.T) {
	ctx := context.Background()
	repo := new(MockProjectRepository)
	svc := application.NewService(repo)

	t.Run("PRJ-004: Project List", func(t *testing.T) {
		expected := []domain.Project{
			{ID: "p-1", Name: "App 1", Status: domain.ProjectActive},
		}
		repo.On("List", ctx, "ws-1", "", 20, 0).Return(expected, int64(1), nil).Once()

		list, total, err := svc.List(ctx, "ws-1", "", 20, 0)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, list, 1)
	})

	t.Run("PRJ-005: Get Project", func(t *testing.T) {
		expected := &domain.Project{ID: "p-1", WorkspaceID: "ws-1", Name: "App 1"}
		repo.On("GetByID", ctx, "ws-1", "p-1").Return(expected, nil).Once()

		p, err := svc.Get(ctx, "ws-1", "p-1")
		assert.NoError(t, err)
		assert.Equal(t, "App 1", p.Name)
	})

	t.Run("PRJ-007: Update Project", func(t *testing.T) {
		newName := "Updated App"
		updateInput := domain.UpdateProjectInput{
			ProjectID: "p-1",
			Name:      &newName,
		}
		expected := &domain.Project{ID: "p-1", Name: "Updated App"}
		repo.On("Update", ctx, "ws-1", updateInput).Return(expected, nil).Once()

		p, err := svc.Update(ctx, "ws-1", updateInput)
		assert.NoError(t, err)
		assert.Equal(t, "Updated App", p.Name)
	})

	t.Run("PRJ-008: Archive Project", func(t *testing.T) {
		repo.On("Archive", ctx, "ws-1", "p-1").Return(nil).Once()
		err := svc.Archive(ctx, "ws-1", "p-1")
		assert.NoError(t, err)
	})

	t.Run("PRJ-009: Restore Project", func(t *testing.T) {
		repo.On("Restore", ctx, "ws-1", "p-1").Return(nil).Once()
		err := svc.Restore(ctx, "ws-1", "p-1")
		assert.NoError(t, err)
	})

	t.Run("PRJ-010: Delete Project", func(t *testing.T) {
		repo.On("Delete", ctx, "ws-1", "p-1").Return(nil).Once()
		err := svc.Delete(ctx, "ws-1", "p-1")
		assert.NoError(t, err)
	})
}
