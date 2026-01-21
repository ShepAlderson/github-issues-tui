package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ErrorComponent handles error display in the TUI
type ErrorComponent struct {
	// Modal error state
	showModal    bool
	modalTitle   string
	modalMessage string

	// Status error state
	statusMessage  string
	statusExpiresAt time.Time

	// Dimensions
	width  int
	height int
}

// NewErrorComponent creates a new error component
func NewErrorComponent() *ErrorComponent {
	return &ErrorComponent{
		showModal:    false,
		modalTitle:   "",
		modalMessage: "",
		statusMessage: "",
		statusExpiresAt: time.Time{},
	}
}

// ShowModal displays a modal error that requires acknowledgment
func (ec *ErrorComponent) ShowModal(title, message string) {
	ec.showModal = true
	ec.modalTitle = title
	ec.modalMessage = message
}

// ShowStatus displays a status bar error that auto-expires
func (ec *ErrorComponent) ShowStatus(message string, duration time.Duration) {
	ec.statusMessage = message
	ec.statusExpiresAt = time.Now().Add(duration)
}

// ClearStatus clears any status message
func (ec *ErrorComponent) ClearStatus() {
	ec.statusMessage = ""
	ec.statusExpiresAt = time.Time{}
}

// DismissModal dismisses the current modal error
func (ec *ErrorComponent) DismissModal() {
	ec.showModal = false
	ec.modalTitle = ""
	ec.modalMessage = ""
}

// Init implements the tea.Model interface
func (ec *ErrorComponent) Init() tea.Cmd {
	return nil
}

// Update implements the tea.Model interface
func (ec *ErrorComponent) Update(msg tea.Msg) (*ErrorComponent, tea.Cmd) {
	// Clear expired status messages
	if !ec.statusExpiresAt.IsZero() && time.Now().After(ec.statusExpiresAt) {
		ec.ClearStatus()
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		ec.width = msg.Width
		ec.height = msg.Height
		return ec, nil

	case tea.KeyMsg:
		// Only handle key messages if modal is shown
		if ec.showModal {
			switch msg.String() {
			case "enter", " ", "esc":
				// Dismiss modal on Enter, Space, or Escape
				ec.DismissModal()
				return ec, nil
			}
		}
	}

	return ec, nil
}

// View implements the tea.Model interface
func (ec *ErrorComponent) View() string {
	var views []string

	// Render modal if shown
	if ec.showModal {
		views = append(views, ec.renderModal())
	}

	// Render status message if active
	if ec.statusMessage != "" && !time.Now().After(ec.statusExpiresAt) {
		views = append(views, ec.renderStatus())
	}

	return strings.Join(views, "\n")
}

// renderModal renders the modal error overlay
func (ec *ErrorComponent) renderModal() string {
	// Calculate modal dimensions (80% of width, centered)
	modalWidth := int(float64(ec.width) * 0.8)
	if modalWidth < 40 {
		modalWidth = 40
	}
	if modalWidth > 80 {
		modalWidth = 80
	}

	// Create modal box
	modalStyle := lipgloss.NewStyle().
		Width(modalWidth).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("9")).
		Padding(1, 2).
		Background(lipgloss.Color("0")).
		Foreground(lipgloss.Color("15"))

	// Title style
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("9")).
		PaddingBottom(1)

	// Message style
	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		PaddingBottom(1)

	// Instruction style
	instructionStyle := lipgloss.NewStyle().
		Italic(true).
		Foreground(lipgloss.Color("8")).
		PaddingTop(1).
		BorderTop(true).
		BorderForeground(lipgloss.Color("8"))

	// Build modal content
	content := titleStyle.Render("⚠️  " + ec.modalTitle) + "\n"
	content += messageStyle.Render(ec.modalMessage) + "\n"
	content += instructionStyle.Render("Press Enter, Space, or Esc to continue")

	return lipgloss.Place(
		ec.width,
		ec.height,
		lipgloss.Center,
		lipgloss.Center,
		modalStyle.Render(content),
	)
}

// renderStatus renders the status bar error message
func (ec *ErrorComponent) renderStatus() string {
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")).
		Background(lipgloss.Color("0")).
		Padding(0, 1)

	return lipgloss.Place(
		ec.width,
		1, // Single line status bar
		lipgloss.Left,
		lipgloss.Top,
		statusStyle.Render("⚠️  "+ec.statusMessage),
	)
}

// CategorizeError categorizes an error and provides actionable guidance
func CategorizeError(errStr string) (title, message string, isCritical bool) {
	lowerErr := strings.ToLower(errStr)

	// Network errors (minor)
	if strings.Contains(lowerErr, "network") ||
		strings.Contains(lowerErr, "timeout") ||
		strings.Contains(lowerErr, "connection") ||
		strings.Contains(lowerErr, "dial") ||
		strings.Contains(lowerErr, "connect") {
		title = "Network Error"
		message = fmt.Sprintf("%s\n\n• Check your internet connection\n• Try again in a few moments\n• If using a VPN, ensure it's connected", errStr)
		isCritical = false
		return
	}

	// Rate limiting errors (minor)
	if strings.Contains(lowerErr, "rate limit") ||
		strings.Contains(lowerErr, "too many requests") ||
		strings.Contains(lowerErr, "429") {
		title = "Rate Limit"
		message = fmt.Sprintf("%s\n\n• GitHub has rate-limited your requests\n• Wait a few minutes before trying again\n• Consider using a personal access token for higher limits", errStr)
		isCritical = false
		return
	}

	// Authentication errors (critical)
	if strings.Contains(lowerErr, "auth") ||
		strings.Contains(lowerErr, "token") ||
		strings.Contains(lowerErr, "unauthorized") ||
		strings.Contains(lowerErr, "401") ||
		strings.Contains(lowerErr, "403") {
		title = "Authentication Error"
		message = fmt.Sprintf("%s\n\n• Check your GitHub token is valid\n• Run 'ghissues config' to reconfigure authentication\n• Ensure your token has proper repository permissions", errStr)
		isCritical = true
		return
	}

	// Database errors (critical)
	if strings.Contains(lowerErr, "database") ||
		strings.Contains(lowerErr, "sql") ||
		strings.Contains(lowerErr, "corrupt") ||
		strings.Contains(lowerErr, "malformed") ||
		strings.Contains(lowerErr, "disk image") {
		title = "Database Error"
		message = fmt.Sprintf("%s\n\n• Your database file may be corrupted\n• Try deleting .ghissues.db and re-syncing\n• Check disk space and permissions", errStr)
		isCritical = true
		return
	}

	// Repository errors (critical)
	if strings.Contains(lowerErr, "repo") ||
		strings.Contains(lowerErr, "repository") ||
		strings.Contains(lowerErr, "not found") ||
		strings.Contains(lowerErr, "404") {
		title = "Repository Error"
		message = fmt.Sprintf("%s\n\n• Check the repository exists and is accessible\n• Verify you have correct permissions\n• Run 'ghissues config' to update repository", errStr)
		isCritical = true
		return
	}

	// Default (minor)
	title = "Error"
	message = errStr
	isCritical = false
	return
}