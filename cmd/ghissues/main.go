package main

import (
	"context"
	"fmt"
	"os"

	"github.com/shepbook/ghissues/internal/auth"
	"github.com/shepbook/ghissues/internal/config"
	"github.com/shepbook/ghissues/internal/github"
)

func main() {
	// Check for 'config' subcommand
	if len(os.Args) > 1 && os.Args[1] == "config" {
		if err := runSetup(); err != nil {
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

	// Validate authentication
	if err := validateAuth(); err != nil {
		fmt.Fprintf(os.Stderr, "Authentication error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Configuration complete. Run 'ghissues' to start the app.")
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
