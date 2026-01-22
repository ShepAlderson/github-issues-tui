package auth

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shepbook/ghissues/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockGhCLI is a test helper that temporarily replaces the gh CLI token function
func mockGhCLI(fn func() (string, error)) func() {
	original := ghCLITokenFunc
	ghCLITokenFunc = fn
	return func() {
		ghCLITokenFunc = original
	}
}

// mockGhCLIUnavailable mocks the gh CLI as unavailable
func mockGhCLIUnavailable() func() {
	return mockGhCLI(func() (string, error) {
		return "", errors.New("gh CLI not available")
	})
}

func TestGetToken_EnvVarFirst(t *testing.T) {
	// Mock gh CLI to be unavailable to isolate test
	restore := mockGhCLIUnavailable()
	defer restore()

	// Set up env var
	originalToken := os.Getenv("GITHUB_TOKEN")
	defer os.Setenv("GITHUB_TOKEN", originalToken)

	os.Setenv("GITHUB_TOKEN", "ghp_env_token_123")

	// Create a config with a different token
	cfg := &config.Config{
		Auth: config.AuthConfig{
			Method: "token",
			Token:  "ghp_config_token_456",
		},
	}

	// GetToken should return the env var token first
	token, source, err := GetToken(cfg)
	require.NoError(t, err)
	assert.Equal(t, "ghp_env_token_123", token)
	assert.Equal(t, SourceEnvVar, source)
}

func TestGetToken_ConfigFileSecond(t *testing.T) {
	// Mock gh CLI to be unavailable to isolate test
	restore := mockGhCLIUnavailable()
	defer restore()

	// Clear env var
	originalToken := os.Getenv("GITHUB_TOKEN")
	defer os.Setenv("GITHUB_TOKEN", originalToken)
	os.Unsetenv("GITHUB_TOKEN")

	// Create a config with a token
	cfg := &config.Config{
		Auth: config.AuthConfig{
			Method: "token",
			Token:  "ghp_config_token_456",
		},
	}

	token, source, err := GetToken(cfg)
	require.NoError(t, err)
	assert.Equal(t, "ghp_config_token_456", token)
	assert.Equal(t, SourceConfig, source)
}

func TestGetToken_ConfigMethodMustBeToken(t *testing.T) {
	// Mock gh CLI to be unavailable
	restore := mockGhCLIUnavailable()
	defer restore()

	// Clear env var
	originalToken := os.Getenv("GITHUB_TOKEN")
	defer os.Setenv("GITHUB_TOKEN", originalToken)
	os.Unsetenv("GITHUB_TOKEN")

	// Config with "env" method but no env var - should fall through
	cfg := &config.Config{
		Auth: config.AuthConfig{
			Method: "env",
			Token:  "ghp_should_not_use", // Should not use this since method is "env"
		},
	}

	// Should not find token from config because method is "env", not "token"
	// and env var is not set, and gh CLI is mocked as unavailable
	_, _, err := GetToken(cfg)
	// We expect an error because no valid auth found
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no valid authentication found")
}

func TestGetToken_GhCLIThird(t *testing.T) {
	// Mock gh CLI to return a token
	restore := mockGhCLI(func() (string, error) {
		return "ghp_cli_token_789", nil
	})
	defer restore()

	// Clear env var
	originalToken := os.Getenv("GITHUB_TOKEN")
	defer os.Setenv("GITHUB_TOKEN", originalToken)
	os.Unsetenv("GITHUB_TOKEN")

	// Config with "gh" method (no env var, no config token)
	cfg := &config.Config{
		Auth: config.AuthConfig{
			Method: "gh",
		},
	}

	// GetToken should fall through to gh CLI
	token, source, err := GetToken(cfg)
	require.NoError(t, err)
	assert.Equal(t, "ghp_cli_token_789", token)
	assert.Equal(t, SourceGhCLI, source)
}

func TestGetToken_NoAuthAvailable(t *testing.T) {
	// Mock gh CLI to be unavailable
	restore := mockGhCLIUnavailable()
	defer restore()

	// Clear env var
	originalToken := os.Getenv("GITHUB_TOKEN")
	defer os.Setenv("GITHUB_TOKEN", originalToken)
	os.Unsetenv("GITHUB_TOKEN")

	// Config with no token and gh CLI unavailable
	cfg := &config.Config{
		Auth: config.AuthConfig{
			Method: "env", // env method but no env var
		},
	}

	_, _, err := GetToken(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no valid authentication found")
}

func TestGetToken_ErrorMessageIsHelpful(t *testing.T) {
	// Mock gh CLI to be unavailable
	restore := mockGhCLIUnavailable()
	defer restore()

	// Clear env var
	originalToken := os.Getenv("GITHUB_TOKEN")
	defer os.Setenv("GITHUB_TOKEN", originalToken)
	os.Unsetenv("GITHUB_TOKEN")

	cfg := &config.Config{
		Auth: config.AuthConfig{
			Method: "env",
		},
	}

	_, _, err := GetToken(cfg)
	require.Error(t, err)

	// Error message should mention all three methods
	errMsg := err.Error()
	assert.Contains(t, errMsg, "GITHUB_TOKEN")
	assert.Contains(t, errMsg, "config")
	assert.Contains(t, errMsg, "gh")
}

func TestGetTokenFromGhCLI_NotInstalled(t *testing.T) {
	// Test with a non-existent command
	// This tests the error handling path
	token, err := getTokenFromGhCLI()
	if err != nil {
		// Expected when gh CLI is not installed or not authenticated
		assert.Empty(t, token)
	} else {
		// If gh CLI is available, token should be non-empty
		assert.NotEmpty(t, token)
	}
}

func TestConfigFilePermissions(t *testing.T) {
	// Verify that config files with tokens are saved with 0600 permissions
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	cfg := &config.Config{
		Repository: "owner/repo",
		Auth: config.AuthConfig{
			Method: "token",
			Token:  "ghp_secret_token",
		},
	}

	err := config.Save(cfg, configPath)
	require.NoError(t, err)

	// Verify permissions are 0600
	info, err := os.Stat(configPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestTokenSource_String(t *testing.T) {
	assert.Equal(t, "environment variable (GITHUB_TOKEN)", SourceEnvVar.String())
	assert.Equal(t, "config file", SourceConfig.String())
	assert.Equal(t, "gh CLI", SourceGhCLI.String())
}

func TestValidateToken_EmptyToken(t *testing.T) {
	err := ValidateToken("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestValidateToken_InvalidToken(t *testing.T) {
	// Use a clearly invalid token
	err := ValidateToken("invalid_token_123")
	require.Error(t, err)
	// Should contain helpful error message
	assert.Contains(t, err.Error(), "invalid")
}

func TestValidateToken_ErrorMessageIsHelpful(t *testing.T) {
	err := ValidateToken("bad_token")
	require.Error(t, err)
	errMsg := err.Error()
	// Error should mention it's an authentication issue
	assert.True(t, strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "401") || strings.Contains(errMsg, "authentication"),
		"Error message should indicate an authentication issue: %s", errMsg)
}

func TestErrNoAuth_IsIdentifiable(t *testing.T) {
	// Mock gh CLI to be unavailable
	restore := mockGhCLIUnavailable()
	defer restore()

	// Clear env var
	originalToken := os.Getenv("GITHUB_TOKEN")
	defer os.Setenv("GITHUB_TOKEN", originalToken)
	os.Unsetenv("GITHUB_TOKEN")

	cfg := &config.Config{
		Auth: config.AuthConfig{
			Method: "env",
		},
	}

	_, _, err := GetToken(cfg)
	require.Error(t, err)

	// Error should wrap ErrNoAuth so callers can check for it
	assert.True(t, errors.Is(err, ErrNoAuth))
}
