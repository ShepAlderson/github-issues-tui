package database

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shepbook/github-issues-tui/internal/config"
)

// hasPathSuffix checks if a path ends with the expected suffix, handling
// platform differences like /private prefix on macOS
func hasPathSuffix(path, suffix string) bool {
	// Clean both paths for consistent comparison
	cleanPath := filepath.Clean(path)
	cleanSuffix := filepath.Clean(suffix)

	// Check if cleanPath ends with cleanSuffix
	return strings.HasSuffix(cleanPath, cleanSuffix)
}

func TestGetDatabasePath(t *testing.T) {
	// tempDir := t.TempDir() - not used in this test

	tests := []struct {
		name          string
		flagPath      string
		configPath    string
		expectedPath  string
		expectedError bool
	}{
		{
			name:          "no flag, no config - uses default",
			flagPath:      "",
			configPath:    "",
			expectedPath:  ".ghissues.db",
			expectedError: false,
		},
		{
			name:          "flag only",
			flagPath:      "/custom/path/db.db",
			configPath:    "",
			expectedPath:  "/custom/path/db.db",
			expectedError: false,
		},
		{
			name:          "config only",
			flagPath:      "",
			configPath:    "/config/path/db.db",
			expectedPath:  "/config/path/db.db",
			expectedError: false,
		},
		{
			name:          "flag overrides config",
			flagPath:      "/flag/path/db.db",
			configPath:    "/config/path/db.db",
			expectedPath:  "/flag/path/db.db",
			expectedError: false,
		},
		{
			name:          "relative path in config",
			flagPath:      "",
			configPath:    "relative.db",
			expectedPath:  "relative.db",
			expectedError: false,
		},
		{
			name:          "relative path in flag",
			flagPath:      "flag-relative.db",
			configPath:    "",
			expectedPath:  "flag-relative.db",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config with database path if specified
			cfg := &config.Config{
				Database: config.Database{
					Path: tt.configPath,
				},
			}

			path := GetDatabasePath(tt.flagPath, cfg)

			// For relative paths, we need to check if they resolve correctly
			if filepath.IsAbs(tt.expectedPath) {
				if path != tt.expectedPath {
					t.Errorf("GetDatabasePath() = %v, want %v", path, tt.expectedPath)
				}
			} else {
				// For relative paths, just check they match the input
				if path != tt.expectedPath {
					t.Errorf("GetDatabasePath() = %v, want %v", path, tt.expectedPath)
				}
			}
		})
	}
}

func TestEnsureDatabasePath(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name          string
		dbPath        string
		setup         func(string) error
		expectedError bool
	}{
		{
			name:   "existing writable directory",
			dbPath: filepath.Join(tempDir, "test.db"),
			setup: func(path string) error {
				dir := filepath.Dir(path)
				return os.MkdirAll(dir, 0755)
			},
			expectedError: false,
		},
		{
			name:   "non-existent parent directory - creates it",
			dbPath: filepath.Join(tempDir, "nested", "deep", "test.db"),
			setup: func(path string) error {
				// Don't create parent directory - let EnsureDatabasePath create it
				return nil
			},
			expectedError: false,
		},
		{
			name:   "writable current directory",
			dbPath: "test.db",
			setup: func(path string) error {
				// No setup needed for current directory
				return nil
			},
			expectedError: false,
		},
		{
			name:   "non-writable parent directory",
			dbPath: filepath.Join(tempDir, "test.db"),
			setup: func(path string) error {
				dir := filepath.Dir(path)
				if err := os.MkdirAll(dir, 0755); err != nil {
					return err
				}
				// Make directory read-only
				return os.Chmod(dir, 0444)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if tt.setup != nil {
				if err := tt.setup(tt.dbPath); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			// Test
			err := EnsureDatabasePath(tt.dbPath)

			if tt.expectedError && err == nil {
				t.Errorf("EnsureDatabasePath() expected error, got nil")
			}

			if !tt.expectedError && err != nil {
				t.Errorf("EnsureDatabasePath() unexpected error: %v", err)
			}

			// Cleanup: restore permissions if we changed them
			if tt.name == "non-writable parent directory" {
				dir := filepath.Dir(tt.dbPath)
				os.Chmod(dir, 0755)
			}
		})
	}
}

func TestAbsolutePath(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already absolute",
			input:    "/absolute/path/db.db",
			expected: "/absolute/path/db.db",
		},
		{
			name:     "relative path",
			input:    "relative.db",
			expected: filepath.Join(tempDir, "relative.db"),
		},
		{
			name:     "empty path uses default",
			input:    "",
			expected: filepath.Join(tempDir, ".ghissues.db"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Change to temp directory for relative path tests
			oldDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get current directory: %v", err)
			}
			defer os.Chdir(oldDir)

			if err := os.Chdir(tempDir); err != nil {
				t.Fatalf("Failed to change directory: %v", err)
			}

			result := AbsolutePath(tt.input)
			// On macOS, os.Getwd() may return path with /private prefix
			// We'll check that the path ends with what we expect
			if !hasPathSuffix(result, tt.expected) {
				t.Errorf("AbsolutePath(%q) = %q, want path ending with %q", tt.input, result, tt.expected)
			}
		})
	}
}
