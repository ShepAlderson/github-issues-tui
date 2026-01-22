package themes

import (
	"fmt"
	"strings"
)

// Theme represents a color theme for the application
type Theme struct {
	Name              string
	Description       string
	PrimaryTextColor  string
	SecondaryTextColor string
	ContrastBackgroundColor string
	PrimaryBackgroundColor string
	BorderColor       string
	HighlightColor    string
	ErrorColor        string
	SuccessColor      string
}

// Default theme - uses terminal-friendly colors
var Default = Theme{
	Name:              "default",
	Description:       "Default terminal colors",
	PrimaryTextColor:  "#ffffff",
	SecondaryTextColor: "#888888",
	ContrastBackgroundColor: "#444444",
	PrimaryBackgroundColor: "#000000",
	BorderColor:       "#666666",
	HighlightColor:    "#0088ff",
	ErrorColor:        "#ff5555",
	SuccessColor:      "#50fa7b",
}

// Dracula theme
var Dracula = Theme{
	Name:              "dracula",
	Description:       "Dracula color scheme",
	PrimaryTextColor:  "#f8f8f2",
	SecondaryTextColor: "#6272a4",
	ContrastBackgroundColor: "#44475a",
	PrimaryBackgroundColor: "#282a36",
	BorderColor:       "#44475a",
	HighlightColor:    "#bd93f9",
	ErrorColor:        "#ff5555",
	SuccessColor:      "#50fa7b",
}

// Gruvbox theme
var Gruvbox = Theme{
	Name:              "gruvbox",
	Description:       "Gruvbox dark color scheme",
	PrimaryTextColor:  "#ebdbb2",
	SecondaryTextColor: "#928374",
	ContrastBackgroundColor: "#504030",
	PrimaryBackgroundColor: "#282828",
	BorderColor:       "#928374",
	HighlightColor:    "#fe8019",
	ErrorColor:        "#fb4934",
	SuccessColor:      "#b8bb26",
}

// Nord theme
var Nord = Theme{
	Name:              "nord",
	Description:       "Nord color scheme",
	PrimaryTextColor:  "#eceff4",
	SecondaryTextColor: "#88c0d0",
	ContrastBackgroundColor: "#4c566a",
	PrimaryBackgroundColor: "#2e3440",
	BorderColor:       "#4c566a",
	HighlightColor:    "#81a1c1",
	ErrorColor:        "#bf616a",
	SuccessColor:      "#a3be8c",
}

// SolarizedDark theme
var SolarizedDark = Theme{
	Name:              "solarized-dark",
	Description:       "Solarized dark color scheme",
	PrimaryTextColor:  "#839496",
	SecondaryTextColor: "#586e75",
	ContrastBackgroundColor: "#073642",
	PrimaryBackgroundColor: "#002b36",
	BorderColor:       "#586e75",
	HighlightColor:    "#268bd2",
	ErrorColor:        "#dc322f",
	SuccessColor:      "#859900",
}

// SolarizedLight theme
var SolarizedLight = Theme{
	Name:              "solarized-light",
	Description:       "Solarized light color scheme",
	PrimaryTextColor:  "#657b83",
	SecondaryTextColor: "#93a1a1",
	ContrastBackgroundColor: "#eee8d5",
	PrimaryBackgroundColor: "#fdf6e3",
	BorderColor:       "#93a1a1",
	HighlightColor:    "#268bd2",
	ErrorColor:        "#dc322f",
	SuccessColor:      "#859900",
}

// All returns all available themes
func All() []Theme {
	return []Theme{Default, Dracula, Gruvbox, Nord, SolarizedDark, SolarizedLight}
}

// Get returns the theme with the given name, or the Default theme if not found
func Get(name string) Theme {
	for _, theme := range All() {
		if theme.Name == name {
			return theme
		}
	}
	return Default
}

// IsValid checks if a theme name is valid
func IsValid(name string) bool {
	for _, theme := range All() {
		if theme.Name == name {
			return true
		}
	}
	return false
}

// DisplayName returns a formatted display name for the theme
func DisplayName(theme Theme) string {
	return fmt.Sprintf("%s - %s", theme.Name, theme.Description)
}

// List returns a formatted string listing all available themes
func List() string {
	var sb strings.Builder
	for i, theme := range All() {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(DisplayName(theme))
	}
	return sb.String()
}