package setup

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/shepbook/ghissues/internal/config"
)

// RunSetupWithValues creates and saves a config with the provided values.
// This is primarily used for testing, but can also be used for non-interactive setup.
func RunSetupWithValues(repository, authMethod, token, configPath string) (*config.Config, error) {
	// Validate inputs
	if err := config.ValidateRepository(repository); err != nil {
		return nil, err
	}
	if err := config.ValidateAuthMethod(authMethod); err != nil {
		return nil, err
	}

	// Create config
	cfg := config.New()
	cfg.Repository = repository
	cfg.Auth.Method = authMethod
	if authMethod == "token" {
		cfg.Auth.Token = token
	}

	// Save config
	if err := config.Save(cfg, configPath); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}

	return cfg, nil
}

// RunInteractiveSetup runs the interactive setup prompts and saves the config.
// Returns the created config or an error.
func RunInteractiveSetup(configPath string) (*config.Config, error) {
	var repository string
	var authMethod string
	var token string

	// Create form for repository input
	repoForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("GitHub Repository").
				Description("Enter the repository to track (owner/repo format)").
				Placeholder("owner/repo").
				Value(&repository).
				Validate(func(s string) error {
					return config.ValidateRepository(s)
				}),
		),
	)

	if err := repoForm.Run(); err != nil {
		return nil, fmt.Errorf("repository input cancelled: %w", err)
	}

	// Create form for auth method selection
	authForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Authentication Method").
				Description("Choose how to authenticate with GitHub").
				Options(
					huh.NewOption("Environment Variable (GITHUB_TOKEN)", "env"),
					huh.NewOption("Personal Access Token (stored in config)", "token"),
					huh.NewOption("GitHub CLI (gh auth token)", "gh"),
				).
				Value(&authMethod),
		),
	)

	if err := authForm.Run(); err != nil {
		return nil, fmt.Errorf("auth method selection cancelled: %w", err)
	}

	// If token method selected, prompt for token
	if authMethod == "token" {
		tokenForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("GitHub Personal Access Token").
					Description("Enter your GitHub PAT (will be stored in config file)").
					EchoMode(huh.EchoModePassword).
					Value(&token).
					Validate(func(s string) error {
						if s == "" {
							return fmt.Errorf("token cannot be empty")
						}
						return nil
					}),
			),
		)

		if err := tokenForm.Run(); err != nil {
			return nil, fmt.Errorf("token input cancelled: %w", err)
		}
	}

	// Use the programmatic setup function to create and save the config
	return RunSetupWithValues(repository, authMethod, token, configPath)
}
