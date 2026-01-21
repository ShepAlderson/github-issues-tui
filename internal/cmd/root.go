package cmd

import (
	"fmt"

	"github.com/shepbook/ghissues/internal/config"
	"github.com/shepbook/ghissues/internal/db"
	"github.com/shepbook/ghissues/internal/setup"
	"github.com/spf13/cobra"
)

var configPath string
var dbPath string

// SetConfigPath sets a custom config path (mainly for testing)
func SetConfigPath(path string) {
	configPath = path
}

// GetConfigPath returns the current config path, defaulting to the standard location
func GetConfigPath() string {
	if configPath == "" {
		return config.DefaultConfigPath()
	}
	return configPath
}

// SetDBPath sets a custom database path (mainly for testing)
func SetDBPath(path string) {
	dbPath = path
}

// GetDBPath returns the current database path (empty string means use default)
func GetDBPath() string {
	return dbPath
}

// ShouldRunSetup returns true if the interactive setup should be run
// (i.e., when config file doesn't exist)
func ShouldRunSetup(path string) bool {
	return !config.Exists(path)
}

// NewRootCmd creates the root command for ghissues
func NewRootCmd() *cobra.Command {
	var dbFlagPath string

	rootCmd := &cobra.Command{
		Use:   "ghissues",
		Short: "GitHub Issues TUI - Browse and review GitHub issues offline",
		Long: `ghissues is a terminal user interface for browsing GitHub issues.
It syncs issues from a GitHub repository to a local database for offline access.

On first run, you'll be prompted to configure your repository and authentication.
You can also run 'ghissues config' to reconfigure at any time.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Set the dbPath global if flag was provided
			if dbFlagPath != "" {
				SetDBPath(dbFlagPath)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			path := GetConfigPath()

			// Check if first-time setup is needed
			if ShouldRunSetup(path) {
				fmt.Fprintln(cmd.OutOrStdout(), "Welcome to ghissues! Let's set up your configuration.")
				fmt.Fprintln(cmd.OutOrStdout())
				_, err := setup.RunInteractiveSetup(path)
				if err != nil {
					return fmt.Errorf("setup failed: %w", err)
				}
				fmt.Fprintln(cmd.OutOrStdout())
				fmt.Fprintln(cmd.OutOrStdout(), "Configuration saved! You can reconfigure anytime with 'ghissues config'")
			}

			// Load config and start TUI (placeholder for now)
			cfg, err := config.Load(path)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Resolve and validate database path
			resolvedDBPath, err := db.ResolveDBPath(GetDBPath(), cfg)
			if err != nil {
				return fmt.Errorf("failed to resolve database path: %w", err)
			}

			if err := db.EnsureDBPath(resolvedDBPath); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Ready to browse issues from %s\n", cfg.Repository)
			fmt.Fprintln(cmd.OutOrStdout(), "(TUI implementation coming soon)")

			return nil
		},
	}

	// Add persistent flags (available to all subcommands)
	rootCmd.PersistentFlags().StringVar(&dbFlagPath, "db", "", "Path to local database file (default: .ghissues.db)")

	// Add subcommands
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newSyncCmd())

	return rootCmd
}

// newConfigCmd creates the config subcommand
func newConfigCmd() *cobra.Command {
	var repo string
	var authMethod string
	var token string

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configure ghissues settings (run interactive configuration)",
		Long: `Configure ghissues settings including repository and authentication method.

When run without flags, starts an interactive setup wizard.
You can also provide flags to configure non-interactively.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := GetConfigPath()

			// If flags provided, use non-interactive mode
			if repo != "" && authMethod != "" {
				_, err := setup.RunSetupWithValues(repo, authMethod, token, path)
				if err != nil {
					return err
				}
				fmt.Println("Configuration saved successfully!")
				return nil
			}

			// Otherwise, run interactive setup
			fmt.Println("Let's configure ghissues!")
			fmt.Println()
			_, err := setup.RunInteractiveSetup(path)
			if err != nil {
				return fmt.Errorf("setup failed: %w", err)
			}
			fmt.Println()
			fmt.Println("Configuration saved!")
			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", "", "GitHub repository in owner/repo format")
	cmd.Flags().StringVar(&authMethod, "auth-method", "", "Authentication method: env, token, or gh")
	cmd.Flags().StringVar(&token, "token", "", "GitHub personal access token (required when auth-method is 'token')")

	return cmd
}

// Execute runs the root command
func Execute() error {
	return NewRootCmd().Execute()
}
