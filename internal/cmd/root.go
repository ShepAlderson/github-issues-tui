package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shepbook/ghissues/internal/auth"
	"github.com/shepbook/ghissues/internal/config"
	"github.com/shepbook/ghissues/internal/db"
	"github.com/shepbook/ghissues/internal/github"
	"github.com/shepbook/ghissues/internal/setup"
	"github.com/shepbook/ghissues/internal/sync"
	"github.com/shepbook/ghissues/internal/tui"
	"github.com/spf13/cobra"
)

var configPath string
var dbPath string
var repoFlag string
var disableTUI bool // For testing - skips TUI startup

// SetDisableTUI sets whether to skip TUI startup (for testing)
func SetDisableTUI(disable bool) {
	disableTUI = disable
}

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

// SetRepoFlag sets the repository flag value (mainly for testing)
func SetRepoFlag(repo string) {
	repoFlag = repo
}

// GetRepoFlag returns the current repository flag value
func GetRepoFlag() string {
	return repoFlag
}

// ShouldRunSetup returns true if the interactive setup should be run
// (i.e., when config file doesn't exist)
func ShouldRunSetup(path string) bool {
	return !config.Exists(path)
}

// NewRootCmd creates the root command for ghissues
func NewRootCmd() *cobra.Command {
	var dbFlagPath string
	var repoFlagPath string

	rootCmd := &cobra.Command{
		Use:   "ghissues",
		Short: "GitHub Issues TUI - Browse and review GitHub issues offline",
		Long: `ghissues is a terminal user interface for browsing GitHub issues.
It syncs issues from a GitHub repository to a local database for offline access.

On first run, you'll be prompted to configure your repository and authentication.
You can also run 'ghissues config' to reconfigure at any time.

When multiple repositories are configured, use --repo to select which one to view.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Set the dbPath global if flag was provided
			if dbFlagPath != "" {
				SetDBPath(dbFlagPath)
			}
			// Set the repoFlag global if flag was provided
			if repoFlagPath != "" {
				SetRepoFlag(repoFlagPath)
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

			// Determine which repository to use
			activeRepo, err := cfg.GetActiveRepository(GetRepoFlag())
			if err != nil {
				return fmt.Errorf("failed to determine repository: %w", err)
			}

			// Resolve database path: flag > repo config > legacy config > default
			resolvedDBPath := GetDBPath()
			if resolvedDBPath == "" {
				resolvedDBPath = activeRepo.GetDatabasePath()
			}

			if err := db.EnsureDBPath(resolvedDBPath); err != nil {
				return err
			}

			// Open database and load issues
			store, err := db.NewStore(resolvedDBPath)
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			defer store.Close()

			issues, err := store.GetAllIssues(context.Background())
			if err != nil {
				return fmt.Errorf("failed to load issues: %w", err)
			}

			// Get display columns (from config or defaults)
			columns := cfg.Display.Columns
			if len(columns) == 0 {
				columns = config.DefaultDisplayColumns()
			}

			// Get sort settings (from config or defaults)
			sortField := cfg.Display.SortField
			sortOrder := cfg.Display.SortOrder
			if sortField == "" {
				sortField, _ = config.DefaultSortConfig()
			}
			if sortOrder == "" {
				_, sortOrder = config.DefaultSortConfig()
			}

			// Load last sync time from database
			lastSyncTime, err := store.GetLastSyncTime(context.Background())
			if err != nil {
				// Non-fatal error - just log and continue with zero time
				lastSyncTime = time.Time{}
			}

			// Skip TUI if disabled (for testing)
			if disableTUI {
				fmt.Fprintf(cmd.OutOrStdout(), "Ready to browse issues from %s (%d issues)\n", activeRepo.Name, len(issues))
				return nil
			}

			// Parse repository for sync
			owner, repo, err := parseRepo(activeRepo.Name)
			if err != nil {
				return fmt.Errorf("invalid repository format: %w", err)
			}

			// Get theme from config
			theme := cfg.Display.Theme

			// Create and run TUI
			model := tui.NewModelWithTheme(issues, columns, sortField, sortOrder, theme)
			model.SetLastSyncTime(lastSyncTime)

			// Set up refresh function for manual refresh (r key)
			// This creates a closure that captures the necessary context
			model.SetRefreshFunc(createRefreshFunc(cfg, store, owner, repo))

			// Set up load comments function for viewing issue comments
			model.SetLoadCommentsFunc(createLoadCommentsFunc(store))

			p := tea.NewProgram(model, tea.WithAltScreen())

			// Trigger auto-refresh on launch by sending RefreshStartMsg
			go func() {
				p.Send(tui.RefreshStartMsg{})
			}()

			finalModel, err := p.Run()
			if err != nil {
				return fmt.Errorf("TUI error: %w", err)
			}

			// Save sort preferences if they changed
			if m, ok := finalModel.(tui.Model); ok && m.SortChanged() {
				newSortField, newSortOrder := m.GetSortConfig()
				cfg.Display.SortField = newSortField
				cfg.Display.SortOrder = newSortOrder
				if err := config.Save(cfg, path); err != nil {
					// Don't fail, just warn - the main operation succeeded
					fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to save sort preferences: %v\n", err)
				}
			}

			return nil
		},
	}

	// Add persistent flags (available to all subcommands)
	rootCmd.PersistentFlags().StringVar(&dbFlagPath, "db", "", "Path to local database file (default: .ghissues.db)")
	rootCmd.PersistentFlags().StringVar(&repoFlagPath, "repo", "", "Repository to view (format: owner/repo)")

	// Add subcommands
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newSyncCmd())
	rootCmd.AddCommand(newThemesCmd())
	rootCmd.AddCommand(newReposCmd())

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

// parseRepo parses a repository string in the format "owner/repo"
func parseRepo(repoStr string) (owner, repo string, err error) {
	parts := strings.Split(repoStr, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository format: %s (expected owner/repo)", repoStr)
	}
	return parts[0], parts[1], nil
}

// createLoadCommentsFunc creates a function that loads comments for an issue from the database
func createLoadCommentsFunc(store *db.Store) func(issueNumber int) tea.Msg {
	return func(issueNumber int) tea.Msg {
		ctx := context.Background()
		comments, err := store.GetComments(ctx, issueNumber)
		if err != nil {
			// Return empty comments on error (non-fatal)
			return tui.CommentsLoadedMsg{Comments: nil}
		}
		return tui.CommentsLoadedMsg{Comments: comments}
	}
}

// createRefreshFunc creates a function that performs a sync operation
// This function is called as a tea.Cmd and returns a tea.Msg
func createRefreshFunc(cfg *config.Config, store *db.Store, owner, repo string) func() tea.Msg {
	return func() tea.Msg {
		ctx := context.Background()

		// Get authentication token
		token, _, err := auth.GetToken(cfg)
		if err != nil {
			return tui.RefreshErrorMsg{Err: fmt.Errorf("authentication failed: %w", err)}
		}

		// Create client and syncer
		client := github.NewClient(token)
		syncer := sync.NewSyncer(client, store)

		// Run sync synchronously (progress callback not used in TUI for simplicity)
		_, err = syncer.Sync(ctx, owner, repo, nil)
		if err != nil {
			return tui.RefreshErrorMsg{Err: fmt.Errorf("sync failed: %w", err)}
		}

		// Re-fetch issues from database after sync
		issues, err := store.GetAllIssues(ctx)
		if err != nil {
			return tui.RefreshErrorMsg{Err: fmt.Errorf("failed to load issues after sync: %w", err)}
		}

		// Get the updated last sync time from the database
		lastSyncTime, err := store.GetLastSyncTime(ctx)
		if err != nil {
			// Non-fatal - use current time as fallback
			lastSyncTime = time.Now()
		}

		return tui.RefreshDoneMsg{Issues: issues, LastSyncTime: lastSyncTime}
	}
}
