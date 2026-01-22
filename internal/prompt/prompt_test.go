package prompt

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func TestPromptRepository(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "Valid repository format",
			input:       "owner/repo\n",
			expected:    "owner/repo",
			expectError: false,
		},
		{
			name:        "Valid repository with hyphens",
			input:       "my-org/my-repo\n",
			expected:    "my-org/my-repo",
			expectError: false,
		},
		{
			name:        "Valid repository with underscores",
			input:       "my_org/my_repo\n",
			expected:    "my_org/my_repo",
			expectError: false,
		},
		{
			name:        "Invalid format - missing slash",
			input:       "no-slash\n",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Invalid format - too many slashes",
			input:       "too/many/slashes\n",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Empty input",
			input:       "\n",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Whitespace input",
			input:       "   \n",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.NewReader(tt.input)
			output := &bytes.Buffer{}
			reader := bufio.NewReader(input)

			result, err := promptRepository(reader, output)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPromptTokenSelection(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		existingToken string
		expected      string
		expectError   bool
	}{
		{
			name:          "Select environment variable with existing token",
			input:         "1\n",
			existingToken: "existing-token",
			expected:      "env",
			expectError:   false,
		},
		{
			name:          "Select config file token with existing token",
			input:         "2\n",
			existingToken: "existing-token",
			expected:      "config",
			expectError:   false,
		},
		{
			name:          "Select manual token entry with existing token",
			input:         "3\n",
			existingToken: "existing-token",
			expected:      "manual",
			expectError:   false,
		},
		{
			name:          "Select config file token without existing token",
			input:         "1\n",
			existingToken: "",
			expected:      "config",
			expectError:   false,
		},
		{
			name:          "Select manual token without existing token",
			input:         "2\n",
			existingToken: "",
			expected:      "manual",
			expectError:   false,
		},
		{
			name:          "Invalid selection then valid",
			input:         "5\n1\n", // First invalid, then valid
			existingToken: "",
			expected:      "config",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.NewReader(tt.input)
			output := &bytes.Buffer{}
			reader := bufio.NewReader(input)

			result, err := promptTokenSelection(reader, output, tt.existingToken)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPromptManualToken(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "Valid token",
			input:       "ghp_xxxxxxxxxxxx\n",
			expected:    "ghp_xxxxxxxxxxxx",
			expectError: false,
		},
		{
			name:        "Token with spaces should be trimmed",
			input:       "  ghp_token  \n",
			expected:    "ghp_token",
			expectError: false,
		},
		{
			name:        "Empty token",
			input:       "\n",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Whitespace only",
			input:       "   \n",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.NewReader(tt.input)
			output := &bytes.Buffer{}
			reader := bufio.NewReader(input)

			result, err := promptManualToken(reader, output)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
