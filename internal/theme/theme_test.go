package theme

import (
	"testing"
)

func TestGetTheme(t *testing.T) {
	tests := []struct {
		name       string
		themeName  string
		wantNotNil bool
	}{
		{
			name:       "default theme",
			themeName:  "default",
			wantNotNil: true,
		},
		{
			name:       "dracula theme",
			themeName:  "dracula",
			wantNotNil: true,
		},
		{
			name:       "gruvbox theme",
			themeName:  "gruvbox",
			wantNotNil: true,
		},
		{
			name:       "nord theme",
			themeName:  "nord",
			wantNotNil: true,
		},
		{
			name:       "solarized-dark theme",
			themeName:  "solarized-dark",
			wantNotNil: true,
		},
		{
			name:       "solarized-light theme",
			themeName:  "solarized-light",
			wantNotNil: true,
		},
		{
			name:       "invalid theme falls back to default",
			themeName:  "invalid",
			wantNotNil: true,
		},
		{
			name:       "empty theme name falls back to default",
			themeName:  "",
			wantNotNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTheme(tt.themeName)
			if !tt.wantNotNil && got == nil {
				t.Errorf("GetTheme() = nil, want non-nil")
			}
			if tt.wantNotNil && got == nil {
				t.Errorf("GetTheme() = nil, want non-nil")
			}
		})
	}
}

func TestThemeValidation(t *testing.T) {
	tests := []struct {
		name      string
		themeName string
		wantValid bool
	}{
		{
			name:      "valid theme - default",
			themeName: "default",
			wantValid: true,
		},
		{
			name:      "valid theme - dracula",
			themeName: "dracula",
			wantValid: true,
		},
		{
			name:      "valid theme - gruvbox",
			themeName: "gruvbox",
			wantValid: true,
		},
		{
			name:      "valid theme - nord",
			themeName: "nord",
			wantValid: true,
		},
		{
			name:      "valid theme - solarized-dark",
			themeName: "solarized-dark",
			wantValid: true,
		},
		{
			name:      "valid theme - solarized-light",
			themeName: "solarized-light",
			wantValid: true,
		},
		{
			name:      "invalid theme",
			themeName: "invalid",
			wantValid: false,
		},
		{
			name:      "empty theme name",
			themeName: "",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidTheme(tt.themeName)
			if got != tt.wantValid {
				t.Errorf("IsValidTheme() = %v, want %v", got, tt.wantValid)
			}
		})
	}
}

func TestGetAllThemes(t *testing.T) {
	themes := GetAllThemes()

	if len(themes) == 0 {
		t.Errorf("GetAllThemes() returned empty list")
	}

	// Check that all expected themes are present
	expectedThemes := []string{"default", "dracula", "gruvbox", "nord", "solarized-dark", "solarized-light"}
	themeMap := make(map[string]bool)
	for _, theme := range themes {
		themeMap[theme] = true
	}

	for _, expected := range expectedThemes {
		if !themeMap[expected] {
			t.Errorf("GetAllThemes() missing theme: %s", expected)
		}
	}
}

func TestThemeColors(t *testing.T) {
	// Test that all themes have required color fields
	themes := GetAllThemes()

	for _, themeName := range themes {
		t.Run(themeName, func(t *testing.T) {
			theme := GetTheme(themeName)
			if theme == nil {
				t.Errorf("GetTheme(%s) returned nil", themeName)
				return
			}

			// Check that no color field is empty
			if theme.StatusBar == "" {
				t.Errorf("theme.StatusBar is empty")
			}
			if theme.StatusKey == "" {
				t.Errorf("theme.StatusKey is empty")
			}
			if theme.Title == "" {
				t.Errorf("theme.Title is empty")
			}
			if theme.Border == "" {
				t.Errorf("theme.Border is empty")
			}
			if theme.ListSelected == "" {
				t.Errorf("theme.ListSelected is empty")
			}
			if theme.ListHeader == "" {
				t.Errorf("theme.ListHeader is empty")
			}
			if theme.Cursor == "" {
				t.Errorf("theme.Cursor is empty")
			}
			if theme.HelpTitle == "" {
				t.Errorf("theme.HelpTitle is empty")
			}
			if theme.HelpSection == "" {
				t.Errorf("theme.HelpSection is empty")
			}
			if theme.HelpKey == "" {
				t.Errorf("theme.HelpKey is empty")
			}
			if theme.Error == "" {
				t.Errorf("theme.Error is empty")
			}
			if theme.Success == "" {
				t.Errorf("theme.Success is empty")
			}
			if theme.Label == "" {
				t.Errorf("theme.Label is empty")
			}
			if theme.Faint == "" {
				t.Errorf("theme.Faint is empty")
			}
		})
	}
}

func TestDefaultTheme(t *testing.T) {
	theme := GetDefaultTheme()

	if theme == nil {
		t.Errorf("GetDefaultTheme() returned nil")
	}

	// Verify default theme has specific colors
	if theme.StatusBar != "242" {
		t.Errorf("GetDefaultTheme().StatusBar = %s, want 242", theme.StatusBar)
	}
	if theme.Title != "blue" {
		t.Errorf("GetDefaultTheme().Title = %s, want blue", theme.Title)
	}
}

func TestThemeUniqueness(t *testing.T) {
	// Test that different themes have different color schemes
	themes := GetAllThemes()

	seen := make(map[string]string)
	for _, themeName := range themes {
		theme := GetTheme(themeName)
		if theme == nil {
			continue
		}

		// Create a signature from key colors
		signature := theme.StatusBar + theme.Title + theme.Border + theme.Error

		if existing, exists := seen[signature]; exists {
			t.Errorf("Theme %s has identical colors to theme %s", themeName, existing)
		}
		seen[signature] = themeName
	}
}
