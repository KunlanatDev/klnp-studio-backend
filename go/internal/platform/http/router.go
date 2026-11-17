package httpserver

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/klnp-studio/backend/internal/config"
	authHttp "github.com/klnp-studio/backend/internal/modules/auth/delivery/http"
	entitlementPresentation "github.com/klnp-studio/backend/internal/modules/entitlement/presentation"
	pagePresentation "github.com/klnp-studio/backend/internal/modules/page/presentation"
	projectPresentation "github.com/klnp-studio/backend/internal/modules/project/presentation"
	workspaceHttp "github.com/klnp-studio/backend/internal/modules/workspace/delivery/http"
	workspaceDomain "github.com/klnp-studio/backend/internal/modules/workspace/domain"
	jwtverify "github.com/klnp-studio/backend/internal/security/jwt"
	"github.com/klnp-studio/backend/internal/security/middleware"
)

type RouterOptions struct {
	Verifier           jwtverify.TokenVerifier
	AuthHandler        *authHttp.AuthHandler
	WorkspaceHandler   *workspaceHttp.WorkspaceHandler
	WorkspaceRepo      workspaceDomain.WorkspaceRepository
	EntitlementHandler *entitlementPresentation.Handler
	ProjectHandler     *projectPresentation.Handler
	PageHandler        *pagePresentation.Handler
}

func NewRouter(cfg config.Config, opts ...RouterOptions) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "KLNP Studio API",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	})

	app.Use(recover.New())
	app.Use(logger.New())

	origins := "http://localhost:5173,http://127.0.0.1:5173,http://localhost:3000,http://127.0.0.1:3000"
	if cfg.FrontendURL != "" && cfg.FrontendURL != "http://localhost:5173" && cfg.FrontendURL != "http://127.0.0.1:5173" {
		origins += "," + cfg.FrontendURL
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins: origins,
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "klnp-studio-api",
		})
	})

	api := app.Group("/api/v1")

	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"version": "v1",
		})
	})

	if len(opts) > 0 {
		opt := opts[0]
		if opt.Verifier != nil {
			authMiddleware := middleware.RequireAuth(opt.Verifier)

			if opt.AuthHandler != nil {
				me := api.Group("/me", authMiddleware)
				me.Get("", opt.AuthHandler.GetMe)
				me.Patch("", opt.AuthHandler.UpdateMe)
			}

			if opt.WorkspaceHandler != nil {
				ws := api.Group("/workspaces", authMiddleware)
				ws.Get("", opt.WorkspaceHandler.List)
				ws.Post("", opt.WorkspaceHandler.Create)
				ws.Get("/:workspaceId", opt.WorkspaceHandler.Get)
				ws.Patch("/:workspaceId", opt.WorkspaceHandler.Update)
			}

			if opt.EntitlementHandler != nil {
				wsEnt := api.Group("/workspaces/:workspaceId", authMiddleware)
				wsEnt.Get("/entitlements", opt.EntitlementHandler.GetEntitlements)
				wsEnt.Get("/plan", opt.EntitlementHandler.GetPlan)
				wsEnt.Get("/usage", opt.EntitlementHandler.GetUsage)
				wsEnt.Get("/features", opt.EntitlementHandler.GetFeatures)
				wsEnt.Get("/limits", opt.EntitlementHandler.GetLimits)
			}

			if opt.WorkspaceRepo != nil {
				// Project Routes
				if opt.ProjectHandler != nil {
					proj := api.Group("/workspaces/:workspaceId/projects", authMiddleware)
					proj.Get("", middleware.RequireWorkspacePermission(opt.WorkspaceRepo, middleware.PermissionProjectRead), opt.ProjectHandler.List)
					proj.Post("", middleware.RequireWorkspacePermission(opt.WorkspaceRepo, middleware.PermissionProjectWrite), opt.ProjectHandler.Create)
					proj.Get("/:projectId", middleware.RequireWorkspacePermission(opt.WorkspaceRepo, middleware.PermissionProjectRead), opt.ProjectHandler.Get)
					proj.Patch("/:projectId", middleware.RequireWorkspacePermission(opt.WorkspaceRepo, middleware.PermissionProjectWrite), opt.ProjectHandler.Update)
					proj.Post("/:projectId/archive", middleware.RequireWorkspacePermission(opt.WorkspaceRepo, middleware.PermissionProjectWrite), opt.ProjectHandler.Archive)
					proj.Post("/:projectId/restore", middleware.RequireWorkspacePermission(opt.WorkspaceRepo, middleware.PermissionProjectWrite), opt.ProjectHandler.Restore)
					proj.Delete("/:projectId", middleware.RequireWorkspacePermission(opt.WorkspaceRepo, middleware.PermissionProjectDelete), opt.ProjectHandler.Delete)
				}

				// Page Routes
				if opt.PageHandler != nil {
					pages := api.Group("/workspaces/:workspaceId/projects/:projectId/pages", authMiddleware)
					pages.Get("", middleware.RequireWorkspacePermission(opt.WorkspaceRepo, middleware.PermissionPageRead), opt.PageHandler.List)
					pages.Post("", middleware.RequireWorkspacePermission(opt.WorkspaceRepo, middleware.PermissionPageWrite), opt.PageHandler.Create)
					pages.Put("/reorder", middleware.RequireWorkspacePermission(opt.WorkspaceRepo, middleware.PermissionPageWrite), opt.PageHandler.Reorder)
					pages.Get("/:pageId", middleware.RequireWorkspacePermission(opt.WorkspaceRepo, middleware.PermissionPageRead), opt.PageHandler.Get)
					pages.Patch("/:pageId", middleware.RequireWorkspacePermission(opt.WorkspaceRepo, middleware.PermissionPageWrite), opt.PageHandler.Update)
					pages.Delete("/:pageId", middleware.RequireWorkspacePermission(opt.WorkspaceRepo, middleware.PermissionPageDelete), opt.PageHandler.Delete)
					pages.Post("/:pageId/duplicate", middleware.RequireWorkspacePermission(opt.WorkspaceRepo, middleware.PermissionPageWrite), opt.PageHandler.Duplicate)
					pages.Post("/:pageId/home", middleware.RequireWorkspacePermission(opt.WorkspaceRepo, middleware.PermissionPageWrite), opt.PageHandler.SetHome)
					pages.Post("/:pageId/archive", middleware.RequireWorkspacePermission(opt.WorkspaceRepo, middleware.PermissionPageWrite), opt.PageHandler.Archive)
					pages.Post("/:pageId/restore", middleware.RequireWorkspacePermission(opt.WorkspaceRepo, middleware.PermissionPageWrite), opt.PageHandler.Restore)
				}
			}
		}
	}

	_ = healthcheck.ConfigDefault

	return app
}
