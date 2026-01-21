package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/shepbook/ghissues/internal/auth"
	"github.com/shepbook/ghissues/internal/config"
	"github.com/shepbook/ghissues/internal/db"
	"github.com/shepbook/ghissues/internal/github"
	"github.com/shepbook/ghissues/internal/sync"
	"github.com/spf13/cobra"
)

// newSyncCmd creates the sync subcommand
func newSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync issues from GitHub to local database",
		Long: `Sync issues from your configured GitHub repository to the local database.

This command fetches all open issues and their comments, storing them locally
for offline access. Progress is displayed during the fetch.

Press Ctrl+C to cancel the sync gracefully.`,
		RunE: runSync,
	}

	return cmd
}

func runSync(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfgPath := GetConfigPath()
	if !config.Exists(cfgPath) {
		return fmt.Errorf("no configuration found, please run 'ghissues config' first")
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get authentication token
	token, source, err := auth.GetToken(cfg)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Using authentication from %s\n", source)

	// Resolve and ensure database path
	resolvedDBPath, err := db.ResolveDBPath(GetDBPath(), cfg)
	if err != nil {
		return fmt.Errorf("failed to resolve database path: %w", err)
	}

	if err := db.EnsureDBPath(resolvedDBPath); err != nil {
		return err
	}

	// Open database
	store, err := db.NewStore(resolvedDBPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer store.Close()

	// Parse repository
	owner, repo, err := parseRepository(cfg.Repository)
	if err != nil {
		return err
	}

	// Set up context with cancellation for Ctrl+C
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Fprintln(cmd.OutOrStdout(), "\nCancelling sync...")
		cancel()
	}()

	// Create client and syncer
	client := github.NewClient(token)
	syncer := sync.NewSyncer(client, store)

	fmt.Fprintf(cmd.OutOrStdout(), "Syncing issues from %s/%s...\n", owner, repo)

	// Run sync with progress callback
	result, err := syncer.Sync(ctx, owner, repo, func(p sync.Progress) {
		printProgress(cmd, p)
	})

	if err != nil {
		if ctx.Err() != nil {
			fmt.Fprintln(cmd.OutOrStdout(), "Sync cancelled.")
			return nil
		}
		return fmt.Errorf("sync failed: %w", err)
	}

	// Clear the progress line and print final result
	fmt.Fprintf(cmd.OutOrStdout(), "\r%-80s\r", "") // Clear line
	fmt.Fprintf(cmd.OutOrStdout(), "Sync complete: %d issues, %d comments fetched in %s\n",
		result.IssuesFetched, result.CommentsFetched, result.Duration.Round(100*1000000))

	return nil
}

func printProgress(cmd *cobra.Command, p sync.Progress) {
	switch p.Phase {
	case "issues":
		bar := progressBar(p.IssuesFetched, p.TotalIssues, 30)
		fmt.Fprintf(cmd.OutOrStdout(), "\rFetching issues: %s %d/%d", bar, p.IssuesFetched, p.TotalIssues)
	case "comments":
		fmt.Fprintf(cmd.OutOrStdout(), "\rFetching comments: %d fetched (issue %d/%d)",
			p.CommentsFetched, p.CurrentIssue, p.TotalIssues)
	}
}

func progressBar(current, total, width int) string {
	if total == 0 {
		return "[" + strings.Repeat(" ", width) + "]"
	}

	filled := (current * width) / total
	if filled > width {
		filled = width
	}

	empty := width - filled
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + "]"
}

func parseRepository(repo string) (owner, name string, err error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository format: %s (expected owner/repo)", repo)
	}
	return parts[0], parts[1], nil
}
