package themes

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/shepbook/ghissues/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestGetTheme(t *testing.T) {
	tests := []struct {
		name      string
		themeName config.Theme
		wantPanic bool
	}{
		{name: "default theme", themeName: config.ThemeDefault, wantPanic: false},
		{name: "dracula theme", themeName: config.ThemeDracula, wantPanic: false},
		{name: "gruvbox theme", themeName: config.ThemeGruvbox, wantPanic: false},
		{name: "nord theme", themeName: config.ThemeNord, wantPanic: false},
		{name: "solarized-dark theme", themeName: config.ThemeSolarizedDark, wantPanic: false},
		{name: "solarized-light theme", themeName: config.ThemeSolarizedLight, wantPanic: false},
		{name: "empty theme returns default", themeName: "", wantPanic: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theme := GetTheme(tt.themeName)
			assert.NotNil(t, theme)
			assert.NotEmpty(t, theme.Name)
		})
	}
}

func TestThemeHasAllRequiredColors(t *testing.T) {
	themes := []config.Theme{
		config.ThemeDefault,
		config.ThemeDracula,
		config.ThemeGruvbox,
		config.ThemeNord,
		config.ThemeSolarizedDark,
		config.ThemeSolarizedLight,
	}

	for _, themeName := range themes {
		t.Run(string(themeName), func(t *testing.T) {
			theme := GetTheme(themeName)

			// Each theme should have all required colors defined
			assert.NotEqual(t, lipgloss.Color(""), theme.Primary, "Primary color should be defined")
			assert.NotEqual(t, lipgloss.Color(""), theme.Secondary, "Secondary color should be defined")
			assert.NotEqual(t, lipgloss.Color(""), theme.Accent, "Accent color should be defined")
			assert.NotEqual(t, lipgloss.Color(""), theme.Text, "Text color should be defined")
			assert.NotEqual(t, lipgloss.Color(""), theme.TextMuted, "TextMuted color should be defined")
			assert.NotEqual(t, lipgloss.Color(""), theme.Error, "Error color should be defined")
			assert.NotEqual(t, lipgloss.Color(""), theme.Warning, "Warning color should be defined")
			assert.NotEqual(t, lipgloss.Color(""), theme.Success, "Success color should be defined")
			assert.NotEqual(t, lipgloss.Color(""), theme.Selected, "Selected color should be defined")
			assert.NotEqual(t, lipgloss.Color(""), theme.Border, "Border color should be defined")
		})
	}
}

func TestThemeStyles(t *testing.T) {
	theme := GetTheme(config.ThemeDefault)

	// Verify styles are created from the theme
	styles := theme.Styles()

	// Check that styles are not nil (they should have been created)
	assert.NotNil(t, styles.Title)
	assert.NotNil(t, styles.Header)
	assert.NotNil(t, styles.Selected)
	assert.NotNil(t, styles.Normal)
	assert.NotNil(t, styles.Muted)
	assert.NotNil(t, styles.Error)
	assert.NotNil(t, styles.Warning)
	assert.NotNil(t, styles.Label)
	assert.NotNil(t, styles.Status)
	assert.NotNil(t, styles.Border)
}

func TestAllThemesReturnStyles(t *testing.T) {
	themes := []config.Theme{
		config.ThemeDefault,
		config.ThemeDracula,
		config.ThemeGruvbox,
		config.ThemeNord,
		config.ThemeSolarizedDark,
		config.ThemeSolarizedLight,
	}

	for _, themeName := range themes {
		t.Run(string(themeName), func(t *testing.T) {
			theme := GetTheme(themeName)
			styles := theme.Styles()
			assert.NotNil(t, styles, "Styles should not be nil for theme %s", themeName)
		})
	}
}

func TestThemeNameMatches(t *testing.T) {
	tests := []struct {
		themeName    config.Theme
		expectedName string
	}{
		{config.ThemeDefault, "default"},
		{config.ThemeDracula, "dracula"},
		{config.ThemeGruvbox, "gruvbox"},
		{config.ThemeNord, "nord"},
		{config.ThemeSolarizedDark, "solarized-dark"},
		{config.ThemeSolarizedLight, "solarized-light"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedName, func(t *testing.T) {
			theme := GetTheme(tt.themeName)
			assert.Equal(t, tt.expectedName, theme.Name)
		})
	}
}
