package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigPath(t *testing.T) {
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "ghissues")
	if got := ConfigPath(); got != expected {
		t.Errorf("ConfigPath() = %q, want %q", got, expected)
	}
}

func TestConfigFilePath(t *testing.T) {
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "ghissues", "config.toml")
	if got := ConfigFilePath(); got != expected {
		t.Errorf("ConfigFilePath() = %q, want %q", got, expected)
	}
}

func TestExists(t *testing.T) {
	// Create a temporary config directory and file
	tmpDir := t.TempDir()

	// Override for test by setting HOME environment variable
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", filepath.Dir(tmpDir))
	defer os.Setenv("HOME", oldHome)

	// Calculate what ConfigPath would return with the new HOME
	testConfigPath := filepath.Join(filepath.Dir(tmpDir), ".config", "ghissues")
	testConfigFile := filepath.Join(testConfigPath, "config.toml")

	// Test when file doesn't exist
	if Exists() {
		t.Error("Exists() = true, want false (file doesn't exist)")
	}

	// Create the config directory and file
	if err := os.MkdirAll(testConfigPath, 0755); err != nil {
		t.Fatalf("Failed to create test config directory: %v", err)
	}
	if err := os.WriteFile(testConfigFile, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test when file exists
	if !Exists() {
		t.Error("Exists() = false, want true (file exists)")
	}
}

func TestLoad_NonExistentFile(t *testing.T) {
	// Create a temp directory that won't have the config file
	tmpDir := t.TempDir()

	// Override HOME to point to our temp directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", filepath.Dir(tmpDir))
	defer os.Setenv("HOME", oldHome)

	_, err := Load()
	if err == nil {
		t.Error("Load() should return error for non-existent file")
	}
}

func TestAuthMethod_Constants(t *testing.T) {
	if AuthMethodEnv != "env" {
		t.Errorf("AuthMethodEnv = %q, want %q", AuthMethodEnv, "env")
	}
	if AuthMethodGhCli != "gh" {
		t.Errorf("AuthMethodGhCli = %q, want %q", AuthMethodGhCli, "gh")
	}
	if AuthMethodToken != "token" {
		t.Errorf("AuthMethodToken = %q, want %q", AuthMethodToken, "token")
	}
}

func TestConfig_Struct(t *testing.T) {
	cfg := Config{
		Repository: "owner/repo",
		AuthMethod: AuthMethodGhCli,
		Token:      "",
	}

	if cfg.Repository != "owner/repo" {
		t.Errorf("Config.Repository = %q, want %q", cfg.Repository, "owner/repo")
	}
	if cfg.AuthMethod != AuthMethodGhCli {
		t.Errorf("Config.AuthMethod = %q, want %q", cfg.AuthMethod, AuthMethodGhCli)
	}
}

func TestSaveAndLoad(t *testing.T) {
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

	// Create and save a config
	cfg := &Config{
		Repository: "anthropics/claude-code",
		AuthMethod: AuthMethodGhCli,
		Token:      "",
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify file was created with correct permissions
	configFile := filepath.Join(configPath, "config.toml")
	info, err := os.Stat(configFile)
	if err != nil {
		t.Fatalf("Config file not created: %v", err)
	}
	if info.Mode() != 0600 {
		t.Errorf("Config file permissions = %o, want 0600", info.Mode())
	}

	// Load and verify
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if loaded.Repository != cfg.Repository {
		t.Errorf("Loaded Repository = %q, want %q", loaded.Repository, cfg.Repository)
	}
	if loaded.AuthMethod != cfg.AuthMethod {
		t.Errorf("Loaded AuthMethod = %q, want %q", loaded.AuthMethod, cfg.AuthMethod)
	}
}

func TestLoad_InvalidTOML(t *testing.T) {
	// Create a temporary config directory
	tmpDir := t.TempDir()

	// Override HOME to point to our temp directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", filepath.Dir(tmpDir))
	defer os.Setenv("HOME", oldHome)

	// Create config directory and invalid config file
	configPath := filepath.Join(filepath.Dir(tmpDir), ".config", "ghissues")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("Failed to create test config directory: %v", err)
	}

	configFile := filepath.Join(configPath, "config.toml")
	if err := os.WriteFile(configFile, []byte("invalid = toml content ["), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	_, err := Load()
	if err == nil {
		t.Error("Load() should return error for invalid TOML")
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	// Create a temporary base directory
	tmpDir := t.TempDir()

	// Override HOME to point to our temp directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", filepath.Dir(tmpDir))
	defer os.Setenv("HOME", oldHome)

	// Ensure the config directory doesn't exist
	configPath := filepath.Join(filepath.Dir(tmpDir), ".config", "ghissues")
	if _, err := os.Stat(configPath); err == nil {
		os.RemoveAll(configPath)
	}

	// Save should create the directory
	cfg := &Config{
		Repository: "test/repo",
		AuthMethod: AuthMethodEnv,
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	if _, err := os.Stat(configPath); err != nil {
		t.Errorf("Config directory not created: %v", err)
	}
}

func TestHasRepositories(t *testing.T) {
	tests := []struct {
		name         string
		repo         string
		repos        []string
		hasRepos     bool
	}{
		{
			name:     "empty config has no repos",
			repo:     "",
			repos:    nil,
			hasRepos: false,
		},
		{
			name:     "legacy repo field counts",
			repo:     "owner/repo",
			repos:    nil,
			hasRepos: true,
		},
		{
			name:     "repositories slice counts",
			repo:     "",
			repos:    []string{"owner/repo"},
			hasRepos: true,
		},
		{
			name:     "both fields set",
			repo:     "owner/repo",
			repos:    []string{"owner/repo2"},
			hasRepos: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Repository:  tt.repo,
				Repositories: tt.repos,
			}
			if got := cfg.HasRepositories(); got != tt.hasRepos {
				t.Errorf("HasRepositories() = %v, want %v", got, tt.hasRepos)
			}
		})
	}
}

func TestGetRepositories(t *testing.T) {
	tests := []struct {
		name         string
		repo         string
		repos        []string
		expectedLen  int
		expectedFirst string
	}{
		{
			name:         "empty returns nil",
			repo:         "",
			repos:        nil,
			expectedLen:  0,
			expectedFirst: "",
		},
		{
			name:         "legacy field used as fallback",
			repo:         "owner/repo",
			repos:        nil,
			expectedLen:  1,
			expectedFirst: "owner/repo",
		},
		{
			name:         "repositories slice preferred",
			repo:         "owner/repo",
			repos:        []string{"owner/repo1", "owner/repo2"},
			expectedLen:  2,
			expectedFirst: "owner/repo1",
		},
		{
			name:         "empty legacy field with repos slice",
			repo:         "",
			repos:        []string{"a/b", "c/d"},
			expectedLen:  2,
			expectedFirst: "a/b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Repository:  tt.repo,
				Repositories: tt.repos,
			}
			repos := cfg.GetRepositories()
			if len(repos) != tt.expectedLen {
				t.Errorf("GetRepositories() length = %d, want %d", len(repos), tt.expectedLen)
			}
			if tt.expectedLen > 0 && repos[0] != tt.expectedFirst {
				t.Errorf("GetRepositories()[0] = %q, want %q", repos[0], tt.expectedFirst)
			}
		})
	}
}

func TestAddRepository(t *testing.T) {
	cfg := &Config{}

	// Add first repo - should become default
	if err := cfg.AddRepository("owner/repo1"); err != nil {
		t.Fatalf("AddRepository() failed: %v", err)
	}
	if len(cfg.Repositories) != 1 {
		t.Errorf("Repositories length = %d, want 1", len(cfg.Repositories))
	}
	if cfg.DefaultRepository != "owner/repo1" {
		t.Errorf("DefaultRepository = %q, want %q", cfg.DefaultRepository, "owner/repo1")
	}

	// Add second repo
	if err := cfg.AddRepository("owner/repo2"); err != nil {
		t.Fatalf("AddRepository() failed: %v", err)
	}
	if len(cfg.Repositories) != 2 {
		t.Errorf("Repositories length = %d, want 2", len(cfg.Repositories))
	}
	if cfg.DefaultRepository != "owner/repo1" {
		t.Errorf("DefaultRepository changed unexpectedly to %q", cfg.DefaultRepository)
	}

	// Add duplicate - should be no-op
	if err := cfg.AddRepository("owner/repo1"); err != nil {
		t.Fatalf("AddRepository() failed for duplicate: %v", err)
	}
	if len(cfg.Repositories) != 2 {
		t.Errorf("Repositories length = %d, want 2 (duplicate ignored)", len(cfg.Repositories))
	}
}

func TestAddRepository_InvalidFormat(t *testing.T) {
	cfg := &Config{}

	tests := []string{"invalid", "owner", "/repo", ""}
	for _, invalid := range tests {
		t.Run(invalid, func(t *testing.T) {
			err := cfg.AddRepository(invalid)
			if err == nil {
				t.Error("AddRepository() should return error for invalid format")
			}
		})
	}
}

func TestRemoveRepository(t *testing.T) {
	cfg := &Config{
		Repositories:      []string{"owner/repo1", "owner/repo2", "owner/repo3"},
		DefaultRepository: "owner/repo2",
	}

	// Remove non-default repo
	if !cfg.RemoveRepository("owner/repo1") {
		t.Error("RemoveRepository() should return true for existing repo")
	}
	if len(cfg.Repositories) != 2 {
		t.Errorf("Repositories length = %d, want 2", len(cfg.Repositories))
	}
	if cfg.DefaultRepository != "owner/repo2" {
		t.Errorf("DefaultRepository changed unexpectedly to %q", cfg.DefaultRepository)
	}

	// Remove default repo
	if !cfg.RemoveRepository("owner/repo2") {
		t.Error("RemoveRepository() should return true for default repo")
	}
	if len(cfg.Repositories) != 1 {
		t.Errorf("Repositories length = %d, want 1", len(cfg.Repositories))
	}
	if cfg.DefaultRepository != "owner/repo3" {
		t.Errorf("DefaultRepository = %q, want %q", cfg.DefaultRepository, "owner/repo3")
	}

	// Remove non-existent repo
	if cfg.RemoveRepository("owner/nonexistent") {
		t.Error("RemoveRepository() should return false for non-existent repo")
	}
}

func TestGetDefaultRepository(t *testing.T) {
	tests := []struct {
		name     string
		repo     string
		repos    []string
		default_ string
		expected string
	}{
		{
			name:     "explicit default takes precedence",
			repo:     "legacy/repo",
			repos:    []string{"owner/repo1", "owner/repo2"},
			default_: "owner/repo2",
			expected: "owner/repo2",
		},
		{
			name:     "falls back to legacy field",
			repo:     "legacy/repo",
			repos:    nil,
			default_: "",
			expected: "legacy/repo",
		},
		{
			name:     "falls back to first in list",
			repo:     "",
			repos:    []string{"owner/repo1", "owner/repo2"},
			default_: "",
			expected: "owner/repo1",
		},
		{
			name:     "empty returns empty",
			repo:     "",
			repos:    nil,
			default_: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Repository:        tt.repo,
				Repositories:      tt.repos,
				DefaultRepository: tt.default_,
			}
			if got := cfg.GetDefaultRepository(); got != tt.expected {
				t.Errorf("GetDefaultRepository() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestSetDefaultRepository(t *testing.T) {
	cfg := &Config{
		Repositories: []string{"owner/repo1", "owner/repo2"},
	}

	// Set to existing repo
	if !cfg.SetDefaultRepository("owner/repo2") {
		t.Error("SetDefaultRepository() should return true for existing repo")
	}
	if cfg.DefaultRepository != "owner/repo2" {
		t.Errorf("DefaultRepository = %q, want %q", cfg.DefaultRepository, "owner/repo2")
	}

	// Set to non-existing repo
	if cfg.SetDefaultRepository("owner/nonexistent") {
		t.Error("SetDefaultRepository() should return false for non-existing repo")
	}
}

func TestSaveAndLoad_MultiRepo(t *testing.T) {
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

	// Create and save a config with multiple repos
	cfg := &Config{
		Repositories:      []string{"anthropics/claude-code", "golang/go"},
		DefaultRepository: "golang/go",
		AuthMethod:        AuthMethodGhCli,
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Load and verify
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if len(loaded.Repositories) != 2 {
		t.Errorf("Loaded Repositories length = %d, want 2", len(loaded.Repositories))
	}
	if loaded.DefaultRepository != "golang/go" {
		t.Errorf("Loaded DefaultRepository = %q, want %q", loaded.DefaultRepository, "golang/go")
	}
}

func TestIsValidRepoFormat(t *testing.T) {
	tests := []struct {
		repo     string
		expected bool
	}{
		{"owner/repo", true},
		{"owner-name/repo-name", true},
		{"owner123/repo456", true},
		{"owner/repo/subpath", true}, // SplitN limits to 2 parts
		{"owner", false},
		{"/repo", false},
		{"owner/", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.repo, func(t *testing.T) {
			if got := isValidRepoFormat(tt.repo); got != tt.expected {
				t.Errorf("isValidRepoFormat(%q) = %v, want %v", tt.repo, got, tt.expected)
			}
		})
	}
}
