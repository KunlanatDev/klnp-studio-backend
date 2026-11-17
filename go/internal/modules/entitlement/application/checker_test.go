package application

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/klnp-studio/backend/internal/modules/entitlement/domain"
	"github.com/stretchr/testify/require"
)

type mockEntitlementRepo struct {
	result *domain.Entitlements
	err    error
}

func (m *mockEntitlementRepo) Resolve(
	ctx context.Context,
	workspaceID string,
) (*domain.Entitlements, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

func TestEntitlementChecker(t *testing.T) {
	ctx := context.Background()

	t.Run("Check Feature Enabled", func(t *testing.T) {
		repo := &mockEntitlementRepo{
			result: &domain.Entitlements{
				WorkspaceID: "ws-1",
				Features: map[string]domain.Feature{
					"workflow.enabled": {
						Key:   "workflow.enabled",
						Type:  "boolean",
						Value: json.RawMessage("true"),
					},
				},
			},
		}

		checker := NewChecker(NewResolver(repo))
		err := checker.CheckFeature(ctx, "ws-1", "workflow.enabled")
		require.NoError(t, err)
	})

	t.Run("Check Feature Disabled", func(t *testing.T) {
		repo := &mockEntitlementRepo{
			result: &domain.Entitlements{
				WorkspaceID: "ws-1",
				Features: map[string]domain.Feature{
					"workflow.enabled": {
						Key:   "workflow.enabled",
						Type:  "boolean",
						Value: json.RawMessage("false"),
					},
				},
			},
		}

		checker := NewChecker(NewResolver(repo))
		err := checker.CheckFeature(ctx, "ws-1", "workflow.enabled")
		require.ErrorIs(t, err, domain.ErrFeatureDisabled)
	})

	t.Run("Check Feature Not Configured", func(t *testing.T) {
		repo := &mockEntitlementRepo{
			result: &domain.Entitlements{
				WorkspaceID: "ws-1",
				Features:    map[string]domain.Feature{},
			},
		}

		checker := NewChecker(NewResolver(repo))
		err := checker.CheckFeature(ctx, "ws-1", "unknown.feature")
		require.ErrorIs(t, err, domain.ErrFeatureNotConfigured)
	})

	t.Run("Check Limit Quota", func(t *testing.T) {
		repo := &mockEntitlementRepo{
			result: &domain.Entitlements{
				WorkspaceID: "ws-1",
				Limits: map[string]int64{
					"pages": 3,
				},
				Usage: domain.Usage{
					Pages: 2,
				},
			},
		}

		checker := NewChecker(NewResolver(repo))

		// 1 page requested -> total 3 -> allowed
		err := checker.CheckLimit(ctx, "ws-1", "pages", 1)
		require.NoError(t, err)

		// 2 pages requested -> total 4 -> quota exceeded
		err = checker.CheckLimit(ctx, "ws-1", "pages", 2)
		require.ErrorIs(t, err, domain.ErrQuotaExceeded)
	})

	t.Run("Check Unlimited Limit", func(t *testing.T) {
		repo := &mockEntitlementRepo{
			result: &domain.Entitlements{
				WorkspaceID: "ws-1",
				Limits: map[string]int64{
					"pages": domain.Unlimited,
				},
				Usage: domain.Usage{
					Pages: 100,
				},
			},
		}

		checker := NewChecker(NewResolver(repo))
		err := checker.CheckLimit(ctx, "ws-1", "pages", 50)
		require.NoError(t, err)
	})
}

