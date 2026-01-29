package theme

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestGetTheme(t *testing.T) {
	tests := []struct {
		name        string
		themeName   string
		wantDefault bool
	}{
		{
			name:        "returns default theme for 'default'",
			themeName:   "default",
			wantDefault: false,
		},
		{
			name:        "returns dracula theme",
			themeName:   "dracula",
			wantDefault: false,
		},
		{
			name:        "returns gruvbox theme",
			themeName:   "gruvbox",
			wantDefault: false,
		},
		{
			name:        "returns nord theme",
			themeName:   "nord",
			wantDefault: false,
		},
		{
			name:        "returns solarized-dark theme",
			themeName:   "solarized-dark",
			wantDefault: false,
		},
		{
			name:        "returns solarized-light theme",
			themeName:   "solarized-light",
			wantDefault: false,
		},
		{
			name:        "returns default theme for unknown theme name",
			themeName:   "unknown",
			wantDefault: true,
		},
		{
			name:        "returns default theme for empty theme name",
			themeName:   "",
			wantDefault: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theme := GetTheme(tt.themeName)

			if theme == nil {
				t.Fatal("GetTheme returned nil")
			}

			if tt.wantDefault && theme.Name != "default" {
				t.Errorf("expected default theme for %q, got %q", tt.themeName, theme.Name)
			}

			// Verify theme has all required colors set
			if theme.Primary == "" {
				t.Error("Primary color is empty")
			}
			if theme.Secondary == "" {
				t.Error("Secondary color is empty")
			}
			if theme.Text == "" {
				t.Error("Text color is empty")
			}
			if theme.Muted == "" {
				t.Error("Muted color is empty")
			}
			if theme.Error == "" {
				t.Error("Error color is empty")
			}
			if theme.Success == "" {
				t.Error("Success color is empty")
			}
			if theme.Open == "" {
				t.Error("Open color is empty")
			}
			if theme.Closed == "" {
				t.Error("Closed color is empty")
			}
			if theme.Label == "" {
				t.Error("Label color is empty")
			}
			if theme.Border == "" {
				t.Error("Border color is empty")
			}
			if theme.Background == "" {
				t.Error("Background color is empty")
			}
		})
	}
}

func TestGetThemeStyles(t *testing.T) {
	theme := GetTheme("default")
	styles := theme.Styles()

	if styles == nil {
		t.Fatal("Styles() returned nil")
	}

	// Verify styles are defined by checking they can render
	styles.Selected.Render("test")
	styles.Normal.Render("test")
	styles.Header.Render("test")
	styles.Status.Render("test")
	styles.Error.Render("test")
	styles.Title.Render("test")
	styles.Meta.Render("test")
	styles.StateOpen.Render("test")
	styles.StateClosed.Render("test")
	styles.Label.Render("test")
	styles.Body.Render("test")
	styles.Footer.Render("test")
	styles.Border.Render("test")
	styles.Section.Render("test")
	styles.Key.Render("test")
	styles.Desc.Render("test")
	styles.Success.Render("test")
	styles.Separator.Render("test")
	styles.CommentHeader.Render("test")
	styles.CommentBody.Render("test")
	styles.ModalBorder.Render("test")
	styles.ModalTitle.Render("test")
	styles.ModalGuidance.Render("test")
	styles.ModalFooter.Render("test")
	styles.Progress.Render("test")
}

func TestThemeColorsAreValid(t *testing.T) {
	themes := []string{"default", "dracula", "gruvbox", "nord", "solarized-dark", "solarized-light"}

	for _, name := range themes {
		t.Run(name, func(t *testing.T) {
			theme := GetTheme(name)

			// Test that colors can be parsed by lipgloss
			colors := []string{
				theme.Primary,
				theme.Secondary,
				theme.Text,
				theme.Muted,
				theme.Error,
				theme.Success,
				theme.Open,
				theme.Closed,
				theme.Label,
				theme.Border,
				theme.Background,
			}

			for _, color := range colors {
				if color == "" {
					t.Error("color is empty")
					continue
				}
				// Verify lipgloss can parse the color
				style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
				_ = style.Render("test")
			}
		})
	}
}

func TestGetAvailableThemes(t *testing.T) {
	themes := GetAvailableThemes()

	if len(themes) != 6 {
		t.Errorf("expected 6 themes, got %d", len(themes))
	}

	expected := map[string]bool{
		"default":         false,
		"dracula":         false,
		"gruvbox":         false,
		"nord":            false,
		"solarized-dark":  false,
		"solarized-light": false,
	}

	for _, theme := range themes {
		if _, ok := expected[theme]; !ok {
			t.Errorf("unexpected theme: %s", theme)
		}
		expected[theme] = true
	}

	for theme, found := range expected {
		if !found {
			t.Errorf("missing expected theme: %s", theme)
		}
	}
}

func TestIsValidTheme(t *testing.T) {
	tests := []struct {
		name     string
		theme    string
		expected bool
	}{
		{"valid default", "default", true},
		{"valid dracula", "dracula", true},
		{"valid gruvbox", "gruvbox", true},
		{"valid nord", "nord", true},
		{"valid solarized-dark", "solarized-dark", true},
		{"valid solarized-light", "solarized-light", true},
		{"invalid theme", "invalid", false},
		{"empty theme", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidTheme(tt.theme)
			if got != tt.expected {
				t.Errorf("IsValidTheme(%q) = %v, want %v", tt.theme, got, tt.expected)
			}
		})
	}
}

func TestThemeStyleRenders(t *testing.T) {
	// Test that all styles can actually render text without panicking
	theme := GetTheme("default")
	styles := theme.Styles()

	testCases := []struct {
		name  string
		style lipgloss.Style
	}{
		{"Selected", styles.Selected},
		{"Normal", styles.Normal},
		{"Header", styles.Header},
		{"Status", styles.Status},
		{"Error", styles.Error},
		{"Title", styles.Title},
		{"Meta", styles.Meta},
		{"StateOpen", styles.StateOpen},
		{"StateClosed", styles.StateClosed},
		{"Label", styles.Label},
		{"Body", styles.Body},
		{"Footer", styles.Footer},
		{"Border", styles.Border},
		{"Section", styles.Section},
		{"Key", styles.Key},
		{"Desc", styles.Desc},
		{"Success", styles.Success},
		{"Separator", styles.Separator},
		{"CommentHeader", styles.CommentHeader},
		{"CommentBody", styles.CommentBody},
		{"ModalBorder", styles.ModalBorder},
		{"ModalTitle", styles.ModalTitle},
		{"ModalGuidance", styles.ModalGuidance},
		{"ModalFooter", styles.ModalFooter},
		{"Progress", styles.Progress},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Should not panic
			result := tc.style.Render("test")
			if result == "" {
				t.Error("rendered empty string")
			}
		})
	}
}
