package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/shepbook/ghissues/internal/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCommand(t *testing.T) {
	cmd := NewRootCmd()
	assert.Equal(t, "ghissues", cmd.Use)
	assert.Contains(t, cmd.Short, "GitHub Issues")
}

func TestRootCommandHasConfigSubcommand(t *testing.T) {
	cmd := NewRootCmd()

	// Look for config subcommand
	var configCmd *cobra.Command
	for _, c := range cmd.Commands() {
		if c.Use == "config" {
			configCmd = c
			break
		}
	}

	require.NotNil(t, configCmd, "config subcommand should exist")
	assert.Contains(t, configCmd.Short, "configuration")
}

func TestConfigCommandCreatesConfig(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")

	// Set up the config path for testing
	SetConfigPath(cfgPath)
	defer SetConfigPath("") // Reset after test

	// Create root command and run config with flags
	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"config", "--repo", "testowner/testrepo", "--auth-method", "env"})

	err := rootCmd.Execute()
	require.NoError(t, err)

	// Verify config was created
	assert.True(t, config.Exists(cfgPath))

	// Verify config contents
	cfg, err := config.Load(cfgPath)
	require.NoError(t, err)
	assert.Equal(t, "testowner/testrepo", cfg.Repository)
	assert.Equal(t, "env", cfg.Auth.Method)
}

func TestConfigCommandWithToken(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")

	SetConfigPath(cfgPath)
	defer SetConfigPath("")

	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"config", "--repo", "owner/repo", "--auth-method", "token", "--token", "ghp_test123"})

	err := rootCmd.Execute()
	require.NoError(t, err)

	// Verify config with token
	cfg, err := config.Load(cfgPath)
	require.NoError(t, err)
	assert.Equal(t, "token", cfg.Auth.Method)
	assert.Equal(t, "ghp_test123", cfg.Auth.Token)
}

func TestConfigCommandInvalidRepo(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")

	SetConfigPath(cfgPath)
	defer SetConfigPath("")

	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"config", "--repo", "invalidrepo", "--auth-method", "env"})

	err := rootCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "owner/repo")
}

func TestConfigCommandInvalidAuthMethod(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")

	SetConfigPath(cfgPath)
	defer SetConfigPath("")

	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"config", "--repo", "owner/repo", "--auth-method", "invalid"})

	err := rootCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid auth method")
}

func TestSkipSetupWhenConfigExists(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")

	// Create existing config
	cfg := &config.Config{
		Repository: "existing/repo",
		Auth: config.AuthConfig{
			Method: "gh",
		},
	}
	err := config.Save(cfg, cfgPath)
	require.NoError(t, err)

	SetConfigPath(cfgPath)
	defer SetConfigPath("")

	// When running root command (not config), should skip setup if config exists
	// The ShouldRunSetup function should return false
	assert.False(t, ShouldRunSetup(cfgPath))
}

func TestShouldRunSetupWhenConfigMissing(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "nonexistent.toml")

	// Should return true when config doesn't exist
	assert.True(t, ShouldRunSetup(cfgPath))
}

func TestGetConfigPath(t *testing.T) {
	// Reset to default
	SetConfigPath("")

	path := GetConfigPath()

	// Should return the default path
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)
	expected := filepath.Join(homeDir, ".config", "ghissues", "config.toml")
	assert.Equal(t, expected, path)
}

func TestSetConfigPathOverrides(t *testing.T) {
	customPath := "/custom/path/config.toml"
	SetConfigPath(customPath)
	defer SetConfigPath("") // Reset after test

	assert.Equal(t, customPath, GetConfigPath())
}

func TestRootCommandWithExistingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")

	// Create existing config
	cfg := &config.Config{
		Repository: "existing/repo",
		Auth: config.AuthConfig{
			Method: "env",
		},
	}
	err := config.Save(cfg, cfgPath)
	require.NoError(t, err)

	SetConfigPath(cfgPath)
	defer SetConfigPath("")

	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{})

	err = rootCmd.Execute()
	require.NoError(t, err)

	// Should have run without prompting for setup
	output := buf.String()
	assert.Contains(t, output, "existing/repo")
}

func TestRootCommandHasDBFlag(t *testing.T) {
	cmd := NewRootCmd()

	// Check that --db flag exists
	flag := cmd.PersistentFlags().Lookup("db")
	require.NotNil(t, flag, "--db flag should exist")
	assert.Equal(t, "string", flag.Value.Type())
}

func TestGetDBPath_Default(t *testing.T) {
	// Reset to default
	SetDBPath("")

	path := GetDBPath()

	// Should return empty string (meaning use default)
	assert.Equal(t, "", path)
}

func TestSetDBPath_Override(t *testing.T) {
	customPath := "/custom/path/db.db"
	SetDBPath(customPath)
	defer SetDBPath("") // Reset after test

	assert.Equal(t, customPath, GetDBPath())
}

func TestRootCommandDBFlagSetsPath(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")
	dbPath := filepath.Join(tmpDir, "custom.db")

	// Create existing config
	cfg := &config.Config{
		Repository: "existing/repo",
		Auth: config.AuthConfig{
			Method: "env",
		},
	}
	err := config.Save(cfg, cfgPath)
	require.NoError(t, err)

	SetConfigPath(cfgPath)
	defer SetConfigPath("")
	defer SetDBPath("") // Reset after test

	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--db", dbPath})

	err = rootCmd.Execute()
	require.NoError(t, err)

	// Verify dbPath was set
	assert.Equal(t, dbPath, GetDBPath())
}

func TestRootCommandDBFlagTakesPrecedenceOverConfig(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")
	flagDBPath := filepath.Join(tmpDir, "flag.db")

	// Create existing config with database.path set
	cfg := &config.Config{
		Repository: "existing/repo",
		Auth: config.AuthConfig{
			Method: "env",
		},
		Database: config.DatabaseConfig{
			Path: filepath.Join(tmpDir, "config.db"),
		},
	}
	err := config.Save(cfg, cfgPath)
	require.NoError(t, err)

	SetConfigPath(cfgPath)
	defer SetConfigPath("")
	defer SetDBPath("")

	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--db", flagDBPath})

	err = rootCmd.Execute()
	require.NoError(t, err)

	// Flag should take precedence
	assert.Equal(t, flagDBPath, GetDBPath())
}

func TestRootCommandDBFlagErrorOnUnwritablePath(t *testing.T) {
	// Skip if running as root
	if os.Getuid() == 0 {
		t.Skip("Test skipped when running as root")
	}

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")
	unwritablePath := "/root/unwritable/path/db.db"

	// Create existing config
	cfg := &config.Config{
		Repository: "existing/repo",
		Auth: config.AuthConfig{
			Method: "env",
		},
	}
	err := config.Save(cfg, cfgPath)
	require.NoError(t, err)

	SetConfigPath(cfgPath)
	defer SetConfigPath("")
	defer SetDBPath("")

	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--db", unwritablePath})

	err = rootCmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not writable")
}
