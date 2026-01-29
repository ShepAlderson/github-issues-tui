package sync

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/shepbook/ghissues/internal/database"
	"github.com/shepbook/ghissues/internal/github"
)

// Status represents the current sync state
type Status int

const (
	StatusIdle Status = iota
	StatusSyncing
	StatusComplete
	StatusError
	StatusCancelled
)

// SyncProgress contains the final sync results
type SyncProgress struct {
	IssuesFetched   int
	CommentsFetched int
	Error           error
	Duration        string
}

// SyncModel represents the sync TUI state
type SyncModel struct {
	db     *sql.DB
	dbPath string
	repo   string
	token  string

	status          Status
	progress        progress.Model
	issuesFetched   int
	issuesTotal     int
	commentsFetched int
	current         string
	err             error
	startTime       int64
	duration        string

	cancelled bool
}

// syncMsg represents an update during sync
type syncMsg struct {
	issuesFetched   int
	issuesTotal     int
	commentsFetched int
	current         string
	status          Status
	err             error
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D4AA"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B"))

	progressStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4"))
)

// NewSyncModel creates a new sync model
func NewSyncModel(dbPath, repo, token string) SyncModel {
	return SyncModel{
		dbPath:    dbPath,
		repo:      repo,
		token:     token,
		status:    StatusIdle,
		progress:  progress.New(progress.WithDefaultGradient()),
		cancelled: false,
	}
}

// Init initializes the sync model
func (m SyncModel) Init() tea.Cmd {
	return tea.Batch(
		m.initializeDatabase(),
		m.startSync(),
	)
}

// Update handles messages
func (m SyncModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.cancelled = true
			m.status = StatusCancelled
			return m, tea.Quit

		case tea.KeyRunes:
			if msg.String() == "q" || msg.String() == "Q" {
				// Only allow quit when done or error
				if m.status == StatusComplete || m.status == StatusError || m.status == StatusCancelled {
					return m, tea.Quit
				}
			}
		}

	case syncMsg:
		m.issuesFetched = msg.issuesFetched
		m.issuesTotal = msg.issuesTotal
		m.commentsFetched = msg.commentsFetched
		m.current = msg.current
		m.status = msg.status
		m.err = msg.err

		if m.status == StatusComplete || m.status == StatusError {
			return m, tea.Quit
		}

		// Update progress bar
		cmd := m.progress.SetPercent(m.progressPercent())
		return m, cmd

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - 4
		if m.progress.Width > 60 {
			m.progress.Width = 60
		}
		return m, nil
	}

	return m, nil
}

// View renders the sync UI
func (m SyncModel) View() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(titleStyle.Render("📥 Syncing GitHub Issues"))
	b.WriteString("\n\n")

	switch m.status {
	case StatusIdle:
		b.WriteString("Initializing...\n")

	case StatusSyncing:
		b.WriteString(subtitleStyle.Render(fmt.Sprintf("Repository: %s", m.repo)))
		b.WriteString("\n\n")
		b.WriteString(progressStyle.Render(m.progress.View()))
		b.WriteString("\n\n")

		if m.issuesTotal > 0 {
			b.WriteString(fmt.Sprintf("Issues: %d / %d\n", m.issuesFetched, m.issuesTotal))
		} else {
			b.WriteString(fmt.Sprintf("Issues: %d fetched\n", m.issuesFetched))
		}

		if m.commentsFetched > 0 {
			b.WriteString(fmt.Sprintf("Comments: %d fetched\n", m.commentsFetched))
		}

		b.WriteString("\n")
		b.WriteString(subtitleStyle.Render(m.current))
		b.WriteString("\n")

		b.WriteString("\n")
		b.WriteString(subtitleStyle.Render("Press Ctrl+C to cancel"))

	case StatusComplete:
		b.WriteString(statusStyle.Render("✓ Sync complete!\n"))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("Issues synced: %d\n", m.issuesFetched))
		if m.commentsFetched > 0 {
			b.WriteString(fmt.Sprintf("Comments synced: %d\n", m.commentsFetched))
		}
		b.WriteString("\n")
		b.WriteString(subtitleStyle.Render("Press 'q' to quit"))

	case StatusError:
		b.WriteString(errorStyle.Render("✗ Sync failed\n"))
		b.WriteString("\n")
		if m.err != nil {
			b.WriteString(fmt.Sprintf("Error: %v\n", m.err))
		}
		b.WriteString("\n")
		b.WriteString(subtitleStyle.Render("Press 'q' to quit"))

	case StatusCancelled:
		b.WriteString(subtitleStyle.Render("⚠ Sync cancelled\n"))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("Issues fetched: %d\n", m.issuesFetched))
		if m.commentsFetched > 0 {
			b.WriteString(fmt.Sprintf("Comments fetched: %d\n", m.commentsFetched))
		}
		b.WriteString("\n")
		b.WriteString(subtitleStyle.Render("Press 'q' to quit"))
	}

	b.WriteString("\n")
	return b.String()
}

// initializeDatabase creates or opens the database
func (m SyncModel) initializeDatabase() tea.Cmd {
	return func() tea.Msg {
		db, err := database.InitializeSchema(m.dbPath)
		if err != nil {
			return syncMsg{
				status: StatusError,
				err:    fmt.Errorf("failed to initialize database: %w", err),
			}
		}
		m.db = db
		return syncMsg{
			status:  StatusIdle,
			current: "Database ready",
		}
	}
}

// startSync begins the sync process
func (m SyncModel) startSync() tea.Cmd {
	return func() tea.Msg {
		if m.db == nil {
			// Database not ready yet, will retry
			return syncMsg{
				status:  StatusIdle,
				current: "Waiting for database...",
			}
		}

		client := github.NewClient(m.token)

		// Create progress channel
		progressChan := make(chan github.FetchProgress, 10)

		// Start fetching in goroutine
		go func() {
			defer close(progressChan)

			issues, err := client.FetchIssues(m.repo, progressChan)
			if err != nil {
				progressChan <- github.FetchProgress{
					Current: fmt.Sprintf("Error: %v", err),
				}
				return
			}

			// Save issues to database
			for _, issue := range issues {
				if err := database.SaveIssue(m.db, m.repo, issue); err != nil {
					progressChan <- github.FetchProgress{
						Current: fmt.Sprintf("Error saving issue #%d: %v", issue.Number, err),
					}
					continue
				}

				// Fetch comments if there are any
				if issue.CommentCount > 0 {
					comments, err := client.FetchComments(m.repo, issue.Number, nil)
					if err != nil {
						progressChan <- github.FetchProgress{
							Current: fmt.Sprintf("Error fetching comments for #%d: %v", issue.Number, err),
						}
						continue
					}

					for _, comment := range comments {
						if err := database.SaveComment(m.db, m.repo, comment); err != nil {
							progressChan <- github.FetchProgress{
								Current: fmt.Sprintf("Error saving comment: %v", err),
							}
						}
					}
				}
			}
		}()

		// Process progress updates
		issuesFetched := 0
		commentsFetched := 0
		issuesTotal := 0

		for progress := range progressChan {
			if m.cancelled {
				return syncMsg{
					issuesFetched:   issuesFetched,
					commentsFetched: commentsFetched,
					status:          StatusCancelled,
					current:         "Sync cancelled",
				}
			}

			if progress.Fetched > 0 {
				issuesFetched = progress.Fetched
			}
			if progress.Total > issuesTotal {
				issuesTotal = progress.Total
			}
		}

		return syncMsg{
			issuesFetched:   issuesFetched,
			issuesTotal:     issuesTotal,
			commentsFetched: commentsFetched,
			status:          StatusComplete,
			current:         "Sync complete",
		}
	}
}

// progressPercent returns the current progress as a float between 0 and 1
func (m SyncModel) progressPercent() float64 {
	if m.issuesTotal == 0 {
		return 0
	}
	percent := float64(m.issuesFetched) / float64(m.issuesTotal)
	if percent > 1 {
		return 1
	}
	return percent
}

// Progress returns the final sync results
func (m SyncModel) Progress() (SyncProgress, error) {
	return SyncProgress{
		IssuesFetched:   m.issuesFetched,
		CommentsFetched: m.commentsFetched,
		Error:           m.err,
		Duration:        m.duration,
	}, nil
}

// RunSync runs the sync TUI and returns the results
func RunSync(dbPath, repo, token string) (SyncProgress, error) {
	// Validate database path is writable
	if err := database.EnsureWritable(dbPath); err != nil {
		return SyncProgress{}, fmt.Errorf("database path is not writable: %w", err)
	}

	// Check for authentication
	if token == "" {
		resolvedToken, err := github.ResolveToken()
		if err != nil {
			return SyncProgress{}, fmt.Errorf("authentication required: %w", err)
		}
		token = resolvedToken
	}

	model := NewSyncModel(dbPath, repo, token)
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return SyncProgress{}, fmt.Errorf("error running sync: %w", err)
	}

	syncModel, ok := finalModel.(SyncModel)
	if !ok {
		return SyncProgress{}, fmt.Errorf("unexpected model type")
	}

	progress, _ := syncModel.Progress()
	return progress, nil
}

// RunSyncCLI runs sync in non-interactive mode for use in CLI
func RunSyncCLI(dbPath, repo, token string) error {
	// Validate database path is writable
	if err := database.EnsureWritable(dbPath); err != nil {
		return fmt.Errorf("database path is not writable: %w", err)
	}

	// Check for authentication
	if token == "" {
		resolvedToken, err := github.ResolveToken()
		if err != nil {
			return fmt.Errorf("authentication required: %w", err)
		}
		token = resolvedToken
	}

	fmt.Printf("Syncing issues from %s...\n", repo)

	// Initialize database
	db, err := database.InitializeSchema(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	client := github.NewClient(token)

	// Fetch issues with progress
	progressChan := make(chan github.FetchProgress, 10)
	done := make(chan bool)

	// Start progress display
	go func() {
		for progress := range progressChan {
			if progress.Current != "" {
				fmt.Printf("\r%s", progress.Current)
			}
		}
		done <- true
	}()

	issues, err := client.FetchIssues(repo, progressChan)
	if err != nil {
		return fmt.Errorf("failed to fetch issues: %w", err)
	}

	<-done
	fmt.Printf("\nFetched %d issues\n", len(issues))

	// Save issues and comments
	commentsFetched := 0
	for _, issue := range issues {
		if err := database.SaveIssue(db, repo, issue); err != nil {
			return fmt.Errorf("failed to save issue #%d: %w", issue.Number, err)
		}

		if issue.CommentCount > 0 {
			comments, err := client.FetchComments(repo, issue.Number, nil)
			if err != nil {
				return fmt.Errorf("failed to fetch comments for issue #%d: %w", issue.Number, err)
			}

			for _, comment := range comments {
				if err := database.SaveComment(db, repo, comment); err != nil {
					return fmt.Errorf("failed to save comment: %w", err)
				}
			}
			commentsFetched += len(comments)
		}
	}

	fmt.Printf("Saved %d issues and %d comments to database\n", len(issues), commentsFetched)
	return nil
}
