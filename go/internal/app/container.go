package app

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/klnp-studio/backend/internal/config"
	"github.com/klnp-studio/backend/internal/platform/database"
	httpserver "github.com/klnp-studio/backend/internal/platform/http"
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

	router := httpserver.NewRouter(cfg)

	return &Container{
		Config: cfg,
		DB:     db,
		HTTP:   router,
	}, nil
}

