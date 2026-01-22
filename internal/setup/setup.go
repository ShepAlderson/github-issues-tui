package setup

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/shepbook/ghissues/internal/config"
)

// Prompter handles interactive setup prompts
type Prompter struct {
	reader *bufio.Reader
	writer *os.File
}

// NewPrompter creates a new prompter for interactive setup
func NewPrompter() *Prompter {
	return &Prompter{
		reader: bufio.NewReader(os.Stdin),
		writer: os.Stdout,
	}
}

// RunSetup runs the interactive setup process
func (p *Prompter) RunSetup(configPath string) error {
	fmt.Fprintln(p.writer, "Welcome to ghissues!")
	fmt.Fprintln(p.writer, "Let's configure your GitHub repository settings.")
	fmt.Fprintln(p.writer)

	// Prompt for repository
	repo, err := p.promptRepository()
	if err != nil {
		return fmt.Errorf("failed to get repository: %w", err)
	}

	// Prompt for authentication method
	authMethod, err := p.promptAuthMethod()
	if err != nil {
		return fmt.Errorf("failed to get authentication method: %w", err)
	}

	// Create config
	cfg := &config.Config{
		Default: config.DefaultConfig{
			Repository: repo,
		},
		Repositories: []config.Repository{
			{
				Owner:    strings.Split(repo, "/")[0],
				Name:     strings.Split(repo, "/")[1],
				Database: ".ghissues.db",
			},
		},
		Display: config.DisplayConfig{
			Theme:   "default",
			Columns: []string{"number", "title", "author", "updated", "comments"},
		},
		Sort: config.SortConfig{
			Field:      "updated",
			Descending: true,
		},
		Database: config.DatabaseConfig{
			Path: ".ghissues.db",
		},
	}

	// If user chose to enter a token, prompt for it
	if authMethod == "token" {
		token, err := p.promptToken()
		if err != nil {
			return fmt.Errorf("failed to get token: %w", err)
		}
		cfg.Auth.Token = token
	}

	// Save config
	fmt.Fprintln(p.writer)
	fmt.Fprintln(p.writer, "Saving configuration...")

	err = config.SaveConfig(configPath, cfg)
	if err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Fprintf(p.writer, "Configuration saved to %s\n", configPath)
	fmt.Fprintln(p.writer)
	fmt.Fprintln(p.writer, "You're all set! Run 'ghissues' to get started.")

	return nil
}

// promptRepository prompts the user for a GitHub repository
func (p *Prompter) promptRepository() (string, error) {
	for {
		fmt.Fprint(p.writer, "Enter GitHub repository (owner/repo): ")

		input, err := p.reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		repo := strings.TrimSpace(input)
		if repo == "" {
			fmt.Fprintln(p.writer, "Repository cannot be empty.")
			continue
		}

		// Validate repository format
		err = config.ValidateRepository(repo)
		if err != nil {
			fmt.Fprintf(p.writer, "Invalid repository format: %v\n", err)
			continue
		}

		return repo, nil
	}
}

// promptAuthMethod prompts the user for authentication method preference
func (p *Prompter) promptAuthMethod() (string, error) {
	fmt.Fprintln(p.writer)
	fmt.Fprintln(p.writer, "Choose authentication method:")
	fmt.Fprintln(p.writer, "  1. GitHub Token (personal access token)")
	fmt.Fprintln(p.writer, "  2. GitHub CLI (use existing 'gh' CLI auth)")
	fmt.Fprintln(p.writer)

	for {
		fmt.Fprint(p.writer, "Enter choice (1 or 2): ")

		input, err := p.reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		choice := strings.TrimSpace(input)
		switch choice {
		case "1":
			return "token", nil
		case "2":
			return "gh", nil
		default:
			fmt.Fprintln(p.writer, "Invalid choice. Please enter 1 or 2.")
		}
	}
}

// promptToken prompts the user for a GitHub token
func (p *Prompter) promptToken() (string, error) {
	fmt.Fprint(p.writer, "Enter GitHub token: ")

	input, err := p.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	token := strings.TrimSpace(input)
	if token == "" {
		fmt.Fprintln(p.writer, "No token entered. You can configure it later in the config file.")
	}

	return token, nil
}
