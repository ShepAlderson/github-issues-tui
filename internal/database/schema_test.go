package database

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestNewDBManager(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	// Test successful creation
	manager, err := NewDBManager(dbPath)
	if err != nil {
		t.Fatalf("NewDBManager failed: %v", err)
	}
	defer manager.Close()

	// Test schema initialization
	err = manager.InitializeSchema()
	if err != nil {
		t.Fatalf("InitializeSchema failed: %v", err)
	}

	// Test closing
	err = manager.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Test database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("Database file was not created")
	}
}

func TestNewDBManagerInvalidPath(t *testing.T) {
	// Test with invalid path (directory that doesn't exist)
	_, err := NewDBManager("/nonexistent/path/test.db")
	if err == nil {
		t.Fatal("Expected error for invalid path, got none")
	}
}

func TestInitializeSchema(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	manager, err := NewDBManager(dbPath)
	if err != nil {
		t.Fatalf("NewDBManager failed: %v", err)
	}
	defer manager.Close()

	// Initialize schema
	err = manager.InitializeSchema()
	if err != nil {
		t.Fatalf("InitializeSchema failed: %v", err)
	}

	// Verify tables exist by querying them
	tables := []string{"issues", "comments", "metadata"}
	for _, table := range tables {
		query := fmt.Sprintf("SELECT name FROM sqlite_master WHERE type='table' AND name='%s'", table)
		var name string
		err := manager.GetDB().QueryRow(query).Scan(&name)
		if err != nil {
			t.Fatalf("Table %s does not exist: %v", table, err)
		}
		if name != table {
			t.Fatalf("Expected table name %s, got %s", table, name)
		}
	}

	// Verify indexes exist
	indexes := []string{"idx_issues_number", "idx_issues_updated", "idx_comments_issue_number", "idx_comments_created"}
	for _, index := range indexes {
		query := fmt.Sprintf("SELECT name FROM sqlite_master WHERE type='index' AND name='%s'", index)
		var name string
		err := manager.GetDB().QueryRow(query).Scan(&name)
		if err != nil {
			t.Fatalf("Index %s does not exist: %v", index, err)
		}
		if name != index {
			t.Fatalf("Expected index name %s, got %s", index, name)
		}
	}
}

func TestInitializeSchemaMultipleTimes(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	manager, err := NewDBManager(dbPath)
	if err != nil {
		t.Fatalf("NewDBManager failed: %v", err)
	}
	defer manager.Close()

	// Initialize schema multiple times - should not error
	for i := 0; i < 3; i++ {
		err = manager.InitializeSchema()
		if err != nil {
			t.Fatalf("InitializeSchema failed on iteration %d: %v", i, err)
		}
	}
}

func TestGetDB(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	manager, err := NewDBManager(dbPath)
	if err != nil {
		t.Fatalf("NewDBManager failed: %v", err)
	}
	defer manager.Close()

	db := manager.GetDB()
	if db == nil {
		t.Fatal("GetDB returned nil")
	}

	// Verify we can use the database
	var version string
	err = db.QueryRow("SELECT sqlite_version()").Scan(&version)
	if err != nil {
		t.Fatalf("Failed to query database: %v", err)
	}
	if version == "" {
		t.Fatal("Empty version string returned")
	}
}

func TestGetLastSyncTime(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	manager, err := NewDBManager(dbPath)
	if err != nil {
		t.Fatalf("NewDBManager failed: %v", err)
	}
	defer manager.Close()

	// Initialize schema
	err = manager.InitializeSchema()
	if err != nil {
		t.Fatalf("InitializeSchema failed: %v", err)
	}

	// Test when no sync time is recorded
	syncTime, err := manager.GetLastSyncTime()
	if err != nil {
		t.Fatalf("GetLastSyncTime failed: %v", err)
	}
	if !syncTime.IsZero() {
		t.Fatalf("Expected zero time for no sync record, got: %v", syncTime)
	}

	// Insert a sync time
	insertTime := "2024-01-15T12:30:45Z"
	_, err = manager.GetDB().Exec("INSERT OR REPLACE INTO metadata (key, value) VALUES ('last_sync', ?)", insertTime)
	if err != nil {
		t.Fatalf("Failed to insert sync time: %v", err)
	}

	// Test retrieving the sync time
	syncTime, err = manager.GetLastSyncTime()
	if err != nil {
		t.Fatalf("GetLastSyncTime failed: %v", err)
	}
	if syncTime.IsZero() {
		t.Fatal("Expected non-zero time after inserting sync record")
	}

	// Verify the time matches what we inserted
	expectedTimeStr := syncTime.Format("2006-01-02T15:04:05Z")
	if expectedTimeStr != insertTime {
		t.Fatalf("Expected sync time %s, got %s", insertTime, expectedTimeStr)
	}
}
