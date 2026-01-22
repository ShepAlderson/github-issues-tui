package theme

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Theme represents a color theme for the TUI
type Theme struct {
	Name                 string
	HeaderStyle          lipgloss.Style
	NormalStyle          lipgloss.Style
	SelectedStyle        lipgloss.Style
	StatusStyle          lipgloss.Style
	FooterStyle          lipgloss.Style
	NoIssuesStyle        lipgloss.Style
	DetailPanelStyle     lipgloss.Style
	DetailHeaderStyle    lipgloss.Style
	DetailTitleStyle     lipgloss.Style
	DetailMetaStyle      lipgloss.Style
	CommentsHeaderStyle  lipgloss.Style
	CommentsMetaStyle    lipgloss.Style
	CommentAuthorStyle   lipgloss.Style
	CommentSeparatorStyle lipgloss.Style
	ErrorStyle           lipgloss.Style
	ModalStyle           lipgloss.Style
	ModalTitleStyle      lipgloss.Style
	HelpOverlayStyle     lipgloss.Style
	HelpTitleStyle       lipgloss.Style
	HelpSectionStyle     lipgloss.Style
	HelpKeyStyle         lipgloss.Style
	HelpDescStyle        lipgloss.Style
	HelpFooterStyle      lipgloss.Style
}

// GetTheme returns the theme with the given name
// Returns default theme if name is empty or unknown
func GetTheme(name string) Theme {
	name = strings.ToLower(strings.TrimSpace(name))

	switch name {
	case "dracula":
		return draculaTheme()
	case "gruvbox":
		return gruvboxTheme()
	case "nord":
		return nordTheme()
	case "solarized-dark":
		return solarizedDarkTheme()
	case "solarized-light":
		return solarizedLightTheme()
	default:
		return defaultTheme()
	}
}

// ListThemes returns a list of all available themes
func ListThemes() []Theme {
	return []Theme{
		defaultTheme(),
		draculaTheme(),
		gruvboxTheme(),
		nordTheme(),
		solarizedDarkTheme(),
		solarizedLightTheme(),
	}
}

// defaultTheme is the original color scheme
func defaultTheme() Theme {
	return Theme{
		Name: "default",
		HeaderStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")),
		NormalStyle: lipgloss.NewStyle(),
		SelectedStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")),
		StatusStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		FooterStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		NoIssuesStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true),
		DetailPanelStyle: lipgloss.NewStyle().
			Padding(1),
		DetailHeaderStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")),
		DetailTitleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")),
		DetailMetaStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		CommentsHeaderStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")),
		CommentsMetaStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		CommentAuthorStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")),
		CommentSeparatorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		ErrorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true),
		ModalStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("196")).
			Padding(1, 2).
			Width(60),
		ModalTitleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196")),
		HelpOverlayStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("170")).
			Padding(2, 4).
			Width(80),
		HelpTitleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")).
			Underline(true),
		HelpSectionStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")),
		HelpKeyStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")),
		HelpDescStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		HelpFooterStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true),
	}
}

// draculaTheme is inspired by the Dracula color scheme
func draculaTheme() Theme {
	return Theme{
		Name: "dracula",
		HeaderStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#bd93f9")), // purple
		NormalStyle: lipgloss.NewStyle(),
		SelectedStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ff79c6")), // pink
		StatusStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272a4")), // comment
		FooterStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272a4")),
		NoIssuesStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272a4")).
			Italic(true),
		DetailPanelStyle: lipgloss.NewStyle().
			Padding(1),
		DetailHeaderStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ff79c6")), // pink
		DetailTitleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#8be9fd")), // cyan
		DetailMetaStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272a4")),
		CommentsHeaderStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ff79c6")),
		CommentsMetaStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272a4")),
		CommentAuthorStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#8be9fd")),
		CommentSeparatorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272a4")),
		ErrorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff5555")). // red
			Bold(true),
		ModalStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#ff5555")).
			Padding(1, 2).
			Width(60),
		ModalTitleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ff5555")),
		HelpOverlayStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#bd93f9")).
			Padding(2, 4).
			Width(80),
		HelpTitleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#bd93f9")).
			Underline(true),
		HelpSectionStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#50fa7b")), // green
		HelpKeyStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ff79c6")),
		HelpDescStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f8f8f2")), // foreground
		HelpFooterStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272a4")).
			Italic(true),
	}
}

// gruvboxTheme is inspired by the Gruvbox color scheme
func gruvboxTheme() Theme {
	return Theme{
		Name: "gruvbox",
		HeaderStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#fabd2f")), // yellow
		NormalStyle: lipgloss.NewStyle(),
		SelectedStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#fe8019")), // orange
		StatusStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#928374")), // gray
		FooterStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#928374")),
		NoIssuesStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#928374")).
			Italic(true),
		DetailPanelStyle: lipgloss.NewStyle().
			Padding(1),
		DetailHeaderStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#fe8019")),
		DetailTitleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#83a598")), // blue
		DetailMetaStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#928374")),
		CommentsHeaderStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#fe8019")),
		CommentsMetaStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#928374")),
		CommentAuthorStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#83a598")),
		CommentSeparatorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#928374")),
		ErrorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#fb4934")). // red
			Bold(true),
		ModalStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#fb4934")).
			Padding(1, 2).
			Width(60),
		ModalTitleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#fb4934")),
		HelpOverlayStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#d3869b")). // purple
			Padding(2, 4).
			Width(80),
		HelpTitleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#d3869b")).
			Underline(true),
		HelpSectionStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#b8bb26")), // green
		HelpKeyStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#fe8019")),
		HelpDescStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ebdbb2")), // fg
		HelpFooterStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#928374")).
			Italic(true),
	}
}

// nordTheme is inspired by the Nord color scheme
func nordTheme() Theme {
	return Theme{
		Name: "nord",
		HeaderStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#88c0d0")), // nord8 (cyan)
		NormalStyle: lipgloss.NewStyle(),
		SelectedStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#81a1c1")), // nord9 (blue)
		StatusStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4c566a")), // nord3 (gray)
		FooterStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4c566a")),
		NoIssuesStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4c566a")).
			Italic(true),
		DetailPanelStyle: lipgloss.NewStyle().
			Padding(1),
		DetailHeaderStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#81a1c1")),
		DetailTitleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#88c0d0")),
		DetailMetaStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4c566a")),
		CommentsHeaderStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#81a1c1")),
		CommentsMetaStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4c566a")),
		CommentAuthorStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#88c0d0")),
		CommentSeparatorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4c566a")),
		ErrorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#bf616a")). // nord11 (red)
			Bold(true),
		ModalStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#bf616a")).
			Padding(1, 2).
			Width(60),
		ModalTitleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#bf616a")),
		HelpOverlayStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#b48ead")). // nord15 (purple)
			Padding(2, 4).
			Width(80),
		HelpTitleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#b48ead")).
			Underline(true),
		HelpSectionStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#a3be8c")), // nord14 (green)
		HelpKeyStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#81a1c1")),
		HelpDescStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#d8dee9")), // nord4 (fg)
		HelpFooterStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4c566a")).
			Italic(true),
	}
}

// solarizedDarkTheme is inspired by the Solarized Dark color scheme
func solarizedDarkTheme() Theme {
	return Theme{
		Name: "solarized-dark",
		HeaderStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#268bd2")), // blue
		NormalStyle: lipgloss.NewStyle(),
		SelectedStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#d33682")), // magenta
		StatusStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#586e75")), // base01
		FooterStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#586e75")),
		NoIssuesStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#586e75")).
			Italic(true),
		DetailPanelStyle: lipgloss.NewStyle().
			Padding(1),
		DetailHeaderStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#d33682")),
		DetailTitleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#2aa198")), // cyan
		DetailMetaStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#586e75")),
		CommentsHeaderStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#d33682")),
		CommentsMetaStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#586e75")),
		CommentAuthorStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#2aa198")),
		CommentSeparatorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#586e75")),
		ErrorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#dc322f")). // red
			Bold(true),
		ModalStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#dc322f")).
			Padding(1, 2).
			Width(60),
		ModalTitleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#dc322f")),
		HelpOverlayStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#6c71c4")). // violet
			Padding(2, 4).
			Width(80),
		HelpTitleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#6c71c4")).
			Underline(true),
		HelpSectionStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#859900")), // green
		HelpKeyStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#d33682")),
		HelpDescStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#93a1a1")), // base1
		HelpFooterStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#586e75")).
			Italic(true),
	}
}

// solarizedLightTheme is inspired by the Solarized Light color scheme
func solarizedLightTheme() Theme {
	return Theme{
		Name: "solarized-light",
		HeaderStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#268bd2")), // blue
		NormalStyle: lipgloss.NewStyle(),
		SelectedStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#d33682")), // magenta
		StatusStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#93a1a1")), // base1
		FooterStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#93a1a1")),
		NoIssuesStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#93a1a1")).
			Italic(true),
		DetailPanelStyle: lipgloss.NewStyle().
			Padding(1),
		DetailHeaderStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#d33682")),
		DetailTitleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#2aa198")), // cyan
		DetailMetaStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#93a1a1")),
		CommentsHeaderStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#d33682")),
		CommentsMetaStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#93a1a1")),
		CommentAuthorStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#2aa198")),
		CommentSeparatorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#93a1a1")),
		ErrorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#dc322f")). // red
			Bold(true),
		ModalStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#dc322f")).
			Padding(1, 2).
			Width(60),
		ModalTitleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#dc322f")),
		HelpOverlayStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#6c71c4")). // violet
			Padding(2, 4).
			Width(80),
		HelpTitleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#6c71c4")).
			Underline(true),
		HelpSectionStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#859900")), // green
		HelpKeyStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#d33682")),
		HelpDescStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#586e75")), // base01
		HelpFooterStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#93a1a1")).
			Italic(true),
	}
}
