package project

import (
	"gorm.io/gorm"

	"github.com/klnp-studio/backend/internal/modules/project/application"
	"github.com/klnp-studio/backend/internal/modules/project/infrastructure/persistence"
	"github.com/klnp-studio/backend/internal/modules/project/presentation"
)

type Module struct {
	Service *application.Service
	Handler *presentation.Handler
}

func New(db *gorm.DB) *Module {
	repo := persistence.NewProjectRepository(db)
	svc := application.NewService(repo)
	handler := presentation.NewHandler(svc)

	return &Module{
		Service: svc,
		Handler: handler,
	}
}
