package theme

import (
	"strings"
)

// Theme defines color scheme for the TUI
type Theme struct {
	// Status bar colors
	StatusBar string
	StatusKey string

	// UI element colors
	Title        string
	Border       string
	ListSelected string
	ListHeader   string
	Cursor       string

	// Help overlay colors
	HelpTitle  string
	HelpKey    string
	HelpSection string

	// Message colors
	Error   string
	Success string
	Label   string
	Faint   string
}

// Built-in themes
var (
	defaultTheme = Theme{
		StatusBar:    "242",    // Light gray
		StatusKey:    "86",     // Cyan
		Title:        "blue",
		Border:       "86",     // Cyan
		ListSelected: "235",    // Dark gray background
		ListHeader:   "blue",
		Cursor:       "235",    // Dark gray background
		HelpTitle:    "86",     // Cyan
		HelpSection:  "228",    // Yellow
		HelpKey:      "212",    // Pink
		Error:        "red",
		Success:      "green",
		Label:        "62",     // Purple
		Faint:        "243",    // Light gray
	}

	draculaTheme = Theme{
		StatusBar:    "248",    // Gray (from Dracula palette)
		StatusKey:    "117",    // Cyan (from Dracula palette)
		Title:        "189",    // Purple (from Dracula palette)
		Border:       "117",    // Cyan
		ListSelected: "235",    // Dark gray background
		ListHeader:   "189",    // Purple
		Cursor:       "235",    // Dark gray background
		HelpTitle:    "189",    // Purple
		HelpSection:  "228",    // Yellow (from Dracula palette)
		HelpKey:      "212",    // Pink (from Dracula palette)
		Error:        "211",    // Red (from Dracula palette)
		Success:      "84",     // Green (from Dracula palette)
		Label:        "141",    // Purple
		Faint:        "242",    // Gray
	}

	gruvboxTheme = Theme{
		StatusBar:    "245",    // Gray (from Gruvbox palette)
		StatusKey:    "109",    // Cyan (from Gruvbox palette)
		Title:        "208",    // Orange (from Gruvbox palette)
		Border:       "109",    // Cyan
		ListSelected: "237",    // Dark gray background (from Gruvbox palette)
		ListHeader:   "208",    // Orange
		Cursor:       "237",    // Dark gray background
		HelpTitle:    "208",    // Orange
		HelpSection:  "223",    // Yellow (from Gruvbox palette)
		HelpKey:      "214",    // Orange (from Gruvbox palette)
		Error:        "203",    // Red (from Gruvbox palette)
		Success:      "142",    // Green (from Gruvbox palette)
		Label:        "175",    // Pink (from Gruvbox palette)
		Faint:        "244",    // Gray
	}

	nordTheme = Theme{
		StatusBar:    "244",    // Gray (from Nord palette)
		StatusKey:    "136",    // Cyan (from Nord palette)
		Title:        "109",    // Blue (from Nord palette)
		Border:       "136",    // Cyan
		ListSelected: "236",    // Dark gray background (from Nord palette)
		ListHeader:   "109",    // Blue
		Cursor:       "236",    // Dark gray background
		HelpTitle:    "109",    // Blue
		HelpSection:  "172",    // Yellow (from Nord palette)
		HelpKey:      "140",    // Pink (from Nord palette)
		Error:        "167",    // Red (from Nord palette)
		Success:      "151",    // Green (from Nord palette)
		Label:        "146",    // Pink (from Nord palette)
		Faint:        "245",    // Gray
	}

	solarizedDarkTheme = Theme{
		StatusBar:    "244",    // Base1 (from Solarized palette)
		StatusKey:    "37",     // Cyan (from Solarized palette)
		Title:        "33",     // Blue (from Solarized palette)
		Border:       "37",     // Cyan
		ListSelected: "235",    // Dark gray background
		ListHeader:   "33",     // Blue
		Cursor:       "235",    // Dark gray background
		HelpTitle:    "33",     // Blue
		HelpSection:  "136",    // Yellow (from Solarized palette)
		HelpKey:      "169",    // Magenta (from Solarized palette)
		Error:        "160",    // Red (from Solarized palette)
		Success:      "106",    // Green (from Solarized palette)
		Label:        "125",    // Magenta (from Solarized palette)
		Faint:        "245",    // Base1
	}

	solarizedLightTheme = Theme{
		StatusBar:    "242",    // Base01 (from Solarized palette)
		StatusKey:    "30",     // Cyan (from Solarized palette)
		Title:        "31",     // Blue (from Solarized palette)
		Border:       "30",     // Cyan
		ListSelected: "230",    // Light background
		ListHeader:   "31",     // Blue
		Cursor:       "230",    // Light background
		HelpTitle:    "31",     // Blue
		HelpSection:  "130",    // Yellow (from Solarized palette)
		HelpKey:      "163",    // Magenta (from Solarized palette)
		Error:        "160",    // Red (from Solarized palette)
		Success:      "64",     // Green (from Solarized palette)
		Label:        "89",     // Magenta (from Solarized palette)
		Faint:        "241",    // Base01
	}
)

// allThemes maintains a list of all available themes
var allThemes = []string{
	"default",
	"dracula",
	"gruvbox",
	"nord",
	"solarized-dark",
	"solarized-light",
}

// themeMap maps theme names to theme definitions
var themeMap = map[string]Theme{
	"default":        defaultTheme,
	"dracula":        draculaTheme,
	"gruvbox":        gruvboxTheme,
	"nord":           nordTheme,
	"solarized-dark": solarizedDarkTheme,
	"solarized-light": solarizedLightTheme,
}

// GetTheme returns a theme by name, or default theme if not found
func GetTheme(name string) *Theme {
	// Normalize theme name
	name = strings.ToLower(strings.TrimSpace(name))

	if theme, ok := themeMap[name]; ok {
		return &theme
	}

	// Return default theme for unknown names
	return &defaultTheme
}

// GetDefaultTheme returns the default theme
func GetDefaultTheme() *Theme {
	return &defaultTheme
}

// IsValidTheme checks if a theme name is valid
func IsValidTheme(name string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	_, ok := themeMap[name]
	return ok
}

// GetAllThemes returns a list of all available theme names
func GetAllThemes() []string {
	return allThemes
}
