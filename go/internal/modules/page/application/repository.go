package application

import (
	"context"

	"github.com/klnp-studio/backend/internal/modules/page/domain"
)

type PageRepository interface {
	Create(
		ctx context.Context,
		input domain.CreatePageInput,
	) (*domain.Page, error)

	GetByID(
		ctx context.Context,
		workspaceID string,
		projectID string,
		pageID string,
	) (*domain.Page, error)

	List(
		ctx context.Context,
		workspaceID string,
		projectID string,
	) ([]domain.Page, error)

	Update(
		ctx context.Context,
		input domain.UpdatePageInput,
	) (*domain.Page, error)

	Delete(
		ctx context.Context,
		workspaceID string,
		projectID string,
		pageID string,
	) error

	Duplicate(
		ctx context.Context,
		workspaceID string,
		projectID string,
		pageID string,
		actorID string,
	) (*domain.Page, error)

	SetHome(
		ctx context.Context,
		workspaceID string,
		projectID string,
		pageID string,
	) error

	Reorder(
		ctx context.Context,
		input domain.ReorderPagesInput,
	) error

	Archive(
		ctx context.Context,
		workspaceID string,
		projectID string,
		pageID string,
	) error

	Restore(
		ctx context.Context,
		workspaceID string,
		projectID string,
		pageID string,
	) error
}
