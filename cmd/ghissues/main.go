package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/shepbook/ghissues/internal/auth"
	"github.com/shepbook/ghissues/internal/config"
	"github.com/shepbook/ghissues/internal/db"
	"github.com/shepbook/ghissues/internal/github"
)

var dbFlag string

func init() {
	flag.StringVar(&dbFlag, "db", "", "Path to the database file (default: .ghissues.db in current directory)")
}

func main() {
	flag.Parse()

	// Check for 'config' subcommand
	if len(os.Args) > 1 && os.Args[1] == "config" {
		if err := runSetup(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Check for 'sync' subcommand
	if len(os.Args) > 1 && os.Args[1] == "sync" {
		if err := runSync(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Check for 'themes' subcommand
	if len(os.Args) > 1 && os.Args[1] == "themes" {
		if err := runThemes(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Check for 'tui' subcommand
	if len(os.Args) > 1 && os.Args[1] == "tui" {
		if err := runTUI(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Check if config exists, if not run setup
	if !config.Exists() {
		fmt.Println("Welcome to ghissues! Let's set up your configuration.")
		fmt.Println()

		if err := runSetup(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Determine database path
	dbPath, err := db.GetPath(dbFlag, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Verify database path is writable
	if err := db.IsWritable(dbPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Database path: %s\n", dbPath)

	// Validate authentication
	if err := validateAuth(); err != nil {
		fmt.Fprintf(os.Stderr, "Authentication error: %v\n", err)
		os.Exit(1)
	}

	// Run TUI with auto-refresh
	fmt.Println("Starting TUI with auto-refresh...")
	if err := RunTUIWithRefresh(dbPath, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// validateAuth attempts to get and validate a GitHub token
func validateAuth() error {
	token, source, err := auth.GetToken()
	if err != nil {
		return err
	}

	fmt.Printf("Using GitHub token from %s...\n", source)

	// Validate token by making an API call
	client := github.NewClient(token)
	ctx := context.Background()
	if err := client.ValidateToken(ctx); err != nil {
		return err
	}

	return nil
}

// runTUI runs the TUI application
func runTUI() error {
	// Check if config exists
	if !config.Exists() {
		return fmt.Errorf("configuration not found. Run 'ghissues config' first")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Determine database path
	dbPath, err := db.GetPath(dbFlag, cfg)
	if err != nil {
		return err
	}

	// Run the TUI
	return RunTUI(dbPath, cfg)
}