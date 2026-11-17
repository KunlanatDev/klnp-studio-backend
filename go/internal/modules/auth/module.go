package auth

import (
	"gorm.io/gorm"

	"github.com/klnp-studio/backend/internal/modules/auth/application"
	"github.com/klnp-studio/backend/internal/modules/auth/domain"
	"github.com/klnp-studio/backend/internal/modules/auth/infrastructure/persistence"
	"github.com/klnp-studio/backend/internal/modules/auth/presentation"
)

type Module struct {
	Repository domain.ProfileRepository
	Service    *application.Service
	Handler    *presentation.AuthHandler
}

func New(db *gorm.DB) *Module {
	repo := persistence.NewPostgresProfileRepository(db)
	svc := application.NewService(repo)
	return &Module{
		Repository: repo,
		Service:    svc,
		Handler:    presentation.NewAuthHandler(svc),
	}
}
