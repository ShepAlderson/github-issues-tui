package database

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shepbook/github-issues-tui/internal/config"
)

func TestGetDatabasePath(t *testing.T) {
	tests := []struct {
		name           string
		cfg            *config.Config
		flagPath       string
		expectedSuffix string // We'll check suffix since cwd might vary
		isAbsolute     bool
	}{
		{
			name:           "default path when no config or flag",
			cfg:            &config.Config{},
			flagPath:       "",
			expectedSuffix: ".ghissues.db",
			isAbsolute:     true,
		},
		{
			name: "config path takes precedence over default",
			cfg: &config.Config{
				Database: config.DatabaseConfig{
					Path: "/custom/path/issues.db",
				},
			},
			flagPath:       "",
			expectedSuffix: "/custom/path/issues.db",
			isAbsolute:     true,
		},
		{
			name: "flag takes precedence over config",
			cfg: &config.Config{
				Database: config.DatabaseConfig{
					Path: "/config/path/issues.db",
				},
			},
			flagPath:       "/flag/path/issues.db",
			expectedSuffix: "/flag/path/issues.db",
			isAbsolute:     true,
		},
		{
			name:           "flag takes precedence over default",
			cfg:            &config.Config{},
			flagPath:       "/flag/path/issues.db",
			expectedSuffix: "/flag/path/issues.db",
			isAbsolute:     true,
		},
		{
			name: "relative path in config is resolved to absolute",
			cfg: &config.Config{
				Database: config.DatabaseConfig{
					Path: "data/issues.db",
				},
			},
			flagPath:       "",
			expectedSuffix: "data/issues.db",
			isAbsolute:     true,
		},
		{
			name:           "relative path in flag is resolved to absolute",
			cfg:            &config.Config{},
			flagPath:       "data/issues.db",
			expectedSuffix: "data/issues.db",
			isAbsolute:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := GetDatabasePath(tt.cfg, tt.flagPath)
			if err != nil {
				t.Fatalf("GetDatabasePath() error = %v", err)
			}

			if tt.isAbsolute && !filepath.IsAbs(path) {
				t.Errorf("GetDatabasePath() = %v, expected absolute path", path)
			}

			if tt.expectedSuffix != "" {
				if tt.isAbsolute && !filepath.IsAbs(tt.expectedSuffix) {
					// For relative paths, check if the result contains the expected suffix
					if !filepath.HasPrefix(path, filepath.Join(mustGetCwd(t), tt.expectedSuffix)) &&
						path != filepath.Join(mustGetCwd(t), tt.expectedSuffix) {
						// Allow for the path to end with the suffix
						if len(path) < len(tt.expectedSuffix) || path[len(path)-len(tt.expectedSuffix):] != tt.expectedSuffix {
							t.Errorf("GetDatabasePath() = %v, expected to contain %v", path, tt.expectedSuffix)
						}
					}
				} else {
					// For absolute paths, check exact match
					if path != tt.expectedSuffix {
						t.Errorf("GetDatabasePath() = %v, want %v", path, tt.expectedSuffix)
					}
				}
			}
		})
	}
}

func TestInitDatabase(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) string // Returns database path to use
		expectError bool
		errorMsg    string
	}{
		{
			name: "creates parent directories if they don't exist",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				return filepath.Join(tmpDir, "nested", "dir", "test.db")
			},
			expectError: false,
		},
		{
			name: "succeeds if database file doesn't exist",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				return filepath.Join(tmpDir, "test.db")
			},
			expectError: false,
		},
		{
			name: "succeeds if database file already exists",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				dbPath := filepath.Join(tmpDir, "test.db")
				// Create the database file
				if err := os.WriteFile(dbPath, []byte("existing data"), 0644); err != nil {
					t.Fatalf("failed to create test database: %v", err)
				}
				return dbPath
			},
			expectError: false,
		},
		{
			name: "fails if parent directory is not writable",
			setup: func(t *testing.T) string {
				if os.Getuid() == 0 {
					t.Skip("skipping test when running as root")
				}
				tmpDir := t.TempDir()
				parentDir := filepath.Join(tmpDir, "readonly")
				if err := os.Mkdir(parentDir, 0555); err != nil {
					t.Fatalf("failed to create readonly dir: %v", err)
				}
				return filepath.Join(parentDir, "test.db")
			},
			expectError: true,
			errorMsg:    "permission denied",
		},
		{
			name: "fails if path is a directory",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				dirPath := filepath.Join(tmpDir, "is-a-dir")
				if err := os.Mkdir(dirPath, 0755); err != nil {
					t.Fatalf("failed to create directory: %v", err)
				}
				return dirPath
			},
			expectError: true,
			errorMsg:    "is a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbPath := tt.setup(t)
			err := InitDatabase(dbPath)

			if tt.expectError {
				if err == nil {
					t.Errorf("InitDatabase() expected error containing %q, got nil", tt.errorMsg)
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("InitDatabase() error = %v, expected to contain %q", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("InitDatabase() unexpected error = %v", err)
				}

				// Verify parent directories were created
				dir := filepath.Dir(dbPath)
				if _, err := os.Stat(dir); os.IsNotExist(err) {
					t.Errorf("InitDatabase() did not create parent directory %v", dir)
				}
			}
		})
	}
}

func TestInitDatabase_Writability(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	err := InitDatabase(dbPath)
	if err != nil {
		t.Fatalf("InitDatabase() error = %v", err)
	}

	// Verify we can write to the database file location
	testData := []byte("test write")
	if err := os.WriteFile(dbPath, testData, 0644); err != nil {
		t.Errorf("Cannot write to database path %v: %v", dbPath, err)
	}

	// Verify we can read back
	data, err := os.ReadFile(dbPath)
	if err != nil {
		t.Errorf("Cannot read from database path %v: %v", dbPath, err)
	}
	if string(data) != string(testData) {
		t.Errorf("Data mismatch: got %v, want %v", string(data), string(testData))
	}
}

// Helper functions

func mustGetCwd(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	return cwd
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
