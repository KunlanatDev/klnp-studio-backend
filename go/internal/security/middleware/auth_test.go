package middleware

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	jwtverify "github.com/klnp-studio/backend/internal/security/jwt"
	"github.com/stretchr/testify/require"
)

type mockVerifier struct {
	claims *jwtverify.Claims
	err    error
}

func (m *mockVerifier) Verify(rawToken string) (*jwtverify.Claims, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.claims, nil
}

func TestRequireAuth(t *testing.T) {
	t.Run("Missing Authorization Header", func(t *testing.T) {
		app := fiber.New()
		app.Use(RequireAuth(&mockVerifier{}))
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendString("ok")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		app := fiber.New()
		app.Use(RequireAuth(&mockVerifier{err: errors.New("invalid token")}))
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendString("ok")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid.jwt.token")
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Valid Token Sets UserID", func(t *testing.T) {
		verifier := &mockVerifier{
			claims: &jwtverify.Claims{
				Role: "authenticated",
			},
		}
		verifier.claims.Subject = "123e4567-e89b-12d3-a456-426614174000"

		app := fiber.New()
		app.Use(RequireAuth(verifier))
		app.Get("/test", func(c *fiber.Ctx) error {
			uid := GetUserID(c)
			return c.SendString(uid)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer valid.jwt.token")
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}
