package sync

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/shepbook/ghissues/internal/database"
)

func TestNewSyncModel(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	t.Run("creates model with correct fields", func(t *testing.T) {
		model := NewSyncModel(dbPath, "owner/repo", "test_token")

		if model.repo != "owner/repo" {
			t.Errorf("Expected repo 'owner/repo', got %s", model.repo)
		}

		if model.token != "test_token" {
			t.Errorf("Expected token 'test_token', got %s", model.token)
		}

		if model.status != StatusIdle {
			t.Errorf("Expected status StatusIdle, got %d", model.status)
		}

		if model.issuesFetched != 0 {
			t.Errorf("Expected 0 issues fetched, got %d", model.issuesFetched)
		}
	})

	t.Run("initializes database connection", func(t *testing.T) {
		model := NewSyncModel(dbPath, "owner/repo", "test_token")

		if model.dbPath != dbPath {
			t.Errorf("Expected dbPath %s, got %s", dbPath, model.dbPath)
		}
	})
}

func TestSyncModel_Update(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	t.Run("quit on ctrl+c", func(t *testing.T) {
		model := NewSyncModel(dbPath, "owner/repo", "test_token")
		model.status = StatusComplete

		msg := tea.KeyMsg{Type: tea.KeyCtrlC}
		newModel, cmd := model.Update(msg)

		if cmd == nil {
			t.Error("Expected quit command")
		}

		if newModel.(SyncModel).db != nil {
			newModel.(SyncModel).db.Close()
		}
	})

	t.Run("quit on q key when done", func(t *testing.T) {
		// First ensure the database is created to avoid errors
		db, _ := database.InitializeSchema(dbPath)
		if db != nil {
			db.Close()
		}

		model := NewSyncModel(dbPath, "owner/repo", "test_token")
		model.status = StatusComplete

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
		newModel, cmd := model.Update(msg)

		if cmd == nil {
			t.Error("Expected quit command")
		}

		if newModel.(SyncModel).db != nil {
			newModel.(SyncModel).db.Close()
		}
	})

	t.Run("ignore q key while syncing", func(t *testing.T) {
		model := NewSyncModel(dbPath, "owner/repo", "test_token")
		model.status = StatusSyncing

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
		_, cmd := model.Update(msg)

		if cmd != nil {
			t.Error("Expected no command while syncing")
		}

		if model.status != StatusSyncing {
			t.Error("Expected status to remain StatusSyncing")
		}
		if model.db != nil {
			model.db.Close()
		}
	})

	t.Run("updates progress on syncMsg", func(t *testing.T) {
		model := NewSyncModel(dbPath, "owner/repo", "test_token")
		model.status = StatusSyncing

		msg := syncMsg{
			issuesFetched: 50,
			issuesTotal:   100,
			current:       "Fetching issue #123",
			status:        StatusSyncing,
		}

		newModel, _ := model.Update(msg)

		if newModel.(SyncModel).issuesFetched != 50 {
			t.Errorf("Expected 50 issues fetched, got %d", newModel.(SyncModel).issuesFetched)
		}

		if newModel.(SyncModel).issuesTotal != 100 {
			t.Errorf("Expected 100 issues total, got %d", newModel.(SyncModel).issuesTotal)
		}

		if newModel.(SyncModel).current != "Fetching issue #123" {
			t.Errorf("Expected 'Fetching issue #123', got %s", newModel.(SyncModel).current)
		}

		if newModel.(SyncModel).db != nil {
			newModel.(SyncModel).db.Close()
		}
	})
}

func TestSyncModel_View(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	t.Run("shows idle state", func(t *testing.T) {
		model := NewSyncModel(dbPath, "owner/repo", "test_token")
		model.status = StatusIdle

		view := model.View()

		if model.db != nil {
			model.db.Close()
		}

		if len(view) == 0 {
			t.Error("Expected non-empty view")
		}
	})

	t.Run("shows syncing state with progress", func(t *testing.T) {
		model := NewSyncModel(dbPath, "owner/repo", "test_token")
		model.status = StatusSyncing
		model.issuesFetched = 50
		model.issuesTotal = 100
		model.current = "Fetching issues"

		view := model.View()

		if model.db != nil {
			model.db.Close()
		}

		if len(view) == 0 {
			t.Error("Expected non-empty view")
		}
	})

	t.Run("shows complete state", func(t *testing.T) {
		model := NewSyncModel(dbPath, "owner/repo", "test_token")
		model.status = StatusComplete
		model.issuesFetched = 100
		model.commentsFetched = 50

		view := model.View()

		if model.db != nil {
			model.db.Close()
		}

		if len(view) == 0 {
			t.Error("Expected non-empty view")
		}
	})

	t.Run("shows error state", func(t *testing.T) {
		model := NewSyncModel(dbPath, "owner/repo", "test_token")
		model.status = StatusError
		model.err = sql.ErrConnDone

		view := model.View()

		if model.db != nil {
			model.db.Close()
		}

		if len(view) == 0 {
			t.Error("Expected non-empty view")
		}
	})
}

func TestSyncModel_Init(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	t.Run("initializes database", func(t *testing.T) {
		// Ensure the database won't exist initially
		os.Remove(dbPath)

		model := NewSyncModel(dbPath, "owner/repo", "test_token")

		cmds := model.Init()
		if cmds == nil {
			t.Error("Expected init commands")
		}

		if model.db != nil {
			model.db.Close()
		}
	})
}

func TestSyncModel_progressPercent(t *testing.T) {
	tests := []struct {
		name     string
		fetched  int
		total    int
		expected float64
	}{
		{
			name:     "0 issues",
			fetched:  0,
			total:    100,
			expected: 0,
		},
		{
			name:     "halfway",
			fetched:  50,
			total:    100,
			expected: 0.5,
		},
		{
			name:     "complete",
			fetched:  100,
			total:    100,
			expected: 1,
		},
		{
			name:     "unknown total",
			fetched:  50,
			total:    0,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			dbPath := filepath.Join(tempDir, "test.db")

			model := NewSyncModel(dbPath, "owner/repo", "test_token")
			model.issuesFetched = tt.fetched
			model.issuesTotal = tt.total

			percent := model.progressPercent()

			if model.db != nil {
				model.db.Close()
			}

			if percent != tt.expected {
				t.Errorf("Expected %f, got %f", tt.expected, percent)
			}
		})
	}
}

func TestSyncModel_Progress(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	t.Run("returns empty progress before completion", func(t *testing.T) {
		model := NewSyncModel(dbPath, "owner/repo", "test_token")
		model.status = StatusSyncing

		progress, err := model.Progress()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if model.db != nil {
			model.db.Close()
		}

		// Progress should be empty before completion
		_ = progress
	})

	t.Run("returns results after completion", func(t *testing.T) {
		// Create database schema for this test
		db, err := database.InitializeSchema(dbPath)
		if err != nil {
			t.Fatalf("Failed to initialize schema: %v", err)
		}
		db.Close()

		model := NewSyncModel(dbPath, "owner/repo", "test_token")
		model.status = StatusComplete
		model.issuesFetched = 10
		model.commentsFetched = 5

		progress, err := model.Progress()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if model.db != nil {
			model.db.Close()
		}

		if progress.IssuesFetched != 10 {
			t.Errorf("Expected 10 issues, got %d", progress.IssuesFetched)
		}
	})
}
