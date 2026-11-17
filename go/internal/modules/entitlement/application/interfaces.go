package application

import (
	"context"

	"github.com/klnp-studio/backend/internal/modules/entitlement/domain"
)

type EntitlementRepository interface {
	Resolve(
		ctx context.Context,
		workspaceID string,
	) (*domain.Entitlements, error)
}

type PlanAssignmentRepository interface {
	ChangePlan(
		ctx context.Context,
		workspaceID string,
		planID string,
		actorID string,
		idempotencyKey string,
	) error
}

