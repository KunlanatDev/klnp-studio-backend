package cache

import (
	"context"
	"testing"
	"time"

	"github.com/klnp-studio/backend/internal/modules/entitlement/application"
	"github.com/klnp-studio/backend/internal/modules/entitlement/domain"
	"github.com/stretchr/testify/require"
)

type countingRepo struct {
	callCount int
	result    *domain.Entitlements
}

func (c *countingRepo) Resolve(ctx context.Context, workspaceID string) (*domain.Entitlements, error) {
	c.callCount++
	return c.result, nil
}

func TestEntitlementCache(t *testing.T) {
	repo := &countingRepo{
		result: &domain.Entitlements{
			WorkspaceID: "ws-test",
			Plan: domain.Plan{
				Code: "free",
				Name: "Free",
			},
		},
	}

	resolver := application.NewResolver(repo)
	c := New(resolver)
	ctx := context.Background()

	// 1. First call hits repo
	res1, err := c.Get(ctx, "ws-test")
	require.NoError(t, err)
	require.Equal(t, "free", res1.Plan.Code)
	require.Equal(t, 1, repo.callCount)

	// 2. Second call hits cache (count remains 1)
	res2, err := c.Get(ctx, "ws-test")
	require.NoError(t, err)
	require.Equal(t, "free", res2.Plan.Code)
	require.Equal(t, 1, repo.callCount)

	// 3. Invalidate workspace cache -> next call hits repo again
	c.Invalidate("ws-test")
	res3, err := c.Get(ctx, "ws-test")
	require.NoError(t, err)
	require.Equal(t, "free", res3.Plan.Code)
	require.Equal(t, 2, repo.callCount)

	// 4. Test TTL expiration
	c.ttl = 10 * time.Millisecond
	c.Clear()
	_, _ = c.Get(ctx, "ws-test")
	require.Equal(t, 3, repo.callCount)

	time.Sleep(20 * time.Millisecond)
	_, _ = c.Get(ctx, "ws-test")
	require.Equal(t, 4, repo.callCount)
}

