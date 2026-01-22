package config

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme represents a color theme for the TUI
type Theme struct {
	Name string

	// Base colors
	Background lipgloss.Color
	Foreground lipgloss.Color

	// Accent colors
	Accent      lipgloss.Color
	AccentLight lipgloss.Color
	AccentDark  lipgloss.Color

	// UI Element colors
	Header      lipgloss.Color
	HeaderText  lipgloss.Color
	Border      lipgloss.Color
	BorderLight lipgloss.Color

	// Status colors
	Success lipgloss.Color
	Warning lipgloss.Color
	Error   lipgloss.Color

	// Text colors
	Text        lipgloss.Color
	TextLight   lipgloss.Color
	TextLighter lipgloss.Color
	TextMuted   lipgloss.Color

	// Selection colors
	SelectedFG lipgloss.Color
	SelectedBG lipgloss.Color
}

// GetTheme returns a theme by name, or default if not found
func GetTheme(name string) Theme {
	switch name {
	case "dracula":
		return DraculaTheme()
	case "gruvbox":
		return GruvboxTheme()
	case "nord":
		return NordTheme()
	case "solarized-dark":
		return SolarizedDarkTheme()
	case "solarized-light":
		return SolarizedLightTheme()
	default:
		return DefaultTheme()
	}
}

// DefaultTheme returns the default theme
func DefaultTheme() Theme {
	return Theme{
		Name:        "default",
		Background:  lipgloss.Color("0"),
		Foreground:  lipgloss.Color("7"),
		Accent:      lipgloss.Color("86"),
		AccentLight: lipgloss.Color("123"),
		AccentDark:  lipgloss.Color("36"),
		Header:      lipgloss.Color("57"),
		HeaderText:  lipgloss.Color("229"),
		Border:      lipgloss.Color("240"),
		BorderLight: lipgloss.Color("252"),
		Success:     lipgloss.Color("47"),
		Warning:     lipgloss.Color("226"),
		Error:       lipgloss.Color("196"),
		Text:        lipgloss.Color("255"),
		TextLight:   lipgloss.Color("252"),
		TextLighter: lipgloss.Color("250"),
		TextMuted:   lipgloss.Color("244"),
		SelectedFG:  lipgloss.Color("229"),
		SelectedBG:  lipgloss.Color("57"),
	}
}

// DraculaTheme returns the Dracula theme
func DraculaTheme() Theme {
	return Theme{
		Name:        "dracula",
		Background:  lipgloss.Color("235"), // Dracula background: #282936
		Foreground:  lipgloss.Color("252"), // Dracula foreground: #f8f8f2
		Accent:      lipgloss.Color("212"), // Dracula pink: #ff79c6
		AccentLight: lipgloss.Color("219"), // Dracula purple: #bd93f9
		AccentDark:  lipgloss.Color("204"), // Dracula red: #ff5555
		Header:      lipgloss.Color("62"),  // Dracula comment: #6272a4
		HeaderText:  lipgloss.Color("255"),
		Border:      lipgloss.Color("240"),
		BorderLight: lipgloss.Color("252"),
		Success:     lipgloss.Color("84"),  // Dracula green: #50fa7b
		Warning:     lipgloss.Color("215"), // Dracula orange: #ffb86c
		Error:       lipgloss.Color("204"), // Dracula red: #ff5555
		Text:        lipgloss.Color("255"),
		TextLight:   lipgloss.Color("252"),
		TextLighter: lipgloss.Color("250"),
		TextMuted:   lipgloss.Color("244"),
		SelectedFG:  lipgloss.Color("230"), // Near white for contrast
		SelectedBG:  lipgloss.Color("63"),  // Dracula cyan-like: #66d9ef
	}
}

// GruvboxTheme returns the Gruvbox dark theme
func GruvboxTheme() Theme {
	return Theme{
		Name:        "gruvbox",
		Background:  lipgloss.Color("235"), // Gruvbox dark0: #282828
		Foreground:  lipgloss.Color("223"), // Gruvbox light1: #ebdbb2
		Accent:      lipgloss.Color("214"), // Gruvbox neutral_orange: #fe8019
		AccentLight: lipgloss.Color("180"), // Gruvbox bright_orange: #fabd2f
		AccentDark:  lipgloss.Color("167"), // Gruvbox neutral_red: #fb4934
		Header:      lipgloss.Color("239"), // Gruvbox dark3: #665c54
		HeaderText:  lipgloss.Color("223"),
		Border:      lipgloss.Color("240"),
		BorderLight: lipgloss.Color("250"),
		Success:     lipgloss.Color("142"), // Gruvbox neutral_green: #b8bb26
		Warning:     lipgloss.Color("214"), // Gruvbox neutral_orange: #fe8019
		Error:       lipgloss.Color("167"), // Gruvbox neutral_red: #fb4934
		Text:        lipgloss.Color("223"),
		TextLight:   lipgloss.Color("250"),
		TextLighter: lipgloss.Color("248"),
		TextMuted:   lipgloss.Color("245"),
		SelectedFG:  lipgloss.Color("230"),
		SelectedBG:  lipgloss.Color("109"), // Gruvbox bright_aqua: #83a598
	}
}

// NordTheme returns the Nord theme
func NordTheme() Theme {
	return Theme{
		Name:        "nord",
		Background:  lipgloss.Color("236"), // Nord polar night 0: #2e3440
		Foreground:  lipgloss.Color("253"), // Nord snow storm 3: #eceff4
		Accent:      lipgloss.Color("110"), // Nord frost 1: #88c0d0
		AccentLight: lipgloss.Color("153"), // Nord frost 2: #81a1c1
		AccentDark:  lipgloss.Color("203"), // Nord aurora red: #bf616a
		Header:      lipgloss.Color("60"),  // Nord polar night 2: #434c5e
		HeaderText:  lipgloss.Color("255"),
		Border:      lipgloss.Color("102"), // Nord polar night 3: #4c566a
		BorderLight: lipgloss.Color("253"),
		Success:     lipgloss.Color("149"), // Nord aurora green: #a3be8c
		Warning:     lipgloss.Color("173"), // Nord aurora yellow: #ebcb8b
		Error:       lipgloss.Color("203"), // Nord aurora red: #bf616a
		Text:        lipgloss.Color("255"),
		TextLight:   lipgloss.Color("253"),
		TextLighter: lipgloss.Color("251"),
		TextMuted:   lipgloss.Color("244"),
		SelectedFG:  lipgloss.Color("255"),
		SelectedBG:  lipgloss.Color("110"), // Nord frost 1: #88c0d0
	}
}

// SolarizedDarkTheme returns the Solarized dark theme
func SolarizedDarkTheme() Theme {
	return Theme{
		Name:        "solarized-dark",
		Background:  lipgloss.Color("234"), // Solarized base03: #002b36
		Foreground:  lipgloss.Color("245"), // Solarized base0: #839496
		Accent:      lipgloss.Color("33"),  // Solarized blue: #268bd2
		AccentLight: lipgloss.Color("37"),  // Solarized cyan: #2aa198
		AccentDark:  lipgloss.Color("124"), // Solarized red: #dc322f
		Header:      lipgloss.Color("66"),  // Solarized base02: #073642
		HeaderText:  lipgloss.Color("254"),
		Border:      lipgloss.Color("102"), // Solarized base01: #586e75
		BorderLight: lipgloss.Color("252"),
		Success:     lipgloss.Color("106"), // Solarized green: #859900
		Warning:     lipgloss.Color("178"), // Solarized yellow: #b58900
		Error:       lipgloss.Color("124"), // Solarized red: #dc322f
		Text:        lipgloss.Color("254"),
		TextLight:   lipgloss.Color("252"),
		TextLighter: lipgloss.Color("250"),
		TextMuted:   lipgloss.Color("244"),
		SelectedFG:  lipgloss.Color("230"),
		SelectedBG:  lipgloss.Color("33"), // Solarized blue: #268bd2
	}
}

// SolarizedLightTheme returns the Solarized light theme
func SolarizedLightTheme() Theme {
	return Theme{
		Name:        "solarized-light",
		Background:  lipgloss.Color("230"), // Solarized base3: #fdf6e3
		Foreground:  lipgloss.Color("240"), // Solarized base00: #657b83
		Accent:      lipgloss.Color("32"),  // Solarized blue: #268bd2
		AccentLight: lipgloss.Color("36"),  // Solarized cyan: #2aa198
		AccentDark:  lipgloss.Color("160"), // Solarized red: #dc322f
		Header:      lipgloss.Color("253"), // Solarized base2: #eee8d5
		HeaderText:  lipgloss.Color("234"),
		Border:      lipgloss.Color("244"), // Solarized base1: #93a1a1
		BorderLight: lipgloss.Color("238"),
		Success:     lipgloss.Color("106"), // Solarized green: #859900
		Warning:     lipgloss.Color("178"), // Solarized yellow: #b58900
		Error:       lipgloss.Color("160"), // Solarized red: #dc322f
		Text:        lipgloss.Color("234"),
		TextLight:   lipgloss.Color("238"),
		TextLighter: lipgloss.Color("244"),
		TextMuted:   lipgloss.Color("242"),
		SelectedFG:  lipgloss.Color("234"),
		SelectedBG:  lipgloss.Color("32"), // Solarized blue: #268bd2
	}
}
