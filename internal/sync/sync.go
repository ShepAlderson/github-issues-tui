package sync

import (
	"context"
	"fmt"
	"time"

	gh "github.com/google/go-github/v62/github"
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

	// Get authenticated GitHub client
	client, err := sm.authManager.GetAuthenticatedClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get authenticated client: %w", err)
	}

	// Get last sync time for incremental updates
	lastSyncTime, err := sm.dbManager.GetLastSyncTime()
	if err != nil {
		return fmt.Errorf("failed to get last sync time: %w", err)
	}

	// Track fetched issue numbers to detect deleted issues
	fetchedIssueNumbers := make(map[int]bool)

	// Fetch issues with pagination
	page := 1
	perPage := 100 // GitHub API max is 100
	for {
		// Check for cancellation
		select {
		case <-opts.CancelChan:
			return fmt.Errorf("sync cancelled")
		default:
			// Continue
		}

		issueOpts := createIssueListOptions(lastSyncTime, page, perPage)
		issues, resp, err := client.Issues.ListByRepo(ctx, owner, repo, issueOpts)
		if err != nil {
			return fmt.Errorf("failed to fetch issues page %d: %w", page, err)
		}

		// Process issues
		for _, issue := range issues {
			// Skip pull requests (issues have PullRequestLinks == nil)
			if issue.PullRequestLinks != nil {
				continue
			}

			// Convert GitHub issue to our IssueDetail structure
			issueDetail := convertGitHubIssue(issue)

			// Store issue in database
			if err := sm.dbManager.UpsertIssue(issueDetail); err != nil {
				return fmt.Errorf("failed to store issue #%d: %w", issueDetail.Number, err)
			}

			// Track fetched issue number
			fetchedIssueNumbers[issueDetail.Number] = true

			// Fetch comments for this issue
			if err := sm.fetchAndStoreComments(ctx, client, issueDetail.Number, owner, repo, lastSyncTime); err != nil {
				return fmt.Errorf("failed to fetch comments for issue #%d: %w", issueDetail.Number, err)
			}
		}

		// Check if we should continue paginating
		if page >= resp.LastPage || len(issues) == 0 {
			break
		}
		page++
	}

	// Handle deleted issues (issues in database but not in fetched list)
	if err := sm.handleDeletedIssues(fetchedIssueNumbers); err != nil {
		return fmt.Errorf("failed to handle deleted issues: %w", err)
	}

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

// convertGitHubIssue converts a GitHub API issue to our IssueDetail structure
func convertGitHubIssue(ghIssue *gh.Issue) *database.IssueDetail {
	issue := &database.IssueDetail{
		Number:       ghIssue.GetNumber(),
		Title:        ghIssue.GetTitle(),
		Body:         ghIssue.GetBody(),
		Author:       ghIssue.User.GetLogin(),
		AuthorURL:    ghIssue.User.GetHTMLURL(),
		CreatedAt:    ghIssue.GetCreatedAt().Time,
		UpdatedAt:    ghIssue.GetUpdatedAt().Time,
		CommentCount: ghIssue.GetComments(),
		State:        ghIssue.GetState(),
	}

	// Extract labels
	var labels []string
	for _, label := range ghIssue.Labels {
		if label.Name != nil {
			labels = append(labels, *label.Name)
		}
	}
	issue.Labels = labels

	// Extract assignees
	var assignees []string
	for _, assignee := range ghIssue.Assignees {
		if assignee.Login != nil {
			assignees = append(assignees, *assignee.Login)
		}
	}
	issue.Assignees = assignees

	return issue
}

// createIssueListOptions creates GitHub issue list options with conditional Since field
func createIssueListOptions(lastSyncTime time.Time, page, perPage int) *gh.IssueListByRepoOptions {
	opts := &gh.IssueListByRepoOptions{
		State:       "open",
		ListOptions: gh.ListOptions{Page: page, PerPage: perPage},
	}
	if !lastSyncTime.IsZero() {
		opts.Since = lastSyncTime
	}
	return opts
}

// createCommentListOptions creates GitHub comment list options with conditional Since field
func createCommentListOptions(lastSyncTime time.Time, page, perPage int) *gh.IssueListCommentsOptions {
	opts := &gh.IssueListCommentsOptions{
		ListOptions: gh.ListOptions{Page: page, PerPage: perPage},
	}
	if !lastSyncTime.IsZero() {
		opts.Since = &lastSyncTime
	}
	return opts
}

// fetchAndStoreComments fetches comments for an issue and stores them in the database
func (sm *SyncManager) fetchAndStoreComments(ctx context.Context, client *gh.Client, issueNumber int, owner, repo string, since time.Time) error {
	page := 1
	perPage := 100

	for {
		commentOpts := createCommentListOptions(since, page, perPage)
		comments, resp, err := client.Issues.ListComments(ctx, owner, repo, issueNumber, commentOpts)
		if err != nil {
			return fmt.Errorf("failed to fetch comments for issue #%d page %d: %w", issueNumber, page, err)
		}

		// Store comments
		for _, comment := range comments {
			commentDetail := &database.Comment{
				ID:        comment.GetID(),
				Author:    comment.User.GetLogin(),
				AuthorURL: comment.User.GetHTMLURL(),
				Body:      comment.GetBody(),
				CreatedAt: comment.GetCreatedAt().Time,
				UpdatedAt: comment.GetUpdatedAt().Time,
			}

			if err := sm.dbManager.UpsertComment(commentDetail, issueNumber); err != nil {
				return fmt.Errorf("failed to store comment %d for issue #%d: %w", commentDetail.ID, issueNumber, err)
			}
		}

		// Check if we should continue paginating
		if page >= resp.LastPage || len(comments) == 0 {
			break
		}
		page++
	}

	return nil
}

// handleDeletedIssues removes issues from the database that weren't fetched (likely closed or deleted)
func (sm *SyncManager) handleDeletedIssues(fetchedIssueNumbers map[int]bool) error {
	// Get all issue numbers from database
	rows, err := sm.dbManager.GetDB().Query("SELECT number FROM issues")
	if err != nil {
		return fmt.Errorf("failed to query existing issues: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var issueNumber int
		if err := rows.Scan(&issueNumber); err != nil {
			return fmt.Errorf("failed to scan issue number: %w", err)
		}

		// If issue wasn't fetched, delete it
		if !fetchedIssueNumbers[issueNumber] {
			if err := sm.dbManager.DeleteIssue(issueNumber); err != nil {
				return fmt.Errorf("failed to delete issue #%d: %w", issueNumber, err)
			}
		}
	}

	return rows.Err()
}
