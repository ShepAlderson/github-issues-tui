package cmd

import (
	"fmt"

	"github.com/shepbook/ghissues/internal/config"
	"github.com/spf13/cobra"
)

// newReposCmd creates the repos subcommand for managing configured repositories
func newReposCmd() *cobra.Command {
	var addRepo string
	var dbPath string
	var setDefault string

	cmd := &cobra.Command{
		Use:   "repos",
		Short: "List and manage configured repositories",
		Long: `List all configured repositories and manage repository settings.

Use --add to add a new repository to track.
Use --set-default to set which repository to use by default.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := GetConfigPath()

			// Load config
			cfg, err := config.Load(path)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Handle --add flag
			if addRepo != "" {
				// Validate repository format
				if err := config.ValidateRepository(addRepo); err != nil {
					return err
				}

				// Add the repository
				if err := cfg.AddRepository(addRepo, dbPath); err != nil {
					return err
				}

				// Save config
				if err := config.Save(cfg, path); err != nil {
					return fmt.Errorf("failed to save config: %w", err)
				}

				fmt.Fprintf(cmd.OutOrStdout(), "Added repository: %s\n", addRepo)
				return nil
			}

			// Handle --set-default flag
			if setDefault != "" {
				if err := cfg.SetDefaultRepository(setDefault); err != nil {
					return err
				}

				// Save config
				if err := config.Save(cfg, path); err != nil {
					return fmt.Errorf("failed to save config: %w", err)
				}

				fmt.Fprintf(cmd.OutOrStdout(), "Default repository set to: %s\n", setDefault)
				return nil
			}

			// List repositories
			repos := cfg.ListRepositories()
			if len(repos) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No repositories configured. Use 'ghissues repos --add owner/repo' to add one.")
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Configured repositories:")
			for _, repo := range repos {
				indicator := "  "
				if repo.Name == cfg.DefaultRepository {
					indicator = "* "
				}
				dbPathDisplay := repo.GetDatabasePath()
				fmt.Fprintf(cmd.OutOrStdout(), "%s%s (db: %s)%s\n", indicator, repo.Name, dbPathDisplay, defaultIndicator(repo.Name, cfg.DefaultRepository))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&addRepo, "add", "", "Add a new repository (format: owner/repo)")
	cmd.Flags().StringVar(&dbPath, "db-path", "", "Database path for the repository (used with --add)")
	cmd.Flags().StringVar(&setDefault, "set-default", "", "Set the default repository")

	return cmd
}

// defaultIndicator returns " (default)" if the repo is the default, empty string otherwise
func defaultIndicator(repoName, defaultRepo string) string {
	if repoName == defaultRepo {
		return " (default)"
	}
	return ""
}
