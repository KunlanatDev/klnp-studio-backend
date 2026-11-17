package document

import (
	"gorm.io/gorm"

	"github.com/klnp-studio/backend/internal/modules/document/application"
	"github.com/klnp-studio/backend/internal/modules/document/infrastructure/persistence"
	"github.com/klnp-studio/backend/internal/modules/document/presentation"
)

type Module struct {
	Service *application.Service
	Handler *presentation.Handler
}

func New(db *gorm.DB) *Module {
	repo := persistence.NewDocumentRepository(db)
	svc := application.NewService(repo)
	return &Module{Service: svc, Handler: presentation.NewHandler(svc)}
}
