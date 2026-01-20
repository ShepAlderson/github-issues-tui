package tui

import (
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shepbook/git/github-issues-tui/internal/config"
)

// ViewType represents which view is currently active
type ViewType int

const (
	ListView ViewType = iota
	DetailView
	CommentsView
)

// String returns a string representation of the ViewType
func (v ViewType) String() string {
	switch v {
	case ListView:
		return "ListView"
	case DetailView:
		return "DetailView"
	case CommentsView:
		return "CommentsView"
	default:
		return "Unknown"
	}
}

// AppModel represents the main application model that can switch between views
type AppModel struct {
	currentView   ViewType
	listModel     *ListModel
	detailModel   *DetailModel
	commentsModel *CommentsModel
	helpModel     *HelpModel
	config        *config.Config
	dbPath        string
	err           error
	width         int
	height        int
	errorMsg      ErrorMessage
	errorModal    ModalDialog
}

// NewAppModel creates a new application model
func NewAppModel(dbPath string, cfg *config.Config) (*AppModel, error) {
	// First try to create the list model
	listModel, err := NewListModel(dbPath, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create list model: %w", err)
	}

	return &AppModel{
		currentView: ListView,
		listModel:   listModel,
		detailModel: nil,
		config:      cfg,
		dbPath:      dbPath,
	}, nil
}

// Init initializes the model
func (m AppModel) Init() tea.Cmd {
	// Initialize the current view
	if m.currentView == ListView && m.listModel != nil {
		return m.listModel.Init()
	}
	return nil
}

// helpDismissedMsg is sent when user dismisses the help overlay
type helpDismissedMsg struct{}

// Update handles messages and updates the model
func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle error acknowledgment first
	if _, ok := msg.(errorAcknowledgedMsg); ok {
		m.clearError()
		return m, nil
	}

	// Handle help dismissal
	if _, ok := msg.(helpDismissedMsg); ok {
		m.helpModel = nil
		return m, nil
	}

	// Handle window size
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height
	}

	// Check if help is active and should handle input
	if m.helpModel != nil && m.helpModel.IsActive() {
		updated, cmd := m.helpModel.Update(msg)
		m.helpModel = updated.(*HelpModel)
		// If help was dismissed, create a message
		if !m.helpModel.IsActive() {
			return m, func() tea.Msg { return helpDismissedMsg{} }
		}
		return m, cmd
	}

	// Check if error modal is active and should handle input
	if m.errorModal.IsActive() {
		updated, cmd := m.errorModal.Update(msg)
		m.errorModal = *(updated.(*ModalDialog))
		return m, cmd
	}

	// Normal view switching logic continues...
	switch m.currentView {
	case ListView:
		// Handle list view
		updated, viewCmd := m.listModel.Update(msg)
		m.listModel = updated.(*ListModel)

		// Check if help should be shown
		if m.listModel.ShouldShowHelp() {
			m.helpModel = NewHelpModel(ListView, m.width, m.height, m.getTheme())
			m.listModel.ClearHelpFlag()
			return m, tea.Batch(viewCmd, m.helpModel.Init())
		}

		// Check if user pressed Enter on an issue
		if msg, ok := msg.(tea.KeyMsg); ok && msg.Type == tea.KeyEnter {
			// Get the selected issue
			selectedIssue := m.listModel.selected
			if selectedIssue < len(m.listModel.issues) && selectedIssue >= 0 {
				issueNum := m.listModel.issues[selectedIssue].Number

				// Create detail model for the selected issue
				detailModel, err := NewDetailModel(m.dbPath, issueNum)
				if err != nil {
					m.err = fmt.Errorf("failed to create detail view: %w", err)
					return m, viewCmd
				}

				// Switch to detail view
				m.currentView = DetailView
				m.detailModel = detailModel
				return m, tea.Batch(viewCmd, detailModel.Init())
			}
		}

		return m, viewCmd

	case DetailView:
		// Handle detail view
		if m.detailModel != nil {
			updated, cmd := m.detailModel.Update(msg)
			m.detailModel = updated.(*DetailModel)

			// Check if help should be shown
			if m.detailModel.ShouldShowHelp() {
				m.helpModel = NewHelpModel(DetailView, m.width, m.height, m.getTheme())
				m.detailModel.ClearHelpFlag()
				return m, tea.Batch(cmd, m.helpModel.Init())
			}

			// Check if user wants to view comments
			if m.detailModel.viewComments {
				// Get the issue to pass to comments model
				issue := m.detailModel.issue

				// Close detail model
				m.detailModel.Close()

				// Create comments model
				commentsModel, err := NewCommentsModel(m.dbPath, issue)
				if err != nil {
					m.err = fmt.Errorf("failed to create comments view: %w", err)
					m.currentView = ListView
					m.detailModel = nil
					return m, nil
				}

				// Switch to comments view
				m.currentView = CommentsView
				m.commentsModel = commentsModel
				m.detailModel = nil
				return m, tea.Batch(cmd, commentsModel.Init())
			}

			// Check if user wants to go back (press 'q' or 'esc')
			if msg, ok := msg.(tea.KeyMsg); ok {
				if msg.String() == "q" || msg.String() == "esc" {
					// Quit detail model
					m.detailModel.Close()

					// Switch back to list view
					m.currentView = ListView
					m.detailModel = nil

					// Return to list view
					return m, nil
				}
			}

			return m, cmd
		}

	case CommentsView:
		// Handle comments view
		if m.commentsModel != nil {
			updated, cmd := m.commentsModel.Update(msg)
			m.commentsModel = updated.(*CommentsModel)

			// Check if help should be shown
			if m.commentsModel.ShouldShowHelp() {
				m.helpModel = NewHelpModel(CommentsView, m.width, m.height, m.getTheme())
				m.commentsModel.ClearHelpFlag()
				return m, tea.Batch(cmd, m.helpModel.Init())
			}

			// Check if user wants to go back (press 'q', 'esc', or Ctrl+C)
			if msg, ok := msg.(tea.KeyMsg); ok {
				if msg.String() == "q" || msg.String() == "esc" || msg.Type == tea.KeyCtrlC {
					// Quit comments model
					m.commentsModel.Close()

					// Switch back to detail view
					m.currentView = DetailView

					// Recreate detail model for the same issue
					detailModel, err := NewDetailModel(m.dbPath, m.commentsModel.issue.Number)
					if err != nil {
						m.err = fmt.Errorf("failed to recreate detail view: %w", err)
						m.currentView = ListView
						m.commentsModel = nil
						return m, nil
					}

					m.detailModel = detailModel
					m.commentsModel = nil
					return m, tea.Batch(cmd, detailModel.Init())
				}
			}

			return m, cmd
		}
	}

	return m, nil
}

// setError sets an error and creates modal if critical
func (m *AppModel) setError(err error) {
	m.errorMsg = newErrorMessage(err)
	if m.errorMsg.severity == ErrorSeverityCritical {
		m.errorModal = NewModalDialog(m.errorMsg, m.width, m.height)
	}
}

// clearError clears the current error
func (m *AppModel) clearError() {
	m.errorMsg = ErrorMessage{}
	m.errorModal = ModalDialog{}
}

// View renders the UI
func (m AppModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	// Show help overlay if active
	if m.helpModel != nil && m.helpModel.IsActive() {
		if os.Getenv("GHISSIES_TEST") == "1" {
			return "Help overlay displayed\n"
		}
		return m.helpModel.View()
	}

	// Show critical error modal if active
	if m.errorModal.IsActive() {
		if os.Getenv("GHISSIES_TEST") == "1" {
			return "Error modal displayed: " + m.errorMsg.userMsg + "\n"
		}
		return m.errorModal.View()
	}

	// Show main content
	var content string
	switch m.currentView {
	case ListView:
		if m.listModel != nil {
			content = m.listModel.View()
		} else {
			content = "List view not initialized\n"
		}

	case DetailView:
		if m.detailModel != nil {
			content = m.detailModel.View()
		} else {
			content = "Detail view not initialized\n"
		}

	case CommentsView:
		if m.commentsModel != nil {
			content = m.commentsModel.View()
		} else {
			content = "Comments view not initialized\n"
		}

	default:
		content = "Unknown view\n"
	}

	// For minor errors that haven't been acknowledged, show in status bar
	if m.errorMsg.severity == ErrorSeverityMinor && m.errorMsg.IsValid() {
		// Append error message to content
		content = strings.TrimRight(content, "\n")
		content += "\n\n" + m.errorMsg.userMsg + "\n"
	}

	return content
}

// Close closes all models
func (m *AppModel) Close() error {
	var err1, err2, err3 error

	if m.listModel != nil {
		err1 = m.listModel.Close()
	}

	if m.detailModel != nil {
		err2 = m.detailModel.Close()
	}

	if m.commentsModel != nil {
		err3 = m.commentsModel.Close()
	}

	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return err3
}

// RunAppView runs the main application view
func RunAppView(dbPath string, cfg *config.Config, output io.Writer) error {
	model, err := NewAppModel(dbPath, cfg)
	if err != nil {
		return fmt.Errorf("failed to create app model: %w", err)
	}
	defer model.Close()

	// Check if we're in a test environment
	if os.Getenv("GHISSIES_TEST") == "1" {
		// In test mode, show simple message and issue count
		fmt.Fprintln(output, "App view would be displayed here")
		issueCount := 0
		if model.listModel != nil {
			issueCount = len(model.listModel.issues)
		}
		fmt.Fprintf(output, "Found %d issues\n", issueCount)
		return nil
	}

	// Create tea program
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running app view: %w", err)
	}

	return nil
}

// getTheme returns the current theme from config
func (m *AppModel) getTheme() config.Theme {
	themeName := "default"
	if m.config != nil && m.config.Display.Theme != "" {
		themeName = m.config.Display.Theme
	}
	return config.GetTheme(themeName)
}
