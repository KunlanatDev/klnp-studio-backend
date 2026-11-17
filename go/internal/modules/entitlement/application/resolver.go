package application

import (
	"context"

	"github.com/klnp-studio/backend/internal/modules/entitlement/domain"
)

type Resolver struct {
	repo EntitlementRepository
}

func NewResolver(repo EntitlementRepository) *Resolver {
	return &Resolver{repo: repo}
}

func (r *Resolver) Resolve(
	ctx context.Context,
	workspaceID string,
) (*domain.Entitlements, error) {
	if workspaceID == "" {
		return nil, domain.ErrNotFound
	}
	return r.repo.Resolve(ctx, workspaceID)
}

