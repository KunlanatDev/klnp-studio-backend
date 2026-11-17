package http

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/klnp-studio/backend/internal/modules/auth/domain"
	"github.com/klnp-studio/backend/internal/security/middleware"
	appErrors "github.com/klnp-studio/backend/internal/shared/errors"
	"github.com/klnp-studio/backend/internal/shared/response"
)

type AuthHandler struct {
	repo domain.ProfileRepository
}

func NewAuthHandler(repo domain.ProfileRepository) *AuthHandler {
	return &AuthHandler{repo: repo}
}

type UpdateProfileDTO struct {
	DisplayName *string `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
}

func (h *AuthHandler) GetMe(c *fiber.Ctx) error {
	userIDStr := middleware.GetUserID(c)
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			response.Fail("UNAUTHENTICATED", "invalid user id in context", nil),
		)
	}

	profile, err := h.repo.GetByID(c.Context(), uid)
	if err != nil {
		if err == appErrors.ErrNotFound {
			// Profile might not exist yet if trigger didn't run, return basic user profile
			return c.JSON(response.Success(fiber.Map{
				"id":           uid,
				"display_name": "",
				"avatar_url":   "",
			}, nil))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.Fail("INTERNAL_ERROR", "failed to fetch user profile", nil),
		)
	}

	return c.JSON(response.Success(profile, nil))
}

func (h *AuthHandler) UpdateMe(c *fiber.Ctx) error {
	userIDStr := middleware.GetUserID(c)
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			response.Fail("UNAUTHENTICATED", "invalid user id in context", nil),
		)
	}

	var dto UpdateProfileDTO
	if err := c.BodyParser(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			response.Fail("VALIDATION_ERROR", "invalid request payload", nil),
		)
	}

	profile, err := h.repo.GetByID(c.Context(), uid)
	if err != nil {
		if err == appErrors.ErrNotFound {
			profile = &domain.Profile{ID: uid}
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(
				response.Fail("INTERNAL_ERROR", "failed to fetch user profile", nil),
			)
		}
	}

	if dto.DisplayName != nil {
		trimmed := strings.TrimSpace(*dto.DisplayName)
		if len(trimmed) > 80 {
			return c.Status(fiber.StatusBadRequest).JSON(
				response.Fail("VALIDATION_ERROR", "display_name must be 80 characters or fewer", nil),
			)
		}
		profile.DisplayName = trimmed
	}

	if dto.AvatarURL != nil {
		profile.AvatarURL = strings.TrimSpace(*dto.AvatarURL)
	}

	if err := h.repo.Update(c.Context(), profile); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.Fail("INTERNAL_ERROR", "failed to update profile", nil),
		)
	}

	return c.JSON(response.Success(profile, nil))
}
