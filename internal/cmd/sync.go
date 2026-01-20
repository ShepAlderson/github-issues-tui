package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/shepbook/git/github-issues-tui/internal/db"
	ghclient "github.com/shepbook/git/github-issues-tui/internal/github"
)

// SyncConfig holds configuration for sync command
type SyncConfig struct {
	Token      string
	Repository string
	GitHubURL  string // For testing with mock servers
}

// RunSyncCommand runs the sync command to fetch and store issues
func RunSyncCommand(dbPath string, config *SyncConfig, output io.Writer) error {
	// Create context with cancellation for Ctrl+C support
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Fprintln(output, "\nSync cancelled by user...")
		cancel()
	}()

	// Initialize database
	database, err := db.NewDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer database.Close()

	// Initialize GitHub client
	client := ghclient.NewClient(config.Token, config.GitHubURL)

	// Check if this is first sync (not used in full sync, but keeping for consistency)
	_, err = database.GetLastSyncDate()
	if err != nil {
		return fmt.Errorf("failed to get last sync date: %w", err)
	}

	// Fetch issues
	fmt.Fprintln(output, "Fetching issues from GitHub...")
	issues, err := fetchIssuesWithProgress(ctx, client, config.Repository, "", output)
	if err != nil {
		return fmt.Errorf("failed to fetch issues: %w", err)
	}

	if len(issues) == 0 {
		fmt.Fprintln(output, "No issues found.")
		return nil
	}

	fmt.Fprintf(output, "\nFound %d open issues\n", len(issues))

	// Clear existing issues before storing new ones (full sync)
	if err := database.ClearAllIssues(); err != nil {
		return fmt.Errorf("failed to clear existing issues: %w", err)
	}

	// Store issues and fetch comments
	fmt.Fprintln(output, "\nStoring issues and comments...")
	if err := storeIssues(ctx, database, client, config.Repository, issues, output); err != nil {
		return fmt.Errorf("failed to store issues: %w", err)
	}

	// Update sync timestamp
	currentTime := time.Now().UTC().Format(time.RFC3339)
	_, err = database.SetLastSyncDate(currentTime)
	if err != nil {
		return fmt.Errorf("failed to update sync timestamp: %w", err)
	}

	fmt.Fprintln(output, "\n✓ Sync complete!")
	return nil
}

// storeIssues stores issues and fetches their comments
func storeIssues(ctx context.Context, database *db.DB, client *ghclient.Client, repo string, issues []ghclient.Issue, output io.Writer) error {
	// Create progress bar
	bar := progressbar.NewOptions(len(issues),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(40),
		progressbar.OptionSetDescription("Processing issues"),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionSetWriter(output),
	)

	for i, issue := range issues {
		// Check for cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("sync cancelled")
		default:
		}

		// Convert GitHub issue to DB issue
		dbIssue := &db.Issue{
			Number:       issue.Number,
			Title:        issue.Title,
			Body:         issue.Body,
			State:        issue.State,
			Author:       issue.User.Login,
			CreatedAt:    issue.CreatedAt,
			UpdatedAt:    issue.UpdatedAt,
			CommentCount: issue.Comments,
			Labels:       extractLabelNames(issue.Labels),
			Assignees:    extractUserLogins(issue.Assignees),
		}

		// Store the issue
		if err := database.StoreIssue(dbIssue); err != nil {
			return fmt.Errorf("failed to store issue %d: %w", issue.Number, err)
		}

		// Fetch and store comments if there are any
		if issue.Comments > 0 {
			comments, err := client.FetchIssueComments(repo, issue.Number)
			if err != nil {
				// Log error but continue with other issues
				fmt.Fprintf(output, "\nWarning: Failed to fetch comments for issue %d: %v\n", issue.Number, err)
			} else {
				// Store comments
				for _, comment := range comments {
					dbComment := &db.Comment{
						ID:        comment.ID,
						IssueNum:  issue.Number,
						Body:      comment.Body,
						Author:    comment.User.Login,
						CreatedAt: comment.CreatedAt,
					}
					if err := database.StoreComment(dbComment); err != nil {
						// Log error but continue
						fmt.Fprintf(output, "\nWarning: Failed to store comment %d: %v\n", comment.ID, err)
					}
				}
			}
		}

		// Update progress bar
		bar.Set(i + 1)
	}

	// Finish the progress bar
	bar.Finish()

	return nil
}

// extractLabelNames extracts label names from GitHub labels
func extractLabelNames(labels []ghclient.Label) []string {
	var names []string
	for _, label := range labels {
		names = append(names, label.Name)
	}
	return names
}

// extractUserLogins extracts user logins from GitHub users
func extractUserLogins(users []ghclient.User) []string {
	var logins []string
	for _, user := range users {
		logins = append(logins, user.Login)
	}
	return logins
}

// RunIncrementalSync performs an incremental sync to update only changed issues
func RunIncrementalSync(dbPath string, config *SyncConfig, output io.Writer) error {
	// Create context with cancellation for Ctrl+C support
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Fprintln(output, "\nSync cancelled by user...")
		cancel()
	}()

	// Initialize database
	database, err := db.NewDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer database.Close()

	// Check if this is first sync
	lastSyncDate, err := database.GetLastSyncDate()
	if err != nil {
		return fmt.Errorf("failed to get last sync date: %w", err)
	}

	isFirstSync := lastSyncDate == "1970-01-01T00:00:00Z"

	if isFirstSync {
		return fmt.Errorf("no previous sync found, run full sync first with 'ghissues sync'")
	}

	// Initialize GitHub client
	client := ghclient.NewClient(config.Token, config.GitHubURL)

	// Start the incremental sync
	fmt.Fprintf(output, "Performing incremental sync since %s...\n", lastSyncDate)
	
	// Get all currently stored issue numbers
	storedIssueNumbers, err := database.GetAllIssueNumbers()
	if err != nil {
		return fmt.Errorf("failed to get stored issue numbers: %w", err)
	}

	// Fetch issues updated since last sync
	updatedIssues, err := fetchIssuesWithProgress(ctx, client, config.Repository, lastSyncDate, output)
	if err != nil {
		return fmt.Errorf("failed to fetch updated issues: %w", err)
	}

	fmt.Fprintf(output, "\nFound %d issues updated since last sync\n", len(updatedIssues))

	if len(updatedIssues) == 0 {
		// Still update sync timestamp
		currentTime := time.Now().UTC().Format(time.RFC3339)
		if _, err := database.SetLastSyncDate(currentTime); err != nil {
			fmt.Fprintf(output, "Warning: Failed to update sync timestamp: %v\n", err)
		}
		fmt.Fprintf(output, "\n✓ Sync complete! No issues were updated.\n")
		return nil
	}

	// Store issues and fetch new comments
	fmt.Fprintln(output, "\nProcessing updated issues...")
	if err := storeIncrementalIssues(ctx, database, client, config.Repository, updatedIssues, output); err != nil {
		return fmt.Errorf("failed to store issues: %w", err)
	}

	// Detect deleted issues by comparing stored vs fetched
	if len(storedIssueNumbers) > 0 {
		fmt.Fprintln(output, "\nChecking for deleted issues...")
		if err := detectAndRemoveDeletedIssues(ctx, client, config.Repository, storedIssueNumbers, database, output); err != nil {
			fmt.Fprintf(output, "Warning: Failed to fully check for deleted issues: %v\n", err)
		}
	}

	// Update sync timestamp
	currentTime := time.Now().UTC().Format(time.RFC3339)
	if _, err := database.SetLastSyncDate(currentTime); err != nil {
		return fmt.Errorf("failed to update sync timestamp: %w", err)
	}

	fmt.Fprintln(output, "\n✓ Sync complete!")
	return nil
}

// detectAndRemoveDeletedIssues checks which issues from our local database no longer exist on GitHub
func detectAndRemoveDeletedIssues(ctx context.Context, client *ghclient.Client, repo string, storedNumbers []int, database *db.DB, output io.Writer) error {
	// Get all issues currently in GitHub to check for deletions
	fmt.Fprintln(output, "Fetching all issues from GitHub to check for deleted ones...")
	allIssues, err := client.FetchOpenIssues(repo)
	if err != nil {
		return fmt.Errorf("failed to fetch all issues for deletion check: %w", err)
	}

	// Build a map of all current issues
	allCurrentOnGitHub := make(map[int]bool)
	for _, issue := range allIssues {
		allCurrentOnGitHub[issue.Number] = true
	}

	// Check which stored issues no longer exist
	var deletedIssues []int
	for _, storedNum := range storedNumbers {
		if !allCurrentOnGitHub[storedNum] {
			deletedIssues = append(deletedIssues, storedNum)
		}
	}

	if len(deletedIssues) > 0 {
		fmt.Fprintf(output, "Found %d deleted issues\n", len(deletedIssues))
		if err := database.RemoveIssues(deletedIssues); err != nil {
			return fmt.Errorf("failed to remove deleted issues: %w", err)
		}
		fmt.Fprintf(output, "Removed %d deleted issues from local database\n", len(deletedIssues))
	} else {
		fmt.Fprintln(output, "No deleted issues found")
	}

	return nil
}

// storeIncrementalIssues stores/fetches updated issues and their new comments
func storeIncrementalIssues(ctx context.Context, database *db.DB, client *ghclient.Client, repo string, issues []ghclient.Issue, output io.Writer) error {
	// Create progress bar
	bar := progressbar.NewOptions(len(issues),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(40),
		progressbar.OptionSetDescription("Processing issues"),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionSetWriter(output),
	)

	// Get the last sync date to check for new comments
	lastSyncDate, err := database.GetLastSyncDate()
	if err != nil {
		lastSyncDate = "1970-01-01T00:00:00Z" // Fallback
	}

	for i, issue := range issues {
		// Check for cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("sync cancelled")
		default:
		}

		// Convert GitHub issue to DB issue
		dbIssue := &db.Issue{
			Number:       issue.Number,
			Title:        issue.Title,
			Body:         issue.Body,
			State:        issue.State,
			Author:       issue.User.Login,
			CreatedAt:    issue.CreatedAt,
			UpdatedAt:    issue.UpdatedAt,
			CommentCount: issue.Comments,
			Labels:       extractLabelNames(issue.Labels),
			Assignees:    extractUserLogins(issue.Assignees),
		}

		// Store the issue (this will update if it exists, or insert if new)
		if err := database.StoreIssue(dbIssue); err != nil {
			return fmt.Errorf("failed to store issue %d: %w", issue.Number, err)
		}

		// Fetch and store comments if there are any and we have a sync date
		if issue.Comments > 0 {
			// Try to fetch only new comments since last sync
			if lastSyncDate != "1970-01-01T00:00:00Z" {
				comments, err := client.FetchIssueCommentsSince(repo, issue.Number, lastSyncDate)
				if err != nil {
					// Log warning but continue - fall back to fetching all comments
					fmt.Fprintf(output, "\nWarning: Failed to fetch new comments for issue %d: %v\n", issue.Number, err)
					// Fall through to fetch all comments
				} else if len(comments) > 0 {
					// Store only new comments
					for _, comment := range comments {
						dbComment := &db.Comment{
							ID:        comment.ID,
							IssueNum:  issue.Number,
							Body:      comment.Body,
							Author:    comment.User.Login,
							CreatedAt: comment.CreatedAt,
						}
						if err := database.StoreComment(dbComment); err != nil {
							fmt.Fprintf(output, "\nWarning: Failed to store comment %d: %v\n", comment.ID, err)
						}
					}
					continue // Skip fetching all comments since we got new ones
				}
			}

			// Fetch all comments (fallback or first sync)
			comments, err := client.FetchIssueComments(repo, issue.Number)
			if err != nil {
				fmt.Fprintf(output, "\nWarning: Failed to fetch comments for issue %d: %v\n", issue.Number, err)
			} else {
				// Store comments
				for _, comment := range comments {
					dbComment := &db.Comment{
						ID:        comment.ID,
						IssueNum:  issue.Number,
						Body:      comment.Body,
						Author:    comment.User.Login,
						CreatedAt: comment.CreatedAt,
					}
					if err := database.StoreComment(dbComment); err != nil {
						fmt.Fprintf(output, "\nWarning: Failed to store comment %d: %v\n", comment.ID, err)
					}
				}
			}
		}

		// Update progress bar
		bar.Set(i + 1)
	}

	// Finish the progress bar
	bar.Finish()

	return nil
}

// fetchIssuesWithProgress fetches issues with optional since parameter
func fetchIssuesWithProgress(ctx context.Context, client *ghclient.Client, repo string, since string, output io.Writer) ([]ghclient.Issue, error) {
	if since != "" {
		fmt.Fprintf(output, "Fetching issues updated since %s...\n", since)
		return client.FetchIssuesSince(repo, since)
	}
	
	fmt.Fprintln(output, "Fetching all issues from GitHub...")
	return client.FetchOpenIssues(repo)
}
