package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	jwtverify "github.com/klnp-studio/backend/internal/security/jwt"
	"github.com/klnp-studio/backend/internal/shared/response"
)

const (
	UserIDKey    = "auth.user_id"
	UserEmailKey = "auth.user_email"
)

func RequireAuth(verifier jwtverify.TokenVerifier) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get(fiber.HeaderAuthorization)

		if !strings.HasPrefix(header, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(
				response.Fail("UNAUTHENTICATED", "missing or invalid authorization header", nil),
			)
		}

		rawToken := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		if rawToken == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(
				response.Fail("UNAUTHENTICATED", "empty authorization token", nil),
			)
		}

		claims, err := verifier.Verify(rawToken)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(
				response.Fail("UNAUTHENTICATED", "invalid or expired token", nil),
			)
		}

		c.Locals(UserIDKey, claims.Subject)
		if claims.Email != "" {
			c.Locals(UserEmailKey, claims.Email)
		}

		return c.Next()
	}
}

func GetUserID(c *fiber.Ctx) string {
	val, ok := c.Locals(UserIDKey).(string)
	if !ok {
		return ""
	}
	return val
}
