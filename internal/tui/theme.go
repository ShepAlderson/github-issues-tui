package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme represents a color theme for the TUI
type Theme struct {
	Name        string
	Description string
	Styles      ThemeStyles
}

// ThemeStyles defines all the style elements for a theme
type ThemeStyles struct {
	// General styles
	TextPrimary   lipgloss.Color
	TextSecondary lipgloss.Color
	TextDim       lipgloss.Color
	TextHighlight lipgloss.Color

	// Status indicators
	StatusSuccess lipgloss.Color
	StatusWarning lipgloss.Color
	StatusError   lipgloss.Color
	StatusInfo    lipgloss.Color

	// UI elements
	Border       lipgloss.Color
	Background   lipgloss.Color
	ModalBorder  lipgloss.Color
	ModalBg      lipgloss.Color
	ModalText    lipgloss.Color

	// Issue list specific
	ListHeader   lipgloss.Color
	ListSelected lipgloss.Color

	// Issue detail specific
	DetailTitle   lipgloss.Color
	DetailLabel   lipgloss.Color
	DetailValue   lipgloss.Color
	DetailStateOpen lipgloss.Color
	DetailStateClosed lipgloss.Color

	// Comments specific
	CommentAuthor lipgloss.Color
	CommentDate   lipgloss.Color

	// Footer/Help specific
	FooterText   lipgloss.Color
	HelpKey      lipgloss.Color
	HelpDesc     lipgloss.Color
	HelpCategory lipgloss.Color
}

// ThemeManager manages color themes
type ThemeManager struct {
	themes map[string]Theme
	currentTheme string
}

// NewThemeManager creates a new theme manager
func NewThemeManager() *ThemeManager {
	tm := &ThemeManager{
		themes: make(map[string]Theme),
		currentTheme: "default",
	}
	tm.registerThemes()
	return tm
}

// registerThemes registers all built-in themes
func (tm *ThemeManager) registerThemes() {
	tm.themes["default"] = Theme{
		Name:        "default",
		Description: "Default terminal colors",
		Styles: ThemeStyles{
			TextPrimary:   lipgloss.Color("15"),      // Bright white
			TextSecondary: lipgloss.Color("250"),     // Light gray
			TextDim:       lipgloss.Color("240"),     // Dark gray
			TextHighlight: lipgloss.Color("205"),     // Magenta

			StatusSuccess: lipgloss.Color("46"),      // Green
			StatusWarning: lipgloss.Color("214"),     // Orange
			StatusError:   lipgloss.Color("9"),       // Red
			StatusInfo:    lipgloss.Color("39"),      // Bright blue

			Border:       lipgloss.Color("8"),        // Dark gray
			Background:   lipgloss.Color("0"),        // Black
			ModalBorder:  lipgloss.Color("39"),       // Bright blue
			ModalBg:      lipgloss.Color("0"),        // Black
			ModalText:    lipgloss.Color("15"),       // White

			ListHeader:   lipgloss.Color("240"),      // Dark gray
			ListSelected: lipgloss.Color("205"),      // Magenta

			DetailTitle:   lipgloss.Color("39"),      // Bright blue
			DetailLabel:   lipgloss.Color("245"),     // Gray
			DetailValue:   lipgloss.Color("214"),     // Orange
			DetailStateOpen: lipgloss.Color("46"),    // Green
			DetailStateClosed: lipgloss.Color("196"), // Red

			CommentAuthor: lipgloss.Color("39"),      // Bright blue
			CommentDate:   lipgloss.Color("214"),     // Orange

			FooterText:   lipgloss.Color("8"),        // Dark gray
			HelpKey:      lipgloss.Color("39"),       // Bright blue
			HelpDesc:     lipgloss.Color("214"),      // Orange
			HelpCategory: lipgloss.Color("46"),       // Green
		},
	}

	tm.themes["dracula"] = Theme{
		Name:        "dracula",
		Description: "Dracula theme - dark purple/cyan",
		Styles: ThemeStyles{
			TextPrimary:   lipgloss.Color("#f8f8f2"),
			TextSecondary: lipgloss.Color("#6272a4"),
			TextDim:       lipgloss.Color("#44475a"),
			TextHighlight: lipgloss.Color("#ff79c6"),

			StatusSuccess: lipgloss.Color("#50fa7b"),
			StatusWarning: lipgloss.Color("#f1fa8c"),
			StatusError:   lipgloss.Color("#ff5555"),
			StatusInfo:    lipgloss.Color("#8be9fd"),

			Border:       lipgloss.Color("#6272a4"),
			Background:   lipgloss.Color("#282a36"),
			ModalBorder:  lipgloss.Color("#8be9fd"),
			ModalBg:      lipgloss.Color("#282a36"),
			ModalText:    lipgloss.Color("#f8f8f2"),

			ListHeader:   lipgloss.Color("#6272a4"),
			ListSelected: lipgloss.Color("#ff79c6"),

			DetailTitle:   lipgloss.Color("#8be9fd"),
			DetailLabel:   lipgloss.Color("#6272a4"),
			DetailValue:   lipgloss.Color("#f1fa8c"),
			DetailStateOpen: lipgloss.Color("#50fa7b"),
			DetailStateClosed: lipgloss.Color("#ff5555"),

			CommentAuthor: lipgloss.Color("#8be9fd"),
			CommentDate:   lipgloss.Color("#f1fa8c"),

			FooterText:   lipgloss.Color("#6272a4"),
			HelpKey:      lipgloss.Color("#8be9fd"),
			HelpDesc:     lipgloss.Color("#f1fa8c"),
			HelpCategory: lipgloss.Color("#50fa7b"),
		},
	}

	tm.themes["gruvbox"] = Theme{
		Name:        "gruvbox",
		Description: "Gruvbox theme - warm earthy colors",
		Styles: ThemeStyles{
			TextPrimary:   lipgloss.Color("#ebdbb2"),
			TextSecondary: lipgloss.Color("#a89984"),
			TextDim:       lipgloss.Color("#665c54"),
			TextHighlight: lipgloss.Color("#d3869b"),

			StatusSuccess: lipgloss.Color("#b8bb26"),
			StatusWarning: lipgloss.Color("#fabd2f"),
			StatusError:   lipgloss.Color("#fb4934"),
			StatusInfo:    lipgloss.Color("#83a598"),

			Border:       lipgloss.Color("#665c54"),
			Background:   lipgloss.Color("#282828"),
			ModalBorder:  lipgloss.Color("#83a598"),
			ModalBg:      lipgloss.Color("#282828"),
			ModalText:    lipgloss.Color("#ebdbb2"),

			ListHeader:   lipgloss.Color("#a89984"),
			ListSelected: lipgloss.Color("#d3869b"),

			DetailTitle:   lipgloss.Color("#83a598"),
			DetailLabel:   lipgloss.Color("#a89984"),
			DetailValue:   lipgloss.Color("#fabd2f"),
			DetailStateOpen: lipgloss.Color("#b8bb26"),
			DetailStateClosed: lipgloss.Color("#fb4934"),

			CommentAuthor: lipgloss.Color("#83a598"),
			CommentDate:   lipgloss.Color("#fabd2f"),

			FooterText:   lipgloss.Color("#665c54"),
			HelpKey:      lipgloss.Color("#83a598"),
			HelpDesc:     lipgloss.Color("#fabd2f"),
			HelpCategory: lipgloss.Color("#b8bb26"),
		},
	}

	tm.themes["nord"] = Theme{
		Name:        "nord",
		Description: "Nord theme - arctic blue/gray",
		Styles: ThemeStyles{
			TextPrimary:   lipgloss.Color("#eceff4"),
			TextSecondary: lipgloss.Color("#81a1c1"),
			TextDim:       lipgloss.Color("#4c566a"),
			TextHighlight: lipgloss.Color("#b48ead"),

			StatusSuccess: lipgloss.Color("#a3be8c"),
			StatusWarning: lipgloss.Color("#ebcb8b"),
			StatusError:   lipgloss.Color("#bf616a"),
			StatusInfo:    lipgloss.Color("#88c0d0"),

			Border:       lipgloss.Color("#4c566a"),
			Background:   lipgloss.Color("#2e3440"),
			ModalBorder:  lipgloss.Color("#88c0d0"),
			ModalBg:      lipgloss.Color("#2e3440"),
			ModalText:    lipgloss.Color("#eceff4"),

			ListHeader:   lipgloss.Color("#4c566a"),
			ListSelected: lipgloss.Color("#b48ead"),

			DetailTitle:   lipgloss.Color("#88c0d0"),
			DetailLabel:   lipgloss.Color("#81a1c1"),
			DetailValue:   lipgloss.Color("#ebcb8b"),
			DetailStateOpen: lipgloss.Color("#a3be8c"),
			DetailStateClosed: lipgloss.Color("#bf616a"),

			CommentAuthor: lipgloss.Color("#88c0d0"),
			CommentDate:   lipgloss.Color("#ebcb8b"),

			FooterText:   lipgloss.Color("#4c566a"),
			HelpKey:      lipgloss.Color("#88c0d0"),
			HelpDesc:     lipgloss.Color("#ebcb8b"),
			HelpCategory: lipgloss.Color("#a3be8c"),
		},
	}

	tm.themes["solarized-dark"] = Theme{
		Name:        "solarized-dark",
		Description: "Solarized dark theme - low contrast",
		Styles: ThemeStyles{
			TextPrimary:   lipgloss.Color("#839496"),
			TextSecondary: lipgloss.Color("#657b83"),
			TextDim:       lipgloss.Color("#586e75"),
			TextHighlight: lipgloss.Color("#d33682"),

			StatusSuccess: lipgloss.Color("#859900"),
			StatusWarning: lipgloss.Color("#b58900"),
			StatusError:   lipgloss.Color("#dc322f"),
			StatusInfo:    lipgloss.Color("#2aa198"),

			Border:       lipgloss.Color("#586e75"),
			Background:   lipgloss.Color("#002b36"),
			ModalBorder:  lipgloss.Color("#2aa198"),
			ModalBg:      lipgloss.Color("#002b36"),
			ModalText:    lipgloss.Color("#839496"),

			ListHeader:   lipgloss.Color("#586e75"),
			ListSelected: lipgloss.Color("#d33682"),

			DetailTitle:   lipgloss.Color("#2aa198"),
			DetailLabel:   lipgloss.Color("#657b83"),
			DetailValue:   lipgloss.Color("#b58900"),
			DetailStateOpen: lipgloss.Color("#859900"),
			DetailStateClosed: lipgloss.Color("#dc322f"),

			CommentAuthor: lipgloss.Color("#2aa198"),
			CommentDate:   lipgloss.Color("#b58900"),

			FooterText:   lipgloss.Color("#586e75"),
			HelpKey:      lipgloss.Color("#2aa198"),
			HelpDesc:     lipgloss.Color("#b58900"),
			HelpCategory: lipgloss.Color("#859900"),
		},
	}

	tm.themes["solarized-light"] = Theme{
		Name:        "solarized-light",
		Description: "Solarized light theme - low contrast",
		Styles: ThemeStyles{
			TextPrimary:   lipgloss.Color("#657b83"),
			TextSecondary: lipgloss.Color("#839496"),
			TextDim:       lipgloss.Color("#93a1a1"),
			TextHighlight: lipgloss.Color("#d33682"),

			StatusSuccess: lipgloss.Color("#859900"),
			StatusWarning: lipgloss.Color("#b58900"),
			StatusError:   lipgloss.Color("#dc322f"),
			StatusInfo:    lipgloss.Color("#2aa198"),

			Border:       lipgloss.Color("#93a1a1"),
			Background:   lipgloss.Color("#fdf6e3"),
			ModalBorder:  lipgloss.Color("#2aa198"),
			ModalBg:      lipgloss.Color("#fdf6e3"),
			ModalText:    lipgloss.Color("#657b83"),

			ListHeader:   lipgloss.Color("#93a1a1"),
			ListSelected: lipgloss.Color("#d33682"),

			DetailTitle:   lipgloss.Color("#2aa198"),
			DetailLabel:   lipgloss.Color("#839496"),
			DetailValue:   lipgloss.Color("#b58900"),
			DetailStateOpen: lipgloss.Color("#859900"),
			DetailStateClosed: lipgloss.Color("#dc322f"),

			CommentAuthor: lipgloss.Color("#2aa198"),
			CommentDate:   lipgloss.Color("#b58900"),

			FooterText:   lipgloss.Color("#93a1a1"),
			HelpKey:      lipgloss.Color("#2aa198"),
			HelpDesc:     lipgloss.Color("#b58900"),
			HelpCategory: lipgloss.Color("#859900"),
		},
	}
}

// SetTheme sets the current theme
func (tm *ThemeManager) SetTheme(name string) bool {
	if _, exists := tm.themes[name]; exists {
		tm.currentTheme = name
		return true
	}
	return false
}

// GetTheme returns the current theme
func (tm *ThemeManager) GetTheme() Theme {
	return tm.themes[tm.currentTheme]
}

// GetThemeNames returns a list of all available theme names
func (tm *ThemeManager) GetThemeNames() []string {
	names := make([]string, 0, len(tm.themes))
	for name := range tm.themes {
		names = append(names, name)
	}
	return names
}

// GetThemeByName returns a theme by name
func (tm *ThemeManager) GetThemeByName(name string) (Theme, bool) {
	theme, exists := tm.themes[name]
	return theme, exists
}

// Style methods for different UI components

// TextPrimary returns a style for primary text
func (ts *ThemeStyles) TextPrimaryStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ts.TextPrimary)
}

// TextSecondary returns a style for secondary text
func (ts *ThemeStyles) TextSecondaryStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ts.TextSecondary)
}

// TextDim returns a style for dim text
func (ts *ThemeStyles) TextDimStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ts.TextDim)
}

// TextHighlight returns a style for highlighted text
func (ts *ThemeStyles) TextHighlightStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ts.TextHighlight)
}

// StatusSuccess returns a style for success status
func (ts *ThemeStyles) StatusSuccessStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ts.StatusSuccess)
}

// StatusWarning returns a style for warning status
func (ts *ThemeStyles) StatusWarningStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ts.StatusWarning)
}

// StatusError returns a style for error status
func (ts *ThemeStyles) StatusErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ts.StatusError)
}

// StatusInfo returns a style for info status
func (ts *ThemeStyles) StatusInfoStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ts.StatusInfo)
}

// ListHeader returns a style for list headers
func (ts *ThemeStyles) ListHeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ts.ListHeader).Bold(true)
}

// ListSelected returns a style for selected list items
func (ts *ThemeStyles) ListSelectedStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ts.ListSelected)
}

// DetailTitle returns a style for detail titles
func (ts *ThemeStyles) DetailTitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ts.DetailTitle).Bold(true)
}

// DetailLabel returns a style for detail labels
func (ts *ThemeStyles) DetailLabelStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ts.DetailLabel)
}

// DetailValue returns a style for detail values
func (ts *ThemeStyles) DetailValueStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ts.DetailValue)
}

// DetailStateOpen returns a style for open state
func (ts *ThemeStyles) DetailStateOpenStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ts.DetailStateOpen)
}

// DetailStateClosed returns a style for closed state
func (ts *ThemeStyles) DetailStateClosedStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ts.DetailStateClosed)
}

// CommentAuthor returns a style for comment authors
func (ts *ThemeStyles) CommentAuthorStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ts.CommentAuthor).Bold(true)
}

// CommentDate returns a style for comment dates
func (ts *ThemeStyles) CommentDateStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ts.CommentDate)
}

// FooterText returns a style for footer text
func (ts *ThemeStyles) FooterTextStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ts.FooterText)
}

// ModalStyle returns a style for modal overlay
func (ts *ThemeStyles) ModalStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		BorderForeground(ts.ModalBorder).
		Border(lipgloss.RoundedBorder()).
		Background(ts.ModalBg).
		Foreground(ts.ModalText).
		Padding(1, 2)
}

// ErrorModalStyle returns a style for error modal
func (ts *ThemeStyles) ErrorModalStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		BorderForeground(ts.StatusError).
		Border(lipgloss.RoundedBorder()).
		Background(ts.ModalBg).
		Foreground(ts.ModalText).
		Padding(1, 2)
}

// HelpModalStyle returns a style for help modal
func (ts *ThemeStyles) HelpModalStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		BorderForeground(ts.StatusInfo).
		Border(lipgloss.RoundedBorder()).
		Background(ts.ModalBg).
		Foreground(ts.ModalText).
		Padding(1, 2)
}

// HelpKeyStyle returns a style for help key bindings
func (ts *ThemeStyles) HelpKeyStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ts.HelpKey).Bold(true)
}

// HelpDescStyle returns a style for help descriptions
func (ts *ThemeStyles) HelpDescStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ts.HelpDesc)
}

// HelpCategoryStyle returns a style for help categories
func (ts *ThemeStyles) HelpCategoryStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ts.HelpCategory).Bold(true)
}