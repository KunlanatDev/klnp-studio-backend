package persistence

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/klnp-studio/backend/internal/modules/entitlement/domain"
	"gorm.io/gorm"
)

type EntitlementRepository struct {
	db *gorm.DB
}

func NewEntitlementRepository(db *gorm.DB) *EntitlementRepository {
	return &EntitlementRepository{db: db}
}

func (r *EntitlementRepository) Resolve(
	ctx context.Context,
	workspaceID string,
) (*domain.Entitlements, error) {
	var raw []byte

	err := r.db.WithContext(ctx).
		Raw(`SELECT public.s2_02_02_resolve_entitlements(?)`, workspaceID).
		Scan(&raw).Error

	if err != nil {
		return nil, fmt.Errorf("resolve entitlement: %w", err)
	}

	if len(raw) == 0 || string(raw) == "null" {
		return nil, domain.ErrNotFound
	}

	var result domain.Entitlements
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("decode entitlement: %w", err)
	}

	if result.WorkspaceID != workspaceID {
		return nil, domain.ErrNotFound
	}

	return &result, nil
}

