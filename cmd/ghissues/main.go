package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shepbook/github-issues-tui/internal/auth"
	"github.com/shepbook/github-issues-tui/internal/config"
	"github.com/shepbook/github-issues-tui/internal/database"
	"github.com/shepbook/github-issues-tui/internal/prompt"
	"github.com/shepbook/github-issues-tui/internal/sync"
	"github.com/shepbook/github-issues-tui/internal/theme"
	"github.com/shepbook/github-issues-tui/internal/tui"
	"golang.org/x/term"
)

// checkTerminalCapabilities verifies that stdin is a terminal
// and returns an error with a clear message if it's not
func checkTerminalCapabilities() error {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return fmt.Errorf("stdin is not a terminal. TUI applications require an interactive terminal environment.\n\nPlease run this command in a terminal (not redirected from a file or pipe).")
	}
	return nil
}

// getTUIOptions returns bubbletea program options based on environment variables
// This allows testing different initialization flags for debugging purposes
func getTUIOptions() []tea.ProgramOption {
	// Always include stdin input configuration (critical for keyboard input)
	options := []tea.ProgramOption{
		tea.WithInput(os.Stdin),
	}

	// Parse GHISSUES_TUI_OPTIONS environment variable for additional flags
	tuiOpts := os.Getenv("GHISSUES_TUI_OPTIONS")
	if tuiOpts == "" {
		// Default: use alt screen
		options = append(options, tea.WithAltScreen())
		return options
	}

	// Parse comma-separated options
	flags := strings.Split(tuiOpts, ",")
	useAltScreen := true

	for _, flag := range flags {
		flag = strings.TrimSpace(flag)
		switch flag {
		case "mouse":
			// Enable mouse support (not added by default)
			options = append(options, tea.WithMouseCellMotion())
		case "noaltscreen":
			// Disable alt screen (useful for debugging - output stays in terminal)
			useAltScreen = false
		}
	}

	if useAltScreen {
		options = append(options, tea.WithAltScreen())
	}

	return options
}

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
	var repoFlag string
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

		// Handle --repo flag
		if arg == "--repo" {
			if i+1 >= len(args) {
				return fmt.Errorf("--repo flag requires a repository argument (owner/repo format)")
			}
			repoFlag = args[i+1]
			// Remove --repo and its value from args
			args = append(args[:i], args[i+2:]...)
			i-- // Adjust index after removal
			continue
		}

		// Handle other flags and commands
		switch arg {
		case "sync":
			// Remove sync command from args
			args = append(args[:i], args[i+1:]...)
			return runSync(configPath, dbPath, repoFlag)
		case "refresh":
			// Remove refresh command from args
			args = append(args[:i], args[i+1:]...)
			return runRefresh(configPath, dbPath, repoFlag)
		case "repos":
			// List configured repositories
			return runRepos(configPath)
		case "config":
			// Force re-run setup
			return runSetup(configPath, true)
		case "themes":
			// List available themes
			return runThemes()
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

	// Resolve which repository to use
	repository, err := config.GetRepository(cfg, repoFlag)
	if err != nil {
		return fmt.Errorf("failed to resolve repository: %w", err)
	}

	// Get database path (use per-repo path unless --db flag is set)
	var resolvedDBPath string
	if dbPath != "" {
		resolvedDBPath, err = database.GetDatabasePath(cfg, dbPath)
		if err != nil {
			return fmt.Errorf("failed to resolve database path: %w", err)
		}
	} else {
		// Use per-repository database path
		resolvedDBPath = config.GetDatabasePathForRepository(repository)
	}

	// Initialize database
	if err := database.InitDatabase(resolvedDBPath); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Auto-refresh: Perform incremental sync on app launch
	fmt.Printf("Checking for updates (%s)...\n", repository)
	syncer, err := sync.NewSyncer(token, resolvedDBPath)
	if err != nil {
		return fmt.Errorf("failed to create syncer: %w", err)
	}

	if err := syncer.RefreshIssues(repository); err != nil {
		fmt.Printf("Warning: Auto-refresh failed: %v\n", err)
		fmt.Println("Continuing with cached data...")
	}
	syncer.Close()

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

	// Get sort preferences from config
	sortBy := config.GetSortBy(cfg)
	sortAscending := config.GetSortAscending(cfg)

	// Get theme from config
	themeName := config.GetTheme(cfg)

	// Get last sync time
	lastSyncTime, err := store.GetLastSyncTime()
	if err != nil {
		return fmt.Errorf("failed to get last sync time: %w", err)
	}

	// Check terminal capabilities before launching TUI
	if err := checkTerminalCapabilities(); err != nil {
		return err
	}

	// Launch TUI
	model := tui.NewModel(issues, columns, sortBy, sortAscending, store, lastSyncTime, themeName)
	tuiOptions := getTUIOptions()
	p := tea.NewProgram(model, tuiOptions...)
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

func runSync(configPath string, dbPath string, repoFlag string) error {
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

	// Resolve which repository to use
	repository, err := config.GetRepository(cfg, repoFlag)
	if err != nil {
		return fmt.Errorf("failed to resolve repository: %w", err)
	}

	// Get database path (use per-repo path unless --db flag is set)
	var resolvedDBPath string
	if dbPath != "" {
		resolvedDBPath, err = database.GetDatabasePath(cfg, dbPath)
		if err != nil {
			return fmt.Errorf("failed to resolve database path: %w", err)
		}
	} else {
		// Use per-repository database path
		resolvedDBPath = config.GetDatabasePathForRepository(repository)
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
	fmt.Printf("Syncing issues from %s...\n\n", repository)
	if err := syncer.SyncIssues(repository); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	return nil
}

func runRefresh(configPath string, dbPath string, repoFlag string) error {
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

	// Resolve which repository to use
	repository, err := config.GetRepository(cfg, repoFlag)
	if err != nil {
		return fmt.Errorf("failed to resolve repository: %w", err)
	}

	// Get database path (use per-repo path unless --db flag is set)
	var resolvedDBPath string
	if dbPath != "" {
		resolvedDBPath, err = database.GetDatabasePath(cfg, dbPath)
		if err != nil {
			return fmt.Errorf("failed to resolve database path: %w", err)
		}
	} else {
		// Use per-repository database path
		resolvedDBPath = config.GetDatabasePathForRepository(repository)
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

	// Run refresh
	fmt.Printf("Refreshing issues from %s...\n\n", repository)
	if err := syncer.RefreshIssues(repository); err != nil {
		return fmt.Errorf("refresh failed: %w", err)
	}

	return nil
}

func runRepos(configPath string) error {
	// Load config
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w\n\nRun 'ghissues config' to configure", err)
	}

	// Get list of configured repositories
	repos := config.ListRepositories(cfg)

	if len(repos) == 0 {
		fmt.Println("No repositories configured.")
		fmt.Println()
		fmt.Println("Run 'ghissues config' to configure repositories.")
		return nil
	}

	fmt.Println("Configured Repositories")
	fmt.Println("=======================")
	fmt.Println()

	// Show default repository if set
	defaultRepo := ""
	if cfg.GitHub.DefaultRepository != "" {
		defaultRepo = cfg.GitHub.DefaultRepository
	} else if len(repos) > 0 {
		defaultRepo = repos[0]
	}

	for _, repo := range repos {
		if repo == defaultRepo {
			fmt.Printf("  %s (default)\n", repo)
		} else {
			fmt.Printf("  %s\n", repo)
		}
	}

	fmt.Println()
	fmt.Println("Use --repo flag to select a specific repository:")
	fmt.Println("  ghissues --repo owner/repo")
	fmt.Println()

	return nil
}

func runThemes() error {
	fmt.Println("Available Themes")
	fmt.Println("================")
	fmt.Println()

	themes := theme.ListThemes()
	for _, t := range themes {
		// Show theme name with a sample of its colors
		fmt.Printf("  %s\n", t.Name)

		// Show sample colored text using the theme
		fmt.Print("    ")
		fmt.Print(t.HeaderStyle.Render("Header"))
		fmt.Print("  ")
		fmt.Print(t.SelectedStyle.Render("Selected"))
		fmt.Print("  ")
		fmt.Print(t.DetailTitleStyle.Render("Title"))
		fmt.Print("  ")
		fmt.Print(t.ErrorStyle.Render("Error"))
		fmt.Println()
	}

	fmt.Println()
	fmt.Println("To use a theme, set display.theme in ~/.config/ghissues/config.toml")
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println("  [display]")
	fmt.Println("  theme = \"dracula\"")
	fmt.Println()

	return nil
}

func printHelp() {
	fmt.Println("ghissues - GitHub Issues TUI")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  ghissues                   Start the TUI (auto-refreshes on launch)")
	fmt.Println("  ghissues sync              Fetch and sync all open issues from GitHub (full sync)")
	fmt.Println("  ghissues refresh           Update issues changed since last sync (incremental)")
	fmt.Println("  ghissues repos             List configured repositories")
	fmt.Println("  ghissues themes            List available color themes")
	fmt.Println("  ghissues --repo owner/repo Select which repository to use")
	fmt.Println("  ghissues --db PATH         Specify database file location")
	fmt.Println("  ghissues config            Run/re-run interactive configuration")
	fmt.Println("  ghissues --help            Show this help message")
	fmt.Println("  ghissues --version         Show version information")
	fmt.Println()
	fmt.Println("CONFIGURATION:")
	fmt.Printf("  Config file: ~/.config/ghissues/config.toml\n")
	fmt.Println()
	fmt.Println("MULTI-REPOSITORY SUPPORT:")
	fmt.Println("  Each repository has its own database file")
	fmt.Println("  Database location: ~/.local/share/ghissues/<owner_repo>.db")
	fmt.Println("  Set default_repository in config to specify which repo to use by default")
	fmt.Println("  Use --repo flag to override and select a specific repository")
	fmt.Println()
	fmt.Println("DATABASE:")
	fmt.Println("  Per-repository databases stored in: ~/.local/share/ghissues/")
	fmt.Println("  Override via: --db flag (applies to selected repository)")
	fmt.Println()
	fmt.Println("AUTHENTICATION METHODS:")
	fmt.Println("  env    - Use GITHUB_TOKEN environment variable")
	fmt.Println("  token  - Store token in config file (secure 0600 permissions)")
	fmt.Println("  gh     - Use GitHub CLI (gh) authentication")
}
