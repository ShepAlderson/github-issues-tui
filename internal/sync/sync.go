package sync

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/shepbook/ghissues/internal/github"
	"github.com/shepbook/ghissues/internal/storage"
)

// Syncer handles the synchronization of issues from GitHub to local storage
type Syncer struct {
	db        *sql.DB
	client    *github.Client
	repo      string
	quietMode bool
}

// NewSyncer creates a new Syncer instance
func NewSyncer(db *sql.DB, client *github.Client, repo string, quietMode bool) *Syncer {
	return &Syncer{
		db:        db,
		client:    client,
		repo:      repo,
		quietMode: quietMode,
	}
}

// Run performs the full sync operation, fetching all issues and comments
// It supports graceful cancellation via Ctrl+C
func (s *Syncer) Run(ctx context.Context) (*SyncResult, error) {
	// Setup signal handling for graceful cancellation
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// Goroutine to handle cancellation signals
	go func() {
		select {
		case <-sigChan:
			cancel()
			if !s.quietMode {
				fmt.Println("\nSync cancelled by user. Cleaning up...")
			}
		case <-ctx.Done():
			return
		}
	}()

	result := &SyncResult{
		StartTime: time.Now(),
	}

	// Create progress channels
	progressChan := make(chan int)

	var wg sync.WaitGroup

	// Start progress display goroutine
	if !s.quietMode {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.displayProgress(ctx, progressChan, result)
		}()
	}

	// Fetch issues in background
	issues, err := func() ([]storage.Issue, error) {
		return s.client.FetchIssues(progressChan, ctx.Done())
	}()

	if err != nil {
		cancel()
		close(progressChan)
		wg.Wait()

		if ctx.Err() == context.Canceled {
			result.Cancelled = true
			return result, fmt.Errorf("sync cancelled")
		}
		return result, fmt.Errorf("failed to fetch issues: %w", err)
	}

	result.TotalIssues = len(issues)

	// Fetch comments for each issue
	commentWG := sync.WaitGroup{}
	issuesChan := make(chan storage.Issue, len(issues))
	errorsChan := make(chan error, len(issues))

	// Start worker goroutines to fetch comments
	workerCount := 5
	for i := 0; i < workerCount; i++ {
		commentWG.Add(1)
		go func() {
			defer commentWG.Done()
			for issue := range issuesChan {
				comments, err := s.client.FetchComments(issue.Number)
				if err != nil {
					errorsChan <- fmt.Errorf("failed to fetch comments for issue %d: %w", issue.Number, err)
					continue
				}

				// Store comments
				for _, comment := range comments {
					if err := storage.StoreComment(s.db, &comment); err != nil {
						errorsChan <- fmt.Errorf("failed to store comment %d: %w", comment.ID, err)
					}
				}

				result.CommentsFetched += len(comments)
			}
		}()
	}

	// Send issues to workers
	for _, issue := range issues {
		// Check for cancellation
		select {
		case <-ctx.Done():
			close(issuesChan)
			commentWG.Wait()
			close(errorsChan)
			result.Cancelled = true
			return result, fmt.Errorf("sync cancelled")
		default:
		}

		// Store issue
		if err := storage.StoreIssue(s.db, &issue); err != nil {
			cancel()
			close(issuesChan)
			commentWG.Wait()
			close(errorsChan)
			return result, fmt.Errorf("failed to store issue %d: %w", issue.Number, err)
		}
		result.IssuesStored++

		issuesChan <- issue
	}

	close(issuesChan)
	commentWG.Wait()
	close(errorsChan)

	// Collect any errors
	for err := range errorsChan {
		result.Errors = append(result.Errors, err)
	}

	// Update last sync time
	if err := storage.UpdateLastSync(s.db, time.Now()); err != nil {
		return result, fmt.Errorf("failed to update last sync time: %w", err)
	}

	close(progressChan)
	wg.Wait()

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	return result, nil
}

// SyncResult contains statistics about the sync operation
type SyncResult struct {
	StartTime        time.Time
	EndTime          time.Time
	Duration         time.Duration
	TotalIssues      int
	IssuesStored     int
	CommentsFetched  int
	Errors           []error
	Cancelled        bool
}

// displayProgress shows a progress bar during sync
func (s *Syncer) displayProgress(ctx context.Context, progressChan <-chan int, result *SyncResult) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	lastCount := 0
	for {
		select {
		case <-ctx.Done():
			// Clear progress line
			fmt.Print("\r\033[K")
			return

		case count, ok := <-progressChan:
			if !ok {
				// Progress channel closed, show final status
				fmt.Print("\r\033[K")
				fmt.Printf("✓ Fetched %d issues\n", result.TotalIssues)
				return
			}

			lastCount = count
			s.printProgress(count)

		case <-ticker.C:
			// Refresh progress display
			s.printProgress(lastCount)
		}
	}
}

// printProgress prints a progress bar
func (s *Syncer) printProgress(count int) {
	fmt.Printf("\r\033[K") // Clear line
	fmt.Printf("Fetching issues... %d", count)
}

// GetDB is a helper to get the database from a syncer
func (s *Syncer) GetDB() *sql.DB {
	return s.db
}
