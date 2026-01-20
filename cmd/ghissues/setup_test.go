package main

import (
	"os"
	"strings"
	"testing"

	"github.com/shepbook/ghissues/internal/config"
)

func TestRepoPattern(t *testing.T) {
	tests := []struct {
		repo  string
		valid bool
	}{
		{"owner/repo", true},
		{"owner-name/repo-name", true},
		{"owner123/repo123", true},
		{"owner/repo_name", true},
		{"owner.name/repo.name", true},
		{"owner/", false},
		{"/repo", false},
		{"owner", false},
		{"owner/repo/extra", false},
		{"owner repo", false},
	}

	for _, tt := range tests {
		got := repoPattern.MatchString(tt.repo)
		if got != tt.valid {
			t.Errorf("repoPattern.MatchString(%q) = %v, want %v", tt.repo, got, tt.valid)
		}
	}
}

func TestAuthMethod_Values(t *testing.T) {
	// Verify expected values for auth methods
	if config.AuthMethodEnv != "env" {
		t.Errorf("AuthMethodEnv = %q, want %q", config.AuthMethodEnv, "env")
	}
	if config.AuthMethodGhCli != "gh" {
		t.Errorf("AuthMethodGhCli = %q, want %q", config.AuthMethodGhCli, "gh")
	}
	if config.AuthMethodToken != "token" {
		t.Errorf("AuthMethodToken = %q, want %q", config.AuthMethodToken, "token")
	}
}

func TestIsReRun_DetectsConfig(t *testing.T) {
	// Save original args
	origArgs := strings.Join(os.Args, " ")
	defer func() { os.Args = strings.Fields(origArgs) }()

	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"empty args", []string{}, false},
		{"config subcommand", []string{"ghissues", "config"}, true},
		{"reconfig subcommand", []string{"ghissues", "reconfig"}, true},
		{"other subcommand", []string{"ghissues", "start"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			got := isReRun()
			if got != tt.want {
				t.Errorf("isReRun() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsReRun_PreservesArgs(t *testing.T) {
	// Ensure isReRun doesn't modify os.Args permanently
	origArgs := make([]string, len(os.Args))
	copy(origArgs, os.Args)

	isReRun()

	for i, arg := range os.Args {
		if arg != origArgs[i] {
			t.Errorf("os.Args was modified at index %d: %q -> %q", i, origArgs[i], arg)
		}
	}
}
