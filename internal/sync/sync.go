package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/shepbook/ghissues/internal/db"
	"github.com/shepbook/ghissues/internal/github"
)

// Progress contains progress information during sync
type Progress struct {
	Phase           string // "issues" or "comments"
	IssuesFetched   int
	TotalIssues     int
	CommentsFetched int
	CurrentIssue    int
}

// ProgressCallback is called during sync operations to report progress
type ProgressCallback func(Progress)

// Result contains the result of a sync operation
type Result struct {
	IssuesFetched   int
	CommentsFetched int
	Duration        time.Duration
}

// Syncer handles syncing issues from GitHub to the local database
type Syncer struct {
	client *github.Client
	store  *db.Store
}

// NewSyncer creates a new Syncer
func NewSyncer(client *github.Client, store *db.Store) *Syncer {
	return &Syncer{
		client: client,
		store:  store,
	}
}

// Sync fetches all open issues and their comments from GitHub
func (s *Syncer) Sync(ctx context.Context, owner, repo string, progress ProgressCallback) (*Result, error) {
	startTime := time.Now()
	result := &Result{}

	// Get current issue numbers from database before sync
	// This allows us to detect issues that have been closed/deleted
	existingNumbers, err := s.store.GetAllIssueNumbers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing issue numbers: %w", err)
	}
	existingSet := make(map[int]bool)
	for _, num := range existingNumbers {
		existingSet[num] = true
	}

	// Fetch issues with progress
	issues, err := s.client.FetchIssues(ctx, owner, repo, func(p github.FetchProgress) {
		if progress != nil {
			progress(Progress{
				Phase:         "issues",
				IssuesFetched: p.Fetched,
				TotalIssues:   p.Total,
			})
		}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issues: %w", err)
	}

	result.IssuesFetched = len(issues)
	result.Duration = time.Since(startTime)

	// Build set of fetched issue numbers
	fetchedSet := make(map[int]bool)
	for _, issue := range issues {
		fetchedSet[issue.Number] = true
	}

	// Find issues that were in DB but not in the fresh fetch (closed/deleted)
	var issuesToRemove []int
	for num := range existingSet {
		if !fetchedSet[num] {
			issuesToRemove = append(issuesToRemove, num)
		}
	}

	// Remove closed/deleted issues from database
	if len(issuesToRemove) > 0 {
		if err := s.store.DeleteIssues(ctx, issuesToRemove); err != nil {
			return nil, fmt.Errorf("failed to remove closed issues: %w", err)
		}
	}

	// Save issues to database
	if err := s.store.SaveIssues(ctx, issues); err != nil {
		return nil, fmt.Errorf("failed to save issues: %w", err)
	}

	// Fetch comments for issues that have them
	for i, issue := range issues {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if issue.CommentCount > 0 {
			comments, err := s.client.FetchIssueComments(ctx, owner, repo, issue.Number)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch comments for issue #%d: %w", issue.Number, err)
			}

			if err := s.store.SaveComments(ctx, issue.Number, comments); err != nil {
				return nil, fmt.Errorf("failed to save comments for issue #%d: %w", issue.Number, err)
			}

			result.CommentsFetched += len(comments)

			if progress != nil {
				progress(Progress{
					Phase:           "comments",
					IssuesFetched:   len(issues),
					TotalIssues:     len(issues),
					CommentsFetched: result.CommentsFetched,
					CurrentIssue:    i + 1,
				})
			}
		}
	}

	// Update last sync time
	if err := s.store.SetLastSyncTime(ctx, time.Now()); err != nil {
		return nil, fmt.Errorf("failed to update sync time: %w", err)
	}

	result.Duration = time.Since(startTime)
	return result, nil
}
