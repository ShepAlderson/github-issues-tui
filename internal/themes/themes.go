package themes

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/shepbook/ghissues/internal/config"
)

// Theme defines colors for the TUI
type Theme struct {
	Name string

	// Base colors
	Primary   lipgloss.Color // Main accent color (titles, headers)
	Secondary lipgloss.Color // Secondary color (links, highlights)
	Accent    lipgloss.Color // Labels, tags
	Text      lipgloss.Color // Normal text
	TextMuted lipgloss.Color // Dimmed/secondary text

	// Semantic colors
	Error   lipgloss.Color // Error messages
	Warning lipgloss.Color // Warning messages
	Success lipgloss.Color // Success indicators

	// UI elements
	Selected lipgloss.Color // Selected item background
	Border   lipgloss.Color // Borders and separators
}

// Styles contains pre-built lipgloss styles for the theme
type Styles struct {
	Title    lipgloss.Style
	Header   lipgloss.Style
	Selected lipgloss.Style
	Normal   lipgloss.Style
	Muted    lipgloss.Style
	Error    lipgloss.Style
	Warning  lipgloss.Style
	Label    lipgloss.Style
	Status   lipgloss.Style
	Border   lipgloss.Style
}

// Styles generates lipgloss styles from the theme colors
func (t *Theme) Styles() *Styles {
	return &Styles{
		Title:    lipgloss.NewStyle().Bold(true).Foreground(t.Primary),
		Header:   lipgloss.NewStyle().Bold(true).Foreground(t.Secondary),
		Selected: lipgloss.NewStyle().Bold(true).Background(t.Selected),
		Normal:   lipgloss.NewStyle().Foreground(t.Text),
		Muted:    lipgloss.NewStyle().Foreground(t.TextMuted),
		Error:    lipgloss.NewStyle().Foreground(t.Error),
		Warning:  lipgloss.NewStyle().Foreground(t.Warning).Italic(true),
		Label:    lipgloss.NewStyle().Foreground(t.Accent),
		Status:   lipgloss.NewStyle().Foreground(t.TextMuted),
		Border:   lipgloss.NewStyle().Foreground(t.Border),
	}
}

// GetTheme returns the theme for the given name
// If the theme name is empty or unknown, returns the default theme
func GetTheme(name config.Theme) *Theme {
	switch name {
	case config.ThemeDracula:
		return themeDracula()
	case config.ThemeGruvbox:
		return themeGruvbox()
	case config.ThemeNord:
		return themeNord()
	case config.ThemeSolarizedDark:
		return themeSolarizedDark()
	case config.ThemeSolarizedLight:
		return themeSolarizedLight()
	default:
		return themeDefault()
	}
}

// Default theme - clean terminal colors
func themeDefault() *Theme {
	return &Theme{
		Name:      "default",
		Primary:   lipgloss.Color("86"),  // Cyan
		Secondary: lipgloss.Color("39"),  // Blue
		Accent:    lipgloss.Color("86"),  // Cyan
		Text:      lipgloss.Color("255"), // White
		TextMuted: lipgloss.Color("241"), // Gray
		Error:     lipgloss.Color("196"), // Red
		Warning:   lipgloss.Color("226"), // Yellow
		Success:   lipgloss.Color("82"),  // Green
		Selected:  lipgloss.Color("238"), // Dark gray background
		Border:    lipgloss.Color("241"), // Gray
	}
}

// Dracula theme - dark purple-based theme
// https://draculatheme.com/
func themeDracula() *Theme {
	return &Theme{
		Name:      "dracula",
		Primary:   lipgloss.Color("#bd93f9"), // Purple
		Secondary: lipgloss.Color("#8be9fd"), // Cyan
		Accent:    lipgloss.Color("#ff79c6"), // Pink
		Text:      lipgloss.Color("#f8f8f2"), // Foreground
		TextMuted: lipgloss.Color("#6272a4"), // Comment
		Error:     lipgloss.Color("#ff5555"), // Red
		Warning:   lipgloss.Color("#f1fa8c"), // Yellow
		Success:   lipgloss.Color("#50fa7b"), // Green
		Selected:  lipgloss.Color("#44475a"), // Current Line
		Border:    lipgloss.Color("#6272a4"), // Comment
	}
}

// Gruvbox theme - retro warm colors
// https://github.com/morhetz/gruvbox
func themeGruvbox() *Theme {
	return &Theme{
		Name:      "gruvbox",
		Primary:   lipgloss.Color("#fabd2f"), // Yellow
		Secondary: lipgloss.Color("#83a598"), // Blue
		Accent:    lipgloss.Color("#8ec07c"), // Aqua
		Text:      lipgloss.Color("#ebdbb2"), // Foreground
		TextMuted: lipgloss.Color("#928374"), // Gray
		Error:     lipgloss.Color("#fb4934"), // Red
		Warning:   lipgloss.Color("#fabd2f"), // Yellow
		Success:   lipgloss.Color("#b8bb26"), // Green
		Selected:  lipgloss.Color("#3c3836"), // bg1
		Border:    lipgloss.Color("#928374"), // Gray
	}
}

// Nord theme - arctic colors
// https://www.nordtheme.com/
func themeNord() *Theme {
	return &Theme{
		Name:      "nord",
		Primary:   lipgloss.Color("#88c0d0"), // Nord8 Frost
		Secondary: lipgloss.Color("#81a1c1"), // Nord9 Frost
		Accent:    lipgloss.Color("#b48ead"), // Nord15 Aurora
		Text:      lipgloss.Color("#eceff4"), // Nord6 Snow Storm
		TextMuted: lipgloss.Color("#4c566a"), // Nord3 Polar Night
		Error:     lipgloss.Color("#bf616a"), // Nord11 Aurora
		Warning:   lipgloss.Color("#ebcb8b"), // Nord13 Aurora
		Success:   lipgloss.Color("#a3be8c"), // Nord14 Aurora
		Selected:  lipgloss.Color("#3b4252"), // Nord1 Polar Night
		Border:    lipgloss.Color("#4c566a"), // Nord3 Polar Night
	}
}

// Solarized Dark theme
// https://ethanschoonover.com/solarized/
func themeSolarizedDark() *Theme {
	return &Theme{
		Name:      "solarized-dark",
		Primary:   lipgloss.Color("#268bd2"), // Blue
		Secondary: lipgloss.Color("#2aa198"), // Cyan
		Accent:    lipgloss.Color("#d33682"), // Magenta
		Text:      lipgloss.Color("#839496"), // Base0
		TextMuted: lipgloss.Color("#586e75"), // Base01
		Error:     lipgloss.Color("#dc322f"), // Red
		Warning:   lipgloss.Color("#b58900"), // Yellow
		Success:   lipgloss.Color("#859900"), // Green
		Selected:  lipgloss.Color("#073642"), // Base02
		Border:    lipgloss.Color("#586e75"), // Base01
	}
}

// Solarized Light theme
// https://ethanschoonover.com/solarized/
func themeSolarizedLight() *Theme {
	return &Theme{
		Name:      "solarized-light",
		Primary:   lipgloss.Color("#268bd2"), // Blue
		Secondary: lipgloss.Color("#2aa198"), // Cyan
		Accent:    lipgloss.Color("#d33682"), // Magenta
		Text:      lipgloss.Color("#657b83"), // Base00
		TextMuted: lipgloss.Color("#93a1a1"), // Base1
		Error:     lipgloss.Color("#dc322f"), // Red
		Warning:   lipgloss.Color("#b58900"), // Yellow
		Success:   lipgloss.Color("#859900"), // Green
		Selected:  lipgloss.Color("#eee8d5"), // Base2
		Border:    lipgloss.Color("#93a1a1"), // Base1
	}
}
