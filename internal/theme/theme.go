package theme

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme represents a color theme for the application
type Theme struct {
	Name       string
	Primary    string
	Secondary  string
	Text       string
	Muted      string
	Error      string
	Success    string
	Open       string
	Closed     string
	Label      string
	Border     string
	Background string
}

// ThemeStyles contains all the lipgloss styles for a theme
type ThemeStyles struct {
	// List styles
	Selected lipgloss.Style
	Normal   lipgloss.Style
	Header   lipgloss.Style
	Status   lipgloss.Style
	Error    lipgloss.Style

	// Detail styles
	Title       lipgloss.Style
	Meta        lipgloss.Style
	StateOpen   lipgloss.Style
	StateClosed lipgloss.Style
	Label       lipgloss.Style
	Body        lipgloss.Style
	Footer      lipgloss.Style

	// Help styles
	Border  lipgloss.Style
	Section lipgloss.Style
	Key     lipgloss.Style
	Desc    lipgloss.Style

	// Comments styles
	CommentHeader lipgloss.Style
	CommentBody   lipgloss.Style
	Separator     lipgloss.Style

	// Error modal styles
	ModalBorder   lipgloss.Style
	ModalTitle    lipgloss.Style
	ModalGuidance lipgloss.Style
	ModalFooter   lipgloss.Style

	// Sync styles
	Progress lipgloss.Style
	Success  lipgloss.Style
}

// availableThemes lists all supported theme names
var availableThemes = []string{
	"default",
	"dracula",
	"gruvbox",
	"nord",
	"solarized-dark",
	"solarized-light",
}

// IsValidTheme checks if a theme name is valid
func IsValidTheme(name string) bool {
	for _, theme := range availableThemes {
		if theme == name {
			return true
		}
	}
	return false
}

// GetAvailableThemes returns a list of all available theme names
func GetAvailableThemes() []string {
	result := make([]string, len(availableThemes))
	copy(result, availableThemes)
	return result
}

// GetTheme returns a theme by name, defaulting to "default" if not found
func GetTheme(name string) *Theme {
	switch name {
	case "default":
		return newDefaultTheme()
	case "dracula":
		return newDraculaTheme()
	case "gruvbox":
		return newGruvboxTheme()
	case "nord":
		return newNordTheme()
	case "solarized-dark":
		return newSolarizedDarkTheme()
	case "solarized-light":
		return newSolarizedLightTheme()
	default:
		return newDefaultTheme()
	}
}

// Styles returns the lipgloss styles for this theme
func (t *Theme) Styles() *ThemeStyles {
	return &ThemeStyles{
		// List styles
		Selected: lipgloss.NewStyle().
			Background(lipgloss.Color(t.Primary)).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true),
		Normal: lipgloss.NewStyle(),
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(t.Primary)),
		Status: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Muted)),
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Error)).
			Bold(true),

		// Detail styles
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(t.Primary)).
			MarginBottom(1),
		Meta: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Muted)),
		StateOpen: lipgloss.NewStyle().
			Background(lipgloss.Color(t.Open)).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1),
		StateClosed: lipgloss.NewStyle().
			Background(lipgloss.Color(t.Closed)).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1),
		Label: lipgloss.NewStyle().
			Background(lipgloss.Color(t.Label)).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			MarginRight(1),
		Body: lipgloss.NewStyle().
			Padding(1, 0),
		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Muted)).
			MarginTop(1),

		// Help styles
		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(t.Primary)).
			Padding(2, 4),
		Section: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(t.Secondary)),
		Key: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Primary)).
			Bold(true),
		Desc: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Muted)),

		// Comments styles
		CommentHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(t.Primary)),
		CommentBody: lipgloss.NewStyle().
			Padding(1, 0),
		Separator: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Border)),

		// Error modal styles
		ModalBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(t.Error)).
			Padding(2, 4),
		ModalTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(t.Error)),
		ModalGuidance: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Secondary)),
		ModalFooter: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Muted)),

		// Sync styles
		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Success)),
		Progress: lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Primary)),
	}
}

// newDefaultTheme creates the default purple theme
func newDefaultTheme() *Theme {
	return &Theme{
		Name:       "default",
		Primary:    "#7D56F4",
		Secondary:  "#CCCCCC",
		Text:       "#FFFFFF",
		Muted:      "#888888",
		Error:      "#FF6B6B",
		Success:    "#00D4AA",
		Open:       "#238636",
		Closed:     "#8957E5",
		Label:      "#1F6FEB",
		Border:     "#444444",
		Background: "#000000",
	}
}

// newDraculaTheme creates the Dracula theme
func newDraculaTheme() *Theme {
	return &Theme{
		Name:       "dracula",
		Primary:    "#BD93F9", // Purple
		Secondary:  "#F8F8F2", // Foreground
		Text:       "#F8F8F2",
		Muted:      "#6272A4", // Comment
		Error:      "#FF5555", // Red
		Success:    "#50FA7B", // Green
		Open:       "#50FA7B", // Green
		Closed:     "#BD93F9", // Purple
		Label:      "#8BE9FD", // Cyan
		Border:     "#44475A", // Selection
		Background: "#282A36", // Background
	}
}

// newGruvboxTheme creates the Gruvbox dark theme
func newGruvboxTheme() *Theme {
	return &Theme{
		Name:       "gruvbox",
		Primary:    "#D79921", // Yellow (accent)
		Secondary:  "#EBDBB2", // Light text
		Text:       "#EBDBB2",
		Muted:      "#928374", // Gray
		Error:      "#FB4934", // Red
		Success:    "#B8BB26", // Green
		Open:       "#B8BB26", // Green
		Closed:     "#CC241D", // Red
		Label:      "#458588", // Blue
		Border:     "#504945", // Dark gray
		Background: "#282828", // Background
	}
}

// newNordTheme creates the Nord theme
func newNordTheme() *Theme {
	return &Theme{
		Name:       "nord",
		Primary:    "#88C0D0", // Frost (light blue)
		Secondary:  "#D8DEE9", // Snow storm (light)
		Text:       "#D8DEE9",
		Muted:      "#5E81AC", // Frost (darker blue)
		Error:      "#BF616A", // Aurora (red)
		Success:    "#A3BE8C", // Aurora (green)
		Open:       "#A3BE8C", // Green
		Closed:     "#B48EAD", // Aurora (purple)
		Label:      "#81A1C1", // Frost (blue)
		Border:     "#4C566A", // Polar night
		Background: "#2E3440", // Polar night (darkest)
	}
}

// newSolarizedDarkTheme creates the Solarized Dark theme
func newSolarizedDarkTheme() *Theme {
	return &Theme{
		Name:       "solarized-dark",
		Primary:    "#268BD2", // Blue
		Secondary:  "#EEE8D5", // Light text
		Text:       "#EEE8D5",
		Muted:      "#93A1A1", // Base1
		Error:      "#DC322F", // Red
		Success:    "#859900", // Green
		Open:       "#859900", // Green
		Closed:     "#D33682", // Magenta
		Label:      "#2AA198", // Cyan
		Border:     "#073642", // Base02
		Background: "#002B36", // Base03
	}
}

// newSolarizedLightTheme creates the Solarized Light theme
func newSolarizedLightTheme() *Theme {
	return &Theme{
		Name:       "solarized-light",
		Primary:    "#268BD2", // Blue
		Secondary:  "#073642", // Dark text
		Text:       "#073642",
		Muted:      "#586E75", // Base01
		Error:      "#DC322F", // Red
		Success:    "#859900", // Green
		Open:       "#859900", // Green
		Closed:     "#D33682", // Magenta
		Label:      "#2AA198", // Cyan
		Border:     "#EEE8D5", // Base2
		Background: "#FDF6E3", // Base3
	}
}
