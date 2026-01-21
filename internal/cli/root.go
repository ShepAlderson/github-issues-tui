package cli

import (
	"fmt"
	"os"

	"github.com/shepbook/github-issues-tui/internal/config"
	"github.com/shepbook/github-issues-tui/internal/database"
	"github.com/shepbook/github-issues-tui/internal/setup"
	"github.com/spf13/cobra"
)

// Execute runs the CLI application
func Execute() error {
	var dbPath string

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

			// Load config to get database path from config file
			cfg, err := cfgMgr.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Get database path following precedence: flag -> config -> default
			finalDbPath := database.GetDatabasePath(dbPath, cfg)

			// Ensure database path is writable and parent directories exist
			if err := database.EnsureDatabasePath(finalDbPath); err != nil {
				return fmt.Errorf("database path error: %w", err)
			}

			fmt.Printf("Database will be stored at: %s\n", database.AbsolutePath(finalDbPath))

			// Config exists, launch TUI (to be implemented in later stories)
			fmt.Println("TUI will be launched here (to be implemented)")
			return nil
		},
	}

	// Add --db flag
	rootCmd.Flags().StringVarP(&dbPath, "db", "d", "", "database file path (overrides config)")

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