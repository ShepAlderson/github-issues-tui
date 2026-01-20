package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/shepbook/git/github-issues-tui/internal/config"
)

func TestRunMain_NoConfigFile(t *testing.T) {
	// Create a temporary directory for config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Mock user input for setup
	input := bytes.NewBufferString("testuser/testrepo\n2\nghp_testtoken123\n")
	output := &bytes.Buffer{}

	// Run main with no config file (should trigger setup)
	err := runMain([]string{"ghissues"}, configPath, input, output)
	if err != nil {
		t.Fatalf("runMain failed: %v", err)
	}

	// Verify config was created
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if cfg == nil {
		t.Fatal("Config should have been created")
	}

	if cfg.Repository != "testuser/testrepo" {
		t.Errorf("Expected repository 'testuser/testrepo', got %q", cfg.Repository)
	}
	if cfg.Token != "ghp_testtoken123" {
		t.Errorf("Expected token 'ghp_testtoken123', got %q", cfg.Token)
	}

	// Check output contains setup messages
	if !bytes.Contains(output.Bytes(), []byte("Enter GitHub repository")) {
		t.Error("Output should contain setup prompt")
	}
}

func TestRunMain_ConfigCommand(t *testing.T) {
	// Create a temporary directory for config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Mock user input for config command
	input := bytes.NewBufferString("anotheruser/anotherrepo\n2\nghp_anothertoken\n")
	output := &bytes.Buffer{}

	// Run with config command
	err := runMain([]string{"ghissues", "config"}, configPath, input, output)
	if err != nil {
		t.Fatalf("runMain failed: %v", err)
	}

	// Verify config was created
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if cfg == nil {
		t.Fatal("Config should have been created")
	}

	if cfg.Repository != "anotheruser/anotherrepo" {
		t.Errorf("Expected repository 'anotheruser/anotherrepo', got %q", cfg.Repository)
	}
}

func TestRunMain_ConfigAlreadyExists(t *testing.T) {
	// Create a temporary directory for config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create an existing config file
	cfg := &config.Config{
		Repository: "existing/repo",
		Token:      "ghp_existing",
	}
	if err := config.SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Run main with existing config (should skip setup)
	input := bytes.NewBufferString("")
	output := &bytes.Buffer{}

	err := runMain([]string{"ghissues"}, configPath, input, output)
	if err != nil {
		t.Fatalf("runMain failed: %v", err)
	}

	// Verify config wasn't changed
	loadedCfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loadedCfg.Repository != "existing/repo" {
		t.Errorf("Expected repository 'existing/repo', got %q", loadedCfg.Repository)
	}
	if loadedCfg.Token != "ghp_existing" {
		t.Errorf("Expected token 'ghp_existing', got %q", loadedCfg.Token)
	}

	// Check that setup was skipped (no prompts in output)
	if bytes.Contains(output.Bytes(), []byte("Enter GitHub repository")) {
		t.Error("Output should not contain setup prompt when config exists")
	}
}

func TestRunMain_TooManyArgs(t *testing.T) {
	// Create a temporary directory for config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Run with too many arguments
	input := bytes.NewBufferString("")
	output := &bytes.Buffer{}

	err := runMain([]string{"ghissues", "config", "extra"}, configPath, input, output)
	if err == nil {
		t.Fatal("Expected error for too many arguments")
	}

	if err.Error() != "too many arguments" {
		t.Errorf("Expected 'too many arguments' error, got: %v", err)
	}
}

func TestGetConfigFilePath_FromEnv(t *testing.T) {
	// Test getting config path from environment variable
	oldEnv := os.Getenv("GHISSUES_CONFIG")
	defer os.Setenv("GHISSUES_CONFIG", oldEnv)

	os.Setenv("GHISSUES_CONFIG", "/custom/path/config.toml")

	path := getConfigFilePath()
	if path != "/custom/path/config.toml" {
		t.Errorf("Expected '/custom/path/config.toml', got %q", path)
	}
}

func TestGetConfigFilePath_Default(t *testing.T) {
	// Test getting default config path
	oldEnv := os.Getenv("GHISSUES_CONFIG")
	defer os.Setenv("GHISSUES_CONFIG", oldEnv)

	os.Unsetenv("GHISSUES_CONFIG")

	path := getConfigFilePath()
	expectedSuffix := ".config/ghissues/config.toml"
	if len(path) < len(expectedSuffix) || path[len(path)-len(expectedSuffix):] != expectedSuffix {
		t.Errorf("Config path should end with %q, got %q", expectedSuffix, path)
	}
}

func TestGetDatabasePath_Default(t *testing.T) {
	// Test default database path
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir("/Users/shepbook/git/github-issues-tui")

	path := getDatabasePath(nil, "")

	// On macOS, paths might be resolved differently (/var vs /private/var)
	// So we check the base name matches and the path is absolute
	if filepath.Base(path) != ".ghissues.db" {
		t.Errorf("Expected basename '.ghissues.db', got %q", filepath.Base(path))
	}
	if !filepath.IsAbs(path) {
		t.Errorf("Expected absolute path, got %q", path)
	}
}

func TestGetDatabasePath_FromConfig(t *testing.T) {
	// Test database path from config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "custom.db")

	cfg := &config.Config{}
	cfg.Database.Path = configPath

	path := getDatabasePath(cfg, "")

	if path != configPath {
		t.Errorf("Expected config path %q, got %q", configPath, path)
	}
}

func TestGetDatabasePath_FromFlag(t *testing.T) {
	// Test database path from --db flag (highest priority)
	tmpDir := t.TempDir()
	flagPath := filepath.Join(tmpDir, "flag.db")

	cfg := &config.Config{}
	cfg.Database.Path = filepath.Join(tmpDir, "config.db")

	path := getDatabasePath(cfg, flagPath)

	if path != flagPath {
		t.Errorf("Expected flag path %q, got %q", flagPath, path)
	}
}

func TestEnsureDatabasePath_CreatesParentDirs(t *testing.T) {
	// Test that parent directories are created
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "nested", "dirs", "test.db")

	err := ensureDatabasePath(dbPath)
	if err != nil {
		t.Fatalf("ensureDatabasePath failed: %v", err)
	}

	// Verify directory was created
	dir := filepath.Dir(dbPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("Parent directories were not created")
	}

	// Verify file doesn't exist yet (we only create dirs, not the file)
	if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
		t.Error("Database file should not be created by ensureDatabasePath")
	}
}

func TestEnsureDatabasePath_PathNotWritable(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping test when running as root")
	}

	// Test error when path is not writable
	tmpDir := t.TempDir()
	// Create a directory and make it read-only
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	if err := os.Mkdir(readOnlyDir, 0555); err != nil {
		t.Fatalf("Failed to create read-only directory: %v", err)
	}

	dbPath := filepath.Join(readOnlyDir, "test.db")

	err := ensureDatabasePath(dbPath)
	if err == nil {
		t.Error("Expected error for non-writable path")
	}
}

func TestRunMain_DatabasePathInOutput(t *testing.T) {
	// Create a temporary directory for config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create an existing config file with database path
	configData := `repository = "test/repo"
token = "ghp_testtoken"
[database]
path = "/tmp/test.db"
`
	if err := os.WriteFile(configPath, []byte(configData), 0600); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Run main and check output includes database path
	input := bytes.NewBufferString("")
	output := &bytes.Buffer{}

	err := runMain([]string{"ghissues"}, configPath, input, output)
	if err != nil {

t.Fatalf("runMain failed: %v", err)
	}

	// Check that database path is displayed
	if !bytes.Contains(output.Bytes(), []byte("Database:")) {
		t.Error("Output should contain 'Database:'")
	}
	if !bytes.Contains(output.Bytes(), []byte("/tmp/test.db")) {
		t.Error("Output should contain the database path")
	}
}

func TestRunMain_AuthenticationFlow(t *testing.T) {
	tests := []struct {
		name         string
		envToken     string
		configToken  string
		expectSource string
		expectError  bool
	}{
		{
			name:         "Uses GITHUB_TOKEN environment variable when set",
			envToken:     "ghp_env_token_123",
			configToken:  "ghp_config_token_456",
			expectSource: "environment",
			expectError:  false,
		},
		{
			name:         "Uses config token when no env token",
			envToken:     "",
			configToken:  "ghp_config_token_789",
			expectSource: "config",
			expectError:  false,
		},
		{
			name:        "Error when no token available",
			envToken:    "",
			configToken: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp config
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.toml")

			// Set up config file with token
			configData := fmt.Sprintf("repository = \"test/repo\"\ntoken = \"%s\"\n", tt.configToken)
			if err := os.WriteFile(configPath, []byte(configData), 0600); err != nil {
				t.Fatalf("Failed to create config: %v", err)
			}

			// Set environment token
			if tt.envToken != "" {
				os.Setenv("GITHUB_TOKEN", tt.envToken)
				defer os.Unsetenv("GITHUB_TOKEN")
			}

			// Override auth package defaults for testing
			// This requires exposing internal variables, so we'll test indirectly
			input := bytes.NewBufferString("")
			output := &bytes.Buffer{}

			err := runMain([]string{"ghissues"}, configPath, input, output)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Check output contains the expected source
			expectedOutput := fmt.Sprintf("Authentication: %s token (validated)", tt.expectSource)
			if !bytes.Contains(output.Bytes(), []byte(expectedOutput)) {
				t.Errorf("Expected output to contain %q, got:\n%s", expectedOutput, output.String())
			}
		})
	}
}
