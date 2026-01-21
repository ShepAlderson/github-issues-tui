package sync

import (
	"context"
	"testing"
	"time"

	gh "github.com/google/go-github/v62/github"
	"github.com/shepbook/github-issues-tui/internal/config"
	"github.com/shepbook/github-issues-tui/internal/database"
)

func TestNewSyncManager(t *testing.T) {
	// Create test dependencies
	configManager := config.NewTestManager(func() (string, error) {
		return t.TempDir(), nil
	})

	// Note: AuthManager requires configManager but we can't instantiate it directly
	// in tests since it has unexported fields. This test will need to be updated
	// when we have a proper test constructor for AuthManager.

	tempDir := t.TempDir()
	dbPath := tempDir + "/test.db"
	dbManager, err := database.NewDBManager(dbPath)
	if err != nil {
		t.Fatalf("Failed to create DB manager: %v", err)
	}
	defer dbManager.Close()

	// Test creating sync manager with nil auth manager (for now)
	manager := NewSyncManager(configManager, nil, dbManager)
	if manager == nil {
		t.Fatal("NewSyncManager returned nil")
	}
}

func TestParseRepo(t *testing.T) {
	tests := []struct {
		name     string
		repoStr  string
		expected []string
	}{
		{"Valid repo", "owner/repo", []string{"owner", "repo"}},
		{"Empty string", "", nil},
		{"No slash", "owner", nil},
		{"Multiple slashes", "owner/repo/extra", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitRepo(tt.repoStr)
			if len(result) != len(tt.expected) {
				t.Errorf("splitRepo(%q) = %v, expected %v", tt.repoStr, result, tt.expected)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("splitRepo(%q) = %v, expected %v", tt.repoStr, result, tt.expected)
					return
				}
			}
		})
	}
}

func TestCreateIssueListOptions(t *testing.T) {
	zeroTime := time.Time{}
	nonZeroTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		lastSyncTime time.Time
		page         int
		perPage      int
		expectSince  bool
	}{
		{"Zero time", zeroTime, 1, 100, false},
		{"Non-zero time", nonZeroTime, 2, 50, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := createIssueListOptions(tt.lastSyncTime, tt.page, tt.perPage)

			// Check State field
			if opts.State != "open" {
				t.Errorf("State = %q, want %q", opts.State, "open")
			}

			// Check pagination
			if opts.Page != tt.page {
				t.Errorf("Page = %d, want %d", opts.Page, tt.page)
			}
			if opts.PerPage != tt.perPage {
				t.Errorf("PerPage = %d, want %d", opts.PerPage, tt.perPage)
			}

			// Check Since field
			if tt.expectSince {
				if opts.Since.IsZero() {
					t.Error("Since field should be set but is zero")
				}
				if !opts.Since.Equal(tt.lastSyncTime) {
					t.Errorf("Since = %v, want %v", opts.Since, tt.lastSyncTime)
				}
			} else {
				if !opts.Since.IsZero() {
					t.Errorf("Since field should be zero but got %v", opts.Since)
				}
			}
		})
	}
}

func TestCreateCommentListOptions(t *testing.T) {
	zeroTime := time.Time{}
	nonZeroTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		lastSyncTime time.Time
		page         int
		perPage      int
		expectSince  bool
	}{
		{"Zero time", zeroTime, 1, 100, false},
		{"Non-zero time", nonZeroTime, 2, 50, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := createCommentListOptions(tt.lastSyncTime, tt.page, tt.perPage)

			// Check pagination
			if opts.Page != tt.page {
				t.Errorf("Page = %d, want %d", opts.Page, tt.page)
			}
			if opts.PerPage != tt.perPage {
				t.Errorf("PerPage = %d, want %d", opts.PerPage, tt.perPage)
			}

			// Check Since field
			if tt.expectSince {
				if opts.Since == nil {
					t.Error("Since pointer should be set but is nil")
				} else if opts.Since.IsZero() {
					t.Error("Since pointer points to zero time")
				} else if !opts.Since.Equal(tt.lastSyncTime) {
					t.Errorf("Since = %v, want %v", opts.Since, tt.lastSyncTime)
				}
			} else {
				if opts.Since != nil {
					t.Errorf("Since pointer should be nil but got %v", opts.Since)
				}
			}
		})
	}
}

// MockAuthManager is a mock implementation of AuthManager for testing
type MockAuthManager struct{}

func (m *MockAuthManager) GetAuthenticatedClient(ctx context.Context) (*gh.Client, error) {
	// Return a nil client for now - in real tests we'd use a proper mock
	return nil, nil
}
