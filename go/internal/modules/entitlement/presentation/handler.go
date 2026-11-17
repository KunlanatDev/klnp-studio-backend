package presentation

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/klnp-studio/backend/internal/modules/entitlement/application"
	"github.com/klnp-studio/backend/internal/modules/entitlement/domain"
	workspaceDomain "github.com/klnp-studio/backend/internal/modules/workspace/domain"
	"github.com/klnp-studio/backend/internal/security/middleware"
	"github.com/klnp-studio/backend/internal/shared/response"
)

type Handler struct {
	resolver      *application.Resolver
	workspaceRepo workspaceDomain.WorkspaceRepository
}

func NewHandler(
	resolver *application.Resolver,
	workspaceRepo workspaceDomain.WorkspaceRepository,
) *Handler {
	return &Handler{
		resolver:      resolver,
		workspaceRepo: workspaceRepo,
	}
}

func (h *Handler) verifyMembership(c *fiber.Ctx, workspaceID string) error {
	userIDStr := middleware.GetUserID(c)
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "UNAUTHENTICATED")
	}

	wsID, err := uuid.Parse(workspaceID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "NOT_FOUND")
	}

	if h.workspaceRepo != nil {
		_, err := h.workspaceRepo.GetMember(c.Context(), wsID, uid)
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, "NOT_FOUND")
		}
	}

	return nil
}

func (h *Handler) GetEntitlements(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")

	if err := h.verifyMembership(c, workspaceID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			response.Fail("NOT_FOUND", "workspace not found", nil),
		)
	}

	result, err := h.resolver.Resolve(c.Context(), workspaceID)
	if err != nil {
		if err == domain.ErrNotFound {
			return c.Status(fiber.StatusServiceUnavailable).JSON(
				response.Fail("ENTITLEMENT_NOT_FOUND", "workspace entitlement unavailable", nil),
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.Fail("INTERNAL_ERROR", "failed to resolve entitlements", nil),
		)
	}

	features := make(map[string]FeatureDTO, len(result.Features))
	for key, feature := range result.Features {
		var val any
		if err := json.Unmarshal(feature.Value, &val); err == nil {
			features[key] = FeatureDTO{
				Type:  feature.Type,
				Value: val,
			}
		}
	}

	limit := result.Limits["pages"]
	used := result.Usage.Pages

	item := UsageItemDTO{
		Used:      used,
		Limit:     limit,
		Unlimited: limit == domain.Unlimited,
	}

	if limit >= 0 {
		rem := limit - used
		if rem < 0 {
			rem = 0
		}
		item.Remaining = &rem
	}

	return c.JSON(response.Success(EntitlementResponse{
		WorkspaceID: result.WorkspaceID,
		Plan: PlanDTO{
			Code: result.Plan.Code,
			Name: result.Plan.Name,
		},
		Limits:   result.Limits,
		Features: features,
		Usage: UsageDTO{
			Pages: item,
		},
	}, nil))
}

func (h *Handler) GetPlan(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")
	if err := h.verifyMembership(c, workspaceID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("NOT_FOUND", "workspace not found", nil))
	}

	res, err := h.resolver.Resolve(c.Context(), workspaceID)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.Fail("ENTITLEMENT_NOT_FOUND", "entitlement not found", nil))
	}

	return c.JSON(response.Success(PlanDTO{
		Code: res.Plan.Code,
		Name: res.Plan.Name,
	}, nil))
}

func (h *Handler) GetUsage(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")
	if err := h.verifyMembership(c, workspaceID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("NOT_FOUND", "workspace not found", nil))
	}

	res, err := h.resolver.Resolve(c.Context(), workspaceID)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.Fail("ENTITLEMENT_NOT_FOUND", "entitlement not found", nil))
	}

	limit := res.Limits["pages"]
	used := res.Usage.Pages
	item := UsageItemDTO{
		Used:      used,
		Limit:     limit,
		Unlimited: limit == domain.Unlimited,
	}
	if limit >= 0 {
		rem := limit - used
		if rem < 0 {
			rem = 0
		}
		item.Remaining = &rem
	}

	return c.JSON(response.Success(fiber.Map{
		"workspace_id": res.WorkspaceID,
		"resources": fiber.Map{
			"pages": item,
		},
	}, nil))
}

func (h *Handler) GetFeatures(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")
	if err := h.verifyMembership(c, workspaceID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("NOT_FOUND", "workspace not found", nil))
	}

	res, err := h.resolver.Resolve(c.Context(), workspaceID)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.Fail("ENTITLEMENT_NOT_FOUND", "entitlement not found", nil))
	}

	features := make(map[string]FeatureDTO, len(res.Features))
	for key, feature := range res.Features {
		var val any
		if err := json.Unmarshal(feature.Value, &val); err == nil {
			features[key] = FeatureDTO{
				Type:  feature.Type,
				Value: val,
			}
		}
	}

	return c.JSON(response.Success(features, nil))
}

func (h *Handler) GetLimits(c *fiber.Ctx) error {
	workspaceID := c.Params("workspaceId")
	if err := h.verifyMembership(c, workspaceID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.Fail("NOT_FOUND", "workspace not found", nil))
	}

	res, err := h.resolver.Resolve(c.Context(), workspaceID)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.Fail("ENTITLEMENT_NOT_FOUND", "entitlement not found", nil))
	}

	return c.JSON(response.Success(res.Limits, nil))
}

