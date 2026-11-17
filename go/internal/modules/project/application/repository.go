package application

import (
	"context"

	"github.com/klnp-studio/backend/internal/modules/project/domain"
)

type ProjectRepository interface {
	Create(
		ctx context.Context,
		input domain.CreateProjectInput,
	) (*domain.Project, error)

	GetByID(
		ctx context.Context,
		workspaceID string,
		projectID string,
	) (*domain.Project, error)

	List(
		ctx context.Context,
		workspaceID string,
		keyword string,
		limit int,
		offset int,
	) ([]domain.Project, int64, error)

	Update(
		ctx context.Context,
		workspaceID string,
		input domain.UpdateProjectInput,
	) (*domain.Project, error)

	Archive(
		ctx context.Context,
		workspaceID string,
		projectID string,
	) error

	Restore(
		ctx context.Context,
		workspaceID string,
		projectID string,
	) error

	Delete(
		ctx context.Context,
		workspaceID string,
		projectID string,
	) error
}

