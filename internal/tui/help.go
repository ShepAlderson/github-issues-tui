package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HelpModel represents the help overlay state
type HelpModel struct {
	Active bool
	Width  int
	Height int
}

// NewHelpModel creates a new help model
func NewHelpModel() HelpModel {
	return HelpModel{
		Active: false,
		Width:  80,
		Height: 24,
	}
}

// Init initializes the help model
func (m HelpModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the help model
func (m HelpModel) Update(msg tea.Msg) (HelpModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Dismiss help on ? or Esc
		if msg.String() == "?" || msg.Type == tea.KeyEsc {
			m.Active = false
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil
	}

	return m, nil
}

// Toggle switches the help overlay on/off
func (m *HelpModel) Toggle() {
	m.Active = !m.Active
}

// View renders the help overlay
func (m HelpModel) View() string {
	if !m.Active {
		return ""
	}

	// Calculate modal dimensions
	modalWidth := 80
	if m.Width > 0 && m.Width < modalWidth+10 {
		modalWidth = m.Width - 10
	}

	// Style definitions
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")). // Cyan
		Width(modalWidth).
		MarginBottom(1)

	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("228")). // Yellow
		Width(modalWidth).
		MarginTop(1).
		MarginBottom(1)

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("212")). // Pink
		Width(20)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("243")). // Light gray
		Width(modalWidth - 22)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("86")).
		Padding(1, 1)

	hintStyle := lipgloss.NewStyle().
		Faint(true).
		MarginTop(1)

	// Build content
	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render("Keyboard Shortcuts"))
	content.WriteString("\n")

	// Main View section
	content.WriteString(sectionStyle.Render("Main View"))
	content.WriteString(keyStyle.Render("j / ↓"))
	content.WriteString(descStyle.Render("Move down"))
	content.WriteString("\n")

	content.WriteString(keyStyle.Render("k / ↑"))
	content.WriteString(descStyle.Render("Move up"))
	content.WriteString("\n")

	content.WriteString(keyStyle.Render("Enter"))
	content.WriteString(descStyle.Render("Open comments view"))
	content.WriteString("\n")

	content.WriteString(keyStyle.Render("Space"))
	content.WriteString(descStyle.Render("Select current issue"))
	content.WriteString("\n")

	content.WriteString(keyStyle.Render("s"))
	content.WriteString(descStyle.Render("Cycle sort field"))
	content.WriteString("\n")

	content.WriteString(keyStyle.Render("S"))
	content.WriteString(descStyle.Render("Toggle sort order"))
	content.WriteString("\n")

	content.WriteString(keyStyle.Render("m"))
	content.WriteString(descStyle.Render("Toggle markdown rendering"))
	content.WriteString("\n")

	content.WriteString(keyStyle.Render("r"))
	content.WriteString(descStyle.Render("Incremental refresh"))
	content.WriteString("\n")

	content.WriteString(keyStyle.Render("R"))
	content.WriteString(descStyle.Render("Full refresh"))
	content.WriteString("\n")

	content.WriteString(keyStyle.Render("?"))
	content.WriteString(descStyle.Render("Show this help"))
	content.WriteString("\n")

	content.WriteString(keyStyle.Render("q / Ctrl+C"))
	content.WriteString(descStyle.Render("Quit"))
	content.WriteString("\n")

	// Comments View section
	content.WriteString(sectionStyle.Render("Comments View"))
	content.WriteString(keyStyle.Render("j / ↓"))
	content.WriteString(descStyle.Render("Scroll down"))
	content.WriteString("\n")

	content.WriteString(keyStyle.Render("k / ↑"))
	content.WriteString(descStyle.Render("Scroll up"))
	content.WriteString("\n")

	content.WriteString(keyStyle.Render("m"))
	content.WriteString(descStyle.Render("Toggle markdown rendering"))
	content.WriteString("\n")

	content.WriteString(keyStyle.Render("Esc / q"))
	content.WriteString(descStyle.Render("Back to issue list"))
	content.WriteString("\n")

	// Dismiss hint
	content.WriteString(hintStyle.Render("Press ? or Esc to dismiss"))

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
