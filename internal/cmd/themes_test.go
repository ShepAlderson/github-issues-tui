package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shepbook/ghissues/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCommandHasThemesSubcommand(t *testing.T) {
	cmd := NewRootCmd()

	// Look for themes subcommand
	var themesCmd bool
	for _, c := range cmd.Commands() {
		if c.Use == "themes" {
			themesCmd = true
			break
		}
	}

	assert.True(t, themesCmd, "themes subcommand should exist")
}

func TestThemesCommandListsAllThemes(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")

	// Create config file
	cfg := &config.Config{
		Repository: "owner/repo",
		Auth:       config.AuthConfig{Method: "env"},
	}
	err := config.Save(cfg, cfgPath)
	require.NoError(t, err)

	SetConfigPath(cfgPath)
	defer SetConfigPath("")

	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"themes"})

	err = rootCmd.Execute()
	require.NoError(t, err)

	output := strings.ToLower(buf.String())

	// Should list all themes (case-insensitive)
	assert.Contains(t, output, "default")
	assert.Contains(t, output, "dracula")
	assert.Contains(t, output, "gruvbox")
	assert.Contains(t, output, "nord")
	assert.Contains(t, output, "solarized")
}

func TestThemesCommandShowsCurrentTheme(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")

	// Create config file with dracula theme
	cfg := &config.Config{
		Repository: "owner/repo",
		Auth:       config.AuthConfig{Method: "env"},
		Display:    config.DisplayConfig{Theme: config.ThemeDracula},
	}
	err := config.Save(cfg, cfgPath)
	require.NoError(t, err)

	SetConfigPath(cfgPath)
	defer SetConfigPath("")

	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"themes"})

	err = rootCmd.Execute()
	require.NoError(t, err)

	output := strings.ToLower(buf.String())

	// Should indicate dracula is current (case-insensitive)
	assert.Contains(t, output, "dracula")
	assert.Contains(t, output, "current") // Should show current indicator
}

func TestThemesCommandSetTheme(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")

	// Create config file with default theme
	cfg := &config.Config{
		Repository: "owner/repo",
		Auth:       config.AuthConfig{Method: "env"},
	}
	err := config.Save(cfg, cfgPath)
	require.NoError(t, err)

	SetConfigPath(cfgPath)
	defer SetConfigPath("")

	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"themes", "--set", "nord"})

	err = rootCmd.Execute()
	require.NoError(t, err)

	// Verify theme was saved to config
	loadedCfg, err := config.Load(cfgPath)
	require.NoError(t, err)
	assert.Equal(t, config.ThemeNord, loadedCfg.Display.Theme)
}

func TestThemesCommandSetInvalidTheme(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")

	// Create config file
	cfg := &config.Config{
		Repository: "owner/repo",
		Auth:       config.AuthConfig{Method: "env"},
	}
	err := config.Save(cfg, cfgPath)
	require.NoError(t, err)

	SetConfigPath(cfgPath)
	defer SetConfigPath("")

	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"themes", "--set", "invalid-theme"})

	err = rootCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid theme")
}

func TestThemesCommandPreviewTheme(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")

	// Create config file
	cfg := &config.Config{
		Repository: "owner/repo",
		Auth:       config.AuthConfig{Method: "env"},
	}
	err := config.Save(cfg, cfgPath)
	require.NoError(t, err)

	SetConfigPath(cfgPath)
	defer SetConfigPath("")

	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"themes", "--preview", "gruvbox"})

	err = rootCmd.Execute()
	require.NoError(t, err)

	output := buf.String()

	// Should show preview of gruvbox theme
	assert.Contains(t, output, "Gruvbox")
	assert.Contains(t, output, "Preview")
}

func TestThemesCommandPreviewInvalidTheme(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")

	// Create config file
	cfg := &config.Config{
		Repository: "owner/repo",
		Auth:       config.AuthConfig{Method: "env"},
	}
	err := config.Save(cfg, cfgPath)
	require.NoError(t, err)

	SetConfigPath(cfgPath)
	defer SetConfigPath("")

	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"themes", "--preview", "invalid-theme"})

	err = rootCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid theme")
}

func TestThemesCommandWithoutConfig(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "nonexistent", "config.toml")

	SetConfigPath(cfgPath)
	defer SetConfigPath("")

	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"themes"})

	err := rootCmd.Execute()
	// Should still work - shows themes but indicates no config
	require.NoError(t, err)
	output := strings.ToLower(buf.String())
	assert.Contains(t, output, "default")
}
