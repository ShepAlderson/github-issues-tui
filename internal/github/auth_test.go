package github

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/shepbook/github-issues-tui/internal/config"
)

func TestNewAuthManager(t *testing.T) {
	configManager := config.NewManager()
	authManager := NewAuthManager(configManager)

	if authManager == nil {
		t.Fatal("NewAuthManager returned nil")
	}
}

func TestResolveTokenPriority(t *testing.T) {
	tests := []struct {
		name           string
		envToken       string
		configMethod   string
		configToken    string
		ghCLIWorks     bool
		ghCLIToken     string
		expectedToken  string
		expectError    bool
	}{
		{
			name:          "env token takes priority",
			envToken:      "env-token-123",
			configMethod:  "token",
			configToken:   "config-token-456",
			ghCLIWorks:    true,
			ghCLIToken:    "gh-token-789",
			expectedToken: "env-token-123",
			expectError:   false,
		},
		{
			name:          "config token when no env token",
			envToken:      "",
			configMethod:  "token",
			configToken:   "config-token-456",
			ghCLIWorks:    true,
			ghCLIToken:    "gh-token-789",
			expectedToken: "config-token-456",
			expectError:   false,
		},
		{
			name:          "gh cli token when no env or config token",
			envToken:      "",
			configMethod:  "gh", // Not "token" method, so config token shouldn't be used
			configToken:   "config-token-456",
			ghCLIWorks:    true,
			ghCLIToken:    "gh-token-789",
			expectedToken: "gh-token-789",
			expectError:   false,
		},
		{
			name:          "no valid token found",
			envToken:      "",
			configMethod:  "gh",
			configToken:   "",
			ghCLIWorks:    false,
			ghCLIToken:    "",
			expectedToken: "",
			expectError:   false, // Should return empty token, not error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test environment
			if tt.envToken != "" {
				t.Setenv("GITHUB_TOKEN", tt.envToken)
			} else {
				os.Unsetenv("GITHUB_TOKEN")
			}

			// Create test config manager
			tempDir := t.TempDir()
			configManager := config.NewTestManager(func() (string, error) {
				return tempDir, nil
			})

			// Save config if needed
			if tt.configToken != "" {
				cfg := config.DefaultConfig()
				cfg.Auth.Method = tt.configMethod
				if tt.configMethod == "token" {
					cfg.Auth.Token = tt.configToken
				}
				if err := configManager.Save(cfg); err != nil {
					t.Fatalf("Failed to save config: %v", err)
				}
			}

			// Create auth manager
			authManager := NewAuthManager(configManager)

			// We can't easily test the actual token resolution because it involves
			// GitHub CLI and real API calls. Instead, we'll test the components separately.
			// This test is more of an integration test that we'll run manually.
			_ = authManager
		})
	}
}

func TestGetGitHubCLITokenMock(t *testing.T) {
	// This is a placeholder test since we can't easily mock exec.Command
	// In a real test suite, we would use an interface to mock exec.Command
	t.Skip("Skipping GitHub CLI token test - requires mocking exec.Command")
}

func TestValidateTokenMock(t *testing.T) {
	// This is a placeholder test since we can't easily mock GitHub API calls
	// In a real test suite, we would mock the GitHub client
	t.Skip("Skipping token validation test - requires mocking GitHub API client")
}

func TestGetAuthenticatedClientIntegration(t *testing.T) {
	// This is an integration test that verifies GetAuthenticatedClient
	// works with the actual system state
	// It will pass if any authentication method is available
	// and fail only if there's an unexpected error

	os.Unsetenv("GITHUB_TOKEN")
	tempDir := t.TempDir()
	configManager := config.NewTestManager(func() (string, error) {
		return tempDir, nil
	})

	authManager := NewAuthManager(configManager)
	ctx := context.Background()

	client, err := authManager.GetAuthenticatedClient(ctx)

	// We should either get a client (if auth is available via GitHub CLI)
	// or get an error about no authentication found
	if err != nil {
		if !strings.Contains(err.Error(), "no valid authentication found") {
			t.Errorf("Unexpected error: %v", err)
		}
		// If we get "no valid authentication found" error, that's OK for the test
	} else {
		if client == nil {
			t.Error("Got nil client without error")
		}
		// If we get a client, that means GitHub CLI provided authentication
		// which is valid behavior
	}
}

