package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigPath(t *testing.T) {
	// Test that ConfigPath returns expected path
	path := ConfigPath()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home dir: %v", err)
	}
	expected := filepath.Join(homeDir, ".config", "ghissues", "config.toml")
	if path != expected {
		t.Errorf("Expected path %s, got %s", expected, path)
	}
}

func TestConfigExists(t *testing.T) {
	// Test non-existent config
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	exists, err := ConfigExists(configPath)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if exists {
		t.Error("Expected config to not exist")
	}

	// Test existing config
	if err := os.WriteFile(configPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	exists, err = ConfigExists(configPath)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !exists {
		t.Error("Expected config to exist")
	}
}

func TestLoadConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	// Test valid config
	configContent := `
[github]
repository = "owner/repo"
auth_method = "token"
token = "ghp_test123"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.GitHub.Repository != "owner/repo" {
		t.Errorf("Expected repository 'owner/repo', got '%s'", cfg.GitHub.Repository)
	}
	if cfg.GitHub.AuthMethod != "token" {
		t.Errorf("Expected auth_method 'token', got '%s'", cfg.GitHub.AuthMethod)
	}
	if cfg.GitHub.Token != "ghp_test123" {
		t.Errorf("Expected token 'ghp_test123', got '%s'", cfg.GitHub.Token)
	}
}

func TestSaveConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	cfg := &Config{
		GitHub: GitHubConfig{
			Repository: "owner/repo",
			AuthMethod: "token",
			Token:      "ghp_test123",
		},
	}

	// Test saving config
	if err := SaveConfig(cfg, configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Verify content can be loaded back
	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loaded.GitHub.Repository != cfg.GitHub.Repository {
		t.Errorf("Expected repository '%s', got '%s'", cfg.GitHub.Repository, loaded.GitHub.Repository)
	}
	if loaded.GitHub.AuthMethod != cfg.GitHub.AuthMethod {
		t.Errorf("Expected auth_method '%s', got '%s'", cfg.GitHub.AuthMethod, loaded.GitHub.AuthMethod)
	}
	if loaded.GitHub.Token != cfg.GitHub.Token {
		t.Errorf("Expected token '%s', got '%s'", cfg.GitHub.Token, loaded.GitHub.Token)
	}
}

func TestSaveConfigCreatesDirectories(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nested", "dirs", "config.toml")

	cfg := &Config{
		GitHub: GitHubConfig{
			Repository: "owner/repo",
			AuthMethod: "env",
		},
	}

	// Test that parent directories are created
	if err := SaveConfig(cfg, configPath); err != nil {
		t.Fatalf("Failed to save config with nested dirs: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created in nested directory")
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config with token",
			cfg: &Config{
				GitHub: GitHubConfig{
					Repository: "owner/repo",
					AuthMethod: "token",
					Token:      "ghp_test",
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with env",
			cfg: &Config{
				GitHub: GitHubConfig{
					Repository: "owner/repo",
					AuthMethod: "env",
				},
			},
			wantErr: false,
		},
		{
			name: "missing repository",
			cfg: &Config{
				GitHub: GitHubConfig{
					AuthMethod: "token",
					Token:      "ghp_test",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid repository format",
			cfg: &Config{
				GitHub: GitHubConfig{
					Repository: "invalid",
					AuthMethod: "token",
					Token:      "ghp_test",
				},
			},
			wantErr: true,
		},
		{
			name: "token method without token",
			cfg: &Config{
				GitHub: GitHubConfig{
					Repository: "owner/repo",
					AuthMethod: "token",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
