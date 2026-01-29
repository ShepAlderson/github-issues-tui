package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigPath(t *testing.T) {
	tests := []struct {
		name     string
		homeEnv  string
		expected string
	}{
		{
			name:     "default config path",
			homeEnv:  "/home/testuser",
			expected: filepath.Join("/home/testuser", ".config", "ghissues", "config.toml"),
		},
		{
			name:     "custom home directory",
			homeEnv:  "/custom/home",
			expected: filepath.Join("/custom/home", ".config", "ghissues", "config.toml"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original HOME
			origHome := os.Getenv("HOME")
			if origHome == "" {
				origHome = os.Getenv("USERPROFILE")
			}
			defer os.Setenv("HOME", origHome)

			// Set test HOME
			os.Setenv("HOME", tt.homeEnv)

			path := ConfigPath()
			if path != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, path)
			}
		})
	}
}

func TestExists(t *testing.T) {
	// Create temp directory for testing
	tempDir := t.TempDir()

	// Save original HOME
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", tempDir)

	// Override config dir to be in temp
	configDir := filepath.Join(tempDir, ".config", "ghissues")
	configPath := filepath.Join(configDir, "config.toml")

	t.Run("config does not exist", func(t *testing.T) {
		// Ensure config dir doesn't exist
		os.RemoveAll(configDir)

		exists := Exists()
		if exists {
			t.Error("Expected config to not exist")
		}
	})

	t.Run("config exists", func(t *testing.T) {
		// Create config directory and file
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config dir: %v", err)
		}
		if err := os.WriteFile(configPath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create config file: %v", err)
		}

		exists := Exists()
		if !exists {
			t.Error("Expected config to exist")
		}
	})
}

func TestLoad(t *testing.T) {
	// Create temp directory for testing
	tempDir := t.TempDir()

	// Save original HOME
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", tempDir)

	configDir := filepath.Join(tempDir, ".config", "ghissues")
	configPath := filepath.Join(configDir, "config.toml")

	t.Run("load empty when config does not exist", func(t *testing.T) {
		os.RemoveAll(configDir)

		cfg, err := Load()
		if err != nil {
			t.Errorf("Expected no error for missing config, got %v", err)
		}
		if cfg == nil {
			t.Error("Expected non-nil config even when file doesn't exist")
		}
	})

	t.Run("load existing config", func(t *testing.T) {
		// Create config directory
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config dir: %v", err)
		}

		configContent := `
[auth]
token = "ghp_test123"

[default]
repository = "owner/repo"

[[repositories]]
owner = "testowner"
name = "testrepo"
database = ".test.db"

[display]
theme = "dracula"
`

		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		cfg, err := Load()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if cfg.Auth.Token != "ghp_test123" {
			t.Errorf("Expected token ghp_test123, got %s", cfg.Auth.Token)
		}

		if cfg.Default.Repository != "owner/repo" {
			t.Errorf("Expected repository owner/repo, got %s", cfg.Default.Repository)
		}

		if len(cfg.Repositories) != 1 {
			t.Errorf("Expected 1 repository, got %d", len(cfg.Repositories))
		}

		if cfg.Repositories[0].Owner != "testowner" {
			t.Errorf("Expected owner testowner, got %s", cfg.Repositories[0].Owner)
		}

		if cfg.Display.Theme != "dracula" {
			t.Errorf("Expected theme dracula, got %s", cfg.Display.Theme)
		}
	})

	t.Run("load invalid toml returns error", func(t *testing.T) {
		invalidContent := "this is not valid toml [[[["
		if err := os.WriteFile(configPath, []byte(invalidContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		_, err := Load()
		if err == nil {
			t.Error("Expected error for invalid TOML, got nil")
		}
	})
}

func TestSave(t *testing.T) {
	// Create temp directory for testing
	tempDir := t.TempDir()

	// Save original HOME
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", tempDir)

	configDir := filepath.Join(tempDir, ".config", "ghissues")

	t.Run("save creates directories and file", func(t *testing.T) {
		// Ensure config dir doesn't exist
		os.RemoveAll(configDir)

		cfg := &Config{
			Auth: AuthConfig{
				Token: "ghp_saved",
			},
			Default: DefaultConfig{
				Repository: "owner/repo",
			},
		}

		err := cfg.Save()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify file exists
		configPath := filepath.Join(configDir, "config.toml")
		_, err = os.Stat(configPath)
		if err != nil {
			t.Errorf("Expected config file to exist, got error: %v", err)
		}

		// Verify we can load it back
		loaded, err := Load()
		if err != nil {
			t.Errorf("Failed to load saved config: %v", err)
		}

		if loaded.Auth.Token != "ghp_saved" {
			t.Errorf("Expected loaded token ghp_saved, got %s", loaded.Auth.Token)
		}

		if loaded.Default.Repository != "owner/repo" {
			t.Errorf("Expected loaded repository owner/repo, got %s", loaded.Default.Repository)
		}
	})

	t.Run("save with file permissions 0600", func(t *testing.T) {
		os.RemoveAll(configDir)

		cfg := &Config{
			Auth: AuthConfig{
				Token: "ghp_secret",
			},
		}

		cfg.Save()

		configPath := filepath.Join(configDir, "config.toml")
		info, err := os.Stat(configPath)
		if err != nil {
			t.Fatalf("Failed to stat config file: %v", err)
		}

		// Check file permissions (should be 0600 - owner read/write only)
		mode := info.Mode().Perm()
		if mode != 0600 {
			t.Errorf("Expected file permissions 0600, got %o", mode)
		}
	})
}

func TestValidateRepository(t *testing.T) {
	tests := []struct {
		name    string
		repo    string
		wantErr bool
	}{
		{
			name:    "valid owner/repo",
			repo:    "owner/repo",
			wantErr: false,
		},
		{
			name:    "valid with hyphens",
			repo:    "my-owner/my-repo",
			wantErr: false,
		},
		{
			name:    "valid with numbers",
			repo:    "owner123/repo456",
			wantErr: false,
		},
		{
			name:    "valid with underscore",
			repo:    "owner_name/repo_name",
			wantErr: false,
		},
		{
			name:    "invalid - no slash",
			repo:    "ownerrepo",
			wantErr: true,
		},
		{
			name:    "invalid - multiple slashes",
			repo:    "owner/repo/extra",
			wantErr: true,
		},
		{
			name:    "invalid - empty owner",
			repo:    "/repo",
			wantErr: true,
		},
		{
			name:    "invalid - empty repo",
			repo:    "owner/",
			wantErr: true,
		},
		{
			name:    "invalid - empty",
			repo:    "",
			wantErr: true,
		},
		{
			name:    "invalid - invalid characters in owner",
			repo:    "owner@name/repo",
			wantErr: true,
		},
		{
			name:    "invalid - invalid characters in repo",
			repo:    "owner/repo@name",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRepository(tt.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRepository(%q) error = %v, wantErr %v", tt.repo, err, tt.wantErr)
			}
		})
	}
}

func TestParseRepository(t *testing.T) {
	tests := []struct {
		name      string
		repo      string
		wantOwner string
		wantName  string
		wantErr   bool
	}{
		{
			name:      "valid owner/repo",
			repo:      "anthropics/claude-code",
			wantOwner: "anthropics",
			wantName:  "claude-code",
			wantErr:   false,
		},
		{
			name:      "valid with multiple dashes",
			repo:      "charmbracelet/bubbletea",
			wantOwner: "charmbracelet",
			wantName:  "bubbletea",
			wantErr:   false,
		},
		{
			name:    "invalid - no slash",
			repo:    "invalidrepo",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, name, err := ParseRepository(tt.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRepository(%q) error = %v, wantErr %v", tt.repo, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if owner != tt.wantOwner {
					t.Errorf("Expected owner %s, got %s", tt.wantOwner, owner)
				}
				if name != tt.wantName {
					t.Errorf("Expected name %s, got %s", tt.wantName, name)
				}
			}
		})
	}
}
