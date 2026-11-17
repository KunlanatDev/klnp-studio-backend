package cache

import (
	"context"
	"sync"
	"time"

	"github.com/klnp-studio/backend/internal/modules/entitlement/application"
	"github.com/klnp-studio/backend/internal/modules/entitlement/domain"
)

type entry struct {
	value     *domain.Entitlements
	expiresAt time.Time
}

type EntitlementCache struct {
	resolver *application.Resolver
	mu       sync.RWMutex
	items    map[string]entry
	ttl      time.Duration
}

func New(resolver *application.Resolver) *EntitlementCache {
	return &EntitlementCache{
		resolver: resolver,
		items:    make(map[string]entry),
		ttl:      30 * time.Second,
	}
}

func (c *EntitlementCache) Get(
	ctx context.Context,
	workspaceID string,
) (*domain.Entitlements, error) {
	c.mu.RLock()
	item, exists := c.items[workspaceID]
	c.mu.RUnlock()

	if exists && time.Now().Before(item.expiresAt) {
		return item.value, nil
	}

	value, err := c.resolver.Resolve(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.items[workspaceID] = entry{
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()

	return value, nil
}

func (c *EntitlementCache) Invalidate(workspaceID string) {
	c.mu.Lock()
	delete(c.items, workspaceID)
	c.mu.Unlock()
}

func (c *EntitlementCache) Clear() {
	c.mu.Lock()
	c.items = make(map[string]entry)
	c.mu.Unlock()
}

