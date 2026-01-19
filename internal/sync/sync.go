package sync

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// Update last sync time
	if err := s.store.SetLastSyncTime(time.Now().UTC()); err != nil {
		return fmt.Errorf("failed to update last sync time: %w", err)
	}

	fmt.Printf("\n✓ Successfully synced %d issues and %d comments\n", len(issues), totalComments)

	return nil
}

// ClearAllData removes all synced data from the database
func (s *Syncer) ClearAllData() error {
	return s.store.ClearData()
}

// RefreshIssues performs an incremental sync, fetching only issues updated since last sync
func (s *Syncer) RefreshIssues(repo string) error {
	// Create a context that can be cancelled with Ctrl+C
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful cancellation
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n\nReceived interrupt signal, cancelling refresh...")
		cancel()
	}()

	return s.refreshWithContext(ctx, repo)
}

// refreshWithContext performs the incremental sync with a cancellable context
func (s *Syncer) refreshWithContext(ctx context.Context, repo string) error {
	// Get last sync time
	lastSync, err := s.store.GetLastSyncTime()
	if err != nil {
		return fmt.Errorf("failed to get last sync time: %w", err)
	}

	// If never synced before, do a full sync instead
	if lastSync.IsZero() {
		fmt.Println("No previous sync found, performing full sync...")
		return s.syncWithContext(ctx, repo)
	}

	fmt.Printf("Fetching issues updated since %s...\n", lastSync.Format("2006-01-02 15:04:05"))

	// Fetch issues updated since last sync
	issues, err := s.client.FetchIssuesSince(ctx, repo, lastSync)
	if err != nil {
		return fmt.Errorf("failed to fetch issues: %w", err)
	}

	if len(issues) == 0 {
		fmt.Println("No updates found")
		return nil
	}

	fmt.Printf("Found %d updated issues\n", len(issues))

	// Create progress bar for issue storage
	bar := progressbar.NewOptions(len(issues),
		progressbar.OptionSetDescription("Refreshing issues"),
		progressbar.OptionSetWidth(40),
		progressbar.OptionShowCount(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionClearOnFinish(),
	)

	// Store updated issues and fetch new comments
	totalComments := 0
	for _, issue := range issues {
		// Check for cancellation
		select {
		case <-ctx.Done():
			fmt.Println("\nRefresh cancelled by user")
			return ctx.Err()
		default:
		}

		// Store the issue (will update if exists)
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

	// Remove closed issues from database
	if err := s.store.RemoveClosedIssues(); err != nil {
		return fmt.Errorf("failed to remove closed issues: %w", err)
	}

	// Update last sync time
	if err := s.store.SetLastSyncTime(time.Now().UTC()); err != nil {
		return fmt.Errorf("failed to update last sync time: %w", err)
	}

	fmt.Printf("\n✓ Successfully refreshed %d issues and %d comments\n", len(issues), totalComments)

	return nil
}
