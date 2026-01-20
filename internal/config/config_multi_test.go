package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMultiRepoConfig(t *testing.T) {
	// Create a temporary config directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create a multi-repo config file
	configContent := `
default_repo = "owner/repo1"

[repositories]
"owner/repo1" = { database_path = "/path/to/db1.db" }
"owner/repo2" = { database_path = "/path/to/db2.db" }
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load the config
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadConfig should return config for existing file")
	}

	// Verify multi-repo fields
	if cfg.DefaultRepo != "owner/repo1" {
		t.Errorf("Expected default_repo 'owner/repo1', got %q", cfg.DefaultRepo)
	}
	if len(cfg.Repositories) != 2 {
		t.Fatalf("Expected 2 repositories, got %d", len(cfg.Repositories))
	}
	if cfg.Repositories["owner/repo1"].DatabasePath != "/path/to/db1.db" {
		t.Errorf("Expected db1 path '/path/to/db1.db', got %q", cfg.Repositories["owner/repo1"].DatabasePath)
	}
	if cfg.Repositories["owner/repo2"].DatabasePath != "/path/to/db2.db" {
		t.Errorf("Expected db2 path '/path/to/db2.db', got %q", cfg.Repositories["owner/repo2"].DatabasePath)
	}
}

func TestLoadSingleRepoConfig_BackwardCompatibility(t *testing.T) {
	// Create a temporary config directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create a single-repo config file (backward compatibility)
	configContent := `
repository = "testuser/testrepo"
token = "test-token"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load the config
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadConfig should return config for existing file")
	}

	// Verify backward compatibility
	if cfg.Repository != "testuser/testrepo" {
		t.Errorf("Expected repository 'testuser/testrepo', got %q", cfg.Repository)
	}
	if cfg.Token != "test-token" {
		t.Errorf("Expected token 'test-token', got %q", cfg.Token)
	}
	// New fields should be empty
	if cfg.DefaultRepo != "" {
		t.Errorf("Expected empty default_repo, got %q", cfg.DefaultRepo)
	}
	if len(cfg.Repositories) != 0 {
		t.Errorf("Expected empty repositories map, got %d entries", len(cfg.Repositories))
	}
}

func TestSaveMultiRepoConfig(t *testing.T) {
	// Create a temporary config directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create a multi-repo config
	cfg := &Config{
		DefaultRepo: "owner/repo1",
		Repositories: map[string]RepositoryConfig{
			"owner/repo1": {DatabasePath: "/path/to/db1.db"},
			"owner/repo2": {DatabasePath: "/path/to/db2.db"},
		},
	}

	// Save the config
	if err := SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Load and verify
	loadedCfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}
	if loadedCfg.DefaultRepo != cfg.DefaultRepo {
		t.Errorf("Expected default_repo %q, got %q", cfg.DefaultRepo, loadedCfg.DefaultRepo)
	}
	if len(loadedCfg.Repositories) != len(cfg.Repositories) {
		t.Fatalf("Expected %d repositories, got %d", len(cfg.Repositories), len(loadedCfg.Repositories))
	}
	for name, repoCfg := range cfg.Repositories {
		loadedRepo, exists := loadedCfg.Repositories[name]
		if !exists {
			t.Errorf("Repository %q not found in loaded config", name)
			continue
		}
		if loadedRepo.DatabasePath != repoCfg.DatabasePath {
			t.Errorf("Expected db path %q for %q, got %q", repoCfg.DatabasePath, name, loadedRepo.DatabasePath)
		}
	}
}

func TestGetDatabasePathForRepo(t *testing.T) {
	tests := []struct {
		name         string
		config       *Config
		repoName     string
		expectedPath string
		shouldError  bool
	}{
		{
			name: "specific repo found",
			config: &Config{
				Repositories: map[string]RepositoryConfig{
					"owner/repo1": {DatabasePath: "/path/to/db1.db"},
				},
			},
			repoName:     "owner/repo1",
			expectedPath: "/path/to/db1.db",
		},
		{
			name: "repo not found",
			config: &Config{
				Repositories: map[string]RepositoryConfig{
					"owner/repo1": {DatabasePath: "/path/to/db1.db"},
				},
			},
			repoName:    "owner/repo2",
			shouldError: true,
		},
		{
			name: "backward compatibility - single repo with database path",
			config: &Config{
				Repository: "owner/repo",
				Database: struct {
					Path string `toml:"path"`
				}{Path: "/path/to/single.db"},
			},
			repoName:     "owner/repo",
			expectedPath: "/path/to/single.db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var path string
			var err error

			if tt.repoName != "" {
				// Test multi-repo path lookup
				path, err = GetRepoDatabasePath(tt.config, tt.repoName)
			} else {
				// Test backward compatibility
				path, err = GetRepoDatabasePath(tt.config, tt.config.Repository)
			}

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if path != tt.expectedPath {
				t.Errorf("Expected path %q, got %q", tt.expectedPath, path)
			}
		})
	}
}

func TestGetDefaultRepo(t *testing.T) {
	tests := []struct {
		name         string
		config       *Config
		expectedRepo string
	}{
		{
			name: "default_repo set in multi-repo config",
			config: &Config{
				DefaultRepo: "owner/repo1",
				Repositories: map[string]RepositoryConfig{
					"owner/repo1": {},
					"owner/repo2": {},
				},
			},
			expectedRepo: "owner/repo1",
		},
		{
			name: "backward compatibility - single repository field",
			config: &Config{
				Repository: "testuser/testrepo",
			},
			expectedRepo: "testuser/testrepo",
		},
		{
			name: "fallback to first repo when no default",
			config: &Config{
				Repositories: map[string]RepositoryConfig{
					"owner/repo1": {},
					"owner/repo2": {},
				},
			},
			expectedRepo: "owner/repo1",
		},
		{
			name:         "no repositories defined",
			config:       &Config{},
			expectedRepo: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := GetDefaultRepo(tt.config)
			if repo != tt.expectedRepo {
				t.Errorf("Expected repo %q, got %q", tt.expectedRepo, repo)
			}
		})
	}
}

func TestListRepositories(t *testing.T) {
	// Create a temporary config directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create a multi-repo config
	cfg := &Config{
		DefaultRepo: "owner/repo1",
		Repositories: map[string]RepositoryConfig{
			"owner/repo1": {DatabasePath: "/path/to/db1.db"},
			"owner/repo2": {DatabasePath: "/path/to/db2.db"},
			"owner/repo3": {DatabasePath: "/path/to/db3.db"},
		},
	}

	// Save the config
	if err := SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Load and list repositories
	loadedCfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	repos := ListRepositories(loadedCfg)
	if len(repos) != 3 {
		t.Fatalf("Expected 3 repositories, got %d", len(repos))
	}

	// Verify all repos are listed
	expectedRepos := map[string]bool{
		"owner/repo1": false,
		"owner/repo2": false,
		"owner/repo3": false,
	}

	for _, repo := range repos {
		if _, exists := expectedRepos[repo]; !exists {
			t.Errorf("Unexpected repository: %q", repo)
		}
		expectedRepos[repo] = true
	}

	// Verify all expected repos were found
	for repo, found := range expectedRepos {
		if !found {
			t.Errorf("Repository not found in list: %q", repo)
		}
	}
}

func TestBackwardCompatibility_SaveAndLoad(t *testing.T) {
	// Create a temporary config directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create single-repo config (old format)
	cfg := &Config{
		Repository: "testuser/testrepo",
		Token:      "test-token",
	}

	// Save the config
	if err := SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Load the config
	loadedCfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	// Verify backward compatibility
	if loadedCfg.Repository != "testuser/testrepo" {
		t.Errorf("Expected repository 'testuser/testrepo', got %q", loadedCfg.Repository)
	}
	if loadedCfg.Token != "test-token" {
		t.Errorf("Expected token 'test-token', got %q", loadedCfg.Token)
	}

	// Verify new fields are empty
	if loadedCfg.DefaultRepo != "" {
		t.Errorf("Expected empty default_repo, got %q", loadedCfg.DefaultRepo)
	}
	if len(loadedCfg.Repositories) != 0 {
		t.Errorf("Expected empty repositories map, got %d entries", len(loadedCfg.Repositories))
	}
}
