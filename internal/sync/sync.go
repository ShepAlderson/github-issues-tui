package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/shepbook/github-issues-tui/internal/config"
	"github.com/shepbook/github-issues-tui/internal/database"
	"github.com/shepbook/github-issues-tui/internal/github"
)

// SyncManager handles syncing issues from GitHub to local database
type SyncManager struct {
	configManager *config.Manager
	authManager   *github.AuthManager
	dbManager     *database.DBManager
}

// NewSyncManager creates a new sync manager
func NewSyncManager(configManager *config.Manager, authManager *github.AuthManager, dbManager *database.DBManager) *SyncManager {
	return &SyncManager{
		configManager: configManager,
		authManager:   authManager,
		dbManager:     dbManager,
	}
}

// SyncOptions contains options for the sync operation
type SyncOptions struct {
	ShowProgress bool
	CancelChan   <-chan struct{}
}

// Sync performs a full sync of issues and comments from GitHub
func (sm *SyncManager) Sync(ctx context.Context, opts SyncOptions) error {
	// Load configuration
	cfg, err := sm.configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get authenticated GitHub client
	_, err = sm.authManager.GetAuthenticatedClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get authenticated client: %w", err)
	}

	// Parse owner and repo from config
	owner, repo := parseRepo(cfg.Repository)
	if owner == "" || repo == "" {
		return fmt.Errorf("invalid repository format in config: %s", cfg.Repository)
	}

	// Initialize database schema if needed
	if err := sm.dbManager.InitializeSchema(); err != nil {
		return fmt.Errorf("failed to initialize database schema: %w", err)
	}

	// TODO: Implement actual GitHub API fetching
	// This would include:
	// 1. Fetching all open issues with pagination
	// 2. Storing issues in database
	// 3. Fetching comments for each issue
	// 4. Storing comments in database
	// 5. Showing progress bar during operations

	// For now, just show what would happen
	if opts.ShowProgress {
		fmt.Println("Progress bar would show here")
	}
	fmt.Printf("Sync would fetch issues from %s/%s\n", owner, repo)
	fmt.Println("Using authenticated GitHub client")

	// Update sync metadata
	if err := sm.updateSyncMetadata(); err != nil {
		return fmt.Errorf("failed to update sync metadata: %w", err)
	}

	return nil
}

// updateSyncMetadata updates the sync metadata in the database
func (sm *SyncManager) updateSyncMetadata() error {
	now := time.Now().UTC()
	_, err := sm.dbManager.GetDB().Exec(`
		INSERT OR REPLACE INTO metadata (key, value)
		VALUES ('last_sync', ?)
	`, now.Format(time.RFC3339))
	return err
}

// parseRepo parses owner/repo string into owner and repo components
func parseRepo(repoStr string) (owner, repo string) {
	parts := splitRepo(repoStr)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

// splitRepo splits owner/repo string
func splitRepo(repoStr string) []string {
	// Only split on first slash
	for i := 0; i < len(repoStr); i++ {
		if repoStr[i] == '/' {
			// Check if there are more slashes
			for j := i + 1; j < len(repoStr); j++ {
				if repoStr[j] == '/' {
					return nil
				}
			}
			return []string{repoStr[:i], repoStr[i+1:]}
		}
	}
	return nil
}

// getIssueCount returns the number of issues in the database
func (sm *SyncManager) getIssueCount() (int, error) {
	var count int
	err := sm.dbManager.GetDB().QueryRow("SELECT COUNT(*) FROM issues").Scan(&count)
	return count, err
}