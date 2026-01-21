package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/shepbook/github-issues-tui/internal/config"
	"github.com/shepbook/github-issues-tui/internal/database"
	"github.com/shepbook/github-issues-tui/internal/github"
	"github.com/shepbook/github-issues-tui/internal/sync"
	"github.com/spf13/cobra"
)

func newSyncCmd() *cobra.Command {
	var showProgress bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync issues from GitHub",
		Long:  "Fetch all open issues and comments from the configured GitHub repository and store them locally",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create context with cancellation support
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Set up signal handling for graceful cancellation
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				<-sigChan
				fmt.Println("\nReceived interrupt signal, cancelling sync...")
				cancel()
			}()

			// Create managers
			configManager := config.NewManager()
			authManager := github.NewAuthManager(configManager)

			// Load config to get database path
			cfg, err := configManager.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Get database path
			dbPath := database.GetDatabasePath("", cfg)

			// Ensure database path is writable
			if err := database.EnsureDatabasePath(dbPath); err != nil {
				return fmt.Errorf("database path error: %w", err)
			}

			// Create database manager
			dbManager, err := database.NewDBManager(dbPath)
			if err != nil {
				return fmt.Errorf("failed to create database manager: %w", err)
			}
			defer dbManager.Close()

			// Create sync manager
			syncManager := sync.NewSyncManager(configManager, authManager, dbManager)

			// Perform sync
			fmt.Printf("Starting sync for repository: %s\n", cfg.Repository)
			fmt.Printf("Database location: %s\n", database.AbsolutePath(dbPath))

			opts := sync.SyncOptions{
				ShowProgress: showProgress,
				CancelChan:   ctx.Done(),
			}

			if err := syncManager.Sync(ctx, opts); err != nil {
				return fmt.Errorf("sync failed: %w", err)
			}

			fmt.Println("Sync completed successfully!")
			return nil
		},
	}

	// Add flags
	cmd.Flags().BoolVarP(&showProgress, "progress", "p", true, "Show progress bar during sync")

	return cmd
}