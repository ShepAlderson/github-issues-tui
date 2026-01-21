package tui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shepbook/github-issues-tui/internal/config"
	"github.com/shepbook/github-issues-tui/internal/database"
	"github.com/shepbook/github-issues-tui/internal/github"
	"github.com/shepbook/github-issues-tui/internal/sync"
)

// App represents the main TUI application
type App struct {
	config          *config.Config
	configManager   *config.Manager
	dbManager       *database.DBManager
	issueList       *IssueList
	issueDetail     *IssueDetailComponent
	comments        *CommentsComponent
	errors          *ErrorComponent
	help            *HelpComponent
	showComments    bool
	width           int
	height          int
	ready           bool
	refreshing     bool
	refreshMessage  string
}

// NewApp creates a new TUI application instance
func NewApp(cfg *config.Config, dbManager *database.DBManager, cfgMgr *config.Manager) *App {
	return &App{
		config:        cfg,
		configManager: cfgMgr,
		dbManager:     dbManager,
		issueList:     NewIssueList(dbManager, cfg, cfgMgr),
		issueDetail:   NewIssueDetailComponent(dbManager),
		comments:      NewCommentsComponent(dbManager),
		errors:        NewErrorComponent(),
		help:          NewHelpComponent(),
		showComments:  false,
		refreshing:    false,
		refreshMessage: "",
	}
}

// Init implements the tea.Model interface
func (a *App) Init() tea.Cmd {
	// Load issues from database
	return a.issueList.Init()
}

// Update implements the tea.Model interface
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// First, update error component to clear expired status messages
	if a.errors != nil {
		var cmd tea.Cmd
		a.errors, cmd = a.errors.Update(msg)
		if cmd != nil {
			return a, cmd
		}
	}

	// Update help component if it exists
	if a.help != nil {
		var cmd tea.Cmd
		a.help, cmd = a.help.Update(msg)
		if cmd != nil {
			return a, cmd
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.ready = true
		// Pass size to error component
		if a.errors != nil {
			a.errors.Update(msg)
		}
		// Pass size to help component
		if a.help != nil {
			a.help.Update(msg)
		}
		// Pass size to both components
		if a.showComments && a.comments != nil {
			var cmd tea.Cmd
			a.comments, cmd = a.comments.Update(msg)
			return a, cmd
		} else if a.issueDetail != nil {
			var cmd tea.Cmd
			a.issueDetail, cmd = a.issueDetail.Update(msg)
			return a, cmd
		}
		return a, nil

	case tea.KeyMsg:
		// If error modal is shown, let error component handle it first
		if a.errors != nil {
			var cmd tea.Cmd
			a.errors, cmd = a.errors.Update(msg)
			if cmd != nil {
				return a, cmd
			}
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		case "esc":
			// Escape returns from comments view to issue list view
			if a.showComments {
				a.showComments = false
				return a, nil
			}
		case "enter":
			// Enter toggles between issue detail and comments view
			if a.issueDetail != nil && a.comments != nil {
				if a.showComments {
					// If showing comments, switch back to detail view
					a.showComments = false
				} else {
					// If showing detail, switch to comments view
					selectedIssue := a.issueList.SelectedIssue()
					if selectedIssue != nil {
						// Get issue details to pass title to comments component
						issue, err := a.dbManager.GetIssueByNumber(selectedIssue.Number)
						if err == nil && issue != nil {
							if err := a.comments.SetIssue(selectedIssue.Number, issue.Title); err != nil {
								// Show error in status bar
								a.handleError(err, false)
							}
							a.showComments = true
						}
					}
				}
				return a, nil
			}
		case "r", "R":
			// Refresh data from GitHub
			if !a.refreshing {
				return a, a.startRefresh()
			}
		case "?":
			// Toggle help overlay
			if a.help != nil {
				if a.help.showHelp {
					a.help.HideHelp()
				} else {
					a.help.ShowHelp()
				}
			}
			return a, nil
		}
	}

	// Update issue list and check if selection changed
	var cmd tea.Cmd
	previousSelected := a.issueList.SelectedIssue()
	a.issueList, cmd = a.issueList.Update(msg)
	currentSelected := a.issueList.SelectedIssue()

	// If selection changed and we're in detail view, update issue detail
	if !a.showComments && a.issueDetail != nil && currentSelected != nil &&
	   (previousSelected == nil || previousSelected.Number != currentSelected.Number) {
		// Load the selected issue in detail view
		if err := a.issueDetail.SetIssue(currentSelected.Number); err != nil {
			// Show error in status bar
			a.handleError(err, false)
		}
	}

	// Pass message to active component for its own keybindings
	if a.showComments && a.comments != nil {
		// In comments view, pass messages to comments component
		var commentsCmd tea.Cmd
		a.comments, commentsCmd = a.comments.Update(msg)
		if commentsCmd != nil {
			// Combine commands if we have multiple
			if cmd != nil {
				cmd = tea.Batch(cmd, commentsCmd)
			} else {
				cmd = commentsCmd
			}
		}
	} else if a.issueDetail != nil {
		// In detail view, pass messages to issue detail component
		var detailCmd tea.Cmd
		a.issueDetail, detailCmd = a.issueDetail.Update(msg)
		if detailCmd != nil {
			// Combine commands if we have multiple
			if cmd != nil {
				cmd = tea.Batch(cmd, detailCmd)
			} else {
				cmd = detailCmd
			}
		}
	}

	return a, cmd
}

// View implements the tea.Model interface
func (a *App) View() string {
	if !a.ready {
		return "Initializing..."
	}

	// Create vertical split layout
	leftPanel := a.issueList.View()
	rightPanel := a.renderRightPanel()

	// Calculate panel widths (left panel takes 60%, right panel takes 40%)
	leftWidth := int(float64(a.width) * 0.6)
	rightWidth := a.width - leftWidth - 1 // -1 for border

	// Style panels
	leftPanelStyled := lipgloss.NewStyle().
		Width(leftWidth).
		Height(a.height).
		Render(leftPanel)

	rightPanelStyled := lipgloss.NewStyle().
		Width(rightWidth).
		Height(a.height).
		Render(rightPanel)

	// Combine with vertical separator
	mainView := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanelStyled,
		lipgloss.NewStyle().
			Border(lipgloss.ThickBorder(), false, true, false, false).
			Height(a.height).
			Render(""),
		rightPanelStyled,
	)

	// Render help component on top (takes precedence over errors)
	if a.help != nil {
		helpView := a.help.View()
		if helpView != "" {
			// The help component handles its own positioning with lipgloss.Place
			// So we need to render it separately
			return helpView
		}
	}

	// Render error component on top
	if a.errors != nil {
		errorView := a.errors.View()
		if errorView != "" {
			// The error component handles its own positioning with lipgloss.Place
			// So we need to render it separately
			return errorView
		}
	}

	return mainView
}

func (a *App) renderRightPanel() string {
	// Show refresh status when refreshing
	if a.refreshing {
		return lipgloss.NewStyle().
			Padding(1, 2).
			Foreground(lipgloss.Color("yellow")).
			Render(fmt.Sprintf("🔄 %s\n\nFetching latest issues from GitHub...", a.refreshMessage))
	}

	if a.showComments && a.comments != nil {
		return a.comments.View()
	} else if a.issueDetail != nil {
		return a.issueDetail.View()
	}

	// Fallback placeholder
	return lipgloss.NewStyle().
		Padding(1, 2).
		Render("Select an issue to view details\n\nPress 'q' or Ctrl+C to quit")
}

// Run starts the TUI application
func Run(cfg *config.Config, dbManager *database.DBManager, cfgMgr *config.Manager) error {
	app := NewApp(cfg, dbManager, cfgMgr)
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}
	return nil
}

// handleError handles an error by categorizing it and displaying appropriately
func (a *App) handleError(err error, forceCritical bool) {
	if a.errors == nil {
		return
	}

	errStr := err.Error()
	title, message, isCritical := CategorizeError(errStr)

	// Force critical if requested
	if forceCritical {
		isCritical = true
	}

	if isCritical {
		a.errors.ShowModal(title, message)
	} else {
		// Show in status bar for 5 seconds
		a.errors.ShowStatus(fmt.Sprintf("%s: %s", title, errStr), 5*time.Second)
	}
}

// startRefresh initiates a background refresh operation
func (a *App) startRefresh() tea.Cmd {
	return func() tea.Msg {
		a.refreshing = true
		a.refreshMessage = "Starting refresh..."

		// Create auth manager
		authManager := github.NewAuthManager(a.configManager)

		// Create sync manager
		syncManager := sync.NewSyncManager(a.configManager, authManager, a.dbManager)

		// Create context for sync operation
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Perform sync
		a.refreshMessage = "Fetching issues from GitHub..."
		err := syncManager.Sync(ctx, sync.SyncOptions{
			ShowProgress: false, // We show progress in TUI instead
			CancelChan:   ctx.Done(),
		})

		if err != nil {
			a.refreshMessage = "Refresh failed"
			// Show error with categorization
			a.handleError(err, false)
		} else {
			a.refreshMessage = "Refresh complete!"
		}

		// Reload issues from database
		a.issueList.loadIssues()

		a.refreshing = false

		// Clear message after delay
		go func() {
			time.Sleep(3 * time.Second)
			a.refreshMessage = ""
		}()

		return nil
	}
}
