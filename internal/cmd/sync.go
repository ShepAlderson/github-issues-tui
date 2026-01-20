package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

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

	// Fetch issues
	fmt.Fprintln(output, "Fetching issues from GitHub...")
	issues, err := fetchIssuesWithProgress(ctx, client, config.Repository, output)
	if err != nil {
		return fmt.Errorf("failed to fetch issues: %w", err)
	}

	if len(issues) == 0 {
		fmt.Fprintln(output, "No issues found.")
		return nil
	}

	fmt.Fprintf(output, "\nFound %d open issues\n", len(issues))

	// Clear existing issues before storing new ones
	if err := database.ClearAllIssues(); err != nil {
		return fmt.Errorf("failed to clear existing issues: %w", err)
	}

	// Store issues and fetch comments
	fmt.Fprintln(output, "\nStoring issues and comments...")
	if err := storeIssues(ctx, database, client, config.Repository, issues, output); err != nil {
		return fmt.Errorf("failed to store issues: %w", err)
	}

	fmt.Fprintln(output, "\n✓ Sync complete!")
	return nil
}

// fetchIssuesWithProgress fetches issues with progress bar
func fetchIssuesWithProgress(ctx context.Context, client *ghclient.Client, repo string, output io.Writer) ([]ghclient.Issue, error) {
	// First, we need to fetch the first page to know how many issues there are
	// Since GitHub API doesn't provide total count, we'll show progress per page
	fmt.Fprintln(output, "Fetching issues...")

	// Create a simple progress indicator
	issues, err := client.FetchOpenIssues(repo)
	if err != nil {
		return nil, err
	}

	return issues, nil
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
