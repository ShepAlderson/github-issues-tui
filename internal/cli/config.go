package cli

import (
	"fmt"

	"github.com/shepbook/github-issues-tui/internal/config"
	"github.com/shepbook/github-issues-tui/internal/setup"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configure ghissues",
		Long:  "Interactive setup for configuring repository and authentication",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgMgr := config.NewManager()

			fmt.Println("Running ghissues configuration...")
			fmt.Println()

			cfg, err := setup.Setup()
			if err != nil {
				return fmt.Errorf("setup failed: %w", err)
			}

			if err := cfgMgr.Save(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Println("\nConfiguration updated successfully!")
			return nil
		},
	}

	return cmd
}