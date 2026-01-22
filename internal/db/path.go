package db

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shepbook/ghissues/internal/config"
)

const defaultDBFilename = ".ghissues.db"

// DefaultDBPath returns the default database path (.ghissues.db in current directory)
func DefaultDBPath() string {
	return defaultDBFilename
}

// ResolveDBPath determines the database path based on precedence:
// 1. Flag value (if provided)
// 2. Config file value (if set)
// 3. Default (.ghissues.db in current directory)
func ResolveDBPath(flagPath string, cfg *config.Config) (string, error) {
	// Flag takes highest precedence
	if flagPath != "" {
		return flagPath, nil
	}

	// Config file takes second precedence
	if cfg != nil && cfg.Database.Path != "" {
		return cfg.Database.Path, nil
	}

	// Default to .ghissues.db in current directory
	return DefaultDBPath(), nil
}

// EnsureDBPath ensures the database path is usable:
// - Creates parent directories if they don't exist
// - Verifies the path is writable
func EnsureDBPath(dbPath string) error {
	// Get the parent directory
	parentDir := filepath.Dir(dbPath)

	// If the path is just a filename, parent is "." (current directory)
	if parentDir == "." {
		// Check if current directory is writable
		if !IsPathWritable(".") {
			return fmt.Errorf("current directory is not writable for database file: %s", dbPath)
		}
		return nil
	}

	// Create parent directories if they don't exist
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("database path not writable: cannot create directory %s: %w", parentDir, err)
	}

	// Verify the directory is writable
	if !IsPathWritable(parentDir) {
		return fmt.Errorf("database path not writable: %s", parentDir)
	}

	return nil
}

// IsPathWritable checks if a path is writable by attempting to create a temp file
func IsPathWritable(path string) bool {
	// First, check if the path exists
	info, err := os.Stat(path)
	if err != nil {
		// Path doesn't exist - check if we can create it
		if os.IsNotExist(err) {
			// Try to create a temp file in the parent
			parentDir := filepath.Dir(path)
			return IsPathWritable(parentDir)
		}
		return false
	}

	// If it's not a directory, check the parent
	if !info.IsDir() {
		return IsPathWritable(filepath.Dir(path))
	}

	// Try to create a temp file in the directory to verify write access
	tmpFile, err := os.CreateTemp(path, ".ghissues-write-test-*")
	if err != nil {
		return false
	}

	// Clean up the temp file
	tmpFile.Close()
	os.Remove(tmpFile.Name())

	return true
}
