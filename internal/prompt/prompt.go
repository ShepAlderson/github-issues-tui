package prompt

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/shepbook/github-issues-tui/internal/config"
)

// ParseRepositoryInput validates repository input format
func ParseRepositoryInput(input string) error {
	input = strings.TrimSpace(input)
	if input == "" {
		return errors.New("repository cannot be empty")
	}

	parts := strings.Split(input, "/")
	if len(parts) != 2 {
		return errors.New("repository must be in owner/repo format")
	}

	if strings.TrimSpace(parts[0]) == "" {
		return errors.New("repository owner cannot be empty")
	}

	if strings.TrimSpace(parts[1]) == "" {
		return errors.New("repository name cannot be empty")
	}

	return nil
}

// ParseAuthMethodInput validates and normalizes auth method input
func ParseAuthMethodInput(input string) (string, error) {
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return "", errors.New("auth method cannot be empty")
	}

	validMethods := map[string]bool{
		"env":   true,
		"token": true,
		"gh":    true,
	}

	if !validMethods[input] {
		return "", errors.New("invalid auth method: must be one of: env, token, gh")
	}

	return input, nil
}

// ParseTokenInput validates token input
func ParseTokenInput(input string) error {
	input = strings.TrimSpace(input)
	if input == "" {
		return errors.New("token cannot be empty")
	}

	return nil
}

// RunInteractiveSetup prompts the user for configuration values
func RunInteractiveSetup() (*config.Config, error) {
	return RunInteractiveSetupWithReader(os.Stdin)
}

// RunInteractiveSetupWithReader allows dependency injection for testing
func RunInteractiveSetupWithReader(reader io.Reader) (*config.Config, error) {
	scanner := bufio.NewScanner(reader)
	cfg := &config.Config{}

	fmt.Println("Welcome to ghissues setup!")
	fmt.Println()

	// Prompt for repository
	var repository string
	for {
		fmt.Print("Enter GitHub repository (owner/repo): ")
		if !scanner.Scan() {
			return nil, errors.New("failed to read repository input")
		}
		repository = scanner.Text()

		if err := ParseRepositoryInput(repository); err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		break
	}
	cfg.GitHub.Repository = strings.TrimSpace(repository)

	// Prompt for auth method
	fmt.Println()
	fmt.Println("Select authentication method:")
	fmt.Println("  env   - Use GITHUB_TOKEN environment variable")
	fmt.Println("  token - Store token in config file")
	fmt.Println("  gh    - Use GitHub CLI (gh) authentication")
	fmt.Println()

	var authMethod string
	for {
		fmt.Print("Auth method [env/token/gh]: ")
		if !scanner.Scan() {
			return nil, errors.New("failed to read auth method input")
		}
		input := scanner.Text()

		var err error
		authMethod, err = ParseAuthMethodInput(input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		break
	}
	cfg.GitHub.AuthMethod = authMethod

	// Prompt for token if auth method is token
	if authMethod == "token" {
		fmt.Println()
		var token string
		for {
			fmt.Print("Enter GitHub personal access token: ")
			if !scanner.Scan() {
				return nil, errors.New("failed to read token input")
			}
			token = scanner.Text()

			if err := ParseTokenInput(token); err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			break
		}
		cfg.GitHub.Token = strings.TrimSpace(token)
	}

	fmt.Println()
	fmt.Println("Configuration complete!")

	return cfg, nil
}

// ShouldRunSetup determines if interactive setup should run
func ShouldRunSetup(configPath string, force bool) (bool, error) {
	// If force flag is set, always run setup
	if force {
		return true, nil
	}

	// Check if config exists
	exists, err := config.ConfigExists(configPath)
	if err != nil {
		return false, err
	}

	// Run setup if config doesn't exist
	return !exists, nil
}
