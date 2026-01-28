package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/shepbook/ghissues/internal/comments"
	"github.com/shepbook/ghissues/internal/config"
	"github.com/shepbook/ghissues/internal/database"
	"github.com/shepbook/ghissues/internal/github"
	"github.com/shepbook/ghissues/internal/list"
	"github.com/shepbook/ghissues/internal/refresh"
	"github.com/shepbook/ghissues/internal/sync"
)

// ConfigAdapter adapts *config.Config to list.Config interface
type ConfigAdapter struct {
	cfg *config.Config
}

func (a *ConfigAdapter) GetDisplayColumns() []string {
	return a.cfg.Display.Columns
}

func (a *ConfigAdapter) GetDefaultRepository() string {
	return a.cfg.Default.Repository
}

func (a *ConfigAdapter) GetSortField() string {
	return a.cfg.Sort.Field
}

func (a *ConfigAdapter) GetSortDescending() bool {
	return a.cfg.Sort.Descending
}

func (a *ConfigAdapter) SaveSort(field string, descending bool) error {
	a.cfg.Sort.Field = field
	a.cfg.Sort.Descending = descending
	return a.cfg.Save()
}

func main() {
	// Parse global flags
	var dbFlag string
	flag.StringVar(&dbFlag, "db", "", "Database file path (overrides config)")

	// Custom flag parsing to allow subcommands
	flag.CommandLine.SetOutput(os.Stdout)
	flag.Parse()

	// Handle subcommands
	if len(flag.Args()) > 0 {
		switch flag.Args()[0] {
		case "config":
			if err := runConfig(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		case "sync":
			if err := runSync(dbFlag); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		case "version", "-v", "--version":
			fmt.Println("ghissues version 0.1.0")
			os.Exit(0)
		case "help", "-h", "--help":
			printHelp()
			os.Exit(0)
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", flag.Args()[0])
			printHelp()
			os.Exit(1)
		}
	}

	// Check if config exists, run setup if not
	if !config.Exists() {
		fmt.Println("🚀 First-time setup required!")
		fmt.Println()

		cfg, err := config.RunSetup()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Setup cancelled or failed: %v\n", err)
			os.Exit(1)
		}

		// Re-run with the new config
		runListView(cfg, dbFlag)
		return
	}

	// Load existing config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Resolve database path (flag > config > default)
	dbPath := database.ResolvePath(dbFlag, cfg.Database.Path)

	// Ensure database path is writable
	if err := database.EnsureWritable(dbPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nDatabase path: %s\n", dbPath)
		fmt.Fprintf(os.Stderr, "\nYou can override the database location using:\n")
		fmt.Fprintf(os.Stderr, "  --db flag:        ghissues --db /path/to/db\n")
		fmt.Fprintf(os.Stderr, "  Config file:      Set database.path in ~/.config/ghissues/config.toml\n")
		os.Exit(1)
	}

	// Run main application with issue list view
	runListView(cfg, dbPath)
}

func runListView(cfg *config.Config, dbPath string) {
	adapter := &ConfigAdapter{cfg: cfg}

	// Resolve authentication token
	token, err := github.ResolveToken()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Authentication error: %v\n", err)
		os.Exit(1)
	}

	// Check if auto-refresh is needed
	shouldRefresh, err := refresh.ShouldAutoRefresh(dbPath, cfg.Default.Repository)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not check last sync time: %v\n", err)
	}

	if shouldRefresh {
		fmt.Println("🔄 Auto-refreshing issues...")
		result, err := refresh.Perform(refresh.Options{
			Repo:   cfg.Default.Repository,
			DBPath: dbPath,
			Token:  token,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: auto-refresh failed: %v\n", err)
		} else {
			fmt.Printf("✅ Refreshed %d issues and %d comments\n", result.IssuesFetched, result.CommentsFetched)
		}
	}

	// Main loop for switching between list and comments views
	for {
		model := list.NewModel(adapter, dbPath, config.ConfigPath())
		model.SetToken(token)
		p := tea.NewProgram(model)
		result, err := p.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
			os.Exit(1)
		}

		// Check if we should open comments view
		finalModel := result.(list.Model)
		if finalModel.ShouldOpenComments() {
			// Get the selected issue and open comments view
			issueNumber, issueTitle, ok := finalModel.GetSelectedIssueForComments()
			if !ok {
				break
			}

			// Run comments view
			if shouldReturnToList := runCommentsView(dbPath, cfg.Default.Repository, issueNumber, issueTitle); !shouldReturnToList {
				break
			}
			// Loop back to show list view
			continue
		}

		// Check if refresh was requested
		if finalModel.ShouldRefresh() {
			fmt.Println("🔄 Refreshing issues...")
			result, err := refresh.Perform(refresh.Options{
				Repo:   cfg.Default.Repository,
				DBPath: dbPath,
				Token:  token,
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: refresh failed: %v\n", err)
			} else {
				fmt.Printf("✅ Refreshed %d issues and %d comments\n", result.IssuesFetched, result.CommentsFetched)
			}
			// Loop back to show list view with updated data
			continue
		}

		// Normal exit
		break
	}
}

func runCommentsView(dbPath, repo string, issueNumber int, issueTitle string) bool {
	model := comments.NewModel(dbPath, repo, issueNumber, issueTitle)
	p := tea.NewProgram(model)
	_, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running comments view: %v\n", err)
		return false
	}
	// Return true to go back to list view
	return true
}

func runConfig() error {
	fmt.Println("🚀 Re-running first-time setup...")
	fmt.Println()

	_, err := config.RunSetup()
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("✅ Configuration updated successfully!")
	return nil
}

func runSync(dbFlag string) error {
	// Load config to get repository and database path
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get repository
	repo := cfg.Default.Repository
	if repo == "" {
		return fmt.Errorf("no default repository configured. Run 'ghissues config' to set one up")
	}

	// Resolve database path
	dbPath := database.ResolvePath(dbFlag, cfg.Database.Path)

	// Run the sync
	if err := sync.RunSyncCLI(dbPath, repo, ""); err != nil {
		return err
	}

	return nil
}

func printHelp() {
	help := `ghissues - A terminal UI for GitHub issues

Usage:
  ghissues              Run the application (setup if first run)
  ghissues config       Configure repository and authentication
  ghissues sync         Sync issues from configured repository
  ghissues help         Show this help message
  ghissues version      Show version

Global Flags:
  --db <path>           Override database file path (default: .ghissues.db)

Configuration:
  The configuration is stored at ~/.config/ghissues/config.toml
  Database location priority:
    1. --db flag (highest priority)
    2. database.path in config file
    3. .ghissues.db in current directory (default)

First-Time Setup:
  On first run, ghissues will prompt you for:
  1. The GitHub repository (owner/repo format)
  2. Authentication method (environment variable or config file token)

Sync:
  The sync command fetches all open issues from your configured repository.
  Supports Ctrl+C to cancel gracefully. All fetched data is stored locally
  in the SQLite database at the configured path.

Data Refresh:
  - Auto-refresh: App auto-refreshes if data is older than 5 minutes
  - Manual refresh: Press 'r' to refresh issues and comments
  - Incremental: Only fetches issues updated since last sync
  - Deleted issues: Removed from local cache during refresh
  - New comments: Re-fetched during refresh for updated issues

Keybindings (Issue List):
  j, ↓    Move down
  k, ↑    Move up
  s       Cycle sort field (updated → created → number → comments)
  S       Toggle sort order (ascending/descending)
  m       Toggle between rendered and raw markdown
  r       Refresh issues (incremental sync)
  Enter   Open comments view for selected issue
  ?       Show help
  q       Quit
`
	fmt.Println(help)
}
