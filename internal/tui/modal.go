package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ModalDialog represents a critical error dialog that must be acknowledged
type ModalDialog struct {
	errMsg ErrorMessage
	width  int
	height int
	active bool
}

// errorAcknowledgedMsg is sent when user acknowledges a critical error
type errorAcknowledgedMsg struct {
	errMsg ErrorMessage
}

// NewModalDialog creates a new modal dialog for a critical error
func NewModalDialog(errMsg ErrorMessage, width, height int) ModalDialog {
	return ModalDialog{
		errMsg: errMsg,
		width:  width,
		height: height,
		active: true,
	}
}

// Init initializes the modal dialog
func (m *ModalDialog) Init() tea.Cmd {
	return nil
}

// Update handles user input for the modal dialog
func (m *ModalDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyEscape:
			return m.handleAcknowledgment()
		default:
			// Check for 'q' key
			if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'q' {
				return m.handleAcknowledgment()
			}
		}
	}

	return m, nil
}

// handleAcknowledgment handles user acknowledgment of the error
func (m *ModalDialog) handleAcknowledgment() (tea.Model, tea.Cmd) {
	m.active = false
	return m, m.acknowledgeError()
}

// acknowledgeError creates a command to mark the error as acknowledged
func (m *ModalDialog) acknowledgeError() tea.Cmd {
	return func() tea.Msg {
		// Mark error as acknowledged
		m.errMsg.MarkAcknowledged()
		return errorAcknowledgedMsg{errMsg: m.errMsg}
	}
}

// View renders the modal dialog
func (m *ModalDialog) View() string {
	if !m.active {
		return ""
	}

	// Create the modal content
	content := m.renderContent()

	// Calculate position to center the modal
	modalWidth := min(60, m.width-4)
	modalHeight := min(12, m.height-4)
	leftPadding := (m.width - modalWidth) / 2
	topPadding := (m.height - modalHeight) / 2

	// Create the modal window
	modalStyle := lipgloss.NewStyle().
		Width(modalWidth).
		Height(modalHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("9")). // Red border for errors
		Background(lipgloss.Color("0")).
		Padding(1, 2)

		// Render the modal
	modal := modalStyle.Render(content)

	// Position the modal in the center of the screen
	return padTopLeft(modal, topPadding, leftPadding)
}

// renderContent renders the content inside the modal
func (m *ModalDialog) renderContent() string {
	var builder strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")).
		Bold(true).
		Width(50).
		Align(lipgloss.Center)

	builder.WriteString(titleStyle.Render("ERROR"))
	builder.WriteString("\n\n")

	// Error message
	msgStyle := lipgloss.NewStyle().
		Width(50).
		Align(lipgloss.Left)

	builder.WriteString(msgStyle.Render(m.errMsg.userMsg))
	builder.WriteString("\n\n")

	// Instruction
	instrStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Width(50).
		Align(lipgloss.Center)

	builder.WriteString(instrStyle.Render("Press Enter or 'q' to continue"))

	return builder.String()
}

// IsActive returns whether the modal is currently active
func (m *ModalDialog) IsActive() bool {
	return m.active
}

// padTopLeft adds padding to the top and left of content
func padTopLeft(content string, top, left int) string {
	lines := strings.Split(content, "\n")

	// Add left padding to each line
	for i, line := range lines {
		lines[i] = strings.Repeat(" ", left) + line
	}

	// Add top padding
	topPadding := strings.Repeat("\n", top)

	return topPadding + strings.Join(lines, "\n")
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
