package help

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the help overlay state
type Model struct {
	showing bool
	width   int
	height  int
}

// Styles for the help overlay
var (
	modalBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7D56F4")).
				Padding(2, 4)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#CCCCCC"))

	keyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))
)

// NewModel creates a new help model
func NewModel() Model {
	return Model{
		showing: false,
		width:   80,
		height:  24,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View renders the help overlay
func (m Model) View() string {
	if !m.showing {
		return ""
	}

	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render("⌨️  Keybindings"))
	content.WriteString("\n\n")

	// Build sections
	sections := []string{
		m.renderNavigationSection(),
		m.renderListSection(),
		m.renderDetailSection(),
		m.renderCommentsSection(),
	}

	// Join sections
	content.WriteString(strings.Join(sections, "\n"))
	content.WriteString("\n")

	// Footer with dismiss hint
	content.WriteString("\n")
	content.WriteString(footerStyle.Render("Press ? or Esc to dismiss"))

	// Apply modal styling
	modal := modalBorderStyle.
		Width(m.width - 10).
		Render(content.String())

	// Center the modal
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		modal,
	)
}

// ToggleHelp toggles the help overlay visibility
func (m *Model) ToggleHelp() {
	m.showing = !m.showing
}

// ShowHelp shows the help overlay
func (m *Model) ShowHelp() {
	m.showing = true
}

// HideHelp hides the help overlay
func (m *Model) HideHelp() {
	m.showing = false
}

// IsShowing returns whether the help overlay is showing
func (m Model) IsShowing() bool {
	return m.showing
}

// SetDimensions updates the model dimensions
func (m *Model) SetDimensions(width, height int) {
	m.width = width
	m.height = height
}

// renderNavigationSection renders the navigation keybindings
func (m Model) renderNavigationSection() string {
	items := []string{
		renderKeybinding("j, k or ↑, ↓", "Move down/up"),
		renderKeybinding("?", "Show/hide this help"),
	}
	return renderSection("Navigation", items)
}

// renderListSection renders the list view keybindings
func (m Model) renderListSection() string {
	items := []string{
		renderKeybinding("Enter", "Open comments view for selected issue"),
		renderKeybinding("s", "Cycle sort field (updated → created → number → comments)"),
		renderKeybinding("S", "Toggle sort order (ascending/descending)"),
		renderKeybinding("r", "Refresh issues (incremental sync)"),
		renderKeybinding("q", "Quit"),
	}
	return renderSection("List View", items)
}

// renderDetailSection renders the detail view keybindings
func (m Model) renderDetailSection() string {
	items := []string{
		renderKeybinding("m", "Toggle markdown rendering (rendered/raw)"),
	}
	return renderSection("Detail View", items)
}

// renderCommentsSection renders the comments view keybindings
func (m Model) renderCommentsSection() string {
	items := []string{
		renderKeybinding("j, k or ↑, ↓", "Scroll down/up"),
		renderKeybinding("m", "Toggle markdown rendering (rendered/raw)"),
		renderKeybinding("q, Esc", "Return to issue list"),
	}
	return renderSection("Comments View", items)
}

// renderKeybinding renders a single keybinding line
func renderKeybinding(key, description string) string {
	return keyStyle.Render(key) + "  " + descStyle.Render(description)
}

// renderSection renders a section with a title and items
func renderSection(title string, items []string) string {
	var b strings.Builder
	b.WriteString(sectionStyle.Render(title))
	b.WriteString("\n")
	for _, item := range items {
		b.WriteString("  ")
		b.WriteString(item)
		b.WriteString("\n")
	}
	return b.String()
}
