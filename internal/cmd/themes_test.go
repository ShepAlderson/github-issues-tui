package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shepbook/git/github-issues-tui/internal/config"
)

func TestRunThemesCommand_NonInteractive(t *testing.T) {
	// Create temp directory for test config
	tmpDir, err := os.MkdirTemp("", "ghissues-themes-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.toml")

	// Create a config file
	cfg := &config.Config{
		Repository: "test/repo",
		Token:      "test-token",
	}
	cfg.Display.Theme = "default"
	if err := config.SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	tests := []struct {
		name         string
		themeName    string
		wantTheme    string
		wantInOutput string
		wantErr      bool
	}{
		{
			name:         "set dracula theme",
			themeName:    "dracula",
			wantTheme:    "dracula",
			wantInOutput: "Theme set to: dracula\n",
			wantErr:      false,
		},
		{
			name:         "set gruvbox theme",
			themeName:    "gruvbox",
			wantTheme:    "gruvbox",
			wantInOutput: "Theme set to: gruvbox\n",
			wantErr:      false,
		},
		{
			name:         "set nord theme",
			themeName:    "nord",
			wantTheme:    "nord",
			wantInOutput: "Theme set to: nord\n",
			wantErr:      false,
		},
		{
			name:         "invalid theme error",
			themeName:    "nonexistent",
			wantTheme:    "",
			wantInOutput: "invalid theme name",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &bytes.Buffer{}
			output := &bytes.Buffer{}

			err := RunThemesCommand(configPath, input, output, true, tt.themeName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				if !strings.Contains(err.Error(), tt.wantInOutput) {
					t.Errorf("expected error %q to contain %q", err.Error(), tt.wantInOutput)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !strings.Contains(output.String(), tt.wantInOutput) {
				t.Errorf("output %q does not contain %q", output.String(), tt.wantInOutput)
			}

			// Verify the config was updated
			updatedCfg, err := config.LoadConfig(configPath)
			if err != nil {
				t.Fatalf("Failed to load updated config: %v", err)
			}

			if updatedCfg.Display.Theme != tt.wantTheme {
				t.Errorf("Theme in config = %v, want %v", updatedCfg.Display.Theme, tt.wantTheme)
			}
		})
	}
}

func TestRunThemesCommand_Interactive(t *testing.T) {
	// Create temp directory for test config
	tmpDir, err := os.MkdirTemp("", "ghissues-themes-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.toml")

	tests := []struct {
		name         string
		input        string
		wantInOutput []string
		wantTheme    string
	}{
		{
			name:  "select dracula",
			input: "2\n", // Select dracula (index 2)
			wantInOutput: []string{
				"Theme Selection",
				"1. default",
				"2. dracula",
				"Theme saved: dracula",
			},
			wantTheme: "dracula",
		},
		{
			name:  "select gruvbox by name",
			input: "gruvbox\n",
			wantInOutput: []string{
				"Theme Selection",
				"Theme saved: gruvbox",
			},
			wantTheme: "gruvbox",
		},
		{
			name:  "press enter to keep current",
			input: "\n",
			wantInOutput: []string{
				"Theme Selection",
				"No theme selected. Keeping current theme.",
			},
			wantTheme: "default", // Should remain default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh config for each test
			cfg := &config.Config{
				Repository: "test/repo",
				Token:      "test-token",
			}
			cfg.Display.Theme = "default"
			if err := config.SaveConfig(configPath, cfg); err != nil {
				t.Fatalf("Failed to save config: %v", err)
			}

			input := bytes.NewBufferString(tt.input)
			output := &bytes.Buffer{}

			err := RunThemesCommand(configPath, input, output, false, "")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			outputStr := output.String()
			for _, want := range tt.wantInOutput {
				if !strings.Contains(outputStr, want) {
					t.Errorf("output does not contain %q\nOutput: %s", want, outputStr)
				}
			}

			// Verify the config was updated correctly
			updatedCfg, err := config.LoadConfig(configPath)
			if err != nil {
				t.Fatalf("Failed to load updated config: %v", err)
			}

			if updatedCfg.Display.Theme != tt.wantTheme {
				t.Errorf("Theme in config = %v, want %v", updatedCfg.Display.Theme, tt.wantTheme)
			}
		})
	}
}

func TestListThemesCommand(t *testing.T) {
	output := &bytes.Buffer{}

	err := ListThemesCommand(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outputStr := output.String()

	// Check that all themes are listed
	expectedThemes := []string{"default", "dracula", "gruvbox", "nord", "solarized-dark", "solarized-light"}
	for _, theme := range expectedThemes {
		if !strings.Contains(outputStr, theme) {
			t.Errorf("output does not contain theme %q", theme)
		}
	}

	// Check for descriptive text
	expectedText := []string{
		"Available themes:",
		"Use 'ghissues themes' to preview",
	}
	for _, text := range expectedText {
		if !strings.Contains(outputStr, text) {
			t.Errorf("output does not contain %q", text)
		}
	}
}

func TestGetCurrentTheme(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want string
	}{
		{
			name: "config with theme",
			cfg:  &config.Config{},
			want: "default",
		},
		{
			name: "config with custom theme",
			cfg: &config.Config{Display: struct {
				Columns []string    "toml:\"columns\""
				Sort    config.Sort "toml:\"sort\""
				Theme   string      "toml:\"theme\""
			}{Theme: "dracula"}},
			want: "dracula",
		},
		{
			name: "nil config",
			cfg:  nil,
			want: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getCurrentTheme(tt.cfg)
			if got != tt.want {
				t.Errorf("getCurrentTheme() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAvailableThemes(t *testing.T) {
	themes := availableThemes()

	if len(themes) != 6 {
		t.Errorf("availableThemes() returned %d themes, want 6", len(themes))
	}

	expected := []string{"default", "dracula", "gruvbox", "nord", "solarized-dark", "solarized-light"}
	for i, theme := range expected {
		if i >= len(themes) || themes[i] != theme {
			t.Errorf("availableThemes()[%d] = %v, want %v", i, themes[i], theme)
		}
	}
}
