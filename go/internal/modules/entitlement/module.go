package entitlement

import (
	"github.com/klnp-studio/backend/internal/modules/entitlement/application"
	"github.com/klnp-studio/backend/internal/modules/entitlement/infrastructure/cache"
	"github.com/klnp-studio/backend/internal/modules/entitlement/infrastructure/persistence"
	"github.com/klnp-studio/backend/internal/modules/entitlement/presentation"
	workspaceDomain "github.com/klnp-studio/backend/internal/modules/workspace/domain"
	"gorm.io/gorm"
)

type Module struct {
	Resolver   *application.Resolver
	Checker    *application.Checker
	Assignment *persistence.PlanAssignmentRepository
	Cache      *cache.EntitlementCache
	Handler    *presentation.Handler
}

func New(db *gorm.DB, wsRepo workspaceDomain.WorkspaceRepository) *Module {
	repository := persistence.NewEntitlementRepository(db)
	resolver := application.NewResolver(repository)
	checker := application.NewChecker(resolver)
	entitlementCache := cache.New(resolver)
	assignment := persistence.NewPlanAssignmentRepository(db)
	handler := presentation.NewHandler(resolver, wsRepo)

	return &Module{
		Resolver:   resolver,
		Checker:    checker,
		Assignment: assignment,
		Cache:      entitlementCache,
		Handler:    handler,
	}
}

