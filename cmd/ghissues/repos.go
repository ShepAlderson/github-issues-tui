package main

import (
	"fmt"

	"github.com/shepbook/ghissues/internal/config"
)

// runRepos lists the configured repositories
func runRepos() error {
	// Check if config exists
	if !config.Exists() {
		return fmt.Errorf("configuration not found. Run 'ghissues config' first")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	repos := cfg.GetRepositories()
	if len(repos) == 0 {
		fmt.Println("No repositories configured.")
		fmt.Println("Use 'ghissues config' to add a repository.")
		return nil
	}

	defaultRepo := cfg.GetDefaultRepository()
	fmt.Println("Configured repositories:")
	fmt.Println()

	for _, repo := range repos {
		if repo == defaultRepo {
			fmt.Printf("  * %s (default)\n", repo)
		} else {
			fmt.Printf("    %s\n", repo)
		}
	}

	fmt.Println()
	if defaultRepo != "" {
		fmt.Printf("Default repository: %s\n", defaultRepo)
	}

	return nil
}

// runRepoAdd adds a repository to the configuration
func runRepoAdd(repo string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		// Create new config if it doesn't exist
		cfg = &config.Config{
			Repositories: []string{},
		}
	}

	if err := cfg.AddRepository(repo); err != nil {
		return err
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Set as default if it's the only one or requested
	if cfg.GetDefaultRepository() == repo {
		fmt.Printf("Added and set '%s' as default repository.\n", repo)
	} else {
		fmt.Printf("Added '%s' to repositories.\n", repo)
	}

	return nil
}

// runRepoRemove removes a repository from the configuration
func runRepoRemove(repo string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if !cfg.RemoveRepository(repo) {
		return fmt.Errorf("repository '%s' not found in configuration", repo)
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Removed '%s' from repositories.\n", repo)
	return nil
}

// runRepoSetDefault sets the default repository
func runRepoSetDefault(repo string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if !cfg.SetDefaultRepository(repo) {
		return fmt.Errorf("repository '%s' not found in configuration", repo)
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Set '%s' as default repository.\n", repo)
	return nil
}

