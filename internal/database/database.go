package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const defaultDatabasePath = ".ghissues.db"

// ResolveDatabasePath determines which database path to use based on flag and config
// Priority: flag > config > default
func ResolveDatabasePath(flagPath, configPath string) (string, error) {
	// Flag takes precedence
	if strings.TrimSpace(flagPath) != "" {
		return flagPath, nil
	}

	// Config file takes second priority
	if strings.TrimSpace(configPath) != "" {
		return configPath, nil
	}

	// Default path
	return defaultDatabasePath, nil
}

// EnsureDatabasePath ensures the parent directory for the database file exists
// If the directory doesn't exist, it will be created
func EnsureDatabasePath(dbPath string) error {
	// Get the parent directory
	parentDir := filepath.Dir(dbPath)

	// If parentDir is ".", the database is in the current directory
	// No directory creation needed
	if parentDir == "." {
		return nil
	}

	// Check if parent directory exists
	info, err := os.Stat(parentDir)
	if err == nil {
		// Path exists, verify it's a directory
		if !info.IsDir() {
			return fmt.Errorf("parent path %q is not a directory", parentDir)
		}
		return nil
	}

	// Directory doesn't exist, create it
	if os.IsNotExist(err) {
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return fmt.Errorf("failed to create database directory %q: %w", parentDir, err)
		}
		return nil
	}

	return fmt.Errorf("failed to access database directory %q: %w", parentDir, err)
}

// CheckDatabaseWritable verifies that the database file can be created/written
// Returns an error if the location is not writable
func CheckDatabaseWritable(dbPath string) error {
	parentDir := filepath.Dir(dbPath)

	// Check if parent directory exists and is writable
	info, err := os.Stat(parentDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Try to create the directory to test writability
			if err := os.MkdirAll(parentDir, 0755); err != nil {
				return fmt.Errorf("cannot create database file at %q: parent directory cannot be created: %w", dbPath, err)
			}
			// Clean up the test directory
			os.Remove(parentDir)
			return nil
		}
		return fmt.Errorf("cannot create database file at %q: %w", dbPath, err)
	}

	// Path exists, check if it's a directory
	if !info.IsDir() {
		return fmt.Errorf("cannot create database file at %q: parent path is not a directory", dbPath)
	}

	// Check if directory is writable by testing file creation
	testFile := filepath.Join(parentDir, ".ghissues-write-test")
	file, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("cannot create database file at %q: %w", dbPath, err)
	}
	file.Close()

	// Clean up test file
	os.Remove(testFile)

	return nil
}
