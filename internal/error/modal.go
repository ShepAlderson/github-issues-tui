package error

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ModalModel represents a modal error dialog that requires acknowledgment
type ModalModel struct {
	Error        AppError
	Width        int
	Height       int
	acknowledged bool
}

// Styles for the modal
var (
	modalBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF6B6B")).
			Padding(2, 4)

	modalTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF6B6B"))

	modalGuidanceStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#CCCCCC"))

	modalFooterStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888"))
)

// NewModalModel creates a new modal error dialog
func NewModalModel(appErr AppError) ModalModel {
	return ModalModel{
		Error:        appErr,
		Width:        60,
		Height:       20,
		acknowledged: false,
	}
}

// Init initializes the model
func (m ModalModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m ModalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEnter, tea.KeySpace:
			// Acknowledge and close
			m.acknowledged = true
			return m, tea.Quit

		case tea.KeyRunes:
			if len(msg.Runes) == 1 && msg.Runes[0] == ' ' || msg.Runes[0] == 'q' {
				m.acknowledged = true
				return m, tea.Quit
			}
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}

	return m, nil
}

// View renders the modal
func (m ModalModel) View() string {
	if m.acknowledged {
		return ""
	}

	var content strings.Builder

	// Title with error indicator
	title := "⚠️  Error"
	if m.Error.Severity.IsCritical() {
		title = "✗ Critical Error"
	}
	content.WriteString(modalTitleStyle.Render(title))
	content.WriteString("\n\n")

	// Error message
	content.WriteString(m.Error.Display)
	content.WriteString("\n")

	// Guidance (if present)
	if m.Error.Guidance != "" {
		content.WriteString("\n")
		content.WriteString(modalGuidanceStyle.Render(m.Error.Guidance))
	}

	// Footer with acknowledgment instruction
	content.WriteString("\n\n")
	content.WriteString(modalFooterStyle.Render("Press Enter or Space to continue"))

	// Apply border style
	modal := modalBorderStyle.Width(m.Width - 10).
		Height(m.Height - 4).
		Render(content.String())

	// Center the modal
	return lipgloss.Place(m.Width, m.Height,
		lipgloss.Center, lipgloss.Center,
		modal,
	)
}

// SetDimensions updates the modal dimensions
func (m *ModalModel) SetDimensions(width, height int) {
	m.Width = width
	m.Height = height
}

// WasAcknowledged returns true if the user acknowledged the error
func (m ModalModel) WasAcknowledged() bool {
	return m.acknowledged
}

// RunModal creates a modal model that can be used with tea.NewProgram
func RunModal(appErr AppError) ModalModel {
	return NewModalModel(appErr)
}
