package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/shepbook/git/github-issues-tui/internal/auth"
	"github.com/shepbook/git/github-issues-tui/internal/cmd"
	"github.com/shepbook/git/github-issues-tui/internal/config"
	"github.com/shepbook/git/github-issues-tui/internal/tui"
)

func main() {
	if err := runMain(os.Args, getConfigFilePath(), os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runMain(args []string, configPath string, input io.Reader, output io.Writer) error {
	// Parse flags
	var dbFlag string
	if len(args) >= 2 && args[1] == "--db" {
		if len(args) < 3 {
			return fmt.Errorf("--db flag requires a path argument")
		}
		dbFlag = args[2]
		// Remove the flag from args for further processing
		args = append(args[:1], args[3:]...)
	}

	// Check number of arguments
	if len(args) > 2 {
		return fmt.Errorf("too many arguments")
	}

	// Check if config command was requested
	if len(args) >= 2 && args[1] == "config" {
		return cmd.RunConfigCommand(configPath, input, output)
	}

	// Check if sync command was requested
	if len(args) >= 2 && args[1] == "sync" {
		// Sync command requires config to be loaded
		exists, err := config.ConfigExists(configPath)
		if err != nil {
			return fmt.Errorf("failed to check config file: %w", err)
		}

		if !exists {
			return fmt.Errorf("configuration not found. Please run 'ghissues config' first")
		}

		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if cfg == nil {
			return fmt.Errorf("config file exists but could not be loaded")
		}

		// Get GitHub token
		token, _, err := auth.GetGitHubToken(cfg)
		if err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		// Validate token
		valid, err := auth.ValidateToken(token)
		if err != nil || !valid {
			return fmt.Errorf("token validation failed: %w", err)
		}

		// Determine database path
		dbPath := getDatabasePath(cfg, dbFlag)
		if err := ensureDatabasePath(dbPath); err != nil {
			return fmt.Errorf("database path error: %w", err)
		}

		// Run sync command
		syncConfig := &cmd.SyncConfig{
			Token:      token,
			Repository: cfg.Repository,
			GitHubURL:  "", // Use default GitHub API or override via env var in tests
		}

		// Allow GitHub URL override via environment variable (for testing)
		if testURL := os.Getenv("GHISSUES_GITHUB_URL"); testURL != "" {
			syncConfig.GitHubURL = testURL
		}

		return cmd.RunSyncCommand(dbPath, syncConfig, output)
	}

	// Check if list command was requested or no command (default to list)
	if len(args) == 1 || (len(args) >= 2 && args[1] == "list") {
		// Check if config exists
		exists, err := config.ConfigExists(configPath)
		if err != nil {
			return fmt.Errorf("failed to check config file: %w", err)
		}

		if !exists {
			// First time setup - run interactive configuration
			fmt.Fprintln(output, "Welcome to ghissues!")
			fmt.Fprintln(output, "It looks like this is your first time running ghissues.")
			fmt.Fprintln(output)
			fmt.Fprintln(output, "Let's set up your configuration...")

			if err := cmd.RunConfigCommand(configPath, input, output); err != nil {
				return fmt.Errorf("setup failed: %w", err)
			}

			return nil
		}

		// Config exists, load it
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if cfg == nil {
			return fmt.Errorf("config file exists but could not be loaded")
		}

		// Determine database path
		dbPath := getDatabasePath(cfg, dbFlag)
		if err := ensureDatabasePath(dbPath); err != nil {
			return fmt.Errorf("database path error: %w", err)
		}

		// Run list command (TUI with list + detail views)
		return tui.RunAppView(dbPath, cfg, output)
	}

	// Check if config file exists
	exists, err := config.ConfigExists(configPath)
	if err != nil {
		return fmt.Errorf("failed to check config file: %w", err)
	}

	if !exists {
		// First time setup - run interactive configuration
		fmt.Fprintln(output, "Welcome to ghissues!")
		fmt.Fprintln(output, "It looks like this is your first time running ghissues.")
		fmt.Fprintln(output)
		fmt.Fprintln(output, "Let's set up your configuration...")

		if err := cmd.RunConfigCommand(configPath, input, output); err != nil {
			return fmt.Errorf("setup failed: %w", err)
		}

		return nil
	}

	// Config exists, try to load it
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg == nil {
		return fmt.Errorf("config file exists but could not be loaded")
	}

	// Determine database path with priority: flag > config > default
	dbPath := getDatabasePath(cfg, dbFlag)

	// Ensure database path is valid and parent directories exist
	if err := ensureDatabasePath(dbPath); err != nil {
		return fmt.Errorf("database path error: %w", err)
	}

	// Get GitHub token with proper priority and validation
	token, source, err := auth.GetGitHubToken(cfg)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Validate the token
	valid, err := auth.ValidateToken(token)
	if err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}
	if !valid {
		return fmt.Errorf("token validation failed: token is invalid")
	}

	fmt.Fprintf(output, "Configuration loaded successfully!\n")
	fmt.Fprintf(output, "Repository: %s\n", cfg.Repository)
	fmt.Fprintf(output, "Database: %s\n", dbPath)
	fmt.Fprintf(output, "Authentication: %s token (validated)\n", source)
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Available commands:")
	fmt.Fprintln(output, "  ghissues config  - Configure repository and authentication")
	fmt.Fprintln(output, "  ghissues sync    - Sync issues from GitHub to local database")
	fmt.Fprintln(output, "  ghissues list    - Show TUI with list of issues (default)")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Run 'ghissues' or 'ghissues list' to launch the TUI")

	return nil
}

func getConfigFilePath() string {
	// Check if config path is set via environment variable
	if envPath := os.Getenv("GHISSUES_CONFIG"); envPath != "" {
		return envPath
	}

	// Use default location
	return config.GetDefaultConfigPath()
}

// getDatabasePath determines the database path with priority: flag > config > default
func getDatabasePath(cfg *config.Config, flagPath string) string {
	// Priority 1: Command line flag
	if flagPath != "" {
		return flagPath
	}

	// Priority 2: Config file
	if cfg != nil && cfg.Database.Path != "" {
		return cfg.Database.Path
	}

	// Priority 3: Default location (.ghissues.db in current directory)
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, ".ghissues.db")
}

// ensureDatabasePath ensures the parent directory for the database path exists and is writable
func ensureDatabasePath(dbPath string) error {
	// Get the parent directory
	dir := filepath.Dir(dbPath)

	// Create parent directories if they don't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directories: %w", err)
	}

	// Check if the directory is writable by trying to create a temporary file
	tmpFile := filepath.Join(dir, ".ghissues-write-test")
	f, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("database path is not writable: %w", err)
	}
	f.Close()
	os.Remove(tmpFile)

	return nil
}
