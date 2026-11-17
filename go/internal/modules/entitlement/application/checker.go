package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/klnp-studio/backend/internal/modules/entitlement/domain"
)

type Checker struct {
	resolver *Resolver
}

func NewChecker(resolver *Resolver) *Checker {
	return &Checker{resolver: resolver}
}

func (c *Checker) CheckFeature(
	ctx context.Context,
	workspaceID string,
	featureKey string,
) error {
	ent, err := c.resolver.Resolve(ctx, workspaceID)
	if err != nil {
		return err
	}

	feature, exists := ent.Features[featureKey]
	if !exists {
		return domain.ErrFeatureNotConfigured
	}

	if feature.Type != "boolean" {
		return fmt.Errorf("feature %s is not boolean", featureKey)
	}

	var enabled bool
	if err := json.Unmarshal(feature.Value, &enabled); err != nil {
		return fmt.Errorf("invalid feature value: %w", err)
	}

	if !enabled {
		return domain.ErrFeatureDisabled
	}

	return nil
}

func (c *Checker) CheckLimit(
	ctx context.Context,
	workspaceID string,
	key string,
	requested int64,
) error {
	if requested <= 0 {
		return domain.ErrInvalidLimit
	}

	ent, err := c.resolver.Resolve(ctx, workspaceID)
	if err != nil {
		return err
	}

	limit, exists := ent.Limits[key]
	if !exists {
		return domain.ErrInvalidLimit
	}

	if limit == domain.Unlimited {
		return nil
	}

	if limit < 0 {
		return domain.ErrInvalidLimit
	}

	var used int64
	switch key {
	case "pages":
		used = ent.Usage.Pages
	default:
		return domain.ErrInvalidLimit
	}

	if used > limit || requested > limit-used {
		return domain.ErrQuotaExceeded
	}

	return nil
}

