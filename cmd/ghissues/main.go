package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/shepbook/ghissues/internal/config"
	"github.com/shepbook/ghissues/internal/database"
	"github.com/shepbook/ghissues/internal/setup"
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

	// TODO: Implement main TUI application
	fmt.Printf("Configuration loaded for repository: %s\n", cfg.Default.Repository)
	fmt.Printf("Database path: %s\n", dbPath)
	fmt.Println("TUI application not yet implemented.")
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
	fmt.Printf("  ghissues [flags]\n")
	fmt.Printf("  ghissues config\n\n")
	fmt.Printf("Flags:\n")
	fmt.Printf("  --config string   Path to config file (default: ~/.config/ghissues/config.toml)\n")
	fmt.Printf("  --db string       Path to database file (default: .ghissues.db)\n")
	fmt.Printf("  --repo string     Repository to use (owner/repo)\n")
	fmt.Printf("  --help            Show help information\n")
	fmt.Printf("  --version         Show version information\n\n")
	fmt.Printf("Commands:\n")
	fmt.Printf("  config            Run interactive configuration\n")
}
