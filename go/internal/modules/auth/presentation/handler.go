package presentation

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/klnp-studio/backend/internal/modules/auth/application"
	"github.com/klnp-studio/backend/internal/security/middleware"
	appErrors "github.com/klnp-studio/backend/internal/shared/errors"
	"github.com/klnp-studio/backend/internal/shared/response"
)

type AuthHandler struct {
	service *application.Service
}

func NewAuthHandler(service *application.Service) *AuthHandler {
	return &AuthHandler{service: service}
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

	profile, err := h.service.GetProfile(c.Context(), uid)
	if err != nil {
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

	profile, err := h.service.UpdateProfile(c.Context(), uid, application.UpdateProfileInput{
		DisplayName: dto.DisplayName,
		AvatarURL:   dto.AvatarURL,
	})
	if err != nil {
		if errors.Is(err, appErrors.ErrValidation) {
			return c.Status(fiber.StatusBadRequest).JSON(
				response.Fail("VALIDATION_ERROR", "display_name must be 80 characters or fewer", nil),
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			response.Fail("INTERNAL_ERROR", "failed to update profile", nil),
		)
	}

	return c.JSON(response.Success(profile, nil))
}
