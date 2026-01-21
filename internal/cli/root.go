package cli

import (
	"fmt"
	"os"

	"github.com/shepbook/github-issues-tui/internal/config"
	"github.com/shepbook/github-issues-tui/internal/setup"
	"github.com/spf13/cobra"
)

// Execute runs the CLI application
func Execute() error {
	rootCmd := &cobra.Command{
		Use:   "ghissues",
		Short: "GitHub Issues TUI",
		Long:  "A terminal UI for browsing GitHub issues offline",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgMgr := config.NewManager()

			// Check if config exists
			exists, err := cfgMgr.Exists()
			if err != nil {
				return fmt.Errorf("failed to check config: %w", err)
			}

			if !exists {
				fmt.Println("No configuration found. Running first-time setup...")
				return runSetup(cfgMgr)
			}

			// Config exists, launch TUI (to be implemented in later stories)
			fmt.Println("TUI will be launched here (to be implemented)")
			return nil
		},
	}

	// Add subcommands
	rootCmd.AddCommand(newConfigCmd())

	return rootCmd.Execute()
}

func runSetup(cfgMgr *config.Manager) error {
	cfg, err := setup.Setup()
	if err != nil {
		return fmt.Errorf("setup failed: %w", err)
	}

	if err := cfgMgr.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("\nConfiguration saved to %s\n", os.ExpandEnv("$HOME/.config/ghissues/config.toml"))
	fmt.Println("You can now run 'ghissues' to start the application.")
	fmt.Println("To reconfigure, run 'ghissues config'.")
	return nil
}