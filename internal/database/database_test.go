package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveDatabasePath(t *testing.T) {
	tests := []struct {
		name           string
		flagPath       string
		configPath     string
		wantPath       string
		errorExpected  bool
		errorContains  string
		setupFunc      func(t *testing.T) (cleanup func())
	}{
		{
			name:          "flag takes precedence over config",
			flagPath:      "/flag/path.db",
			configPath:    "/config/path.db",
			wantPath:      "/flag/path.db",
			errorExpected: false,
		},
		{
			name:          "config used when flag is empty",
			flagPath:      "",
			configPath:    "/config/path.db",
			wantPath:      "/config/path.db",
			errorExpected: false,
		},
		{
			name:          "default used when both flag and config are empty",
			flagPath:      "",
			configPath:    "",
			wantPath:      ".ghissues.db",
			errorExpected: false,
		},
		{
			name:          "config with relative path",
			flagPath:      "",
			configPath:    "custom.db",
			wantPath:      "custom.db",
			errorExpected: false,
		},
		{
			name:          "flag with relative path",
			flagPath:      "flag.db",
			configPath:    "config.db",
			wantPath:      "flag.db",
			errorExpected: false,
		},
		{
			name:          "default path is used when config path is whitespace only",
			flagPath:      "",
			configPath:    "   ",
			wantPath:      ".ghissues.db",
			errorExpected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, err := ResolveDatabasePath(tt.flagPath, tt.configPath)

			if tt.errorExpected {
				if err == nil {
					t.Errorf("ResolveDatabasePath() expected error containing %q, got nil", tt.errorContains)
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("ResolveDatabasePath() error = %v, want error containing %q", err, tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ResolveDatabasePath() unexpected error: %v", err)
				return
			}

			if gotPath != tt.wantPath {
				t.Errorf("ResolveDatabasePath() = %v, want %v", gotPath, tt.wantPath)
			}
		})
	}
}

func TestEnsureDatabasePath(t *testing.T) {
	tests := []struct {
		name          string
		dbPath        string
		errorExpected bool
		errorContains string
	}{
		{
			name:          "simple filename in current directory",
			dbPath:        "test.db",
			errorExpected: false,
		},
		{
			name:          "relative path with existing parent directory",
			dbPath:        "testdata/test.db",
			errorExpected: false,
		},
		{
			name:          "relative path with non-existent parent directory",
			dbPath:        "testdir/subdir/test.db",
			errorExpected: false,
		},
		{
			name:          "absolute path with non-existent parent directory",
			dbPath:        "/tmp/ghissues-test/subdir/test.db",
			errorExpected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temp directory for test isolation
			tempDir := t.TempDir()

			// Convert test paths to be relative to temp directory
			testPath := filepath.Join(tempDir, tt.dbPath)

			err := EnsureDatabasePath(testPath)

			if tt.errorExpected {
				if err == nil {
					t.Errorf("EnsureDatabasePath(%q) expected error containing %q, got nil", testPath, tt.errorContains)
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("EnsureDatabasePath() error = %v, want error containing %q", err, tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("EnsureDatabasePath(%q) unexpected error: %v", testPath, err)
				return
			}

			// Verify parent directory was created
			parentDir := filepath.Dir(testPath)
			info, err := os.Stat(parentDir)
			if err != nil {
				t.Errorf("EnsureDatabasePath() parent directory not created: %v", err)
				return
			}

			if !info.IsDir() {
				t.Errorf("EnsureDatabasePath() parent path is not a directory")
			}
		})
	}
}

func TestCheckDatabaseWritable(t *testing.T) {
	tests := []struct {
		name          string
		setupPath     func(t *testing.T) string
		errorExpected bool
		errorContains string
	}{
		{
			name: "writable directory",
			setupPath: func(t *testing.T) string {
				tempDir := t.TempDir()
				return filepath.Join(tempDir, "test.db")
			},
			errorExpected: false,
		},
		{
			name: "non-existent parent directory",
			setupPath: func(t *testing.T) string {
				return "/root/ghissues/test.db"
			},
			errorExpected: true,
			errorContains: "cannot create database file",
		},
		{
			name: "read-only directory",
			setupPath: func(t *testing.T) string {
				tempDir := t.TempDir()
				// Create a read-only subdirectory
				readOnlyDir := filepath.Join(tempDir, "readonly")
				if err := os.Mkdir(readOnlyDir, 0444); err != nil {
					t.Skipf("Cannot create read-only directory: %v", err)
					return ""
				}
				return filepath.Join(readOnlyDir, "test.db")
			},
			errorExpected: true,
			errorContains: "cannot create database file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbPath := tt.setupPath(t)
			if dbPath == "" {
				return
			}

			err := CheckDatabaseWritable(dbPath)

			if tt.errorExpected {
				if err == nil {
					t.Errorf("CheckDatabaseWritable(%q) expected error containing %q, got nil", dbPath, tt.errorContains)
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("CheckDatabaseWritable() error = %v, want error containing %q", err, tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("CheckDatabaseWritable(%q) unexpected error: %v", dbPath, err)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
