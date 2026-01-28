package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/shepbook/ghissues/internal/comments"
	"github.com/shepbook/ghissues/internal/config"
	"github.com/shepbook/ghissues/internal/database"
	apperror "github.com/shepbook/ghissues/internal/error"
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

func (a *ConfigAdapter) GetTheme() string {
	return a.cfg.Display.Theme
}

func (a *ConfigAdapter) GetRepositories() []list.RepositoryInfo {
	var repos []list.RepositoryInfo
	for _, r := range a.cfg.Repositories {
		repos = append(repos, list.RepositoryInfo{
			Owner:    r.Owner,
			Name:     r.Name,
			FullName: r.Owner + "/" + r.Name,
		})
	}
	return repos
}

func (a *ConfigAdapter) GetRepositoryDatabase(repo string) string {
	for _, r := range a.cfg.Repositories {
		if r.Owner+"/"+r.Name == repo {
			return r.Database
		}
	}
	return ""
}

func main() {
	// Parse global flags
	var dbFlag string
	var repoFlag string
	flag.StringVar(&dbFlag, "db", "", "Database file path (overrides config)")
	flag.StringVar(&repoFlag, "repo", "", "Repository to view (owner/repo format, overrides default)")

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
		case "themes":
			if err := runThemes(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		case "sync":
			if err := runSync(dbFlag, repoFlag); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		case "repos":
			if err := runRepos(); err != nil {
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
		runListView(cfg, dbFlag, "")
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
	runListView(cfg, dbPath, repoFlag)
}

func runListView(cfg *config.Config, dbPath string, overrideRepo string) {
	adapter := &ConfigAdapter{cfg: cfg}

	// Determine which repository to use
	targetRepo := cfg.Default.Repository
	if overrideRepo != "" {
		// Validate the override repo format
		if err := config.ValidateRepository(overrideRepo); err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid repository format: %v\n", err)
			os.Exit(1)
		}
		targetRepo = overrideRepo
	}

	if targetRepo == "" {
		fmt.Fprintf(os.Stderr, "Error: no repository configured. Run 'ghissues config' to set up.\n")
		os.Exit(1)
	}

	// Check if the repository is configured
	var repoDBPath string
	if overrideRepo != "" {
		// For override repo, check if it exists in config or use default database
		repoDBPath = findRepoDatabase(cfg, targetRepo)
		if repoDBPath == "" {
			repoDBPath = dbPath
		}
	} else {
		repoDBPath = dbPath
	}

	// Resolve authentication token
	token, err := github.ResolveToken()
	if err != nil {
		// Show critical error modal for authentication issues
		appErr := apperror.Classify(err)
		if appErr.Severity.IsCritical() {
			runErrorModal(appErr)
		}
		fmt.Fprintf(os.Stderr, "Authentication error: %v\n", err)
		os.Exit(1)
	}

	// Check if auto-refresh is needed
	shouldRefresh, err := refresh.ShouldAutoRefresh(repoDBPath, targetRepo)
	if err != nil {
		// Classify and handle error appropriately
		appErr := apperror.Classify(err)
		if appErr.Severity.IsCritical() {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "Warning: could not check last sync time: %v\n", err)
		}
	}

	if shouldRefresh {
		fmt.Println("🔄 Auto-refreshing issues...")
		result, err := refresh.Perform(refresh.Options{
			Repo:   targetRepo,
			DBPath: repoDBPath,
			Token:  token,
		})
		if err != nil {
			// Classify the error
			appErr := apperror.Classify(err)
			if appErr.Severity.IsCritical() {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "Warning: auto-refresh failed: %v\n", err)
			}
		} else {
			fmt.Printf("✅ Refreshed %d issues and %d comments\n", result.IssuesFetched, result.CommentsFetched)
		}
	}

	// Main loop for switching between list and comments views
	for {
		model := list.NewModel(adapter, repoDBPath, config.ConfigPath())
		model.SetRepository(targetRepo)
		model.SetToken(token)
		p := tea.NewProgram(model)
		result, err := p.Run()
		if err != nil {
			// Classify and handle error
			appErr := apperror.Classify(err)
			if appErr.Severity.IsCritical() {
				runErrorModal(appErr)
			}
			fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
			os.Exit(1)
		}

		// Check if we should open comments view
		finalModel := result.(list.Model)

		// Check for critical error that needs modal display
		if finalModel.HasCriticalError() {
			errInfo := finalModel.GetCriticalError()
			if errInfo != nil {
				runErrorModal(*errInfo)
			}
			// Continue to show list view after acknowledgment
			continue
		}

		if finalModel.ShouldOpenComments() {
			// Get the selected issue and open comments view
			issueNumber, issueTitle, ok := finalModel.GetSelectedIssueForComments()
			if !ok {
				break
			}

			// Run comments view
			if shouldReturnToList := runCommentsView(repoDBPath, targetRepo, issueNumber, issueTitle); !shouldReturnToList {
				break
			}
			// Loop back to show list view
			continue
		}

		// Check if refresh was requested
		if finalModel.ShouldRefresh() {
			fmt.Println("🔄 Refreshing issues...")
			result, err := refresh.Perform(refresh.Options{
				Repo:   targetRepo,
				DBPath: repoDBPath,
				Token:  token,
			})
			if err != nil {
				// Classify the error
				appErr := apperror.Classify(err)
				if appErr.Severity.IsCritical() {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				} else {
					fmt.Fprintf(os.Stderr, "Warning: refresh failed: %v\n", err)
				}
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

// findRepoDatabase finds the database path for a specific repository in the config
func findRepoDatabase(cfg *config.Config, repo string) string {
	for _, r := range cfg.Repositories {
		if r.Owner+"/"+r.Name == repo {
			return r.Database
		}
	}
	return ""
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

func runErrorModal(appErr apperror.AppError) {
	// Create and run the error modal
	model := apperror.RunModal(appErr)
	p := tea.NewProgram(model)
	_, err := p.Run()
	if err != nil {
		// If modal fails, just log the error to stderr
		fmt.Fprintf(os.Stderr, "Error: %s\n", appErr.Display)
		if appErr.Guidance != "" {
			fmt.Fprintf(os.Stderr, "%s\n", appErr.Guidance)
		}
	}
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

func runThemes() error {
	// Load current config to get the current theme
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	currentTheme := cfg.Display.Theme
	if currentTheme == "" {
		currentTheme = "default"
	}

	// Run the theme picker
	selectedTheme, saved, err := config.RunThemePicker(currentTheme)
	if err != nil {
		return fmt.Errorf("failed to run theme picker: %w", err)
	}

	if !saved {
		// User cancelled
		return nil
	}

	// Save the selected theme to config
	if err := config.SaveThemeToConfig(selectedTheme); err != nil {
		return fmt.Errorf("failed to save theme: %w", err)
	}

	fmt.Printf("Theme set to: %s\n", selectedTheme)
	return nil
}

func runSync(dbFlag string, repoFlag string) error {
	// Load config to get repository and database path
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Determine which repository to sync
	targetRepo := cfg.Default.Repository
	if repoFlag != "" {
		targetRepo = repoFlag
	}

	if targetRepo == "" {
		return fmt.Errorf("no repository specified. Use --repo flag or set a default repository with 'ghissues config'")
	}

	// Validate the repository format
	if err := config.ValidateRepository(targetRepo); err != nil {
		return fmt.Errorf("invalid repository format: %w", err)
	}

	// Find the database path for this repository, or use default
	dbPath := database.ResolvePath(dbFlag, cfg.Database.Path)
	if repoPath := findRepoDatabase(cfg, targetRepo); repoPath != "" {
		// If a per-repo database is configured, use it
		// but still allow the --db flag to override
		if dbFlag == "" {
			dbPath = database.ResolvePath(repoPath, cfg.Database.Path)
		}
	}

	// Run the sync
	if err := sync.RunSyncCLI(dbPath, targetRepo, ""); err != nil {
		return err
	}

	return nil
}

func runRepos() error {
	// Load config to get repositories
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Configured Repositories:")
	fmt.Println()

	if len(cfg.Repositories) == 0 {
		fmt.Println("  No repositories configured.")
		fmt.Println()
		fmt.Println("  Run 'ghissues config' to add your first repository.")
		return nil
	}

	// Find the default repository
	defaultRepo := cfg.Default.Repository

	for _, repo := range cfg.Repositories {
		repoName := repo.Owner + "/" + repo.Name
		marker := "  "
		if repoName == defaultRepo {
			marker = "* "
		}
		fmt.Printf("%s%s\n", marker, repoName)
		if repo.Database != "" && repo.Database != cfg.Database.Path {
			fmt.Printf("    Database: %s\n", repo.Database)
		}
	}

	fmt.Println()
	fmt.Println("* = default repository")
	fmt.Println()
	fmt.Printf("Total: %d repositories\n", len(cfg.Repositories))

	return nil
}

func printHelp() {
	help := `ghissues - A terminal UI for GitHub issues

Usage:
  ghissues                    Run the application (setup if first run)
  ghissues config             Configure repository and authentication
  ghissues themes             Preview and change color theme
  ghissues sync               Sync issues from configured repository
  ghissues repos              List configured repositories
  ghissues help               Show this help message
  ghissues version            Show version

Global Flags:
  --db <path>                 Override database file path (default: .ghissues.db)
  --repo <owner/repo>         Override repository to view (e.g., ghissues --repo owner/repo)

Configuration:
  The configuration is stored at ~/.config/ghissues/config.toml
  Database location priority:
    1. --db flag (highest priority)
    2. Per-repo database path in repositories[].database
    3. database.path in config file
    4. .ghissues.db in current directory (default)

Multiple Repositories:
  Configure multiple repositories in config to view issues from different projects:
  - Set default: [default] section in config
  - Override temporarily: use --repo flag
  - Each repository can have its own database file

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
