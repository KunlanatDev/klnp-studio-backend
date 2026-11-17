package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/klnp-studio/backend/internal/app"
	"github.com/klnp-studio/backend/internal/config"
)

func Execute() {
	root := &cobra.Command{
		Use:   "klp",
		Short: "KLNP Studio CLI",
	}

	root.AddCommand(&cobra.Command{
		Use:   "serve",
		Short: "Start API Server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			container, err := app.New(cfg)
			if err != nil {
				return err
			}

			defer container.Close()

			return container.Run()
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "doctor",
		Short: "Check application configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			fmt.Println("KLNP Studio Environment Check")
			fmt.Println("Environment:", cfg.AppEnv)
			fmt.Println("HTTP Address:", cfg.HTTPAddr)
			fmt.Println("Database URL: configured")

			return nil
		},
	})

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
