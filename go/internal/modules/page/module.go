package page

import (
	"gorm.io/gorm"

	"github.com/klnp-studio/backend/internal/modules/page/application"
	"github.com/klnp-studio/backend/internal/modules/page/infrastructure/persistence"
	"github.com/klnp-studio/backend/internal/modules/page/presentation"
)

type Module struct {
	Service *application.Service
	Handler *presentation.Handler
}

func New(db *gorm.DB) *Module {
	repo := persistence.NewPageRepository(db)
	svc := application.NewService(repo)
	handler := presentation.NewHandler(svc)

	return &Module{
		Service: svc,
		Handler: handler,
	}
}
