package refresh

import (
	"testing"
	"time"

	"github.com/shepbook/ghissues/internal/database"
)

func TestPerform(t *testing.T) {
	// Skip this test as it requires real API interaction
	// Integration tests for Perform are better done at a higher level
	t.Skip("Perform requires real GitHub API - skipping in unit tests")
}

func TestShouldAutoRefresh(t *testing.T) {
	// Create a test database
	tempDir := t.TempDir()
	dbPath := tempDir + "/test.db"

	// Initialize the database
	db, err := database.InitializeSchema(dbPath)
	if err != nil {
		t.Fatalf("InitializeSchema failed: %v", err)
	}
	db.Close()

	t.Run("should auto refresh when no sync exists", func(t *testing.T) {
		should, err := ShouldAutoRefresh(dbPath, "owner/repo")
		if err != nil {
			t.Fatalf("ShouldAutoRefresh failed: %v", err)
		}
		if !should {
			t.Error("expected ShouldAutoRefresh to return true when no sync exists")
		}
	})

	t.Run("should auto refresh when sync is old", func(t *testing.T) {
		// Re-initialize database
		db, err := database.InitializeSchema(dbPath)
		if err != nil {
			t.Fatalf("InitializeSchema failed: %v", err)
		}

		// Set last sync time to 10 minutes ago
		oldTime := time.Now().UTC().Add(-10 * time.Minute).Format(time.RFC3339)
		err = database.SaveLastSyncTime(db, "owner/repo", oldTime)
		if err != nil {
			t.Fatalf("SaveLastSyncTime failed: %v", err)
		}
		db.Close()

		should, err := ShouldAutoRefresh(dbPath, "owner/repo")
		if err != nil {
			t.Fatalf("ShouldAutoRefresh failed: %v", err)
		}
		if !should {
			t.Error("expected ShouldAutoRefresh to return true when sync is old")
		}
	})

	t.Run("should not auto refresh when sync is recent", func(t *testing.T) {
		// Re-initialize database
		db, err := database.InitializeSchema(dbPath)
		if err != nil {
			t.Fatalf("InitializeSchema failed: %v", err)
		}

		// Set last sync time to 1 minute ago
		recentTime := time.Now().UTC().Add(-1 * time.Minute).Format(time.RFC3339)
		err = database.SaveLastSyncTime(db, "owner/repo", recentTime)
		if err != nil {
			t.Fatalf("SaveLastSyncTime failed: %v", err)
		}
		db.Close()

		should, err := ShouldAutoRefresh(dbPath, "owner/repo")
		if err != nil {
			t.Fatalf("ShouldAutoRefresh failed: %v", err)
		}
		if should {
			t.Error("expected ShouldAutoRefresh to return false when sync is recent")
		}
	})
}
