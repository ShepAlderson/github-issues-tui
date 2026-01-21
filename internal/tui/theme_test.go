package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestNewThemeManager(t *testing.T) {
	tm := NewThemeManager()
	if tm == nil {
		t.Fatal("Expected ThemeManager, got nil")
	}

	// Check that default theme is set
	defaultTheme := tm.GetTheme()
	if defaultTheme.Name != "default" {
		t.Errorf("Expected default theme, got %s", defaultTheme.Name)
	}

	// Check that all required themes exist
	requiredThemes := []string{"default", "dracula", "gruvbox", "nord", "solarized-dark", "solarized-light"}
	for _, themeName := range requiredThemes {
		theme, exists := tm.GetThemeByName(themeName)
		if !exists {
			t.Errorf("Expected theme %s to exist", themeName)
		}
		if theme.Name != themeName {
			t.Errorf("Expected theme name %s, got %s", themeName, theme.Name)
		}
	}
}

func TestSetTheme(t *testing.T) {
	tm := NewThemeManager()

	// Test setting to a valid theme
	if !tm.SetTheme("dracula") {
		t.Error("Expected SetTheme to return true for valid theme")
	}

	// Verify theme changed
	currentTheme := tm.GetTheme()
	if currentTheme.Name != "dracula" {
		t.Errorf("Expected theme 'dracula', got %s", currentTheme.Name)
	}

	// Test setting to an invalid theme
	if tm.SetTheme("invalid-theme") {
		t.Error("Expected SetTheme to return false for invalid theme")
	}

	// Verify theme didn't change
	currentTheme = tm.GetTheme()
	if currentTheme.Name != "dracula" {
		t.Errorf("Expected theme to remain 'dracula', got %s", currentTheme.Name)
	}
}

func TestGetThemeNames(t *testing.T) {
	tm := NewThemeManager()
	names := tm.GetThemeNames()

	// Check that we have the expected number of themes
	expectedCount := 6
	if len(names) != expectedCount {
		t.Errorf("Expected %d theme names, got %d", expectedCount, len(names))
	}

	// Check that all required themes are in the list
	requiredThemes := map[string]bool{
		"default": true, "dracula": true, "gruvbox": true,
		"nord": true, "solarized-dark": true, "solarized-light": true,
	}

	for _, name := range names {
		if !requiredThemes[name] {
			t.Errorf("Unexpected theme name in list: %s", name)
		}
		delete(requiredThemes, name)
	}

	// Check that we found all required themes
	if len(requiredThemes) > 0 {
		t.Errorf("Missing themes in names list: %v", requiredThemes)
	}
}

func TestThemeStyles(t *testing.T) {
	tm := NewThemeManager()

	// Test each theme has valid styles
	for _, themeName := range tm.GetThemeNames() {
		theme, exists := tm.GetThemeByName(themeName)
		if !exists {
			t.Errorf("Theme %s should exist", themeName)
			continue
		}

		// Check that all style fields are set
		styles := theme.Styles

		// General styles
		if styles.TextPrimary == "" {
			t.Errorf("Theme %s: TextPrimary not set", themeName)
		}
		if styles.TextSecondary == "" {
			t.Errorf("Theme %s: TextSecondary not set", themeName)
		}
		if styles.TextDim == "" {
			t.Errorf("Theme %s: TextDim not set", themeName)
		}
		if styles.TextHighlight == "" {
			t.Errorf("Theme %s: TextHighlight not set", themeName)
		}

		// Status indicators
		if styles.StatusSuccess == "" {
			t.Errorf("Theme %s: StatusSuccess not set", themeName)
		}
		if styles.StatusWarning == "" {
			t.Errorf("Theme %s: StatusWarning not set", themeName)
		}
		if styles.StatusError == "" {
			t.Errorf("Theme %s: StatusError not set", themeName)
		}
		if styles.StatusInfo == "" {
			t.Errorf("Theme %s: StatusInfo not set", themeName)
		}

		// UI elements
		if styles.Border == "" {
			t.Errorf("Theme %s: Border not set", themeName)
		}
		if styles.Background == "" {
			t.Errorf("Theme %s: Background not set", themeName)
		}
		if styles.ModalBorder == "" {
			t.Errorf("Theme %s: ModalBorder not set", themeName)
		}
		if styles.ModalBg == "" {
			t.Errorf("Theme %s: ModalBg not set", themeName)
		}
		if styles.ModalText == "" {
			t.Errorf("Theme %s: ModalText not set", themeName)
		}

		// Issue list specific
		if styles.ListHeader == "" {
			t.Errorf("Theme %s: ListHeader not set", themeName)
		}
		if styles.ListSelected == "" {
			t.Errorf("Theme %s: ListSelected not set", themeName)
		}

		// Issue detail specific
		if styles.DetailTitle == "" {
			t.Errorf("Theme %s: DetailTitle not set", themeName)
		}
		if styles.DetailLabel == "" {
			t.Errorf("Theme %s: DetailLabel not set", themeName)
		}
		if styles.DetailValue == "" {
			t.Errorf("Theme %s: DetailValue not set", themeName)
		}
		if styles.DetailStateOpen == "" {
			t.Errorf("Theme %s: DetailStateOpen not set", themeName)
		}
		if styles.DetailStateClosed == "" {
			t.Errorf("Theme %s: DetailStateClosed not set", themeName)
		}

		// Comments specific
		if styles.CommentAuthor == "" {
			t.Errorf("Theme %s: CommentAuthor not set", themeName)
		}
		if styles.CommentDate == "" {
			t.Errorf("Theme %s: CommentDate not set", themeName)
		}

		// Footer/Help specific
		if styles.FooterText == "" {
			t.Errorf("Theme %s: FooterText not set", themeName)
		}
		if styles.HelpKey == "" {
			t.Errorf("Theme %s: HelpKey not set", themeName)
		}
		if styles.HelpDesc == "" {
			t.Errorf("Theme %s: HelpDesc not set", themeName)
		}
		if styles.HelpCategory == "" {
			t.Errorf("Theme %s: HelpCategory not set", themeName)
		}
	}
}

func TestLipglossColorFormat(t *testing.T) {
	// Test that colors are valid lipgloss.Color format
	tm := NewThemeManager()
	defaultTheme, exists := tm.GetThemeByName("default")
	if !exists {
		t.Fatal("Default theme should exist")
	}
	styles := defaultTheme.Styles

	// Test that we can create styles with these colors
	// This will panic if colors are invalid
	style := lipgloss.NewStyle().
		Foreground(styles.TextPrimary).
		Background(styles.Background).
		BorderForeground(styles.Border)

	// Just creating the style is enough to test validity
	_ = style
}