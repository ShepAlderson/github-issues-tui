package tui

import (
	"fmt"
	"io"
	"os"

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

// AppModel represents the main application model that can switch between views
type AppModel struct {
	currentView   ViewType
	listModel     *ListModel
	detailModel   *DetailModel
	commentsModel *CommentsModel
	config        *config.Config
	dbPath        string
	err           error
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

// Update handles messages and updates the model
func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ListView:
		// Handle list view
		updated, viewCmd := m.listModel.Update(msg)
		m.listModel = updated.(*ListModel)

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

// View renders the UI
func (m AppModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	switch m.currentView {
	case ListView:
		if m.listModel != nil {
			return m.listModel.View()
		}
		return "List view not initialized\n"

	case DetailView:
		if m.detailModel != nil {
			return m.detailModel.View()
		}
		return "Detail view not initialized\n"

	case CommentsView:
		if m.commentsModel != nil {
			return m.commentsModel.View()
		}
		return "Comments view not initialized\n"

	default:
		return "Unknown view\n"
	}
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
