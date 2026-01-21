package setup

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/shepbook/github-issues-tui/internal/config"
	"github.com/manifoldco/promptui"
)

// Setup runs the interactive setup wizard
func Setup() (*config.Config, error) {
	fmt.Println("Welcome to ghissues setup!")
	fmt.Println("Let's configure your GitHub repository and authentication.")
	fmt.Println()

	// Get repository
	repo, err := promptRepository()
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	// Get authentication method
	authMethod, err := promptAuthMethod()
	if err != nil {
		return nil, fmt.Errorf("failed to get authentication method: %w", err)
	}

	// Create config
	cfg := config.DefaultConfig()
	cfg.Repository = repo
	cfg.Auth.Method = authMethod

	return cfg, nil
}

func promptRepository() (string, error) {
	validate := func(input string) error {
		if input == "" {
			return fmt.Errorf("repository cannot be empty")
		}

		// Validate owner/repo format
		parts := strings.Split(input, "/")
		if len(parts) != 2 {
			return fmt.Errorf("repository must be in owner/repo format")
		}

		owner := parts[0]
		repo := parts[1]

		// Validate owner and repo names (GitHub allows letters, numbers, hyphens)
		ownerRegex := regexp.MustCompile(`^[a-zA-Z0-9\-]+$`)
		repoRegex := regexp.MustCompile(`^[a-zA-Z0-9\-\._]+$`)

		if !ownerRegex.MatchString(owner) {
			return fmt.Errorf("owner name contains invalid characters")
		}

		if !repoRegex.MatchString(repo) {
			return fmt.Errorf("repository name contains invalid characters")
		}

		return nil
	}

	prompt := promptui.Prompt{
		Label:    "GitHub repository (owner/repo)",
		Validate: validate,
		Default:  "",
	}

	result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return result, nil
}

func promptAuthMethod() (string, error) {
	prompt := promptui.Select{
		Label: "Authentication method",
		Items: []string{
			"gh (use GitHub CLI token)",
			"token (store token in config)",
			"env (use GITHUB_TOKEN environment variable)",
		},
	}

	index, _, err := prompt.Run()
	if err != nil {
		return "", err
	}

	// Map index to method value
	methods := []string{"gh", "token", "env"}
	return methods[index], nil
}