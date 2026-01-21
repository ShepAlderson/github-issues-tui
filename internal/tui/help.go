package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HelpComponent handles help display in the TUI
type HelpComponent struct {
	// Modal help state
	showHelp bool

	// Dimensions
	width  int
	height int
}

// NewHelpComponent creates a new help component
func NewHelpComponent() *HelpComponent {
	return &HelpComponent{
		showHelp: false,
	}
}

// ShowHelp displays the help overlay
func (hc *HelpComponent) ShowHelp() {
	hc.showHelp = true
}

// HideHelp hides the help overlay
func (hc *HelpComponent) HideHelp() {
	hc.showHelp = false
}

// Init implements the tea.Model interface
func (hc *HelpComponent) Init() tea.Cmd {
	return nil
}

// Update implements the tea.Model interface
func (hc *HelpComponent) Update(msg tea.Msg) (*HelpComponent, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		hc.width = msg.Width
		hc.height = msg.Height
		return hc, nil

	case tea.KeyMsg:
		// Only handle key messages if help is shown
		if hc.showHelp {
			switch msg.String() {
			case "esc", "enter", " ":
				// Hide help on Esc, Enter, or Space
				hc.HideHelp()
				return hc, nil
			}
		}
	}

	return hc, nil
}

// View implements the tea.Model interface
func (hc *HelpComponent) View() string {
	if !hc.showHelp {
		return ""
	}

	return hc.renderHelp()
}

// renderHelp renders the help overlay
func (hc *HelpComponent) renderHelp() string {
	// Calculate modal dimensions (90% of width, 80% of height, centered)
	modalWidth := int(float64(hc.width) * 0.9)
	if modalWidth < 50 {
		modalWidth = 50
	}
	if modalWidth > 100 {
		modalWidth = 100
	}

	modalHeight := int(float64(hc.height) * 0.8)
	if modalHeight < 30 {
		modalHeight = 30
	}
	if modalHeight > 40 {
		modalHeight = 40
	}

	// Create modal box
	modalStyle := lipgloss.NewStyle().
		Width(modalWidth).
		Height(modalHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2).
		Background(lipgloss.Color("0")).
		Foreground(lipgloss.Color("15"))

	// Title style
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		PaddingBottom(1)

	// Section style
	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214")).
		PaddingTop(1).
		PaddingBottom(1)

	// Key style
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")). // Green
		Bold(true)

	// Description style
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("250"))

	// Instruction style
	instructionStyle := lipgloss.NewStyle().
		Italic(true).
		Foreground(lipgloss.Color("8")).
		PaddingTop(1).
		BorderTop(true).
		BorderForeground(lipgloss.Color("8"))

	// Build help content
	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render("📖 Keybindings Help"))
	content.WriteString("\n\n")

	// Global keybindings
	content.WriteString(sectionStyle.Render("Global"))
	content.WriteString("\n")
	content.WriteString(fmt.Sprintf("  %s: %s\n", keyStyle.Render("q / Ctrl+C"), descStyle.Render("quit application")))
	content.WriteString(fmt.Sprintf("  %s: %s\n", keyStyle.Render("r / R"), descStyle.Render("refresh data from GitHub")))
	content.WriteString(fmt.Sprintf("  %s: %s\n", keyStyle.Render("?"), descStyle.Render("show/hide this help")))
	content.WriteString("\n")

	// Navigation
	content.WriteString(sectionStyle.Render("Navigation"))
	content.WriteString("\n")
	content.WriteString(fmt.Sprintf("  %s: %s\n", keyStyle.Render("↑ / k"), descStyle.Render("move up")))
	content.WriteString(fmt.Sprintf("  %s: %s\n", keyStyle.Render("↓ / j"), descStyle.Render("move down")))
	content.WriteString(fmt.Sprintf("  %s: %s\n", keyStyle.Render("Enter"), descStyle.Render("open selected issue / view comments")))
	content.WriteString(fmt.Sprintf("  %s: %s\n", keyStyle.Render("Esc"), descStyle.Render("back to issue list / close help")))
	content.WriteString("\n")

	// Issue List
	content.WriteString(sectionStyle.Render("Issue List"))
	content.WriteString("\n")
	content.WriteString(fmt.Sprintf("  %s: %s\n", keyStyle.Render("s"), descStyle.Render("cycle sort field (updated, created, number, comments)")))
	content.WriteString(fmt.Sprintf("  %s: %s\n", keyStyle.Render("S"), descStyle.Render("toggle sort order (ascending/descending)")))
	content.WriteString("\n")

	// Issue Detail
	content.WriteString(sectionStyle.Render("Issue Detail"))
	content.WriteString("\n")
	content.WriteString(fmt.Sprintf("  %s: %s\n", keyStyle.Render("m"), descStyle.Render("toggle between raw/rendered markdown")))
	content.WriteString(fmt.Sprintf("  %s: %s\n", keyStyle.Render("g"), descStyle.Render("scroll to top")))
	content.WriteString(fmt.Sprintf("  %s: %s\n", keyStyle.Render("G"), descStyle.Render("scroll to bottom")))
	content.WriteString("\n")

	// Comments View
	content.WriteString(sectionStyle.Render("Comments View"))
	content.WriteString("\n")
	content.WriteString(fmt.Sprintf("  %s: %s\n", keyStyle.Render("m"), descStyle.Render("toggle between raw/rendered markdown")))
	content.WriteString(fmt.Sprintf("  %s: %s\n", keyStyle.Render("g"), descStyle.Render("scroll to first comment")))
	content.WriteString(fmt.Sprintf("  %s: %s\n", keyStyle.Render("G"), descStyle.Render("scroll to last comment")))
	content.WriteString("\n")

	// Error Modal
	content.WriteString(sectionStyle.Render("Error Modal"))
	content.WriteString("\n")
	content.WriteString(fmt.Sprintf("  %s: %s\n", keyStyle.Render("Enter / Space / Esc"), descStyle.Render("dismiss error modal")))
	content.WriteString("\n")

	// Instructions
	content.WriteString(instructionStyle.Render("Press ?, Esc, Enter, or Space to close"))

	return lipgloss.Place(
		hc.width,
		hc.height,
		lipgloss.Center,
		lipgloss.Center,
		modalStyle.Render(content.String()),
	)
}

// FooterContext represents the context for footer hints
type FooterContext int

const (
	// FooterContextList is the issue list view context
	FooterContextList FooterContext = iota
	// FooterContextDetail is the issue detail view context
	FooterContextDetail
	// FooterContextComments is the comments view context
	FooterContextComments
)

// GetFooterHints returns context-sensitive footer hints
func GetFooterHints(context FooterContext) string {
	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)

	var hints []string

	// Always show these hints
	hints = append(hints, "?: help")
	hints = append(hints, "q: quit")

	switch context {
	case FooterContextList:
		hints = append(hints, "↑/↓: navigate")
		hints = append(hints, "Enter: open issue")
		hints = append(hints, "s/S: sort")
		hints = append(hints, "r: refresh")

	case FooterContextDetail:
		hints = append(hints, "↑/↓: scroll")
		hints = append(hints, "Enter: view comments")
		hints = append(hints, "Esc: back to list")
		hints = append(hints, "m: toggle markdown")

	case FooterContextComments:
		hints = append(hints, "↑/↓: scroll comments")
		hints = append(hints, "Esc: back to list")
		hints = append(hints, "m: toggle markdown")
	}

	return hintStyle.Render(strings.Join(hints, " • "))
}