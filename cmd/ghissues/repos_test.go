package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shepbook/ghissues/internal/config"
)

func TestRunRepos_NoConfig(t *testing.T) {
	// Create a temporary config directory
	tmpDir := t.TempDir()

	// Override HOME to point to our temp directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", filepath.Dir(tmpDir))
	defer os.Setenv("HOME", oldHome)

	// Ensure no config exists
	configPath := filepath.Join(filepath.Dir(tmpDir), ".config", "ghissues")
	configFile := filepath.Join(configPath, "config.toml")
	if _, err := os.Stat(configFile); err == nil {
		os.Remove(configFile)
	}

	err := runRepos()
	if err == nil {
		t.Error("runRepos() should return error when config doesn't exist")
	}
}

func TestRunRepos_EmptyConfig(t *testing.T) {
	// Create a temporary config directory
	tmpDir := t.TempDir()

	// Override HOME to point to our temp directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", filepath.Dir(tmpDir))
	defer os.Setenv("HOME", oldHome)

	// Create config directory and empty config
	configPath := filepath.Join(filepath.Dir(tmpDir), ".config", "ghissues")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("Failed to create test config directory: %v", err)
	}

	cfg := &config.Config{}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	err := runRepos()
	if err != nil {
		t.Errorf("runRepos() returned error for empty config: %v", err)
	}
}

func TestRunRepos_WithRepos(t *testing.T) {
	// Create a temporary config directory
	tmpDir := t.TempDir()

	// Override HOME to point to our temp directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", filepath.Dir(tmpDir))
	defer os.Setenv("HOME", oldHome)

	// Create config directory and config with repos
	configPath := filepath.Join(filepath.Dir(tmpDir), ".config", "ghissues")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("Failed to create test config directory: %v", err)
	}

	cfg := &config.Config{
		Repositories:      []string{"owner/repo1", "owner/repo2"},
		DefaultRepository: "owner/repo2",
	}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	err := runRepos()
	if err != nil {
		t.Errorf("runRepos() returned error: %v", err)
	}
}

func TestRunRepoAdd(t *testing.T) {
	// Create a temporary config directory
	tmpDir := t.TempDir()

	// Override HOME to point to our temp directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", filepath.Dir(tmpDir))
	defer os.Setenv("HOME", oldHome)

	// Create config directory
	configPath := filepath.Join(filepath.Dir(tmpDir), ".config", "ghissues")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("Failed to create test config directory: %v", err)
	}

	// Create empty config
	cfg := &config.Config{}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Add a repository
	if err := runRepoAdd("owner/repo"); err != nil {
		t.Errorf("runRepoAdd() returned error: %v", err)
	}

	// Verify
	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if len(loaded.Repositories) != 1 {
		t.Errorf("Repositories length = %d, want 1", len(loaded.Repositories))
	}
	if loaded.DefaultRepository != "owner/repo" {
		t.Errorf("DefaultRepository = %q, want %q", loaded.DefaultRepository, "owner/repo")
	}
}

func TestRunRepoAdd_InvalidFormat(t *testing.T) {
	tests := []string{"invalid", "owner", "/repo", ""}
	for _, invalid := range tests {
		t.Run(invalid, func(t *testing.T) {
			err := runRepoAdd(invalid)
			if err == nil {
				t.Error("runRepoAdd() should return error for invalid format")
			}
		})
	}
}

func TestRunRepoRemove(t *testing.T) {
	// Create a temporary config directory
	tmpDir := t.TempDir()

	// Override HOME to point to our temp directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", filepath.Dir(tmpDir))
	defer os.Setenv("HOME", oldHome)

	// Create config directory and config with repos
	configPath := filepath.Join(filepath.Dir(tmpDir), ".config", "ghissues")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("Failed to create test config directory: %v", err)
	}

	cfg := &config.Config{
		Repositories:      []string{"owner/repo1", "owner/repo2"},
		DefaultRepository: "owner/repo1",
	}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Remove a repository
	if err := runRepoRemove("owner/repo1"); err != nil {
		t.Errorf("runRepoRemove() returned error: %v", err)
	}

	// Verify
	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if len(loaded.Repositories) != 1 {
		t.Errorf("Repositories length = %d, want 1", len(loaded.Repositories))
	}
	if loaded.DefaultRepository != "owner/repo2" {
		t.Errorf("DefaultRepository = %q, want %q", loaded.DefaultRepository, "owner/repo2")
	}
}

func TestRunRepoRemove_NotFound(t *testing.T) {
	// Create a temporary config directory
	tmpDir := t.TempDir()

	// Override HOME to point to our temp directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", filepath.Dir(tmpDir))
	defer os.Setenv("HOME", oldHome)

	// Create config directory and config with repos
	configPath := filepath.Join(filepath.Dir(tmpDir), ".config", "ghissues")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("Failed to create test config directory: %v", err)
	}

	cfg := &config.Config{
		Repositories: []string{"owner/repo1"},
	}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	err := runRepoRemove("owner/nonexistent")
	if err == nil {
		t.Error("runRepoRemove() should return error for non-existent repo")
	}
}

func TestRunRepoSetDefault(t *testing.T) {
	// Create a temporary config directory
	tmpDir := t.TempDir()

	// Override HOME to point to our temp directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", filepath.Dir(tmpDir))
	defer os.Setenv("HOME", oldHome)

	// Create config directory and config with repos
	configPath := filepath.Join(filepath.Dir(tmpDir), ".config", "ghissues")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("Failed to create test config directory: %v", err)
	}

	cfg := &config.Config{
		Repositories: []string{"owner/repo1", "owner/repo2"},
	}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Set default
	if err := runRepoSetDefault("owner/repo2"); err != nil {
		t.Errorf("runRepoSetDefault() returned error: %v", err)
	}

	// Verify
	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if loaded.DefaultRepository != "owner/repo2" {
		t.Errorf("DefaultRepository = %q, want %q", loaded.DefaultRepository, "owner/repo2")
	}
}

func TestRunRepoSetDefault_NotFound(t *testing.T) {
	// Create a temporary config directory
	tmpDir := t.TempDir()

	// Override HOME to point to our temp directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", filepath.Dir(tmpDir))
	defer os.Setenv("HOME", oldHome)

	// Create config directory and config with repos
	configPath := filepath.Join(filepath.Dir(tmpDir), ".config", "ghissues")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("Failed to create test config directory: %v", err)
	}

	cfg := &config.Config{
		Repositories: []string{"owner/repo1"},
	}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	err := runRepoSetDefault("owner/nonexistent")
	if err == nil {
		t.Error("runRepoSetDefault() should return error for non-existent repo")
	}
}