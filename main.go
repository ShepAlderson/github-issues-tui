package main

import (
	"fmt"
	"io"
	"os"

	"github.com/shepbook/git/github-issues-tui/internal/auth"
	"github.com/shepbook/git/github-issues-tui/internal/cmd"
	"github.com/shepbook/git/github-issues-tui/internal/config"
)

func main() {
	if err := runMain(os.Args, getConfigFilePath(), os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runMain(args []string, configPath string, input io.Reader, output io.Writer) error {
	// Check number of arguments
	if len(args) > 2 {
		return fmt.Errorf("too many arguments")
	}

	// Check if config command was requested
	if len(args) == 2 && args[1] == "config" {
		return cmd.RunConfigCommand(configPath, input, output)
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

	// TODO: In future stories, we'll launch the TUI here
	fmt.Fprintf(output, "Configuration loaded successfully!\n")
	fmt.Fprintf(output, "Repository: %s\n", cfg.Repository)
	fmt.Fprintf(output, "Authentication: %s token (validated)\n", source)
	fmt.Fprintln(output)
	fmt.Fprintln(output, "TUI implementation coming in future stories...")

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
