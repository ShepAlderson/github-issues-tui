package auth

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/shepbook/ghissues/internal/config"
)

func TestGetToken_FromEnvVar(t *testing.T) {
	// Create a temp directory for config
	tmpDir := t.TempDir()

	// Override HOME
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", filepath.Dir(tmpDir))
	defer os.Setenv("HOME", oldHome)

	// Set GITHUB_TOKEN env var
	oldToken := os.Getenv("GITHUB_TOKEN")
	testToken := "ghp_testenvtoken123"
	os.Setenv("GITHUB_TOKEN", testToken)
	defer os.Setenv("GITHUB_TOKEN", oldToken)

	token, source, err := GetToken()
	if err != nil {
		t.Fatalf("GetToken() unexpected error: %v", err)
	}
	if token != testToken {
		t.Errorf("GetToken() = %q, want %q", token, testToken)
	}
	if source != TokenSourceEnv {
		t.Errorf("GetToken() source = %q, want %q", source, TokenSourceEnv)
	}
}

func TestGetToken_FromConfigFile(t *testing.T) {
	// Create a temp directory for config
	tmpDir := t.TempDir()

	// Override HOME
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", filepath.Dir(tmpDir))
	defer os.Setenv("HOME", oldHome)

	// Clear GITHUB_TOKEN env var
	oldToken := os.Getenv("GITHUB_TOKEN")
	os.Unsetenv("GITHUB_TOKEN")
	defer os.Setenv("GITHUB_TOKEN", oldToken)

	// Create config directory and file with token
	configPath := filepath.Join(filepath.Dir(tmpDir), ".config", "ghissues")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	testToken := "ghp_testconfigtoken456"
	cfg := &config.Config{
		Repository: "test/repo",
		AuthMethod: config.AuthMethodToken,
		Token:      testToken,
	}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	token, source, err := GetToken()
	if err != nil {
		t.Fatalf("GetToken() unexpected error: %v", err)
	}
	if token != testToken {
		t.Errorf("GetToken() = %q, want %q", token, testToken)
	}
	if source != TokenSourceConfig {
		t.Errorf("GetToken() source = %q, want %q", source, TokenSourceConfig)
	}
}

func TestGetToken_FromGhCli(t *testing.T) {
	// Create a temp directory for config
	tmpDir := t.TempDir()

	// Override HOME
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", filepath.Dir(tmpDir))
	defer os.Setenv("HOME", oldHome)

	// Clear GITHUB_TOKEN env var
	oldToken := os.Getenv("GITHUB_TOKEN")
	os.Unsetenv("GITHUB_TOKEN")
	defer os.Setenv("GITHUB_TOKEN", oldToken)

	// Create config directory with no token
	configPath := filepath.Join(filepath.Dir(tmpDir), ".config", "ghissues")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	cfg := &config.Config{
		Repository: "test/repo",
		AuthMethod: config.AuthMethodGhCli,
		Token:      "", // No token in config
	}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Create a mock gh script in a temp bin directory
	binDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("Failed to create bin dir: %v", err)
	}

	// Override PATH to include our mock bin
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir)
	defer os.Setenv("PATH", oldPath)

	// Create mock gh script that returns a token
	mockGhScript := `#!/bin/sh
if [ "$1" = "auth" ] && [ "$2" = "token" ]; then
  echo "ghp_testclitoken789"
  exit 0
fi
exit 1
`
	ghPath := filepath.Join(binDir, "gh")
	if err := os.WriteFile(ghPath, []byte(mockGhScript), 0755); err != nil {
		t.Fatalf("Failed to write mock gh script: %v", err)
	}

	token, source, err := GetToken()
	if err != nil {
		t.Fatalf("GetToken() unexpected error: %v", err)
	}
	if token != "ghp_testclitoken789" {
		t.Errorf("GetToken() = %q, want %q", token, "ghp_testclitoken789")
	}
	if source != TokenSourceGhCli {
		t.Errorf("GetToken() source = %q, want %q", source, TokenSourceGhCli)
	}
}

func TestGetToken_NoAuthFound(t *testing.T) {
	// Create a temp directory for config
	tmpDir := t.TempDir()

	// Override HOME
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", filepath.Dir(tmpDir))
	defer os.Setenv("HOME", oldHome)

	// Clear GITHUB_TOKEN env var
	oldToken := os.Getenv("GITHUB_TOKEN")
	os.Unsetenv("GITHUB_TOKEN")
	defer os.Setenv("GITHUB_TOKEN", oldToken)

	// Create config directory with no token
	configPath := filepath.Join(filepath.Dir(tmpDir), ".config", "ghissues")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	cfg := &config.Config{
		Repository: "test/repo",
		AuthMethod: config.AuthMethodEnv,
		Token:      "", // No token
	}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Create a mock gh script that fails
	binDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("Failed to create bin dir: %v", err)
	}

	// Override PATH to include our mock bin
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir)
	defer os.Setenv("PATH", oldPath)

	// Create mock gh script that fails
	mockGhScript := `#!/bin/sh
exit 1
`
	ghPath := filepath.Join(binDir, "gh")
	if err := os.WriteFile(ghPath, []byte(mockGhScript), 0755); err != nil {
		t.Fatalf("Failed to write mock gh script: %v", err)
	}

	_, _, err := GetToken()
	if err == nil {
		t.Error("GetToken() expected error, got nil")
	}
	if err != ErrNoAuthFound {
		t.Errorf("GetToken() error = %v, want %v", err, ErrNoAuthFound)
	}
}

func TestGetToken_EnvVarTakesPrecedence(t *testing.T) {
	// Create a temp directory for config
	tmpDir := t.TempDir()

	// Override HOME
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", filepath.Dir(tmpDir))
	defer os.Setenv("HOME", oldHome)

	// Set GITHUB_TOKEN env var
	oldToken := os.Getenv("GITHUB_TOKEN")
	envToken := "ghp_envprecedence"
	os.Setenv("GITHUB_TOKEN", envToken)
	defer os.Setenv("GITHUB_TOKEN", oldToken)

	// Create config with a different token
	configPath := filepath.Join(filepath.Dir(tmpDir), ".config", "ghissues")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	cfg := &config.Config{
		Repository: "test/repo",
		AuthMethod: config.AuthMethodToken,
		Token:      "ghp_configdifferent",
	}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	token, source, err := GetToken()
	if err != nil {
		t.Fatalf("GetToken() unexpected error: %v", err)
	}
	if token != envToken {
		t.Errorf("GetToken() = %q, want env var token %q", token, envToken)
	}
	if source != TokenSourceEnv {
		t.Errorf("GetToken() source = %q, want %q (env var should take precedence)", source, TokenSourceEnv)
	}
}

func TestStringTrim(t *testing.T) {
	tests := []struct {
		input    []byte
		expected string
	}{
		{[]byte("token\n"), "token"},
		{[]byte("token\r\n"), "token"},
		{[]byte(" token "), "token"},
		{[]byte("token"), "token"},
		{[]byte("\ntoken\n"), "token"},
	}

	for _, tt := range tests {
		result := stringTrim(tt.input)
		if result != tt.expected {
			t.Errorf("stringTrim(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestGetGhAuthToken_GhNotInstalled(t *testing.T) {
	// Create a temp directory for PATH
	tmpDir := t.TempDir()

	// Create a bin dir without gh
	binDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("Failed to create bin dir: %v", err)
	}

	// Override PATH to exclude gh
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir)
	defer os.Setenv("PATH", oldPath)

	_, err := getGhAuthToken()
	if err == nil {
		t.Error("getGhAuthToken() expected error when gh not found, got nil")
	}
	if !isExecError(err) {
		t.Errorf("getGhAuthToken() error = %v, want exec.Error", err)
	}
}

// isExecError checks if the error is an exec.Error (command not found)
func isExecError(err error) bool {
	_, ok := err.(*exec.Error)
	return ok
}