package app

import (
	"fmt"
	"log"
)

func (c *Container) Run() error {
	log.Printf("KLNP Studio API listening on %s",
		c.Config.HTTPAddr,
	)

	if err := c.HTTP.Listen(c.Config.HTTPAddr); err != nil {
		return fmt.Errorf("http server: %w", err)
	}

	return nil
}

func (c *Container) Close() error {
	if err := c.HTTP.Shutdown(); err != nil {
		return err
	}

	sqlDB, err := c.DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

