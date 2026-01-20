package auth

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shepbook/git/github-issues-tui/internal/config"
)

func TestGetGitHubToken_Priority(t *testing.T) {
	tests := []struct {
		name           string
		envToken       string
		configToken    string
		setupGhCLI     bool
		expectedToken  string
		expectedSource string
		expectError    bool
	}{
		{
			name:           "GITHUB_TOKEN environment variable has highest priority",
			envToken:       "env_token_123",
			configToken:    "config_token_456",
			setupGhCLI:     true,
			expectedToken:  "env_token_123",
			expectedSource: "environment",
		},
		{
			name:           "Config token used if no env token",
			envToken:       "",
			configToken:    "config_token_789",
			setupGhCLI:     true,
			expectedToken:  "config_token_789",
			expectedSource: "config",
		},
		{
			name:        "Error when no token found",
			envToken:    "",
			configToken: "",
			setupGhCLI:  false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.toml")

			// Save config file
			data := ""
			if tt.configToken != "" {
				data += "token = \"" + tt.configToken + "\"\n"
			}

			err := os.WriteFile(configPath, []byte(data), 0600)
			if err != nil {
				t.Fatalf("Failed to create config file: %v", err)
			}

			// Set up gh CLI token if requested
			if tt.setupGhCLI {
				ghDir := filepath.Join(tmpDir, "gh")
				err := os.MkdirAll(ghDir, 0755)
				if err != nil {
					t.Fatalf("Failed to create gh dir: %v", err)
				}

				// Write a mock gh CLI hosts.yml
				hostsContent := `
github.com:
    oauth_token: ghcli_token_abc
`
				err = os.WriteFile(filepath.Join(ghDir, "hosts.yml"), []byte(hostsContent), 0600)
				if err != nil {
					t.Fatalf("Failed to create hosts.yml: %v", err)
				}
			}

			// Set environment token
			if tt.envToken != "" {
				os.Setenv("GITHUB_TOKEN", tt.envToken)
				defer os.Unsetenv("GITHUB_TOKEN")
			}

			// Override default paths for testing
			oldGetDefaultConfigPath := defaultGetDefaultConfigPath
			oldGetDefaultGhHostsPath := defaultGetDefaultGhHostsPath
			defer func() {
				defaultGetDefaultConfigPath = oldGetDefaultConfigPath
				defaultGetDefaultGhHostsPath = oldGetDefaultGhHostsPath
			}()

			defaultGetDefaultConfigPath = func() string { return configPath }
			defaultGetDefaultGhHostsPath = func() string {
				if tt.setupGhCLI {
					return filepath.Join(tmpDir, "gh", "hosts.yml")
				}
				return filepath.Join(tmpDir, "gh", "hosts.yml")
			}

			// Create config struct for testing
			testConfig := &config.Config{
				Repository: "test/repo",
				Token:      tt.configToken,
			}

			// Try to get token
			token, source, err := GetGitHubToken(testConfig)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if token != tt.expectedToken {
				t.Errorf("Expected token %q, got %q", tt.expectedToken, token)
			}

			if source != tt.expectedSource {
				t.Errorf("Expected source %q, got %q", tt.expectedSource, source)
			}
		})
	}
}

func TestGetTokenFromEnv(t *testing.T) {
	tests := []struct {
		name        string
		envValue    string
		expectEmpty bool
	}{
		{
			name:        "Returns token when GITHUB_TOKEN is set",
			envValue:    "test_token_123",
			expectEmpty: false,
		},
		{
			name:        "Returns empty when GITHUB_TOKEN is not set",
			envValue:    "",
			expectEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("GITHUB_TOKEN", tt.envValue)
				defer os.Unsetenv("GITHUB_TOKEN")
			}

			token := getTokenFromEnv()

			if tt.expectEmpty && token != "" {
				t.Error("Expected empty token but got value")
			}
			if !tt.expectEmpty && token == "" {
				t.Error("Expected token value but got empty")
			}
			if !tt.expectEmpty && token != tt.envValue {
				t.Errorf("Expected %q, got %q", tt.envValue, token)
			}
		})
	}
}

func TestGetTokenFromConfig(t *testing.T) {
	tests := []struct {
		name          string
		configContent string
		expectToken   string
		expectError   bool
	}{
		{
			name:          "Extracts token from config file",
			configContent: "token = \"config_token_123\"\n",
			expectToken:   "config_token_123",
			expectError:   false,
		},
		{
			name:          "Returns empty if no token in config",
			configContent: "repository = \"test/repo\"\n",
			expectToken:   "",
			expectError:   false,
		},
		{
			name:          "Returns empty for missing config file",
			configContent: "",
			expectToken:   "",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.toml")

			if tt.configContent != "" {
				err := os.WriteFile(configPath, []byte(tt.configContent), 0600)
				if err != nil {
					t.Fatalf("Failed to create config file: %v", err)
				}
			}

			// Override the default config path function
			oldGetDefaultConfigPath := defaultGetDefaultConfigPath
			defer func() { defaultGetDefaultConfigPath = oldGetDefaultConfigPath }()
			defaultGetDefaultConfigPath = func() string { return configPath }

			token, err := getTokenFromConfig()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if token != tt.expectToken {
				t.Errorf("Expected token %q, got %q", tt.expectToken, token)
			}
		})
	}
}

func TestGetTokenFromGhCLI(t *testing.T) {
	tests := []struct {
		name         string
		hostsContent string
		expectToken  string
		expectError  bool
	}{
		{
			name: "Extracts token from gh CLI hosts.yml",
			hostsContent: `
github.com:
    oauth_token: ghcli_token_xyz
`,
			expectToken: "ghcli_token_xyz",
			expectError: false,
		},
		{
			name:         "Returns empty if hosts.yml doesn't exist",
			hostsContent: "",
			expectToken:  "",
			expectError:  false,
		},
		{
			name: "Returns empty if no github.com entry",
			hostsContent: `
github.enterprise.com:
    oauth_token: enterprise_token
`,
			expectToken: "",
			expectError: false,
		},
		{
			name: "Returns empty if no oauth_token field",
			hostsContent: `
github.com:
    user: testuser
`,
			expectToken: "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			var hostsPath string

			if tt.hostsContent != "" {
				ghDir := filepath.Join(tmpDir, "gh")
				err := os.MkdirAll(ghDir, 0755)
				if err != nil {
					t.Fatalf("Failed to create gh dir: %v", err)
				}

				hostsPath = filepath.Join(ghDir, "hosts.yml")
				err = os.WriteFile(hostsPath, []byte(tt.hostsContent), 0600)
				if err != nil {
					t.Fatalf("Failed to create hosts.yml: %v", err)
				}
			} else {
				hostsPath = filepath.Join(tmpDir, "nonexistent", "hosts.yml")
			}

			// Override the default gh hosts path function
			oldGetDefaultGhHostsPath := defaultGetDefaultGhHostsPath
			defer func() { defaultGetDefaultGhHostsPath = oldGetDefaultGhHostsPath }()
			defaultGetDefaultGhHostsPath = func() string { return hostsPath }

			token, err := getTokenFromGhCLI()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if token != tt.expectToken {
				t.Errorf("Expected token %q, got %q", tt.expectToken, token)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		expectError bool
		expectValid bool
	}{
		{
			name:        "Empty token returns error without making API call",
			token:       "",
			expectError: true,
			expectValid: false,
		},
		{
			name:        "Valid token format should attempt validation",
			token:       "ghp_1234567890abcdefghijklmnopqrst",
			expectError: false,
			expectValid: true,
		},
		{
			name:        "Token with github_ prefix should attempt validation",
			token:       "github_pat_1234567890abcdefghijklmnopqrst",
			expectError: false,
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := ValidateToken(tt.token)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if valid != tt.expectValid {
				t.Errorf("Expected valid=%v, got valid=%v", tt.expectValid, valid)
			}
		})
	}
}
