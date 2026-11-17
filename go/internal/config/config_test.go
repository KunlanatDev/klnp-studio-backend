package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigLoadSuccess(t *testing.T) {
	origDB := os.Getenv("DATABASE_URL")
	defer func() {
		if origDB != "" {
			_ = os.Setenv("DATABASE_URL", origDB)
		} else {
			_ = os.Unsetenv("DATABASE_URL")
		}
	}()

	_ = os.Setenv("DATABASE_URL", "postgresql://postgres:postgres@127.0.0.1:54322/postgres?sslmode=disable")
	_ = os.Setenv("HTTP_ADDR", ":8080")

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, ":8080", cfg.HTTPAddr)
	require.NotEmpty(t, cfg.DatabaseURL)
}

func TestConfigMissingDatabaseURL(t *testing.T) {
	origDB := os.Getenv("DATABASE_URL")
	defer func() {
		if origDB != "" {
			_ = os.Setenv("DATABASE_URL", origDB)
		}
	}()

	_ = os.Unsetenv("DATABASE_URL")

	_, err := Load()
	require.Error(t, err)
	require.Contains(t, err.Error(), "DATABASE_URL is required")
}

