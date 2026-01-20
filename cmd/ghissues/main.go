package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shepbook/ghissues/internal/auth"
	"github.com/shepbook/ghissues/internal/config"
	"github.com/shepbook/ghissues/internal/database"
	"github.com/shepbook/ghissues/internal/github"
	"github.com/shepbook/ghissues/internal/setup"
	"github.com/shepbook/ghissues/internal/storage"
	"github.com/shepbook/ghissues/internal/sync"
	"github.com/shepbook/ghissues/internal/tui"
)

const version = "0.1.0"

func main() {
	// Define flags
	var (
		showVersion bool
		showHelp    bool
		configPath  string
		repoFlag    string
		dbFlag      string
	)

	flag.BoolVar(&showVersion, "version", false, "show version information")
	flag.BoolVar(&showHelp, "help", false, "show help information")
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.StringVar(&repoFlag, "repo", "", "repository to use (owner/repo)")
	flag.StringVar(&dbFlag, "db", "", "path to database file")

	flag.Parse()

	// Handle version flag
	if showVersion {
		fmt.Printf("ghissues version %s\n", version)
		os.Exit(0)
	}

	// Handle help flag
	if showHelp {
		printHelp()
		os.Exit(0)
	}

	// Determine config path
	if configPath == "" {
		configPath = config.GetDefaultConfigPath()
	}

	// Get command arguments
	args := flag.Args()
	command := ""
	if len(args) > 0 {
		command = args[0]
	}

	// Handle config command
	if command == "config" {
		handleConfigCommand(configPath)
		return
	}

	// Check if config exists
	if !config.ConfigExists(configPath) {
		// Run first-time setup
		prompter := setup.NewPrompter()
		err := prompter.RunSetup(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Setup failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Load config
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		fmt.Fprintln(os.Stderr, "Run 'ghissues config' to reconfigure.")
		os.Exit(1)
	}

	// Resolve database path
	dbPath, err := database.ResolveDatabasePath(dbFlag, cfg.Database.Path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to resolve database path: %v\n", err)
		os.Exit(1)
	}

	// Ensure database directory exists
	if err := database.EnsureDatabasePath(dbPath); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create database directory: %v\n", err)
		os.Exit(1)
	}

	// Check if database location is writable
	if err := database.CheckDatabaseWritable(dbPath); err != nil {
		fmt.Fprintf(os.Stderr, "Database path not writable: %v\n", err)
		os.Exit(1)
	}

	// Get authentication token
	token, _, err := auth.GetToken(configPath, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get authentication: %v\n", err)
		os.Exit(1)
	}

	// Initialize database
	db, err := storage.InitializeDatabase(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Determine which repository to use
	repo := cfg.Default.Repository
	if repoFlag != "" {
		if err := github.ValidateRepo(repoFlag); err != nil {
			fmt.Fprintf(os.Stderr, "Invalid repository: %v\n", err)
			os.Exit(1)
		}
		repo = repoFlag
	} else {
		// Validate the configured repository
		if err := github.ValidateRepo(repo); err != nil {
			fmt.Fprintf(os.Stderr, "Invalid repository in config: %v\n", err)
			fmt.Fprintln(os.Stderr, "Run 'ghissues config' to reconfigure.")
			os.Exit(1)
		}
	}

	// Create GitHub client
	ghClient := github.NewClient(token, repo, "")

	// Create syncer
	syncer := sync.NewSyncer(db, ghClient, repo, false)

	// Handle sync command
	if command == "sync" {
		handleSyncCommand(syncer)
		return
	}

	// Auto-sync on first run or if database is empty
	issues, err := storage.GetIssues(db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to check database: %v\n", err)
		os.Exit(1)
	}

	if len(issues) == 0 {
		fmt.Println("No issues found. Running initial sync...")
		handleSyncCommand(syncer)

		// Reload issues after sync
		issues, err = storage.GetIssues(db)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to reload issues: %v\n", err)
			os.Exit(1)
		}
	}

	// Get column configuration
	columns := tui.GetDefaultColumns(cfg)

	// Get sort configuration from config file
	sortField := cfg.Sort.Field
	sortDescending := cfg.Sort.Descending

	// Use defaults if not configured
	if sortField == "" {
		sortField = "updated" // Default sort field
		sortDescending = true // Default to descending (most recent first)
	}

	// Create and start TUI with sort configuration
	model := tui.NewModelWithSort(issues, columns, sortField, sortDescending)
	program := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

func handleSyncCommand(syncer *sync.Syncer) {
	fmt.Println("Syncing issues from GitHub...")
	result, err := syncer.Run(context.Background())
	if err != nil && !result.Cancelled {
		fmt.Fprintf(os.Stderr, "\nSync failed: %v\n", err)
		os.Exit(1)
	}

	if result.Cancelled {
		fmt.Println("Sync cancelled.")
		return
	}

	fmt.Printf("\n✓ Sync completed in %v\n", result.Duration)
	fmt.Printf("  Issues stored: %d\n", result.IssuesStored)
	fmt.Printf("  Comments fetched: %d\n", result.CommentsFetched)
	if len(result.Errors) > 0 {
		fmt.Printf("  Errors: %d\n", len(result.Errors))
		for _, e := range result.Errors {
			fmt.Fprintf(os.Stderr, "    - %v\n", e)
		}
	}
}

func handleConfigCommand(configPath string) {
	prompter := setup.NewPrompter()
	err := prompter.RunSetup(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration failed: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Printf("ghissues - A terminal-based GitHub issues viewer\n\n")
	fmt.Printf("Usage:\n")
	fmt.Printf("  ghissues [flags] [command]\n\n")
	fmt.Printf("Flags:\n")
	fmt.Printf("  --config string   Path to config file (default: ~/.config/ghissues/config.toml)\n")
	fmt.Printf("  --db string       Path to database file (default: .ghissues.db)\n")
	fmt.Printf("  --repo string     Repository to use (owner/repo)\n")
	fmt.Printf("  --help            Show help information\n")
	fmt.Printf("  --version         Show version information\n\n")
	fmt.Printf("Commands:\n")
	fmt.Printf("  config            Run interactive configuration\n")
	fmt.Printf("  sync              Sync issues from GitHub (also runs automatically on first use)\n")
}
