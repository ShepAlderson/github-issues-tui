package auth

import (
	"os"
	"strings"
	"testing"

	"github.com/shepbook/github-issues-tui/internal/config"
)

func TestGetToken_EnvironmentVariable(t *testing.T) {
	// Set up environment variable
	os.Setenv("GITHUB_TOKEN", "env_token_123")
	defer os.Unsetenv("GITHUB_TOKEN")

	cfg := &config.Config{
		GitHub: config.GitHubConfig{
			AuthMethod: "env",
		},
	}

	token, err := GetToken(cfg)
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if token != "env_token_123" {
		t.Errorf("expected token 'env_token_123', got '%s'", token)
	}
}

func TestGetToken_ConfigFile(t *testing.T) {
	// Ensure no environment variable
	os.Unsetenv("GITHUB_TOKEN")

	cfg := &config.Config{
		GitHub: config.GitHubConfig{
			AuthMethod: "token",
			Token:      "config_token_456",
		},
	}

	token, err := GetToken(cfg)
	if err != nil {
		t.Fatalf("GetToken failed: %v", err)
	}

	if token != "config_token_456" {
		t.Errorf("expected token 'config_token_456', got '%s'", token)
	}
}

func TestGetToken_FallbackOrder(t *testing.T) {
	tests := []struct {
		name           string
		envToken       string
		configToken    string
		authMethod     string
		expectedToken  string
		expectedError  bool
	}{
		{
			name:          "env variable takes precedence",
			envToken:      "env_token",
			configToken:   "config_token",
			authMethod:    "env",
			expectedToken: "env_token",
		},
		{
			name:          "config file when env not set",
			envToken:      "",
			configToken:   "config_token",
			authMethod:    "token",
			expectedToken: "config_token",
		},
		{
			name:          "error when no token available",
			envToken:      "",
			configToken:   "",
			authMethod:    "token",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.envToken != "" {
				os.Setenv("GITHUB_TOKEN", tt.envToken)
				defer os.Unsetenv("GITHUB_TOKEN")
			} else {
				os.Unsetenv("GITHUB_TOKEN")
			}

			cfg := &config.Config{
				GitHub: config.GitHubConfig{
					AuthMethod: tt.authMethod,
					Token:      tt.configToken,
				},
			}

			token, err := GetToken(cfg)

			if tt.expectedError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if token != tt.expectedToken {
				t.Errorf("expected token '%s', got '%s'", tt.expectedToken, token)
			}
		})
	}
}

func TestGetToken_ClearErrorMessage(t *testing.T) {
	os.Unsetenv("GITHUB_TOKEN")

	cfg := &config.Config{
		GitHub: config.GitHubConfig{
			AuthMethod: "env",
		},
	}

	_, err := GetToken(cfg)
	if err == nil {
		t.Fatal("expected error when no token available")
	}

	// Check that error message is helpful
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("error message should not be empty")
	}

	// Should mention authentication and provide guidance
	if len(errMsg) < 20 {
		t.Errorf("error message too short, might not be helpful: %s", errMsg)
	}
}

func TestGetToken_GhCLI(t *testing.T) {
	// Skip this test if gh CLI is not available
	// This is a placeholder for the gh CLI integration
	// We'll implement this to call `gh auth token` command
	t.Skip("gh CLI integration test - implement when gh CLI auth is added")
}

func TestValidateToken_InvalidToken(t *testing.T) {
	// Test with an obviously invalid token
	err := ValidateToken("invalid_token_123")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}

	// Check error message is helpful
	errMsg := err.Error()
	if !strings.Contains(errMsg, "invalid") && !strings.Contains(errMsg, "token") {
		t.Errorf("error message should mention invalid token, got: %s", errMsg)
	}
}

func TestValidateToken_EmptyToken(t *testing.T) {
	// Test with empty token
	err := ValidateToken("")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

// Note: Testing ValidateToken with a real valid token would require
// network access and a real token, which is not suitable for unit tests.
// Integration tests or manual testing should be used for that scenario.
