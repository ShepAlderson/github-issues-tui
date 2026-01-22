package prompt

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestRunInteractiveSetup_ManualToken(t *testing.T) {
	// Ensure GITHUB_TOKEN is not set
	oldToken := os.Getenv("GITHUB_TOKEN")
	os.Unsetenv("GITHUB_TOKEN")
	defer func() {
		if oldToken != "" {
			os.Setenv("GITHUB_TOKEN", oldToken)
		}
	}()

	// Test input: repo, select option 2, enter token
	input := bytes.NewBufferString("testuser/testrepo\n2\nghp_testtoken123\n")
	output := &bytes.Buffer{}

	repository, token, err := RunInteractiveSetup(input, output)
	if err != nil {
		t.Fatalf("RunInteractiveSetup failed: %v", err)
	}

	if repository != "testuser/testrepo" {
		t.Errorf("Expected repository 'testuser/testrepo', got %q", repository)
	}
	if token != "ghp_testtoken123" {
		t.Errorf("Expected token 'ghp_testtoken123', got %q", token)
	}

	// Check that output contains expected prompts
	outputStr := output.String()
	if !strings.Contains(outputStr, "repository") {
		t.Error("Output should contain repository prompt")
	}
	if !strings.Contains(outputStr, "authentication method") {
		t.Error("Output should contain token selection prompt")
	}
}
