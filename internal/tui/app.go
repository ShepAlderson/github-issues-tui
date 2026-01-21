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
	config        *config.Config
	dbManager     *database.DBManager
	issueList     *IssueList
	issueDetail   *IssueDetailComponent
	width         int
	height        int
	ready         bool
}

// NewApp creates a new TUI application instance
func NewApp(cfg *config.Config, dbManager *database.DBManager, cfgMgr *config.Manager) *App {
	return &App{
		config:      cfg,
		dbManager:   dbManager,
		issueList:   NewIssueList(dbManager, cfg, cfgMgr),
		issueDetail: NewIssueDetailComponent(dbManager),
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
		if a.issueDetail != nil {
			var cmd tea.Cmd
			a.issueDetail, cmd = a.issueDetail.Update(msg)
			return a, cmd
		}
		return a, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		case "enter":
			// Handle Enter key for comments view (future US-008)
			// For now, just pass to issue detail
			if a.issueDetail != nil {
				var cmd tea.Cmd
				a.issueDetail, cmd = a.issueDetail.Update(msg)
				return a, cmd
			}
			return a, nil
		}
	}

	// Update issue list and check if selection changed
	var cmd tea.Cmd
	previousSelected := a.issueList.SelectedIssue()
	a.issueList, cmd = a.issueList.Update(msg)
	currentSelected := a.issueList.SelectedIssue()

	// If selection changed, update issue detail
	if a.issueDetail != nil && currentSelected != nil &&
	   (previousSelected == nil || previousSelected.Number != currentSelected.Number) {
		// Load the selected issue in detail view
		if err := a.issueDetail.SetIssue(currentSelected.Number); err != nil {
			// Log error but continue
			// In production, we might want to show an error message
		}
	}

	// Also pass message to issue detail for its own keybindings
	if a.issueDetail != nil {
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
	if a.issueDetail != nil {
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
