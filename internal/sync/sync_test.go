package sync

import (
	"context"
	"testing"

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

// MockAuthManager is a mock implementation of AuthManager for testing
type MockAuthManager struct{}

func (m *MockAuthManager) GetAuthenticatedClient(ctx context.Context) (*gh.Client, error) {
	// Return a nil client for now - in real tests we'd use a proper mock
	return nil, nil
}
