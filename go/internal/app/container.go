package app

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/klnp-studio/backend/internal/config"
	"github.com/klnp-studio/backend/internal/modules/auth"
	"github.com/klnp-studio/backend/internal/modules/document"
	"github.com/klnp-studio/backend/internal/modules/entitlement"
	"github.com/klnp-studio/backend/internal/modules/page"
	"github.com/klnp-studio/backend/internal/modules/project"
	"github.com/klnp-studio/backend/internal/modules/workspace"
	"github.com/klnp-studio/backend/internal/platform/database"
	httpserver "github.com/klnp-studio/backend/internal/platform/http"
	jwtverify "github.com/klnp-studio/backend/internal/security/jwt"
)

type Container struct {
	Config config.Config
	DB     *gorm.DB
	HTTP   *fiber.App
}

func New(cfg config.Config) (*Container, error) {
	db, err := database.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	var verifier jwtverify.TokenVerifier
	if cfg.SupabaseJWKSURL != "" {
		v, err := jwtverify.New(cfg.SupabaseJWKSURL, cfg.SupabaseIssuer, cfg.SupabaseAudience)
		if err != nil {
			log.Printf("[WARN] Failed to initialize JWKS verifier from %s: %v", cfg.SupabaseJWKSURL, err)
		} else {
			verifier = v
		}
	}

	authModule := auth.New(db)
	workspaceModule := workspace.New(db)

	entitlementModule := entitlement.New(db, workspaceModule.Repository)
	projectModule := project.New(db)
	pageModule := page.New(db)
	documentModule := document.New(db)

	router := httpserver.NewRouter(cfg, httpserver.RouterOptions{
		Verifier:           verifier,
		AuthHandler:        authModule.Handler,
		WorkspaceHandler:   workspaceModule.Handler,
		WorkspaceRepo:      workspaceModule.Repository,
		EntitlementHandler: entitlementModule.Handler,
		ProjectHandler:     projectModule.Handler,
		PageHandler:        pageModule.Handler,
		DocumentHandler:    documentModule.Handler,
	})

	return &Container{
		Config: cfg,
		DB:     db,
		HTTP:   router,
	}, nil
}
