package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv           string
	HTTPAddr         string
	DatabaseURL      string
	FrontendURL      string
	SupabaseURL      string
	SupabaseJWKSURL  string
	SupabaseIssuer   string
	SupabaseAudience string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		AppEnv:           getEnv("APP_ENV", "development"),
		HTTPAddr:         getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		FrontendURL:      getEnv("FRONTEND_URL", "http://localhost:5173"),
		SupabaseURL:      getEnv("SUPABASE_URL", "http://127.0.0.1:54321"),
		SupabaseJWKSURL:  getEnv("SUPABASE_JWKS_URL", "http://127.0.0.1:54321/auth/v1/.well-known/jwks.json"),
		SupabaseIssuer:   getEnv("SUPABASE_ISSUER", "http://127.0.0.1:54321/auth/v1"),
		SupabaseAudience: getEnv("SUPABASE_AUDIENCE", "authenticated"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
