package config

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestGetTheme(t *testing.T) {
	tests := []struct {
		name       string
		themeName  string
		wantName   string
		checkField func(Theme) bool
	}{
		{
			name:      "default theme",
			themeName: "default",
			wantName:  "default",
			checkField: func(t Theme) bool {
				return t.Accent == lipgloss.Color("86")
			},
		},
		{
			name:      "dracula theme",
			themeName: "dracula",
			wantName:  "dracula",
			checkField: func(t Theme) bool {
				return t.Accent == lipgloss.Color("212")
			},
		},
		{
			name:      "gruvbox theme",
			themeName: "gruvbox",
			wantName:  "gruvbox",
			checkField: func(t Theme) bool {
				return t.Accent == lipgloss.Color("214")
			},
		},
		{
			name:      "nord theme",
			themeName: "nord",
			wantName:  "nord",
			checkField: func(t Theme) bool {
				return t.Accent == lipgloss.Color("110")
			},
		},
		{
			name:      "solarized-dark theme",
			themeName: "solarized-dark",
			wantName:  "solarized-dark",
			checkField: func(t Theme) bool {
				return t.Accent == lipgloss.Color("33")
			},
		},
		{
			name:      "solarized-light theme",
			themeName: "solarized-light",
			wantName:  "solarized-light",
			checkField: func(t Theme) bool {
				return t.Accent == lipgloss.Color("32")
			},
		},
		{
			name:      "non-existent theme returns default",
			themeName: "nonexistent",
			wantName:  "default",
			checkField: func(t Theme) bool {
				return t.Name == "default"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTheme(tt.themeName)
			if got.Name != tt.wantName {
				t.Errorf("GetTheme() name = %v, want %v", got.Name, tt.wantName)
			}
			if !tt.checkField(got) {
				t.Errorf("GetTheme() field check failed for theme %s", tt.name)
			}
		})
	}
}

func TestDefaultTheme(t *testing.T) {
	theme := DefaultTheme()

	if theme.Name != "default" {
		t.Errorf("DefaultTheme() name = %v, want 'default'", theme.Name)
	}

	// Check that all colors are set (lipgloss.Color is never nil)
	if theme.Background == "" {
		t.Error("DefaultTheme() Background not set")
	}
	if theme.Foreground == "" {
		t.Error("DefaultTheme() Foreground not set")
	}
	if theme.Accent == "" {
		t.Error("DefaultTheme() Accent not set")
	}
	if theme.Header == "" {
		t.Error("DefaultTheme() Header not set")
	}
	if theme.Border == "" {
		t.Error("DefaultTheme() Border not set")
	}
	if theme.Success == "" {
		t.Error("DefaultTheme() Success not set")
	}
	if theme.Error == "" {
		t.Error("DefaultTheme() Error not set")
	}
}

func TestDraculaTheme(t *testing.T) {
	theme := DraculaTheme()

	if theme.Name != "dracula" {
		t.Errorf("DraculaTheme() name = %v, want 'dracula'", theme.Name)
	}

	// Check Dracula-specific colors
	if theme.Accent != lipgloss.Color("212") {
		t.Errorf("DraculaTheme() Accent = %v, want Dracula pink", theme.Accent)
	}
	if theme.Background != lipgloss.Color("235") {
		t.Errorf("DraculaTheme() Background should be dark")
	}
}

func TestGruvboxTheme(t *testing.T) {
	theme := GruvboxTheme()

	if theme.Name != "gruvbox" {
		t.Errorf("GruvboxTheme() name = %v, want 'gruvbox'", theme.Name)
	}

	// Check Gruvbox-specific colors
	if theme.Accent != lipgloss.Color("214") {
		t.Errorf("GruvboxTheme() Accent = %v, want Gruvbox orange", theme.Accent)
	}
	if theme.Background != lipgloss.Color("235") {
		t.Errorf("GruvboxTheme() Background should be dark")
	}
}

func TestNordTheme(t *testing.T) {
	theme := NordTheme()

	if theme.Name != "nord" {
		t.Errorf("NordTheme() name = %v, want 'nord'", theme.Name)
	}

	// Check Nord-specific colors
	if theme.Accent != lipgloss.Color("110") {
		t.Errorf("NordTheme() Accent = %v, want Nord frost", theme.Accent)
	}
	if theme.Background != lipgloss.Color("236") {
		t.Errorf("NordTheme() Background should be dark blue-gray")
	}
}

func TestSolarizedDarkTheme(t *testing.T) {
	theme := SolarizedDarkTheme()

	if theme.Name != "solarized-dark" {
		t.Errorf("SolarizedDarkTheme() name = %v, want 'solarized-dark'", theme.Name)
	}

	// Check Solarized dark-specific colors
	if theme.Accent != lipgloss.Color("33") {
		t.Errorf("SolarizedDarkTheme() Accent = %v, want Solarized blue", theme.Accent)
	}
	if theme.Background != lipgloss.Color("234") {
		t.Errorf("SolarizedDarkTheme() Background should be very dark")
	}
}

func TestSolarizedLightTheme(t *testing.T) {
	theme := SolarizedLightTheme()

	if theme.Name != "solarized-light" {
		t.Errorf("SolarizedLightTheme() name = %v, want 'solarized-light'", theme.Name)
	}

	// Check Solarized light-specific colors
	if theme.Accent != lipgloss.Color("32") {
		t.Errorf("SolarizedLightTheme() Accent = %v, want Solarized blue", theme.Accent)
	}
	if theme.Background != lipgloss.Color("230") {
		t.Errorf("SolarizedLightTheme() Background should be light")
	}
	if theme.Foreground != lipgloss.Color("240") {
		t.Errorf("SolarizedLightTheme() Foreground should be dark for light theme")
	}
}

func TestGetDefaultTheme(t *testing.T) {
	want := "default"
	if got := GetDefaultTheme(); got != want {
		t.Errorf("GetDefaultTheme() = %v, want %v", got, want)
	}
}
