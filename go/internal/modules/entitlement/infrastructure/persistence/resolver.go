package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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
	var raw string

	row := r.db.WithContext(ctx).
		Raw(`SELECT public.s2_02_02_resolve_entitlements(?)`, workspaceID).
		Row()

	if err := row.Scan(&raw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("resolve entitlement: %w", err)
	}

	if raw == "" || raw == "null" {
		return nil, domain.ErrNotFound
	}

	var result domain.Entitlements
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("decode entitlement: %w", err)
	}

	if result.WorkspaceID != workspaceID {
		return nil, domain.ErrNotFound
	}

	return &result, nil
}
