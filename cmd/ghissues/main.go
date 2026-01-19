package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shepbook/github-issues-tui/internal/auth"
	"github.com/shepbook/github-issues-tui/internal/config"
	"github.com/shepbook/github-issues-tui/internal/database"
	"github.com/shepbook/github-issues-tui/internal/prompt"
	"github.com/shepbook/github-issues-tui/internal/sync"
	"github.com/shepbook/github-issues-tui/internal/tui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Get config path
	configPath := config.ConfigPath()
	if configPath == "" {
		return fmt.Errorf("failed to determine config path")
	}

	// Parse command line arguments
	var dbPath string
	args := os.Args[1:]

	// Parse flags
	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Handle --db flag
		if arg == "--db" {
			if i+1 >= len(args) {
				return fmt.Errorf("--db flag requires a path argument")
			}
			dbPath = args[i+1]
			// Remove --db and its value from args
			args = append(args[:i], args[i+2:]...)
			i-- // Adjust index after removal
			continue
		}

		// Handle other flags and commands
		switch arg {
		case "sync":
			// Remove sync command from args
			args = append(args[:i], args[i+1:]...)
			return runSync(configPath, dbPath)
		case "config":
			// Force re-run setup
			return runSetup(configPath, true)
		case "--help", "-h":
			printHelp()
			return nil
		case "--version", "-v":
			fmt.Println("ghissues v0.1.0")
			return nil
		default:
			return fmt.Errorf("unknown command: %s\n\nRun 'ghissues --help' for usage", arg)
		}
	}

	// Check if setup is needed
	shouldSetup, err := prompt.ShouldRunSetup(configPath, false)
	if err != nil {
		return fmt.Errorf("failed to check setup status: %w", err)
	}

	if shouldSetup {
		return runSetup(configPath, false)
	}

	// Load existing config
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate config
	if err := config.ValidateConfig(cfg); err != nil {
		return fmt.Errorf("invalid config: %w\n\nRun 'ghissues config' to reconfigure", err)
	}

	// Get authentication token
	token, err := auth.GetToken(cfg)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Validate token with GitHub API
	fmt.Println("Validating GitHub token...")
	if err := auth.ValidateToken(token); err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	// Get database path (flag takes precedence over config)
	resolvedDBPath, err := database.GetDatabasePath(cfg, dbPath)
	if err != nil {
		return fmt.Errorf("failed to resolve database path: %w", err)
	}

	// Initialize database
	if err := database.InitDatabase(resolvedDBPath); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Load issues from database
	store, err := sync.NewIssueStore(resolvedDBPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer store.Close()

	issues, err := store.LoadIssues()
	if err != nil {
		return fmt.Errorf("failed to load issues: %w", err)
	}

	// Get display columns from config
	columns := config.GetDisplayColumns(cfg)

	// Launch TUI
	model := tui.NewModel(issues, columns)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}

func runSetup(configPath string, force bool) error {
	if force {
		fmt.Println("Running configuration setup...")
		fmt.Println()
	}

	// Run interactive setup
	cfg, err := prompt.RunInteractiveSetup()
	if err != nil {
		return fmt.Errorf("setup failed: %w", err)
	}

	// Validate config
	if err := config.ValidateConfig(cfg); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Save config
	if err := config.SaveConfig(cfg, configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("\nConfiguration saved to: %s\n", configPath)
	fmt.Println("\nRun 'ghissues' to start the application.")

	return nil
}

func runSync(configPath string, dbPath string) error {
	// Load config
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w\n\nRun 'ghissues config' to configure", err)
	}

	// Validate config
	if err := config.ValidateConfig(cfg); err != nil {
		return fmt.Errorf("invalid config: %w\n\nRun 'ghissues config' to reconfigure", err)
	}

	// Get authentication token
	token, err := auth.GetToken(cfg)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Validate token with GitHub API
	fmt.Println("Validating GitHub token...")
	if err := auth.ValidateToken(token); err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}
	fmt.Println("Authentication: ✓")
	fmt.Println()

	// Get database path
	resolvedDBPath, err := database.GetDatabasePath(cfg, dbPath)
	if err != nil {
		return fmt.Errorf("failed to resolve database path: %w", err)
	}

	// Initialize database
	if err := database.InitDatabase(resolvedDBPath); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Create syncer
	syncer, err := sync.NewSyncer(token, resolvedDBPath)
	if err != nil {
		return fmt.Errorf("failed to create syncer: %w", err)
	}
	defer syncer.Close()

	// Run sync
	fmt.Printf("Syncing issues from %s...\n\n", cfg.GitHub.Repository)
	if err := syncer.SyncIssues(cfg.GitHub.Repository); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	return nil
}

func printHelp() {
	fmt.Println("ghissues - GitHub Issues TUI")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  ghissues                Start the TUI (runs setup if needed)")
	fmt.Println("  ghissues sync           Fetch and sync all open issues from GitHub")
	fmt.Println("  ghissues --db PATH      Specify database file location")
	fmt.Println("  ghissues config         Run/re-run interactive configuration")
	fmt.Println("  ghissues --help         Show this help message")
	fmt.Println("  ghissues --version      Show version information")
	fmt.Println()
	fmt.Println("CONFIGURATION:")
	fmt.Printf("  Config file: ~/.config/ghissues/config.toml\n")
	fmt.Println()
	fmt.Println("DATABASE:")
	fmt.Println("  Default location: .ghissues.db (in current directory)")
	fmt.Println("  Override via: --db flag or database.path in config file")
	fmt.Println("  Flag takes precedence over config file")
	fmt.Println()
	fmt.Println("AUTHENTICATION METHODS:")
	fmt.Println("  env    - Use GITHUB_TOKEN environment variable")
	fmt.Println("  token  - Store token in config file (secure 0600 permissions)")
	fmt.Println("  gh     - Use GitHub CLI (gh) authentication")
}
