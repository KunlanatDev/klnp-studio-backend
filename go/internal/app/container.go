package app

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/klnp-studio/backend/internal/config"
	authDelivery "github.com/klnp-studio/backend/internal/modules/auth/delivery/http"
	authInfra "github.com/klnp-studio/backend/internal/modules/auth/infrastructure"
	"github.com/klnp-studio/backend/internal/modules/entitlement"
	"github.com/klnp-studio/backend/internal/modules/page"
	"github.com/klnp-studio/backend/internal/modules/project"
	workspaceApp "github.com/klnp-studio/backend/internal/modules/workspace/application"
	workspaceDelivery "github.com/klnp-studio/backend/internal/modules/workspace/delivery/http"
	workspaceInfra "github.com/klnp-studio/backend/internal/modules/workspace/infrastructure"
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

	profileRepo := authInfra.NewPostgresProfileRepository(db)
	authHandler := authDelivery.NewAuthHandler(profileRepo)

	wsRepo := workspaceInfra.NewPostgresWorkspaceRepository(db)
	wsUseCase := workspaceApp.NewWorkspaceUseCase(wsRepo)
	wsHandler := workspaceDelivery.NewWorkspaceHandler(wsUseCase)

	entitlementModule := entitlement.New(db, wsRepo)
	projectModule := project.New(db)
	pageModule := page.New(db)

	router := httpserver.NewRouter(cfg, httpserver.RouterOptions{
		Verifier:           verifier,
		AuthHandler:        authHandler,
		WorkspaceHandler:   wsHandler,
		WorkspaceRepo:      wsRepo,
		EntitlementHandler: entitlementModule.Handler,
		ProjectHandler:     projectModule.Handler,
		PageHandler:        pageModule.Handler,
	})

	return &Container{
		Config: cfg,
		DB:     db,
		HTTP:   router,
	}, nil
}
