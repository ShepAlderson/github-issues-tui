package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shepbook/git/github-issues-tui/internal/config"
)

// keybinding represents a single keybinding
// with the key and its description
type keybinding struct {
	key         string
	description string
}

// HelpModel represents the TUI model for the help overlay
// It displays keybindings organized by context
type HelpModel struct {
	viewType     ViewType
	width        int
	height       int
	active       bool
	theme        config.Theme
	keybindings  map[string][]keybinding
}

// NewHelpModel creates a new help model for the specified view type
// It loads keybindings appropriate for the current context
func NewHelpModel(viewType ViewType, width, height int, theme config.Theme) *HelpModel {
	return &HelpModel{
		viewType:    viewType,
		width:       width,
		height:      height,
		active:      true,
		theme:       theme,
		keybindings: getContextSensitiveKeybindings(viewType),
	}
}

// Init initializes the model
func (m HelpModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m *HelpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "?", "esc", "q", "enter":
			// Dismiss help with any of these keys
			m.active = false
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the help overlay
func (m HelpModel) View() string {
	if !m.active {
		return ""
	}

	// Create the help content
	content := m.renderHelpContent()

	// Center and style the modal
	return m.renderModal(content)
}

// renderHelpContent creates the formatted keybinding content
func (m HelpModel) renderHelpContent() string {
	var builder strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Accent).
		Align(lipgloss.Center).
		Width(50)

	builder.WriteString(titleStyle.Render("Keybinding Help"))
	builder.WriteString("\n\n")

	// Render each section
	for section, bindings := range m.keybindings {
		// Section header
		sectionStyle := lipgloss.NewStyle().
			Bold(true).
			Underline(true).
			Foreground(m.theme.AccentLight).
			Width(50)

		builder.WriteString(sectionStyle.Render(section))
		builder.WriteString("\n")

		// Render each keybinding in the section
		for _, kb := range bindings {
			// Key
			keyStyle := lipgloss.NewStyle().
				Foreground(m.theme.Warning). // Yellow/orange for key
				Width(12).
				Align(lipgloss.Left)

			builder.WriteString(keyStyle.Render(kb.key))

			// Description
			descStyle := lipgloss.NewStyle().
				Foreground(m.theme.Text). // White
				Width(38)

			builder.WriteString(descStyle.Render(kb.description))
			builder.WriteString("\n")
		}

		builder.WriteString("\n")
	}

	// Footer instruction
	footerStyle := lipgloss.NewStyle().
		Foreground(m.theme.TextMuted).
		Italic(true).
		Align(lipgloss.Center).
		Width(50)

	builder.WriteString(footerStyle.Render("Press ? or Esc to close"))

	return builder.String()
}

// renderModal centers and styles the help content as an overlay
func (m HelpModel) renderModal(content string) string {
	// Create modal window style
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Accent).
		Background(m.theme.Background).
		Padding(1, 2)

	// Calculate size based on content
	lines := strings.Split(content, "\n")
	contentHeight := len(lines)
	contentWidth := 0
	for _, line := range lines {
		if len(line) > contentWidth {
			contentWidth = len(line)
		}
	}

	// Add some padding to the content width
	contentWidth += 4

	// Apply styling
	modal := modalStyle.Width(contentWidth).Render(content)

	// Center the modal on screen
	return m.centerModal(modal, contentWidth, contentHeight)
}

// centerModal centers content on the screen
func (m HelpModel) centerModal(content string, width, height int) string {
	// Calculate padding
	topPadding := (m.height - height) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	leftPadding := (m.width - width) / 2
	if leftPadding < 0 {
		leftPadding = 0
	}

	// Add left padding to each line
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = strings.Repeat(" ", leftPadding) + line
	}

	// Add top padding
	topPad := strings.Repeat("\n", topPadding)

	return topPad + strings.Join(lines, "\n")
}

// IsActive returns whether the help overlay is currently active
func (m HelpModel) IsActive() bool {
	return m.active
}

// getContextSensitiveKeybindings returns keybindings organized by section
// based on the current view type
func getContextSensitiveKeybindings(viewType ViewType) map[string][]keybinding {
	bindings := make(map[string][]keybinding)

	// Always include global bindings
	bindings["Global"] = getGlobalKeybindings()

	// Add view-specific bindings
	switch viewType {
	case ListView:
		bindings["List View"] = getListKeybindings()
		bindings["Issue Detail"] = getDetailKeybindings()
		bindings["Comments"] = getCommentsKeybindings()
	case DetailView:
		bindings["Issue Detail"] = getDetailKeybindings()
		bindings["List View"] = getListKeybindings()
		bindings["Comments"] = getCommentsKeybindings()
	case CommentsView:
		bindings["Comments"] = getCommentsKeybindings()
		bindings["Issue Detail"] = getDetailKeybindings()
		bindings["List View"] = getListKeybindings()
	}

	return bindings
}

// getListKeybindings returns keybindings for the list view
func getListKeybindings() []keybinding {
	return []keybinding{
		{key: "j/k", description: "Move down/up"},
		{key: "↑/↓", description: "Move down/up"},
		{key: "enter", description: "Open issue detail"},
		{key: "s", description: "Cycle sort field"},
		{key: "S", description: "Toggle sort direction"},
		{key: "r", description: "Refresh issue list"},
	}
}

// getDetailKeybindings returns keybindings for the detail view
func getDetailKeybindings() []keybinding {
	return []keybinding{
		{key: "j/k", description: "Scroll down/up"},
		{key: "↑/↓", description: "Scroll down/up"},
		{key: "m", description: "Toggle markdown/raw format"},
		{key: "c", description: "View comments"},
		{key: "q/esc", description: "Return to list"},
	}
}

// getCommentsKeybindings returns keybindings for the comments view
func getCommentsKeybindings() []keybinding {
	return []keybinding{
		{key: "j/k", description: "Scroll down/up"},
		{key: "↑/↓", description: "Scroll down/up"},
		{key: "m", description: "Toggle markdown/raw format"},
		{key: "q/esc", description: "Return to issue detail"},
	}
}

// getGlobalKeybindings returns keybindings available in all views
func getGlobalKeybindings() []keybinding {
	return []keybinding{
		{key: "?", description: "Show/hide this help"},
		{key: "ctrl+c", description: "Quit application"},
	}
}

// getFooter returns a context-sensitive footer showing common keys
// for the current view
func getFooter(viewType ViewType) string {
	var footer string

	switch viewType {
	case ListView:
		footer = "? : help | q : quit | enter : view issue | j/k : navigate | s/S : sort | r : refresh"
	case DetailView:
		footer = "? : help | q : back | m : toggle markdown | c : comments | j/k : scroll"
	case CommentsView:
		footer = "? : help | q : back | m : toggle markdown | j/k : scroll"
	}

	// Style the footer
	footerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("57")).
		Foreground(lipgloss.Color("229")).
		Padding(0, 1)

	return footerStyle.Render(footer)
}
