package prompt

import (
	"strings"
	"testing"

	"github.com/shepbook/github-issues-tui/internal/config"
)

func TestParseRepositoryInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid owner/repo format",
			input:   "owner/repo",
			wantErr: false,
		},
		{
			name:    "valid with hyphens",
			input:   "my-owner/my-repo",
			wantErr: false,
		},
		{
			name:    "valid with underscores",
			input:   "my_owner/my_repo",
			wantErr: false,
		},
		{
			name:    "invalid missing owner",
			input:   "/repo",
			wantErr: true,
		},
		{
			name:    "invalid missing repo",
			input:   "owner/",
			wantErr: true,
		},
		{
			name:    "invalid no slash",
			input:   "ownerrepo",
			wantErr: true,
		},
		{
			name:    "invalid empty",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid multiple slashes",
			input:   "owner/repo/extra",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ParseRepositoryInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRepositoryInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseAuthMethodInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "env variable choice",
			input:   "env",
			want:    "env",
			wantErr: false,
		},
		{
			name:    "token choice",
			input:   "token",
			want:    "token",
			wantErr: false,
		},
		{
			name:    "gh cli choice",
			input:   "gh",
			want:    "gh",
			wantErr: false,
		},
		{
			name:    "case insensitive",
			input:   "ENV",
			want:    "env",
			wantErr: false,
		},
		{
			name:    "invalid choice",
			input:   "invalid",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAuthMethodInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAuthMethodInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseAuthMethodInput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseTokenInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid GitHub token",
			input:   "ghp_1234567890abcdefghijklmnopqrstuv",
			wantErr: false,
		},
		{
			name:    "valid classic token",
			input:   "ghp_abcdefghij",
			wantErr: false,
		},
		{
			name:    "empty token",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ParseTokenInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTokenInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunInteractiveSetup(t *testing.T) {
	// Test successful setup with token auth
	input := "owner/repo\ntoken\nghp_test123\n"
	cfg, err := runInteractiveSetupWithInput(strings.NewReader(input))
	if err != nil {
		t.Fatalf("RunInteractiveSetup failed: %v", err)
	}

	if cfg.GitHub.Repository != "owner/repo" {
		t.Errorf("Expected repository 'owner/repo', got '%s'", cfg.GitHub.Repository)
	}
	if cfg.GitHub.AuthMethod != "token" {
		t.Errorf("Expected auth_method 'token', got '%s'", cfg.GitHub.AuthMethod)
	}
	if cfg.GitHub.Token != "ghp_test123" {
		t.Errorf("Expected token 'ghp_test123', got '%s'", cfg.GitHub.Token)
	}
}

func TestRunInteractiveSetupWithEnv(t *testing.T) {
	// Test setup with env auth (no token prompt)
	input := "owner/repo\nenv\n"
	cfg, err := runInteractiveSetupWithInput(strings.NewReader(input))
	if err != nil {
		t.Fatalf("RunInteractiveSetup failed: %v", err)
	}

	if cfg.GitHub.Repository != "owner/repo" {
		t.Errorf("Expected repository 'owner/repo', got '%s'", cfg.GitHub.Repository)
	}
	if cfg.GitHub.AuthMethod != "env" {
		t.Errorf("Expected auth_method 'env', got '%s'", cfg.GitHub.AuthMethod)
	}
	if cfg.GitHub.Token != "" {
		t.Errorf("Expected empty token for env auth, got '%s'", cfg.GitHub.Token)
	}
}

func TestRunInteractiveSetupWithGH(t *testing.T) {
	// Test setup with gh CLI auth (no token prompt)
	input := "owner/repo\ngh\n"
	cfg, err := runInteractiveSetupWithInput(strings.NewReader(input))
	if err != nil {
		t.Fatalf("RunInteractiveSetup failed: %v", err)
	}

	if cfg.GitHub.Repository != "owner/repo" {
		t.Errorf("Expected repository 'owner/repo', got '%s'", cfg.GitHub.Repository)
	}
	if cfg.GitHub.AuthMethod != "gh" {
		t.Errorf("Expected auth_method 'gh', got '%s'", cfg.GitHub.AuthMethod)
	}
	if cfg.GitHub.Token != "" {
		t.Errorf("Expected empty token for gh auth, got '%s'", cfg.GitHub.Token)
	}
}

func TestRunInteractiveSetupInvalidInputRetry(t *testing.T) {
	// Test that invalid inputs are rejected and user can retry
	// First attempt: invalid repo format, then valid
	input := "invalid\nowner/repo\ntoken\nghp_test\n"
	cfg, err := runInteractiveSetupWithInput(strings.NewReader(input))
	if err != nil {
		t.Fatalf("RunInteractiveSetup failed: %v", err)
	}

	if cfg.GitHub.Repository != "owner/repo" {
		t.Errorf("Expected repository 'owner/repo', got '%s'", cfg.GitHub.Repository)
	}
}

// Helper function for testing with custom input
func runInteractiveSetupWithInput(input *strings.Reader) (*config.Config, error) {
	return RunInteractiveSetupWithReader(input)
}
