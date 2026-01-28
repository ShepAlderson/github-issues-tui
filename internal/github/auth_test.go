package github

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestResolveToken(t *testing.T) {
	// Create temp directory for testing
	tempDir := t.TempDir()

	// Save original environment
	origHome := os.Getenv("HOME")
	origToken := os.Getenv("GITHUB_TOKEN")
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("GITHUB_TOKEN", origToken)
	}()

	os.Setenv("HOME", tempDir)

	t.Run("returns error when no authentication available", func(t *testing.T) {
		// Clear environment variable
		os.Unsetenv("GITHUB_TOKEN")

		// Ensure no config exists
		configDir := filepath.Join(tempDir, ".config", "ghissues")
		os.RemoveAll(configDir)

		_, err := ResolveToken()
		if err == nil {
			t.Error("Expected error when no authentication available")
		}

		expectedMsg := "no GitHub authentication found"
		if err != nil && !contains(err.Error(), expectedMsg) {
			t.Errorf("Expected error message to contain %q, got %q", expectedMsg, err.Error())
		}
	})

	t.Run("environment variable takes priority", func(t *testing.T) {
		// Set environment variable
		os.Setenv("GITHUB_TOKEN", "env_token_123")

		// Create config with different token
		configDir := filepath.Join(tempDir, ".config", "ghissues")
		os.MkdirAll(configDir, 0755)

		// Manually write config with token
		configPath := filepath.Join(configDir, "config.toml")
		content := `[auth]
token = "config_token_456"
`
		os.WriteFile(configPath, []byte(content), 0600)

		token, err := ResolveToken()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if token != "env_token_123" {
			t.Errorf("Expected env token, got %s", token)
		}
	})

	t.Run("config file token used when env var not set", func(t *testing.T) {
		// Clear environment variable
		os.Unsetenv("GITHUB_TOKEN")

		// Create config with token
		configDir := filepath.Join(tempDir, ".config", "ghissues")
		os.MkdirAll(configDir, 0755)
		configPath := filepath.Join(configDir, "config.toml")
		content := `[auth]
token = "config_token_789"
`
		os.WriteFile(configPath, []byte(content), 0600)

		token, err := ResolveToken()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if token != "config_token_789" {
			t.Errorf("Expected config token, got %s", token)
		}
	})

	t.Run("empty environment variable falls through to config", func(t *testing.T) {
		// Set empty environment variable
		os.Setenv("GITHUB_TOKEN", "")

		// Create config with token
		configDir := filepath.Join(tempDir, ".config", "ghissues")
		os.MkdirAll(configDir, 0755)
		configPath := filepath.Join(configDir, "config.toml")
		content := `[auth]
token = "fallback_token"
`
		os.WriteFile(configPath, []byte(content), 0600)

		token, err := ResolveToken()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if token != "fallback_token" {
			t.Errorf("Expected fallback token, got %s", token)
		}
	})
}

func TestGetEnvToken(t *testing.T) {
	// Save original environment
	origToken := os.Getenv("GITHUB_TOKEN")
	defer os.Setenv("GITHUB_TOKEN", origToken)

	t.Run("returns token when set", func(t *testing.T) {
		os.Setenv("GITHUB_TOKEN", "my_token_123")

		token, found := getEnvToken()
		if !found {
			t.Error("Expected to find token")
		}
		if token != "my_token_123" {
			t.Errorf("Expected token my_token_123, got %s", token)
		}
	})

	t.Run("returns not found when not set", func(t *testing.T) {
		os.Unsetenv("GITHUB_TOKEN")

		_, found := getEnvToken()
		if found {
			t.Error("Expected token not to be found")
		}
	})

	t.Run("returns not found when empty", func(t *testing.T) {
		os.Setenv("GITHUB_TOKEN", "")

		_, found := getEnvToken()
		if found {
			t.Error("Expected empty token not to be found")
		}
	})
}

func TestGetConfigToken(t *testing.T) {
	// Create temp directory for testing
	tempDir := t.TempDir()

	// Save original HOME
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", tempDir)

	t.Run("returns token from config file", func(t *testing.T) {
		configDir := filepath.Join(tempDir, ".config", "ghissues")
		os.MkdirAll(configDir, 0755)
		configPath := filepath.Join(configDir, "config.toml")
		content := `[auth]
token = "config_file_token"
`
		os.WriteFile(configPath, []byte(content), 0600)

		token, found := getConfigToken()
		if !found {
			t.Error("Expected to find token")
		}
		if token != "config_file_token" {
			t.Errorf("Expected token config_file_token, got %s", token)
		}
	})

	t.Run("returns not found when config does not exist", func(t *testing.T) {
		configDir := filepath.Join(tempDir, ".config", "ghissues")
		os.RemoveAll(configDir)

		_, found := getConfigToken()
		if found {
			t.Error("Expected token not to be found when config missing")
		}
	})

	t.Run("returns not found when token is empty", func(t *testing.T) {
		configDir := filepath.Join(tempDir, ".config", "ghissues")
		os.MkdirAll(configDir, 0755)
		configPath := filepath.Join(configDir, "config.toml")
		content := `[auth]
token = ""
`
		os.WriteFile(configPath, []byte(content), 0600)

		_, found := getConfigToken()
		if found {
			t.Error("Expected empty token not to be found")
		}
	})
}

func TestGetGhCliToken(t *testing.T) {
	t.Run("returns token when gh CLI is available", func(t *testing.T) {
		// Skip if gh is not installed
		if _, err := exec.LookPath("gh"); err != nil {
			t.Skip("gh CLI not installed, skipping test")
		}

		// We can't actually test getting a real token without being logged in
		// Just verify the function doesn't panic
		_, _ = getGhCliToken()
	})

	t.Run("returns not found when gh CLI is not installed", func(t *testing.T) {
		// Temporarily modify PATH to exclude gh
		origPath := os.Getenv("PATH")
		defer os.Setenv("PATH", origPath)

		// Set PATH to empty - gh won't be found
		os.Setenv("PATH", "")

		_, found := getGhCliToken()
		if found {
			t.Error("Expected token not to be found when gh CLI unavailable")
		}
	})
}

func TestValidateToken(t *testing.T) {
	t.Run("returns error for empty token", func(t *testing.T) {
		err := ValidateToken("")
		if err == nil {
			t.Error("Expected error for empty token")
		}
	})

	t.Run("returns nil for valid-looking token", func(t *testing.T) {
		// We can't actually validate without making an API call
		// Just verify the function accepts the format
		err := ValidateToken("ghp_valid_token_format")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}

func TestGetTokenSource(t *testing.T) {
	// Create temp directory for testing
	tempDir := t.TempDir()

	// Save original environment
	origHome := os.Getenv("HOME")
	origToken := os.Getenv("GITHUB_TOKEN")
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("GITHUB_TOKEN", origToken)
	}()

	os.Setenv("HOME", tempDir)

	t.Run("reports environment variable source", func(t *testing.T) {
		os.Setenv("GITHUB_TOKEN", "env_token")

		_, source, err := GetTokenWithSource()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if source != "GITHUB_TOKEN environment variable" {
			t.Errorf("Expected env var source, got %s", source)
		}
	})

	t.Run("reports config file source", func(t *testing.T) {
		os.Unsetenv("GITHUB_TOKEN")

		configDir := filepath.Join(tempDir, ".config", "ghissues")
		os.MkdirAll(configDir, 0755)
		configPath := filepath.Join(configDir, "config.toml")
		content := `[auth]
token = "config_token"
`
		os.WriteFile(configPath, []byte(content), 0600)

		_, source, err := GetTokenWithSource()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if source != "config file" {
			t.Errorf("Expected config file source, got %s", source)
		}
	})
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
