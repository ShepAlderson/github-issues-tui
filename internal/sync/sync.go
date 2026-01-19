package sync

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/schollz/progressbar/v3"
)

// Syncer coordinates the issue synchronization process
type Syncer struct {
	client *GitHubClient
	store  *IssueStore
}

// NewSyncer creates a new syncer
func NewSyncer(token string, dbPath string) (*Syncer, error) {
	store, err := NewIssueStore(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue store: %w", err)
	}

	return &Syncer{
		client: NewGitHubClient(token),
		store:  store,
	}, nil
}

// Close closes the syncer and releases resources
func (s *Syncer) Close() error {
	return s.store.Close()
}

// SyncIssues performs a full sync of issues and their comments
func (s *Syncer) SyncIssues(repo string) error {
	// Create a context that can be cancelled with Ctrl+C
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful cancellation
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n\nReceived interrupt signal, cancelling sync...")
		cancel()
	}()

	return s.syncWithContext(ctx, repo)
}

// syncWithContext performs the sync with a cancellable context
func (s *Syncer) syncWithContext(ctx context.Context, repo string) error {
	fmt.Println("Fetching issues from GitHub...")

	// Fetch all issues
	issues, err := s.client.FetchIssues(ctx, repo)
	if err != nil {
		return fmt.Errorf("failed to fetch issues: %w", err)
	}

	if len(issues) == 0 {
		fmt.Println("No open issues found in repository")
		return nil
	}

	fmt.Printf("Found %d open issues\n", len(issues))

	// Create progress bar for issue storage
	bar := progressbar.NewOptions(len(issues),
		progressbar.OptionSetDescription("Syncing issues"),
		progressbar.OptionSetWidth(40),
		progressbar.OptionShowCount(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionClearOnFinish(),
	)

	// Store issues and fetch comments
	totalComments := 0
	for _, issue := range issues {
		// Check for cancellation
		select {
		case <-ctx.Done():
			fmt.Println("\nSync cancelled by user")
			return ctx.Err()
		default:
		}

		// Store the issue
		if err := s.store.StoreIssue(issue); err != nil {
			return fmt.Errorf("failed to store issue #%d: %w", issue.Number, err)
		}

		// Fetch and store comments if the issue has any
		if issue.CommentCount > 0 {
			comments, err := s.client.FetchComments(ctx, repo, issue.Number)
			if err != nil {
				return fmt.Errorf("failed to fetch comments for issue #%d: %w", issue.Number, err)
			}

			for _, comment := range comments {
				if err := s.store.StoreComment(comment); err != nil {
					return fmt.Errorf("failed to store comment: %w", err)
				}
			}

			totalComments += len(comments)
		}

		bar.Add(1)
	}

	bar.Finish()
	fmt.Printf("\n✓ Successfully synced %d issues and %d comments\n", len(issues), totalComments)

	return nil
}

// ClearAllData removes all synced data from the database
func (s *Syncer) ClearAllData() error {
	return s.store.ClearData()
}
