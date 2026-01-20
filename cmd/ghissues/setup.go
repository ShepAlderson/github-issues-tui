package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/shepbook/ghissues/internal/config"
)

// repoPattern validates owner/repo format
var repoPattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$`)

func runSetup() error {
	scanner := bufio.NewScanner(os.Stdin)

	// If config already exists and not explicitly re-running, skip
	if config.Exists() && !isReRun() {
		fmt.Println("Configuration already exists. Run 'ghissues config' to re-configure.")
		return nil
	}

	fmt.Println("=== ghissues Setup ===")
	fmt.Println()

	// Get repository
	repository, err := promptRepository(scanner)
	if err != nil {
		return err
	}

	// Get auth method
	authMethod, err := promptAuthMethod(scanner)
	if err != nil {
		return err
	}

	// Get token if using token auth method
	var token string
	if authMethod == config.AuthMethodToken {
		token, err = promptToken(scanner)
		if err != nil {
			return err
		}
	}

	// Save configuration
	cfg := &config.Config{
		Repository: repository,
		AuthMethod: authMethod,
		Token:      token,
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println()
	fmt.Println("Configuration saved successfully!")
	fmt.Printf("Repository: %s\n", repository)
	fmt.Printf("Auth method: %s\n", authMethod)

	if authMethod == config.AuthMethodEnv {
		fmt.Println()
		fmt.Println("Tip: Make sure to set the GITHUB_TOKEN environment variable.")
	} else if authMethod == config.AuthMethodToken {
		fmt.Println("Token has been securely saved.")
	}

	return nil
}

func isReRun() bool {
	// Check if 'config' command was explicitly invoked
	if len(os.Args) < 2 {
		return false
	}
	for _, arg := range os.Args[1:] {
		if arg == "config" || arg == "reconfig" {
			return true
		}
	}
	return false
}

func promptRepository(scanner *bufio.Scanner) (string, error) {
	fmt.Print("GitHub repository (owner/repo): ")

	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}
		return "", fmt.Errorf("no input received")
	}

	repo := strings.TrimSpace(scanner.Text())
	if repo == "" {
		return "", fmt.Errorf("repository cannot be empty")
	}

	if !repoPattern.MatchString(repo) {
		return "", fmt.Errorf("invalid repository format. Use owner/repo format (e.g., 'anthropics/claude-code')")
	}

	return repo, nil
}

func promptAuthMethod(scanner *bufio.Scanner) (config.AuthMethod, error) {
	fmt.Println("Authentication method:")
	fmt.Println("  1. Environment variable (GITHUB_TOKEN)")
	fmt.Println("  2. GitHub CLI (gh auth token)")
	fmt.Println("  3. Personal access token (saved to config)")

	fmt.Print("Choose [1-3]: ")

	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}
		return "", fmt.Errorf("no input received")
	}

	choice := strings.TrimSpace(scanner.Text())

	switch choice {
	case "1":
		return config.AuthMethodEnv, nil
	case "2":
		return config.AuthMethodGhCli, nil
	case "3":
		return config.AuthMethodToken, nil
	default:
		return "", fmt.Errorf("invalid choice: %s. Please enter 1, 2, or 3", choice)
	}
}

func promptToken(scanner *bufio.Scanner) (string, error) {
	fmt.Println()
	fmt.Println("Personal Access Token:")
	fmt.Println("  - Create at: https://github.com/settings/tokens")
	fmt.Println("  - Required scopes: repo (for private repos)")

	fmt.Print("Enter your GitHub personal access token: ")

	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}
		return "", fmt.Errorf("no input received")
	}

	token := strings.TrimSpace(scanner.Text())
	if token == "" {
		return "", fmt.Errorf("token cannot be empty")
	}

	// Basic validation - GitHub tokens are typically 40+ characters (classic) or variable length (fine-grained)
	if len(token) < 10 {
		return "", fmt.Errorf("token appears too short. Please enter a valid GitHub personal access token")
	}

	return token, nil
}
