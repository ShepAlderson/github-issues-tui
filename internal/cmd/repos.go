package cmd

import (
	"fmt"
	"io"

	"github.com/shepbook/git/github-issues-tui/internal/config"
)

// RunReposCommand runs the repos command to list configured repositories
func RunReposCommand(configPath string, output io.Writer) error {
	// Check if config file exists
	exists, err := config.ConfigExists(configPath)
	if err != nil {
		return fmt.Errorf("failed to check config file: %w", err)
	}

	if !exists {
		fmt.Fprintln(output, "Configuration file not found.")
		fmt.Fprintf(output, "Expected: %s\n", configPath)
		fmt.Fprintln(output)
		fmt.Fprintln(output, "Please run 'ghissues config' to set up your configuration.")
		return fmt.Errorf("config file not found")
	}

	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg == nil {
		return fmt.Errorf("config file exists but could not be loaded")
	}

	// Get the list of repositories
	repos := config.ListRepositories(cfg)

	if len(repos) == 0 {
		fmt.Fprintln(output, "No repositories configured.")
		fmt.Fprintln(output)
		fmt.Fprintln(output, "Run 'ghissues config' to add a repository.")
		return nil
	}

	// Get the default repository
	defaultRepo := config.GetDefaultRepo(cfg)

	// Print header
	fmt.Fprintln(output, "=== Configured Repositories ===")
	fmt.Fprintln(output)

	// Print each repository
	for _, repo := range repos {
		marker := "  "
		if repo == defaultRepo {
			marker = "* "
		}
		fmt.Fprintf(output, "%s%s", marker, repo)

		// Add (default) indicator
		if repo == defaultRepo {
			fmt.Fprint(output, " (default)")
		}
		fmt.Fprintln(output)
	}

	fmt.Fprintln(output)
	fmt.Fprintln(output, "Use --repo <repository> to view a specific repository")
	fmt.Fprintln(output, "Example: ghissues --repo owner/repo list")

	return nil
}
