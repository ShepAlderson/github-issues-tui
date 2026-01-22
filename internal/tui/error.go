package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shepbook/ghissues/internal/theme"
)

// ErrorSeverity represents the severity level of an error
type ErrorSeverity int

const (
	// ErrorSeverityMinor represents minor errors that don't block the UI
	// Examples: network timeout, rate limit, individual operation failures
	ErrorSeverityMinor ErrorSeverity = iota

	// ErrorSeverityCritical represents critical errors that require user attention
	// Examples: invalid token, database corruption, repository not found
	ErrorSeverityCritical
)

// ErrorModel represents an error state in the TUI
type ErrorModel struct {
	Active    bool           // Whether the error is currently being displayed
	Complete  bool           // Whether the error has been acknowledged (for critical errors)
	Message   string         // The error message
	Guidance  string         // Actionable guidance for resolving the error
	Severity  ErrorSeverity  // The severity level of the error
	Width     int            // Window width for centering the modal
	Height    int            // Window height for centering the modal
	Timestamp string         // When the error occurred (for minor errors in status bar)
}

// NewErrorModel creates a new error model with automatic classification
func NewErrorModel(err error, severity ErrorSeverity) ErrorModel {
	if err == nil {
		return ErrorModel{}
	}

	// If severity is not explicitly set, classify the error
	if severity == ErrorSeverityMinor && IsCriticalError(err) {
		severity = ErrorSeverityCritical
	}

	// Get actionable guidance based on error type
	guidance := GetActionableGuidance(err)

	model := ErrorModel{
		Active:   severity == ErrorSeverityCritical, // Only critical errors show modal
		Complete: false,
		Message:  err.Error(),
		Guidance: guidance,
		Severity: severity,
	}

	return model
}

// ClassifyError determines the severity and provides guidance for an error
func ClassifyError(err error) (ErrorSeverity, string) {
	if err == nil {
		return ErrorSeverityMinor, ""
	}

	// Check for critical errors
	if IsCriticalError(err) {
		guidance := GetActionableGuidance(err)
		return ErrorSeverityCritical, guidance
	}

	return ErrorSeverityMinor, ""
}

// IsCriticalError determines if an error is critical (shows modal) vs minor (status bar)
func IsCriticalError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())

	// Critical: Authentication errors
	if strings.Contains(errMsg, "authentication failed") ||
		strings.Contains(errMsg, "invalid token") ||
		strings.Contains(errMsg, "unauthorized") {
		return true
	}

	// Critical: Repository errors
	if strings.Contains(errMsg, "repository not found") {
		return true
	}

	// Critical: Database corruption errors
	if strings.Contains(errMsg, "database") &&
		(strings.Contains(errMsg, "corrupt") ||
			strings.Contains(errMsg, "malformed") ||
			strings.Contains(errMsg, "schema") ||
			strings.Contains(errMsg, "no such table")) {
		return true
	}

	// Critical: Permission errors
	if strings.Contains(errMsg, "permission denied") {
		return true
	}

	// Minor: Network errors
	if strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "network") ||
		strings.Contains(errMsg, "connection") ||
		strings.Contains(errMsg, "dns") {
		return false
	}

	// Minor: Rate limiting
	if strings.Contains(errMsg, "rate limit") {
		return false
	}

	// Minor: User cancellation
	if strings.Contains(errMsg, "cancelled") {
		return false
	}

	// Default to critical for unknown errors (better safe than sorry)
	return true
}

// GetActionableGuidance returns helpful guidance for resolving an error
func GetActionableGuidance(err error) string {
	if err == nil {
		return ""
	}

	errMsg := strings.ToLower(err.Error())

	// Authentication errors
	if strings.Contains(errMsg, "authentication failed") ||
		strings.Contains(errMsg, "invalid token") ||
		strings.Contains(errMsg, "unauthorized") {

		return "Please check your GitHub token:\n  • Set GITHUB_TOKEN environment variable\n  • Run 'ghissues config' to update token\n  • Verify token has 'repo' scope"
	}

	// Repository not found
	if strings.Contains(errMsg, "repository not found") {
		return "Please check the repository:\n  • Verify repository name is correct (owner/repo format)\n  • Ensure you have access to the repository\n  • Run 'ghissues config' to update repository"
	}

	// Database errors
	if strings.Contains(errMsg, "database") &&
		(strings.Contains(errMsg, "corrupt") ||
			strings.Contains(errMsg, "malformed") ||
			strings.Contains(errMsg, "schema") ||
			strings.Contains(errMsg, "no such table")) {

		return "Database issue detected:\n  • Try deleting the database file and syncing again\n  • Run 'ghissues sync' to re-fetch all data\n  • Check database file permissions"
	}

	// Network errors
	if strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "network") ||
		strings.Contains(errMsg, "connection") ||
		strings.Contains(errMsg, "dns") {

		return "Network issue detected:\n  • Check your internet connection\n  • Verify GitHub.com is accessible\n  • Try again in a moment (press 'r' to refresh)"
	}

	// Rate limiting
	if strings.Contains(errMsg, "rate limit") {
		return "GitHub rate limit exceeded:\n  • Wait a few minutes before trying again\n  • Authenticate with a GitHub token for higher limits\n  • Check your remaining quota at: https://api.github.com/rate_limit"
	}

	// Default guidance
	return "An unexpected error occurred.\n  • Try again (press 'r' to refresh)\n  • Run 'ghissues sync' to re-sync data\n  • Check logs for more details"
}

// GetErrorMessageForStatusBar returns a formatted error message for the status bar
// Returns empty string for critical errors (they show modals instead)
func GetErrorMessageForStatusBar(err error, severity ErrorSeverity) string {
	if err == nil {
		return ""
	}

	// Critical errors show modal, not status bar
	if severity == ErrorSeverityCritical || IsCriticalError(err) {
		return ""
	}

	// Minor errors show in status bar
	return "Error: " + err.Error()
}

// Init initializes the error model
func (m ErrorModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the error model
func (m ErrorModel) Update(msg tea.Msg) (ErrorModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Any key dismisses the error when complete
		if m.Complete {
			m.Active = false
			return m, nil
		}

		// Any key marks the error as complete (requires acknowledgment)
		if m.Active {
			m.Complete = true
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil
	}

	return m, nil
}

// View renders the error modal (for critical errors)
func (m ErrorModel) View(theme *theme.Theme) string {
	if !m.Active {
		return ""
	}

	// Calculate modal dimensions
	modalWidth := 70
	if m.Width > 0 && m.Width < modalWidth+10 {
		modalWidth = m.Width - 10
	}

	// Build the error modal content
	title := "⚠ Error"
	if m.Severity == ErrorSeverityCritical {
		title = "✕ Critical Error"
	}

	// Style definitions
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(theme.Error)).
		Width(modalWidth).
		MarginBottom(1)

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("white")).
		Width(modalWidth).
		MarginBottom(1)

	guidanceStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Faint)).
		Width(modalWidth).
		MarginBottom(1)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(theme.Error)).
		Padding(1, 1)

	// Build content
	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n")

	// Error message (truncate if needed)
	message := m.Message
	if len(message) > modalWidth {
		message = message[:modalWidth-3] + "..."
	}
	content.WriteString(messageStyle.Render(message))
	content.WriteString("\n")

	// Guidance (if available)
	if m.Guidance != "" {
		content.WriteString(guidanceStyle.Render(m.Guidance))
		content.WriteString("\n")
	}

	// Dismiss hint
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Faint)).MarginTop(1)
	hint := "Press any key to continue"
	if m.Complete {
		hint = "Press any key to dismiss"
	}
	content.WriteString(hintStyle.Render(hint))

	// Wrap in border
	modal := borderStyle.Render(content.String())

	// Center the modal
	lines := strings.Split(modal, "\n")
	var centeredLines []string
	for _, line := range lines {
		padding := (m.Width - lipgloss.Width(line)) / 2
		if padding > 0 {
			centeredLines = append(centeredLines, strings.Repeat(" ", padding)+line)
		} else {
			centeredLines = append(centeredLines, line)
		}
	}

	// Vertical centering
	verticalPadding := (m.Height - len(lines)) / 2
	if verticalPadding < 0 {
		verticalPadding = 0
	}

	result := strings.Repeat("\n", verticalPadding)
	result += strings.Join(centeredLines, "\n")

	return result
}

// Show displays an error in the model
func (m *ErrorModel) Show(err error, severity ErrorSeverity) {
	newModel := NewErrorModel(err, severity)
	m.Active = newModel.Active
	m.Complete = newModel.Complete
	m.Message = newModel.Message
	m.Guidance = newModel.Guidance
	m.Severity = newModel.Severity
}

// Hide hides the error modal
func (m *ErrorModel) Hide() {
	m.Active = false
	m.Complete = false
}

// IsCritical returns whether the error is critical
func (m ErrorModel) IsCritical() bool {
	return m.Severity == ErrorSeverityCritical
}

// IsMinor returns whether the error is minor
func (m ErrorModel) IsMinor() bool {
	return m.Severity == ErrorSeverityMinor
}
