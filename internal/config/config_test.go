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

func TestGetRepository(t *testing.T) {
	cfg := &Config{
		Default: DefaultConfig{
			Repository: "default/repo",
		},
		Repositories: []Repository{
			{Owner: "owner1", Name: "repo1", Database: "db1.db"},
			{Owner: "owner2", Name: "repo2", Database: "db2.db"},
		},
	}

	t.Run("returns repository when owner/name matches", func(t *testing.T) {
		repo, err := cfg.GetRepository("owner1/repo1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if repo.Owner != "owner1" || repo.Name != "repo1" {
			t.Errorf("expected owner1/repo1, got %s/%s", repo.Owner, repo.Name)
		}
		if repo.Database != "db1.db" {
			t.Errorf("expected database db1.db, got %s", repo.Database)
		}
	})

	t.Run("returns error when repository not found", func(t *testing.T) {
		_, err := cfg.GetRepository("unknown/repo")
		if err == nil {
			t.Error("expected error for unknown repository, got nil")
		}
	})

	t.Run("returns error for invalid format", func(t *testing.T) {
		_, err := cfg.GetRepository("invalid-format")
		if err == nil {
			t.Error("expected error for invalid format, got nil")
		}
	})
}

func TestListRepositories(t *testing.T) {
	cfg := &Config{
		Default: DefaultConfig{
			Repository: "default/repo",
		},
		Repositories: []Repository{
			{Owner: "owner1", Name: "repo1", Database: "db1.db"},
			{Owner: "owner2", Name: "repo2", Database: "db2.db"},
		},
	}

	repos := cfg.ListRepositories()

	if len(repos) != 2 {
		t.Errorf("expected 2 repositories, got %d", len(repos))
	}

	if repos[0] != "owner1/repo1" {
		t.Errorf("expected first repo to be owner1/repo1, got %s", repos[0])
	}

	if repos[1] != "owner2/repo2" {
		t.Errorf("expected second repo to be owner2/repo2, got %s", repos[1])
	}
}

func TestAddRepository(t *testing.T) {
	t.Run("adds new repository", func(t *testing.T) {
		cfg := &Config{
			Repositories: []Repository{
				{Owner: "owner1", Name: "repo1", Database: "db1.db"},
			},
		}

		err := cfg.AddRepository("owner2/repo2", "db2.db")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(cfg.Repositories) != 2 {
			t.Errorf("expected 2 repositories, got %d", len(cfg.Repositories))
		}

		if cfg.Repositories[1].Owner != "owner2" || cfg.Repositories[1].Name != "repo2" {
			t.Errorf("expected owner2/repo2, got %s/%s", cfg.Repositories[1].Owner, cfg.Repositories[1].Name)
		}
	})

	t.Run("returns error for invalid repository format", func(t *testing.T) {
		cfg := &Config{}

		err := cfg.AddRepository("invalid-format", "db.db")
		if err == nil {
			t.Error("expected error for invalid format, got nil")
		}
	})

	t.Run("returns error when repository already exists", func(t *testing.T) {
		cfg := &Config{
			Repositories: []Repository{
				{Owner: "owner1", Name: "repo1", Database: "db1.db"},
			},
		}

		err := cfg.AddRepository("owner1/repo1", "db2.db")
		if err == nil {
			t.Error("expected error for duplicate repository, got nil")
		}
	})
}

func TestRemoveRepository(t *testing.T) {
	t.Run("removes existing repository", func(t *testing.T) {
		cfg := &Config{
			Repositories: []Repository{
				{Owner: "owner1", Name: "repo1", Database: "db1.db"},
				{Owner: "owner2", Name: "repo2", Database: "db2.db"},
			},
		}

		err := cfg.RemoveRepository("owner1/repo1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(cfg.Repositories) != 1 {
			t.Errorf("expected 1 repository, got %d", len(cfg.Repositories))
		}

		if cfg.Repositories[0].Owner != "owner2" {
			t.Errorf("expected owner2, got %s", cfg.Repositories[0].Owner)
		}
	})

	t.Run("returns error when repository not found", func(t *testing.T) {
		cfg := &Config{
			Repositories: []Repository{
				{Owner: "owner1", Name: "repo1", Database: "db1.db"},
			},
		}

		err := cfg.RemoveRepository("unknown/repo")
		if err == nil {
			t.Error("expected error for unknown repository, got nil")
		}
	})
}

func TestGetRepositoryDatabase(t *testing.T) {
	cfg := &Config{
		Repositories: []Repository{
			{Owner: "owner1", Name: "repo1", Database: "custom.db"},
		},
	}

	t.Run("returns custom database when set", func(t *testing.T) {
		db, err := cfg.GetRepositoryDatabase("owner1/repo1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if db != "custom.db" {
			t.Errorf("expected custom.db, got %s", db)
		}
	})

	t.Run("returns default database when not set", func(t *testing.T) {
		db, err := cfg.GetRepositoryDatabase("owner2/repo2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := ".ghissues-owner2-repo2.db"
		if db != expected {
			t.Errorf("expected %s, got %s", expected, db)
		}
	})
}
