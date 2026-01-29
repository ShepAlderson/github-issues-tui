package refresh

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/shepbook/ghissues/internal/database"
	"github.com/shepbook/ghissues/internal/github"
)

// Result contains the results of a refresh operation
type Result struct {
	IssuesFetched   int
	IssuesDeleted   int
	CommentsFetched int
	Error           error
	Duration        time.Duration
}

// Options contains options for the refresh operation
type Options struct {
	Repo    string
	DBPath  string
	Token   string
	Since   string // Optional: only fetch issues updated since this time
}

// Perform performs a refresh operation, fetching updated issues and comments
func Perform(opts Options) (Result, error) {
	startTime := time.Now()
	result := Result{}

	// Initialize database
	db, err := database.InitializeSchema(opts.DBPath)
	if err != nil {
		return result, fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Get last sync time if not provided
	since := opts.Since
	if since == "" {
		lastSync, err := database.GetLastSyncTime(db, opts.Repo)
		if err != nil {
			return result, fmt.Errorf("failed to get last sync time: %w", err)
		}
		since = lastSync
	}

	// Create GitHub client
	client := github.NewClient(opts.Token)

	// Fetch issues
	var issues []database.Issue
	if since != "" {
		issues, err = client.FetchIssuesSince(opts.Repo, since, nil)
	} else {
		issues, err = client.FetchIssues(opts.Repo, nil)
	}
	if err != nil {
		return result, fmt.Errorf("failed to fetch issues: %w", err)
	}

	result.IssuesFetched = len(issues)

	// Save issues and fetch comments
	for _, issue := range issues {
		if err := database.SaveIssue(db, opts.Repo, issue); err != nil {
			return result, fmt.Errorf("failed to save issue #%d: %w", issue.Number, err)
		}

		// Fetch comments for this issue
		if issue.CommentCount > 0 {
			// Delete existing comments first (to handle deleted comments)
			if err := database.DeleteCommentsForIssue(db, opts.Repo, issue.Number); err != nil {
				return result, fmt.Errorf("failed to delete comments for issue #%d: %w", issue.Number, err)
			}

			// Fetch all comments
			comments, err := client.FetchComments(opts.Repo, issue.Number, nil)
			if err != nil {
				return result, fmt.Errorf("failed to fetch comments for issue #%d: %w", issue.Number, err)
			}

			for _, comment := range comments {
				if err := database.SaveComment(db, opts.Repo, comment); err != nil {
					return result, fmt.Errorf("failed to save comment: %w", err)
				}
			}
			result.CommentsFetched += len(comments)
		}
	}

	// Handle deleted issues - fetch all current open issue numbers from GitHub
	// and compare with local database
	if err := handleDeletedIssues(db, client, opts.Repo); err != nil {
		return result, fmt.Errorf("failed to handle deleted issues: %w", err)
	}

	// Update last sync time
	now := time.Now().UTC().Format(time.RFC3339)
	if err := database.SaveLastSyncTime(db, opts.Repo, now); err != nil {
		return result, fmt.Errorf("failed to save last sync time: %w", err)
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// handleDeletedIssues removes issues that are no longer present in the GitHub repository
func handleDeletedIssues(db *sql.DB, client *github.Client, repo string) error {
	// Get all local issue numbers
	localNumbers, err := database.GetAllIssueNumbers(db, repo)
	if err != nil {
		return fmt.Errorf("failed to get local issue numbers: %w", err)
	}

	// If no local issues, nothing to check
	if len(localNumbers) == 0 {
		return nil
	}

	// Fetch current open issues from GitHub
	currentIssues, err := client.FetchIssues(repo, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch current issues: %w", err)
	}

	// Build a set of current issue numbers
	currentNumbers := make(map[int]bool)
	for _, issue := range currentIssues {
		currentNumbers[issue.Number] = true
	}

	// Find and delete issues that are no longer present
	for _, num := range localNumbers {
		if !currentNumbers[num] {
			// Issue no longer exists on GitHub, delete it locally
			if err := database.DeleteIssue(db, repo, num); err != nil {
				return fmt.Errorf("failed to delete issue #%d: %w", num, err)
			}
		}
	}

	return nil
}

// ShouldAutoRefresh returns true if auto-refresh should be performed
// based on time since last sync (e.g., if more than 5 minutes have passed)
func ShouldAutoRefresh(dbPath, repo string) (bool, error) {
	db, err := database.InitializeSchema(dbPath)
	if err != nil {
		return false, fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	lastSync, err := database.GetLastSyncTime(db, repo)
	if err != nil {
		return false, fmt.Errorf("failed to get last sync time: %w", err)
	}

	// If never synced, should auto-refresh
	if lastSync == "" {
		return true, nil
	}

	// Parse last sync time
	lastSyncTime, err := time.Parse(time.RFC3339, lastSync)
	if err != nil {
		return false, fmt.Errorf("failed to parse last sync time: %w", err)
	}

	// Auto-refresh if more than 5 minutes have passed
	return time.Since(lastSyncTime) > 5*time.Minute, nil
}
