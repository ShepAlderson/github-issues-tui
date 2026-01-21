package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shepbook/github-issues-tui/internal/config"
	"github.com/shepbook/github-issues-tui/internal/database"
)

// App represents the main TUI application
type App struct {
	config          *config.Config
	dbManager       *database.DBManager
	issueList       *IssueList
	issueDetail     *IssueDetailComponent
	comments        *CommentsComponent
	showComments    bool
	width           int
	height          int
	ready           bool
}

// NewApp creates a new TUI application instance
func NewApp(cfg *config.Config, dbManager *database.DBManager, cfgMgr *config.Manager) *App {
	return &App{
		config:        cfg,
		dbManager:     dbManager,
		issueList:     NewIssueList(dbManager, cfg, cfgMgr),
		issueDetail:   NewIssueDetailComponent(dbManager),
		comments:      NewCommentsComponent(dbManager),
		showComments:  false,
	}
}

// Init implements the tea.Model interface
func (a *App) Init() tea.Cmd {
	// Load issues from database
	return a.issueList.Init()
}

// Update implements the tea.Model interface
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.ready = true
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
								// Log error but continue
							}
							a.showComments = true
						}
					}
				}
				return a, nil
			}
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
			// Log error but continue
			// In production, we might want to show an error message
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
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanelStyled,
		lipgloss.NewStyle().
			Border(lipgloss.ThickBorder(), false, true, false, false).
			Height(a.height).
			Render(""),
		rightPanelStyled,
	)
}

func (a *App) renderRightPanel() string {
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
