package config

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/shepbook/ghissues/internal/theme"
)

// ThemeModel represents the theme picker TUI state
type ThemeModel struct {
	availableThemes []string
	currentTheme    string
	selected        int
	width           int
	height          int
	quitting        bool
	saved           bool
}

// themeStyles contains styles specific to the theme picker
var themeStyles struct {
	title         lipgloss.Style
	header        lipgloss.Style
	selected      lipgloss.Style
	normal        lipgloss.Style
	muted         lipgloss.Style
	key           lipgloss.Style
	preview       lipgloss.Style
	currentMarker lipgloss.Style
}

func init() {
	// Initialize styles
	themeStyles.title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		MarginBottom(1)

	themeStyles.header = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		MarginBottom(1)

	themeStyles.selected = lipgloss.NewStyle().
		Background(lipgloss.Color("#7D56F4")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)

	themeStyles.normal = lipgloss.NewStyle()

	themeStyles.muted = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888"))

	themeStyles.key = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true)

	themeStyles.preview = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2)

	themeStyles.currentMarker = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D4AA")).
		Bold(true)
}

// NewThemeModel creates a new theme picker model
func NewThemeModel(currentTheme string) ThemeModel {
	themes := theme.GetAvailableThemes()

	// Find the index of the current theme
	selected := 0
	for i, t := range themes {
		if t == currentTheme {
			selected = i
			break
		}
	}

	return ThemeModel{
		availableThemes: themes,
		currentTheme:    currentTheme,
		selected:        selected,
		width:           80,
		height:          24,
		quitting:        false,
		saved:           false,
	}
}

// Init initializes the model
func (m ThemeModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m ThemeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit

		case tea.KeyUp:
			if m.selected > 0 {
				m.selected--
			} else {
				m.selected = len(m.availableThemes) - 1
			}

		case tea.KeyDown:
			if m.selected < len(m.availableThemes)-1 {
				m.selected++
			} else {
				m.selected = 0
			}

		case tea.KeyRunes:
			switch msg.String() {
			case "q", "Q":
				m.quitting = true
				return m, tea.Quit

			case "j":
				if m.selected < len(m.availableThemes)-1 {
					m.selected++
				} else {
					m.selected = 0
				}

			case "k":
				if m.selected > 0 {
					m.selected--
				} else {
					m.selected = len(m.availableThemes) - 1
				}

			case "enter", "Enter":
				// Save the selected theme
				m.saved = true
				m.quitting = true
				return m, tea.Quit
			}

		case tea.KeyEnter:
			// Save the selected theme
			m.saved = true
			m.quitting = true
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View renders the theme picker
func (m ThemeModel) View() string {
	if m.quitting {
		if m.saved {
			return fmt.Sprintf("Theme set to: %s\n", m.GetSelectedTheme())
		}
		return "Theme selection cancelled.\n"
	}

	var b strings.Builder

	// Title
	b.WriteString(themeStyles.title.Render("🎨 Theme Picker"))
	b.WriteString("\n")

	// Header
	b.WriteString(themeStyles.header.Render("Select a color theme for ghissues"))
	b.WriteString("\n\n")

	// Current theme indicator
	fmt.Fprintf(&b, "Current theme: %s\n", themeStyles.currentMarker.Render(m.currentTheme))
	b.WriteString("\n")

	// Theme list
	for i, t := range m.availableThemes {
		var line string
		if i == m.selected {
			line = themeStyles.selected.Render("> " + t)
		} else if t == m.currentTheme {
			line = fmt.Sprintf("  %s %s", themeStyles.currentMarker.Render("●"), t)
		} else {
			line = themeStyles.normal.Render("   " + t)
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Preview section
	previewTheme := theme.GetTheme(m.availableThemes[m.selected])
	previewStyles := previewTheme.Styles()

	var preview strings.Builder
	preview.WriteString("Preview:\n")
	preview.WriteString(previewStyles.Header.Render("Header") + " ")
	preview.WriteString(previewStyles.Key.Render("Key") + " ")
	preview.WriteString(previewStyles.Success.Render("Success") + " ")
	preview.WriteString(previewStyles.Error.Render("Error") + "\n")
	preview.WriteString(previewStyles.Status.Render("Status text") + "\n")
	preview.WriteString(previewStyles.StateOpen.Render(" open ") + " ")
	preview.WriteString(previewStyles.StateClosed.Render(" closed ") + " ")
	preview.WriteString(previewStyles.Label.Render(" label ") + "\n")

	previewBox := themeStyles.preview.Width(50).Render(preview.String())
	b.WriteString(previewBox)
	b.WriteString("\n\n")

	// Footer with keybindings
	footer := fmt.Sprintf("%s to select, %s/%s to navigate, %s to cancel",
		themeStyles.key.Render("Enter"),
		themeStyles.key.Render("j"),
		themeStyles.key.Render("k"),
		themeStyles.key.Render("q"))
	b.WriteString(themeStyles.muted.Render(footer))
	b.WriteString("\n")

	return b.String()
}

// GetSelectedTheme returns the currently selected theme name
func (m ThemeModel) GetSelectedTheme() string {
	return m.availableThemes[m.selected]
}

// IsSaved returns whether the theme was saved
func (m ThemeModel) IsSaved() bool {
	return m.saved
}

// SaveThemeToConfig saves the selected theme to the config file
func SaveThemeToConfig(themeName string) error {
	// Validate theme name
	if !theme.IsValidTheme(themeName) {
		return fmt.Errorf("invalid theme: %s", themeName)
	}

	// Load existing config
	cfg, err := Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Update theme
	cfg.Display.Theme = themeName

	// Save config
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// RunThemePicker runs the theme picker TUI and returns the selected theme
func RunThemePicker(currentTheme string) (string, bool, error) {
	model := NewThemeModel(currentTheme)
	p := tea.NewProgram(model)

	result, err := p.Run()
	if err != nil {
		return "", false, fmt.Errorf("error running theme picker: %w", err)
	}

	themeModel, ok := result.(ThemeModel)
	if !ok {
		return "", false, fmt.Errorf("unexpected model type")
	}

	return themeModel.GetSelectedTheme(), themeModel.IsSaved(), nil
}
