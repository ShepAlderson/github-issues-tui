package theme

import (
	"testing"
)

func TestGetTheme(t *testing.T) {
	tests := []struct {
		name      string
		themeName string
		wantName  string
	}{
		{"default theme", "default", "default"},
		{"dracula theme", "dracula", "dracula"},
		{"gruvbox theme", "gruvbox", "gruvbox"},
		{"nord theme", "nord", "nord"},
		{"solarized-dark theme", "solarized-dark", "solarized-dark"},
		{"solarized-light theme", "solarized-light", "solarized-light"},
		{"empty string defaults to default", "", "default"},
		{"unknown theme defaults to default", "unknown", "default"},
		{"case insensitive", "DRACULA", "dracula"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTheme(tt.themeName)
			if got.Name != tt.wantName {
				t.Errorf("GetTheme(%q).Name = %v, want %v", tt.themeName, got.Name, tt.wantName)
			}
		})
	}
}

func TestListThemes(t *testing.T) {
	themes := ListThemes()

	// Should have at least 6 themes
	if len(themes) < 6 {
		t.Errorf("ListThemes() returned %d themes, want at least 6", len(themes))
	}

	// Check that expected themes are present
	expectedThemes := []string{"default", "dracula", "gruvbox", "nord", "solarized-dark", "solarized-light"}
	themeMap := make(map[string]bool)
	for _, theme := range themes {
		themeMap[theme.Name] = true
	}

	for _, expected := range expectedThemes {
		if !themeMap[expected] {
			t.Errorf("ListThemes() missing expected theme: %s", expected)
		}
	}
}

func TestThemeHasAllStyles(t *testing.T) {
	themes := ListThemes()

	for _, theme := range themes {
		t.Run(theme.Name, func(t *testing.T) {
			// Check that styles can render (simple test to ensure they're initialized)
			// We just verify that rendering doesn't panic
			_ = theme.HeaderStyle.Render("test")
			_ = theme.NormalStyle.Render("test")
			_ = theme.SelectedStyle.Render("test")
			_ = theme.StatusStyle.Render("test")
			_ = theme.FooterStyle.Render("test")
			_ = theme.NoIssuesStyle.Render("test")
			_ = theme.DetailPanelStyle.Render("test")
			_ = theme.DetailHeaderStyle.Render("test")
			_ = theme.DetailTitleStyle.Render("test")
			_ = theme.DetailMetaStyle.Render("test")
			_ = theme.CommentsHeaderStyle.Render("test")
			_ = theme.CommentsMetaStyle.Render("test")
			_ = theme.CommentAuthorStyle.Render("test")
			_ = theme.CommentSeparatorStyle.Render("test")
			_ = theme.ErrorStyle.Render("test")
			_ = theme.ModalStyle.Render("test")
			_ = theme.ModalTitleStyle.Render("test")
			_ = theme.HelpOverlayStyle.Render("test")
			_ = theme.HelpTitleStyle.Render("test")
			_ = theme.HelpSectionStyle.Render("test")
			_ = theme.HelpKeyStyle.Render("test")
			_ = theme.HelpDescStyle.Render("test")
			_ = theme.HelpFooterStyle.Render("test")
		})
	}
}

func TestThemeColorsAreDifferent(t *testing.T) {
	// Verify that different themes have different colors
	defaultTheme := GetTheme("default")
	draculaTheme := GetTheme("dracula")

	// Get foreground colors from selected style
	defaultColor := defaultTheme.SelectedStyle.GetForeground()
	draculaColor := draculaTheme.SelectedStyle.GetForeground()

	if defaultColor == draculaColor {
		t.Error("Default and Dracula themes should have different selected colors")
	}
}
