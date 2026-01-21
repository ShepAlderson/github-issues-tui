package database

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shepbook/github-issues-tui/internal/config"
)

// GetDatabasePath returns the database path following precedence:
// 1. Command line flag (if provided)
// 2. Config file value (if provided)
// 3. Default (".ghissues.db")
func GetDatabasePath(flagPath string, cfg *config.Config) string {
	if flagPath != "" {
		return flagPath
	}
	if cfg.Database.Path != "" {
		return cfg.Database.Path
	}
	return ".ghissues.db"
}

// AbsolutePath converts a relative database path to an absolute path.
// If the path is empty, it returns the absolute path to the default database.
func AbsolutePath(dbPath string) string {
	if dbPath == "" {
		dbPath = ".ghissues.db"
	}

	if filepath.IsAbs(dbPath) {
		return dbPath
	}

	cwd, err := os.Getwd()
	if err != nil {
		// If we can't get the current directory, return the relative path as-is
		return dbPath
	}

	return filepath.Join(cwd, dbPath)
}

// EnsureDatabasePath ensures the database path is writable and creates
// parent directories if they don't exist.
func EnsureDatabasePath(dbPath string) error {
	if dbPath == "" {
		dbPath = ".ghissues.db"
	}

	// Get absolute path
	absPath := AbsolutePath(dbPath)

	// Ensure parent directory exists
	parentDir := filepath.Dir(absPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory for database: %w", err)
	}

	// Check if we can write to the directory
	// Try to create a test file to check writability
	testFile := filepath.Join(parentDir, ".ghissues-write-test")
	f, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("database directory is not writable: %w", err)
	}
	f.Close()
	os.Remove(testFile)

	return nil
}
