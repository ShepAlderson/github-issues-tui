package main

import (
	"fmt"
	"os"

	"github.com/shepbook/github-issues-tui/internal/auth"
	"github.com/shepbook/github-issues-tui/internal/config"
	"github.com/shepbook/github-issues-tui/internal/prompt"
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

	// Check command line arguments
	if len(os.Args) > 1 {
		switch os.Args[1] {
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
			return fmt.Errorf("unknown command: %s\n\nRun 'ghissues --help' for usage", os.Args[1])
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

	// TODO: Launch TUI (to be implemented in later user stories)
	fmt.Printf("Configuration loaded successfully!\n")
	fmt.Printf("Repository: %s\n", cfg.GitHub.Repository)
	fmt.Printf("Auth method: %s\n", cfg.GitHub.AuthMethod)
	fmt.Println("Authentication: ✓")
	fmt.Println("\nTUI not yet implemented. Stay tuned!")

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

func printHelp() {
	fmt.Println("ghissues - GitHub Issues TUI")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  ghissues           Start the TUI (runs setup if needed)")
	fmt.Println("  ghissues config    Run/re-run interactive configuration")
	fmt.Println("  ghissues --help    Show this help message")
	fmt.Println("  ghissues --version Show version information")
	fmt.Println()
	fmt.Println("CONFIGURATION:")
	fmt.Printf("  Config file: ~/.config/ghissues/config.toml\n")
	fmt.Println()
	fmt.Println("AUTHENTICATION METHODS:")
	fmt.Println("  env    - Use GITHUB_TOKEN environment variable")
	fmt.Println("  token  - Store token in config file (secure 0600 permissions)")
	fmt.Println("  gh     - Use GitHub CLI (gh) authentication")
}
