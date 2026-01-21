package setup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shepbook/ghissues/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunSetup(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Test with valid inputs using the programmatic interface
	cfg, err := RunSetupWithValues("owner/repo", "env", "", configPath)
	require.NoError(t, err)

	assert.Equal(t, "owner/repo", cfg.Repository)
	assert.Equal(t, "env", cfg.Auth.Method)

	// Verify file was created
	assert.True(t, config.Exists(configPath))
}

func TestRunSetupWithTokenMethod(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Test with token method - should include token in config
	cfg, err := RunSetupWithValues("owner/repo", "token", "ghp_mytoken123", configPath)
	require.NoError(t, err)

	assert.Equal(t, "owner/repo", cfg.Repository)
	assert.Equal(t, "token", cfg.Auth.Method)
	assert.Equal(t, "ghp_mytoken123", cfg.Auth.Token)

	// Verify the token is saved to the config file
	loadedCfg, err := config.Load(configPath)
	require.NoError(t, err)
	assert.Equal(t, "ghp_mytoken123", loadedCfg.Auth.Token)
}

func TestRunSetupValidatesRepository(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Test with invalid repository format
	_, err := RunSetupWithValues("invalidrepo", "env", "", configPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "owner/repo")
}

func TestRunSetupValidatesAuthMethod(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Test with invalid auth method
	_, err := RunSetupWithValues("owner/repo", "invalid", "", configPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid auth method")
}

func TestSetupCreatesDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nested", "deep", "config.toml")

	cfg, err := RunSetupWithValues("owner/repo", "env", "", configPath)
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Verify file was created in nested directory
	assert.True(t, config.Exists(configPath))
}

func TestSetupOverwritesExistingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create an existing config
	oldCfg := &config.Config{
		Repository: "old/repo",
		Auth: config.AuthConfig{
			Method: "gh",
		},
	}
	err := config.Save(oldCfg, configPath)
	require.NoError(t, err)

	// Run setup with new values
	newCfg, err := RunSetupWithValues("new/repo", "env", "", configPath)
	require.NoError(t, err)

	// Verify the config was updated
	assert.Equal(t, "new/repo", newCfg.Repository)
	assert.Equal(t, "env", newCfg.Auth.Method)

	// Verify the file was updated
	loadedCfg, err := config.Load(configPath)
	require.NoError(t, err)
	assert.Equal(t, "new/repo", loadedCfg.Repository)
}

func TestSetupSecurePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	_, err := RunSetupWithValues("owner/repo", "token", "secret", configPath)
	require.NoError(t, err)

	// Verify file permissions are secure (0600)
	info, err := os.Stat(configPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}
