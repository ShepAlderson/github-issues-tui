package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/shepbook/ghissues/internal/auth"
	"github.com/shepbook/ghissues/internal/config"
	"github.com/shepbook/ghissues/internal/db"
	"github.com/shepbook/ghissues/internal/github"
)

// ProgressCallback is called during sync to report progress
type ProgressCallback func(current, total int, status string)

// RefreshSync performs an incremental sync that:
// - Fetches only issues updated since the last sync
// - Removes issues that have been deleted from GitHub
// - Updates comments for existing issues
// - Calls the progress callback with current status
func RefreshSync(dbPath string, cfg *config.Config, progress ProgressCallback) error {
	// Parse owner/repo
	parts := strings.SplitN(cfg.Repository, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository format: %q (expected owner/repo)", cfg.Repository)
	}
	owner, repo := parts[0], parts[1]

	// Open database
	database, err := db.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	// Get authentication token
	token, source, err := auth.GetToken()
	if err != nil {
		return err
	}
	if progress != nil {
		progress(0, 0, fmt.Sprintf("Using GitHub token from %s...", source))
	}

	// Create GitHub client
	client := github.NewClient(token)

	// Get last sync time
	lastSync, err := db.GetLastSyncTime(database, owner, repo)
	if err != nil {
		return fmt.Errorf("failed to get last sync time: %w", err)
	}

	// Get current issue numbers before fetching (to detect deleted issues)
	existingNumbers, err := db.GetIssueNumbers(database, owner, repo)
	if err != nil {
		return fmt.Errorf("failed to get existing issue numbers: %w", err)
	}

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Fetch issues updated since last sync
	if progress != nil {
		progress(0, 0, fmt.Sprintf("Fetching issues updated since %s...", lastSync))
	}

	var issues []github.Issue
	var totalCount int
	if lastSync == "" {
		// First sync - fetch all issues
		issues, totalCount, err = client.FetchIssues(ctx, owner, repo)
	} else {
		// Incremental sync
		issues, totalCount, err = client.FetchIssuesSince(ctx, owner, repo, lastSync)
	}
	if err != nil {
		if ctx.Err() == context.Canceled {
			return fmt.Errorf("sync cancelled")
		}
		return fmt.Errorf("failed to fetch issues: %w", err)
	}

	if progress != nil {
		progress(0, totalCount, fmt.Sprintf("Processing %d issues...", len(issues)))
	}

	// Track which issues are still present
	stillPresent := make(map[int]bool)

	// Process each issue
	for i := range issues {
		select {
		case <-ctx.Done():
			return fmt.Errorf("sync cancelled")
		default:
		}

		issue := &issues[i]
		stillPresent[issue.Number] = true

		// Delete existing labels and assignees before upserting
		if err := db.DeleteLabels(database, issue.Number); err != nil {
			return fmt.Errorf("failed to delete labels for issue %d: %w", issue.Number, err)
		}
		if err := db.DeleteAssignees(database, issue.Number); err != nil {
			return fmt.Errorf("failed to delete assignees for issue %d: %w", issue.Number, err)
		}

		// Upsert the issue
		if err := db.UpsertIssue(database, owner, repo, issue); err != nil {
			return fmt.Errorf("failed to store issue %d: %w", issue.Number, err)
		}

		// Insert labels
		for j := range issue.Labels {
			if err := db.InsertLabel(database, issue.Number, &issue.Labels[j]); err != nil {
				return fmt.Errorf("failed to insert label for issue %d: %w", issue.Number, err)
			}
		}

		// Insert assignees
		for j := range issue.Assignees {
			if err := db.InsertAssignee(database, issue.Number, &issue.Assignees[j]); err != nil {
				return fmt.Errorf("failed to insert assignee for issue %d: %w", issue.Number, err)
			}
		}

		// Always fetch and update comments for updated issues
		comments, err := client.FetchComments(ctx, owner, repo, issue.Number)
		if err != nil {
			if ctx.Err() == context.Canceled {
				return fmt.Errorf("sync cancelled")
			}
			return fmt.Errorf("failed to fetch comments for issue %d: %w", issue.Number, err)
		}

		for j := range comments {
			if err := db.UpsertComment(database, issue.Number, &comments[j]); err != nil {
				return fmt.Errorf("failed to store comment for issue %d: %w", issue.Number, err)
			}
		}

		if progress != nil {
			progress(i+1, len(issues), fmt.Sprintf("Processed issue #%d", issue.Number))
		}
	}

	// Delete issues that are no longer present in GitHub
	var deletedCount int
	for _, num := range existingNumbers {
		if !stillPresent[num] {
			if err := db.DeleteIssue(database, owner, repo, num); err != nil {
				return fmt.Errorf("failed to delete issue %d: %w", num, err)
			}
			deletedCount++
		}
	}

	if progress != nil {
		progress(len(issues), len(issues), fmt.Sprintf("Sync complete! %d issues updated, %d deleted.", len(issues), deletedCount))
	}

	return nil
}

// runSync fetches issues and comments from GitHub and stores them in the database
func runSync() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate repository format
	if !strings.Contains(cfg.Repository, "/") {
		return fmt.Errorf("invalid repository format: %q (expected owner/repo)", cfg.Repository)
	}
	parts := strings.SplitN(cfg.Repository, "/", 2)
	owner, repo := parts[0], parts[1]

	// Get database path
	dbPath, err := db.GetPath(dbFlag, cfg)
	if err != nil {
		return fmt.Errorf("failed to get database path: %w", err)
	}

	// Ensure database directory exists
	if err := db.EnsureDir(dbPath); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database
	database, err := db.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	// Get authentication token
	token, source, err := auth.GetToken()
	if err != nil {
		return err
	}
	fmt.Printf("Using GitHub token from %s...\n", source)

	// Create GitHub client
	client := github.NewClient(token)

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nCancelling sync...")
		cancel()
	}()

	fmt.Printf("Fetching issues from %s/%s...\n", owner, repo)

	// Fetch issues with progress
	progress := &ProgressBar{Total: 0, Current: 0}
	issues, totalCount, err := client.FetchIssues(ctx, owner, repo)
	if err != nil {
		if ctx.Err() == context.Canceled {
			fmt.Println("Sync cancelled.")
			return nil
		}
		return fmt.Errorf("failed to fetch issues: %w", err)
	}

	progress.Total = totalCount
	progress.Show()

	// Store issues and fetch comments
	for i := range issues {
		select {
		case <-ctx.Done():
			fmt.Println("\nSync cancelled.")
			return nil
		default:
		}

		issue := &issues[i]

		// Delete existing labels and assignees before upserting
		if err := db.DeleteLabels(database, issue.Number); err != nil {
			return fmt.Errorf("failed to delete labels for issue %d: %w", issue.Number, err)
		}
		if err := db.DeleteAssignees(database, issue.Number); err != nil {
			return fmt.Errorf("failed to delete assignees for issue %d: %w", issue.Number, err)
		}

		// Upsert the issue
		if err := db.UpsertIssue(database, owner, repo, issue); err != nil {
			return fmt.Errorf("failed to store issue %d: %w", issue.Number, err)
		}

		// Insert labels
		for j := range issue.Labels {
			if err := db.InsertLabel(database, issue.Number, &issue.Labels[j]); err != nil {
				return fmt.Errorf("failed to insert label for issue %d: %w", issue.Number, err)
			}
		}

		// Insert assignees
		for j := range issue.Assignees {
			if err := db.InsertAssignee(database, issue.Number, &issue.Assignees[j]); err != nil {
				return fmt.Errorf("failed to insert assignee for issue %d: %w", issue.Number, err)
			}
		}

		// Fetch and store comments if there are any
		if issue.Comments > 0 {
			comments, err := client.FetchComments(ctx, owner, repo, issue.Number)
			if err != nil {
				if ctx.Err() == context.Canceled {
					fmt.Println("\nSync cancelled.")
					return nil
				}
				return fmt.Errorf("failed to fetch comments for issue %d: %w", issue.Number, err)
			}

			for j := range comments {
				if err := db.UpsertComment(database, issue.Number, &comments[j]); err != nil {
					return fmt.Errorf("failed to store comment for issue %d: %w", issue.Number, err)
				}
			}
		}

		progress.Current++
		progress.Update()
	}

	progress.Finish()
	fmt.Printf("\nSync complete! Fetched %d issues.\n", len(issues))

	return nil
}

// ProgressBar displays a progress bar in the terminal
type ProgressBar struct {
	Total   int
	Current int
	width   int
}

func (p *ProgressBar) Show() {
	p.width = 40
	fmt.Printf("\rFetching: [%s] %d/%d", strings.Repeat("=", p.width), p.Current, p.Total)
}

func (p *ProgressBar) Update() {
	if p.Total == 0 {
		fmt.Printf("\rFetching: [%s] %d/?", strings.Repeat("=", p.width), p.Current)
		return
	}

	percent := float64(p.Current) / float64(p.Total)
	filled := int(float64(p.width) * percent)
	empty := p.width - filled

	fmt.Printf("\rFetching: [%s%s] %d/%d",
		strings.Repeat("=", filled),
		strings.Repeat(" ", empty),
		p.Current,
		p.Total)
}

func (p *ProgressBar) Finish() {
	if p.Total > 0 {
		fmt.Printf("\rFetching: [%s] %d/%d - Done!\n", strings.Repeat("=", p.width), p.Current, p.Total)
	} else {
		fmt.Printf("\rFetching: [%s] %d/? - Done!\n", strings.Repeat("=", p.width), p.Current)
	}
}

// SyncProgress represents sync progress for reporting
type SyncProgress struct {
	IssuesFetched  int
	CommentsFetched int
	TotalIssues    int
	StartTime      time.Time
}

// NewSyncProgress creates a new progress tracker
func NewSyncProgress() *SyncProgress {
	return &SyncProgress{
		StartTime: time.Now(),
	}
}

// UpdateIssue updates the progress when an issue is fetched
func (p *SyncProgress) UpdateIssue() {
	p.IssuesFetched++
}

// UpdateComment updates the progress when a comment is fetched
func (p *SyncProgress) UpdateComment() {
	p.CommentsFetched++
}

// SetTotal sets the total number of issues expected
func (p *SyncProgress) SetTotal(n int) {
	p.TotalIssues = n
}

// Summary returns a summary of the sync operation
func (p *SyncProgress) Summary() string {
	elapsed := time.Since(p.StartTime)
	return fmt.Sprintf("Issues: %d, Comments: %d, Time: %v", p.IssuesFetched, p.CommentsFetched, elapsed.Round(time.Second))
}