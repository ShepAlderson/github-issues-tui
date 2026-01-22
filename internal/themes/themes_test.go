package themes

import (
	"testing"
)

func TestAll(t *testing.T) {
	themes := All()
	expected := 6
	if len(themes) != expected {
		t.Errorf("Expected %d themes, got %d", expected, len(themes))
	}

	// Check that all themes have unique names
	names := make(map[string]bool)
	for _, theme := range themes {
		if names[theme.Name] {
			t.Errorf("Duplicate theme name: %s", theme.Name)
		}
		names[theme.Name] = true
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"default", "default"},
		{"dracula", "dracula"},
		{"gruvbox", "gruvbox"},
		{"nord", "nord"},
		{"solarized-dark", "solarized-dark"},
		{"solarized-light", "solarized-light"},
		{"unknown", "default"}, // Unknown theme returns Default
		{"", "default"},        // Empty string returns Default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theme := Get(tt.name)
			if theme.Name != tt.expected {
				t.Errorf("Get(%q) = %s, want %s", tt.name, theme.Name, tt.expected)
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	validThemes := []string{"default", "dracula", "gruvbox", "nord", "solarized-dark", "solarized-light"}
	invalidThemes := []string{"", "unknown", "dark", "light", " Dracula "}

	for _, name := range validThemes {
		t.Run("valid_"+name, func(t *testing.T) {
			if !IsValid(name) {
				t.Errorf("IsValid(%q) = false, want true", name)
			}
		})
	}

	for _, name := range invalidThemes {
		t.Run("invalid_"+name, func(t *testing.T) {
			if IsValid(name) {
				t.Errorf("IsValid(%q) = true, want false", name)
			}
		})
	}
}

func TestDisplayName(t *testing.T) {
	tests := []struct {
		themeName string
		expected  string
	}{
		{"default", "default - Default terminal colors"},
		{"dracula", "dracula - Dracula color scheme"},
		{"gruvbox", "gruvbox - Gruvbox dark color scheme"},
		{"nord", "nord - Nord color scheme"},
		{"solarized-dark", "solarized-dark - Solarized dark color scheme"},
		{"solarized-light", "solarized-light - Solarized light color scheme"},
	}

	for _, tt := range tests {
		t.Run(tt.themeName, func(t *testing.T) {
			theme := Get(tt.themeName)
			displayName := DisplayName(theme)
			if displayName != tt.expected {
				t.Errorf("DisplayName(%s) = %q, want %q", tt.themeName, displayName, tt.expected)
			}
		})
	}
}

func TestList(t *testing.T) {
	list := List()
	// Should contain all theme names
	expectedThemes := []string{"default", "dracula", "gruvbox", "nord", "solarized-dark", "solarized-light"}
	for _, name := range expectedThemes {
		if !contains(list, name) {
			t.Errorf("List() should contain %q", name)
		}
	}
	// Should contain descriptions
	for _, theme := range All() {
		if !contains(list, theme.Description) {
			t.Errorf("List() should contain description for %s", theme.Name)
		}
	}
}

func TestDefaultTheme(t *testing.T) {
	theme := Default
	if theme.Name != "default" {
		t.Errorf("Default theme name = %s, want default", theme.Name)
	}
	if theme.PrimaryTextColor == "" {
		t.Error("Default theme should have PrimaryTextColor")
	}
	if theme.SecondaryTextColor == "" {
		t.Error("Default theme should have SecondaryTextColor")
	}
	if theme.ContrastBackgroundColor == "" {
		t.Error("Default theme should have ContrastBackgroundColor")
	}
	if theme.BorderColor == "" {
		t.Error("Default theme should have BorderColor")
	}
	if theme.HighlightColor == "" {
		t.Error("Default theme should have HighlightColor")
	}
}

func TestThemeColorFormats(t *testing.T) {
	for _, theme := range All() {
		// All hex colors should be in #RRGGGBB format
		colors := []string{
			theme.PrimaryTextColor,
			theme.SecondaryTextColor,
			theme.ContrastBackgroundColor,
			theme.PrimaryBackgroundColor,
			theme.BorderColor,
			theme.HighlightColor,
			theme.ErrorColor,
			theme.SuccessColor,
		}
		for i, color := range colors {
			if len(color) != 7 || color[0] != '#' {
				t.Errorf("Theme %s has invalid color format at index %d: %q (expected #RRGGBB)", theme.Name, i, color)
			}
			// Check that the rest are valid hex digits
			for j := 1; j < 7; j++ {
				if !isHexDigit(color[j]) {
					t.Errorf("Theme %s has invalid hex digit %c in color %q", theme.Name, color[j], color)
				}
			}
		}
	}
}

func isHexDigit(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}