package workspace

import (
	"gorm.io/gorm"

	"github.com/klnp-studio/backend/internal/modules/workspace/application"
	"github.com/klnp-studio/backend/internal/modules/workspace/domain"
	"github.com/klnp-studio/backend/internal/modules/workspace/infrastructure/persistence"
	"github.com/klnp-studio/backend/internal/modules/workspace/presentation"
)

type Module struct {
	Repository domain.WorkspaceRepository
	UseCase    *application.WorkspaceUseCase
	Handler    *presentation.WorkspaceHandler
}

func New(db *gorm.DB) *Module {
	repo := persistence.NewPostgresWorkspaceRepository(db)
	uc := application.NewWorkspaceUseCase(repo)
	return &Module{
		Repository: repo,
		UseCase:    uc,
		Handler:    presentation.NewWorkspaceHandler(uc),
	}
}
