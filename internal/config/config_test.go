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

func TestGetDisplayColumns(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *Config
		expected []string
	}{
		{
			name: "default columns when not configured",
			cfg: &Config{
				GitHub: GitHubConfig{
					Repository: "owner/repo",
					AuthMethod: "token",
					Token:      "test",
				},
			},
			expected: []string{"number", "title", "author", "date", "comments"},
		},
		{
			name: "custom columns from config",
			cfg: &Config{
				GitHub: GitHubConfig{
					Repository: "owner/repo",
					AuthMethod: "token",
					Token:      "test",
				},
				Display: DisplayConfig{
					Columns: []string{"number", "title", "author"},
				},
			},
			expected: []string{"number", "title", "author"},
		},
		{
			name: "empty columns list returns defaults",
			cfg: &Config{
				GitHub: GitHubConfig{
					Repository: "owner/repo",
					AuthMethod: "token",
					Token:      "test",
				},
				Display: DisplayConfig{
					Columns: []string{},
				},
			},
			expected: []string{"number", "title", "author", "date", "comments"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			columns := GetDisplayColumns(tt.cfg)
			if len(columns) != len(tt.expected) {
				t.Errorf("Expected %d columns, got %d", len(tt.expected), len(columns))
			}
			for i, col := range columns {
				if col != tt.expected[i] {
					t.Errorf("Column %d: expected %s, got %s", i, tt.expected[i], col)
				}
			}
		})
	}
}

func TestLoadConfigWithDisplayColumns(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	// Test config with display.columns
	configContent := `
[github]
repository = "owner/repo"
auth_method = "token"
token = "ghp_test123"

[display]
columns = ["number", "title", "author"]
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(cfg.Display.Columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(cfg.Display.Columns))
	}
	if cfg.Display.Columns[0] != "number" {
		t.Errorf("Expected first column 'number', got '%s'", cfg.Display.Columns[0])
	}
	if cfg.Display.Columns[1] != "title" {
		t.Errorf("Expected second column 'title', got '%s'", cfg.Display.Columns[1])
	}
	if cfg.Display.Columns[2] != "author" {
		t.Errorf("Expected third column 'author', got '%s'", cfg.Display.Columns[2])
	}
}

func TestGetSortBy(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *Config
		expected string
	}{
		{
			name: "default sort when not configured",
			cfg: &Config{
				GitHub: GitHubConfig{
					Repository: "owner/repo",
					AuthMethod: "token",
					Token:      "test",
				},
			},
			expected: "updated",
		},
		{
			name: "custom sort from config",
			cfg: &Config{
				GitHub: GitHubConfig{
					Repository: "owner/repo",
					AuthMethod: "token",
					Token:      "test",
				},
				Display: DisplayConfig{
					SortBy: "number",
				},
			},
			expected: "number",
		},
		{
			name: "invalid sort returns default",
			cfg: &Config{
				GitHub: GitHubConfig{
					Repository: "owner/repo",
					AuthMethod: "token",
					Token:      "test",
				},
				Display: DisplayConfig{
					SortBy: "invalid",
				},
			},
			expected: "updated",
		},
		{
			name: "valid sort by created",
			cfg: &Config{
				GitHub: GitHubConfig{
					Repository: "owner/repo",
					AuthMethod: "token",
					Token:      "test",
				},
				Display: DisplayConfig{
					SortBy: "created",
				},
			},
			expected: "created",
		},
		{
			name: "valid sort by comments",
			cfg: &Config{
				GitHub: GitHubConfig{
					Repository: "owner/repo",
					AuthMethod: "token",
					Token:      "test",
				},
				Display: DisplayConfig{
					SortBy: "comments",
				},
			},
			expected: "comments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sortBy := GetSortBy(tt.cfg)
			if sortBy != tt.expected {
				t.Errorf("Expected sortBy %s, got %s", tt.expected, sortBy)
			}
		})
	}
}

func TestGetSortAscending(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *Config
		expected bool
	}{
		{
			name: "default sort ascending is false",
			cfg: &Config{
				GitHub: GitHubConfig{
					Repository: "owner/repo",
					AuthMethod: "token",
					Token:      "test",
				},
			},
			expected: false,
		},
		{
			name: "sort ascending true from config",
			cfg: &Config{
				GitHub: GitHubConfig{
					Repository: "owner/repo",
					AuthMethod: "token",
					Token:      "test",
				},
				Display: DisplayConfig{
					SortAscending: true,
				},
			},
			expected: true,
		},
		{
			name: "sort ascending false from config",
			cfg: &Config{
				GitHub: GitHubConfig{
					Repository: "owner/repo",
					AuthMethod: "token",
					Token:      "test",
				},
				Display: DisplayConfig{
					SortAscending: false,
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sortAscending := GetSortAscending(tt.cfg)
			if sortAscending != tt.expected {
				t.Errorf("Expected sortAscending %v, got %v", tt.expected, sortAscending)
			}
		})
	}
}

func TestLoadConfigWithSortPreferences(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	// Test config with sort preferences
	configContent := `
[github]
repository = "owner/repo"
auth_method = "token"
token = "ghp_test123"

[display]
sort_by = "number"
sort_ascending = true
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Display.SortBy != "number" {
		t.Errorf("Expected SortBy 'number', got '%s'", cfg.Display.SortBy)
	}
	if !cfg.Display.SortAscending {
		t.Errorf("Expected SortAscending true, got false")
	}
}
