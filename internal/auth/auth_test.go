package auth

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetToken(t *testing.T) {
	// Save original environment to restore later
	oldToken := os.Getenv("GITHUB_TOKEN")
	defer func() {
		if oldToken != "" {
			os.Setenv("GITHUB_TOKEN", oldToken)
		} else {
			os.Unsetenv("GITHUB_TOKEN")
		}
	}()

	t.Run("returns token from GITHUB_TOKEN environment variable", func(t *testing.T) {
		// Unset any existing token
		os.Unsetenv("GITHUB_TOKEN")

		expectedToken := "ghp_env_token_12345"
		os.Setenv("GITHUB_TOKEN", expectedToken)
		defer os.Unsetenv("GITHUB_TOKEN")

		token, source, err := GetToken("", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if token != expectedToken {
			t.Errorf("expected token '%s', got '%s'", expectedToken, token)
		}

		if source != "environment variable" {
			t.Errorf("expected source 'environment variable', got '%s'", source)
		}
	})

	t.Run("environment variable takes precedence over config file", func(t *testing.T) {
		os.Unsetenv("GITHUB_TOKEN")

		// Create a config file with a token
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.toml")
		configContent := `
[auth]
token = "ghp_config_token_67890"
`
		err := os.WriteFile(configPath, []byte(configContent), 0600)
		if err != nil {
			t.Fatalf("failed to create test config: %v", err)
		}

		// Set environment variable
		envToken := "ghp_env_token_12345"
		os.Setenv("GITHUB_TOKEN", envToken)
		defer os.Unsetenv("GITHUB_TOKEN")

		token, source, err := GetToken(configPath, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if token != envToken {
			t.Errorf("expected env token '%s', got '%s'", envToken, token)
		}

		if source != "environment variable" {
			t.Errorf("expected source 'environment variable', got '%s'", source)
		}
	})

	t.Run("returns token from config file when no env var set", func(t *testing.T) {
		os.Unsetenv("GITHUB_TOKEN")

		// Create a config file with a token
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.toml")
		configContent := `
[auth]
token = "ghp_config_token_67890"
`
		err := os.WriteFile(configPath, []byte(configContent), 0600)
		if err != nil {
			t.Fatalf("failed to create test config: %v", err)
		}

		token, source, err := GetToken(configPath, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedToken := "ghp_config_token_67890"
		if token != expectedToken {
			t.Errorf("expected token '%s', got '%s'", expectedToken, token)
		}

		if source != "config file" {
			t.Errorf("expected source 'config file', got '%s'", source)
		}
	})

	t.Run("returns token from gh CLI when no env var or config", func(t *testing.T) {
		os.Unsetenv("GITHUB_TOKEN")

		// Create a temporary directory without config
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "nonexistent.toml")

		// Create a temporary gh CLI config file
		ghConfigDir := t.TempDir()
		hostsConfigPath := filepath.Join(ghConfigDir, "hosts.yml")
		hostsConfigContent := `github.com:
    oauth_token: ghp_gh_cli_token_abcde
    user: testuser
`
		err := os.WriteFile(hostsConfigPath, []byte(hostsConfigContent), 0600)
		if err != nil {
			t.Fatalf("failed to create gh config: %v", err)
		}

		token, source, err := GetToken(configPath, ghConfigDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedToken := "ghp_gh_cli_token_abcde"
		if token != expectedToken {
			t.Errorf("expected token '%s', got '%s'", expectedToken, token)
		}

		if source != "gh CLI" {
			t.Errorf("expected source 'gh CLI', got '%s'", source)
		}
	})

	t.Run("returns error when no authentication method found", func(t *testing.T) {
		os.Unsetenv("GITHUB_TOKEN")

		// Create a temporary directory without config
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "nonexistent.toml")

		// Use empty gh config dir
		emptyConfigDir := t.TempDir()

		_, _, err := GetToken(configPath, emptyConfigDir)
		if err == nil {
			t.Error("expected error when no auth found, got nil")
		}

		// Check that error message contains the expected text
		expectedErrMsg := "no valid GitHub authentication found"
		if !strings.Contains(err.Error(), expectedErrMsg) {
			t.Errorf("expected error message to contain '%s', got '%s'", expectedErrMsg, err.Error())
		}
	})

	t.Run("returns token from gh CLI when config file exists but has no token", func(t *testing.T) {
		os.Unsetenv("GITHUB_TOKEN")

		// Create a config file without auth token
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.toml")
		configContent := `
[default]
repository = "owner/repo"
`
		err := os.WriteFile(configPath, []byte(configContent), 0600)
		if err != nil {
			t.Fatalf("failed to create test config: %v", err)
		}

		// Create a temporary gh CLI config file
		ghConfigDir := t.TempDir()
		hostsConfigPath := filepath.Join(ghConfigDir, "hosts.yml")
		hostsConfigContent := `github.com:
    oauth_token: ghp_gh_cli_token_xyz
    user: testuser
`
		err = os.WriteFile(hostsConfigPath, []byte(hostsConfigContent), 0600)
		if err != nil {
			t.Fatalf("failed to create gh config: %v", err)
		}

		token, source, err := GetToken(configPath, ghConfigDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedToken := "ghp_gh_cli_token_xyz"
		if token != expectedToken {
			t.Errorf("expected token '%s', got '%s'", expectedToken, token)
		}

		if source != "gh CLI" {
			t.Errorf("expected source 'gh CLI', got '%s'", source)
		}
	})

	t.Run("prioritizes config file over gh CLI when env var not set", func(t *testing.T) {
		os.Unsetenv("GITHUB_TOKEN")

		// Create a config file with a token
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.toml")
		configContent := `
[auth]
token = "ghp_config_token_priority"
`
		err := os.WriteFile(configPath, []byte(configContent), 0600)
		if err != nil {
			t.Fatalf("failed to create test config: %v", err)
		}

		// Create a temporary gh CLI config file
		ghConfigDir := t.TempDir()
		hostsConfigPath := filepath.Join(ghConfigDir, "hosts.yml")
		hostsConfigContent := `github.com:
    oauth_token: ghp_gh_cli_token_deferred
    user: testuser
`
		err = os.WriteFile(hostsConfigPath, []byte(hostsConfigContent), 0600)
		if err != nil {
			t.Fatalf("failed to create gh config: %v", err)
		}

		token, source, err := GetToken(configPath, ghConfigDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedToken := "ghp_config_token_priority"
		if token != expectedToken {
			t.Errorf("expected token '%s', got '%s'", expectedToken, token)
		}

		if source != "config file" {
			t.Errorf("expected source 'config file', got '%s'", source)
		}
	})
}

func TestGetTokenFromEnv(t *testing.T) {
	// Save original environment
	oldToken := os.Getenv("GITHUB_TOKEN")
	defer func() {
		if oldToken != "" {
			os.Setenv("GITHUB_TOKEN", oldToken)
		} else {
			os.Unsetenv("GITHUB_TOKEN")
		}
	}()

	t.Run("returns token when GITHUB_TOKEN is set", func(t *testing.T) {
		os.Unsetenv("GITHUB_TOKEN")

		expectedToken := "ghp_test_token"
		os.Setenv("GITHUB_TOKEN", expectedToken)

		token, ok := GetTokenFromEnv()
		if !ok {
			t.Error("expected ok=true, got false")
		}

		if token != expectedToken {
			t.Errorf("expected token '%s', got '%s'", expectedToken, token)
		}
	})

	t.Run("returns false when GITHUB_TOKEN is not set", func(t *testing.T) {
		os.Unsetenv("GITHUB_TOKEN")

		_, ok := GetTokenFromEnv()
		if ok {
			t.Error("expected ok=false, got true")
		}
	})

	t.Run("returns false when GITHUB_TOKEN is empty string", func(t *testing.T) {
		os.Setenv("GITHUB_TOKEN", "")

		_, ok := GetTokenFromEnv()
		if ok {
			t.Error("expected ok=false when token is empty, got true")
		}
	})
}

func TestGetTokenFromConfig(t *testing.T) {
	t.Run("returns token from config file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.toml")
		configContent := `
[auth]
token = "ghp_config_token_123"
`
		err := os.WriteFile(configPath, []byte(configContent), 0600)
		if err != nil {
			t.Fatalf("failed to create test config: %v", err)
		}

		token, ok := GetTokenFromConfig(configPath)
		if !ok {
			t.Error("expected ok=true, got false")
		}

		if token != "ghp_config_token_123" {
			t.Errorf("expected token 'ghp_config_token_123', got '%s'", token)
		}
	})

	t.Run("returns false when config file does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "nonexistent.toml")

		_, ok := GetTokenFromConfig(configPath)
		if ok {
			t.Error("expected ok=false for nonexistent config, got true")
		}
	})

	t.Run("returns false when auth section missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.toml")
		configContent := `
[default]
repository = "owner/repo"
`
		err := os.WriteFile(configPath, []byte(configContent), 0600)
		if err != nil {
			t.Fatalf("failed to create test config: %v", err)
		}

		_, ok := GetTokenFromConfig(configPath)
		if ok {
			t.Error("expected ok=false when auth section missing, got true")
		}
	})

	t.Run("returns false when token is empty string", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.toml")
		configContent := `
[auth]
token = ""
`
		err := os.WriteFile(configPath, []byte(configContent), 0600)
		if err != nil {
			t.Fatalf("failed to create test config: %v", err)
		}

		_, ok := GetTokenFromConfig(configPath)
		if ok {
			t.Error("expected ok=false when token is empty, got true")
		}
	})
}

func TestGetTokenFromGhCLI(t *testing.T) {
	t.Run("returns token from gh CLI hosts.yml", func(t *testing.T) {
		ghConfigDir := t.TempDir()
		hostsConfigPath := filepath.Join(ghConfigDir, "hosts.yml")
		hostsConfigContent := `github.com:
    oauth_token: ghp_gh_cli_token
    user: testuser
`
		err := os.WriteFile(hostsConfigPath, []byte(hostsConfigContent), 0600)
		if err != nil {
			t.Fatalf("failed to create gh config: %v", err)
		}

		token, ok := GetTokenFromGhCLI(ghConfigDir)
		if !ok {
			t.Error("expected ok=true, got false")
		}

		if token != "ghp_gh_cli_token" {
			t.Errorf("expected token 'ghp_gh_cli_token', got '%s'", token)
		}
	})

	t.Run("returns false when gh config dir does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		nonexistentDir := filepath.Join(tmpDir, "nonexistent")

		_, ok := GetTokenFromGhCLI(nonexistentDir)
		if ok {
			t.Error("expected ok=false for nonexistent dir, got true")
		}
	})

	t.Run("returns false when hosts.yml does not exist", func(t *testing.T) {
		ghConfigDir := t.TempDir()

		_, ok := GetTokenFromGhCLI(ghConfigDir)
		if ok {
			t.Error("expected ok=false when hosts.yml missing, got true")
		}
	})

	t.Run("returns false when oauth_token is missing", func(t *testing.T) {
		ghConfigDir := t.TempDir()
		hostsConfigPath := filepath.Join(ghConfigDir, "hosts.yml")
		hostsConfigContent := `github.com:
    user: testuser
`
		err := os.WriteFile(hostsConfigPath, []byte(hostsConfigContent), 0600)
		if err != nil {
			t.Fatalf("failed to create gh config: %v", err)
		}

		_, ok := GetTokenFromGhCLI(ghConfigDir)
		if ok {
			t.Error("expected ok=false when oauth_token missing, got true")
		}
	})

	t.Run("returns false when oauth_token is empty", func(t *testing.T) {
		ghConfigDir := t.TempDir()
		hostsConfigPath := filepath.Join(ghConfigDir, "hosts.yml")
		hostsConfigContent := `github.com:
    oauth_token: ""
    user: testuser
`
		err := os.WriteFile(hostsConfigPath, []byte(hostsConfigContent), 0600)
		if err != nil {
			t.Fatalf("failed to create gh config: %v", err)
		}

		_, ok := GetTokenFromGhCLI(ghConfigDir)
		if ok {
			t.Error("expected ok=false when oauth_token is empty, got true")
		}
	})
}
