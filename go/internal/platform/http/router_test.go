package httpserver

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	"github.com/klnp-studio/backend/internal/config"
)

func TestHealthEndpoint(t *testing.T) {
	app := NewRouter(config.Config{
		AppEnv:      "test",
		HTTPAddr:    ":8080",
		FrontendURL: "http://localhost:5173",
	})

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "klnp-studio-api")
}

func TestApiV1HealthEndpoint(t *testing.T) {
	app := NewRouter(config.Config{
		AppEnv:      "test",
		HTTPAddr:    ":8080",
		FrontendURL: "http://localhost:5173",
	})

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	resp, err := app.Test(req)

	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "v1")
}

