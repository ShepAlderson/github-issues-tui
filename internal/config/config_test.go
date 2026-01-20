package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigExists(t *testing.T) {
	t.Run("returns false when config file does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.toml")

		exists := ConfigExists(configPath)
		if exists != false {
			t.Errorf("expected config to not exist, got %v", exists)
		}
	})

	t.Run("returns true when config file exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.toml")

		// Create the config file
		err := os.WriteFile(configPath, []byte("# test config"), 0644)
		if err != nil {
			t.Fatalf("failed to create test config: %v", err)
		}

		exists := ConfigExists(configPath)
		if exists != true {
			t.Errorf("expected config to exist, got %v", exists)
		}
	})
}

func TestGetDefaultConfigPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot get home directory")
	}

	expected := filepath.Join(home, ".config", "ghissues", "config.toml")
	got := GetDefaultConfigPath()

	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestLoadConfig(t *testing.T) {
	t.Run("returns error when config file does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "nonexistent.toml")

		_, err := LoadConfig(configPath)
		if err == nil {
			t.Error("expected error when loading nonexistent config, got nil")
		}
	})

	t.Run("loads valid config file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.toml")
		configContent := `
[default]
repository = "owner/repo"

[[repositories]]
owner = "testowner"
name = "testrepo"
database = ".test.db"

[display]
theme = "dracula"
`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		if err != nil {
			t.Fatalf("failed to create test config: %v", err)
		}

		cfg, err := LoadConfig(configPath)
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		if cfg.Default.Repository != "owner/repo" {
			t.Errorf("expected repository 'owner/repo', got '%s'", cfg.Default.Repository)
		}

		if len(cfg.Repositories) != 1 {
			t.Errorf("expected 1 repository, got %d", len(cfg.Repositories))
		}

		if cfg.Display.Theme != "dracula" {
			t.Errorf("expected theme 'dracula', got '%s'", cfg.Display.Theme)
		}
	})
}

func TestSaveConfig(t *testing.T) {
	t.Run("saves config to file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.toml")

		cfg := &Config{
			Default: DefaultConfig{
				Repository: "testowner/testrepo",
			},
			Repositories: []Repository{
				{
					Owner:    "testowner",
					Name:     "testrepo",
					Database: ".test.db",
				},
			},
			Display: DisplayConfig{
				Theme: "default",
			},
		}

		err := SaveConfig(configPath, cfg)
		if err != nil {
			t.Fatalf("failed to save config: %v", err)
		}

		// Verify file was created
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("config file was not created")
		}

		// Verify file permissions
		info, _ := os.Stat(configPath)
		perms := info.Mode().Perm()
		if perms != 0600 {
			t.Errorf("expected file permissions 0600, got %04o", perms)
		}

		// Verify we can load it back
		loaded, err := LoadConfig(configPath)
		if err != nil {
			t.Fatalf("failed to load saved config: %v", err)
		}

		if loaded.Default.Repository != "testowner/testrepo" {
			t.Errorf("expected repository 'testowner/testrepo', got '%s'", loaded.Default.Repository)
		}
	})
}

func TestEnsureConfigDir(t *testing.T) {
	t.Run("creates config directory if it does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "ghissues", "config.toml")

		err := EnsureConfigDir(configPath)
		if err != nil {
			t.Fatalf("failed to ensure config dir: %v", err)
		}

		// Check that directory was created
		configDir := filepath.Dir(configPath)
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			t.Error("config directory was not created")
		}
	})

	t.Run("does not error if directory already exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "ghissues", "config.toml")

		// Create directory first
		configDir := filepath.Dir(configPath)
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			t.Fatalf("failed to create test directory: %v", err)
		}

		// Should not error
		err = EnsureConfigDir(configPath)
		if err != nil {
			t.Errorf("unexpected error ensuring existing config dir: %v", err)
		}
	})
}

func TestValidateRepository(t *testing.T) {
	tests := []struct {
		name    string
		repo    string
		wantErr bool
	}{
		{"valid repository", "owner/repo", false},
		{"valid repository with hyphen", "owner-name/repo-name", false},
		{"valid repository with dot", "owner.name/repo.name", false},
		{"missing owner", "/repo", true},
		{"missing repo", "owner/", true},
		{"missing both", "/", true},
		{"no separator", "ownerrepo", true},
		{"too many separators", "owner/repo/extra", true},
		{"empty string", "", true},
		{"extra slashes", "owner//repo", true},
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
