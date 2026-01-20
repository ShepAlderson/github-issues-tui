# Ralph Progress Log

This file tracks progress across iterations. It's automatically updated
after each iteration and included in agent prompts for context.

## Codebase Patterns (Study These First)

**Config Module Pattern**: Use a dedicated `internal/config` package that:
- Provides `ConfigPath()`, `ConfigFilePath()`, `Exists()` helpers
- Uses `os.UserHomeDir()` and `os.Getenv("HOME")` for cross-platform home directory
- Creates config directory with `os.MkdirAll(dir, 0755)` before saving
- Uses secure file permissions `0600` for config files
- Uses TOML for configuration format with `github.com/BurntSushi/toml`

**Test Pattern for Config**: Override `HOME` environment variable in tests to control config path without modifying global state.

**CLI Subcommand Pattern**: Check `os.Args` for subcommands before main logic, allowing `ghissues config` to trigger setup.

**Auth Package Pattern**: Use a dedicated `internal/auth` package that:
- Provides `GetToken()` function that returns (token, source, error)
- Attempts auth in order: env var -> config file -> gh CLI
- Returns `TokenSource` enum for user feedback
- Uses `exec.Command("gh", "auth", "token")` for gh CLI integration

**Test Pattern for External Commands**: Override PATH environment variable in tests to include a temp directory containing mock scripts for testing exec.Command calls.

**GitHub API Client Pattern**: Use `net/http` for GitHub API calls with:
- Context support for cancellation
- Proper Accept header: `application/vnd.github.v3+json`
- User-Agent header: `ghissues-tui`
- Check for 401 (bad credentials) and 403 (rate limit or forbidden)

---

## 2026-01-20 - US-001
- What was implemented:
  - Go project structure with cmd/ghissues and internal/config packages
  - Config module for loading/saving TOML config to ~/.config/ghissues/config.toml
  - Interactive setup command that prompts for repository (owner/repo) and auth method
  - Config exists check to skip setup on subsequent runs
  - `ghissues config` subcommand to re-run setup

- Files changed:
  - cmd/ghissues/main.go (main entry point)
  - cmd/ghissues/setup.go (interactive prompts)
  - cmd/ghissues/setup_test.go (setup tests)
  - internal/config/config.go (config loading/saving)
  - internal/config/config_test.go (config tests)
  - go.mod, go.sum (dependencies)

- **Learnings:**
  - Patterns discovered:
    - TOML encoding with `BurntSushi/toml` using `NewEncoder` with `bytes.Buffer`
    - Testing file operations by overriding HOME environment variable
    - Slice bounds checking before accessing `os.Args[1:]` to avoid panics
    - `os.MkdirAll` ensures parent directories exist before file write
    - Config file permissions should be 0600 for security
  - Gotchas encountered:
    - `os.Args[1:]` panics when len(os.Args) < 2 - need bounds check
    - `toml.NewEncoder` requires `io.Writer`, not `*[]byte` - use `bytes.Buffer`
    - Go tests run in parallel by default, need to properly restore env vars with defer

---

## ✓ Iteration 1 - US-001: First-Time Setup
*2026-01-20T08:39:35.031Z (218s)*

**Status:** Completed

---

## 2026-01-20 - US-002
- What was implemented:
  - Created internal/auth package with GetToken() for authentication chain
  - Authentication order: GITHUB_TOKEN env var -> config file -> gh CLI token
  - Clear error message when no authentication found
  - Created internal/github package with token validation via GitHub API
  - Updated setup to prompt for token when using token auth method
  - Token validated on startup with helpful error if invalid

- Files changed:
  - internal/auth/auth.go (GetToken function with auth chain)
  - internal/auth/auth_test.go (7 tests for auth precedence)
  - internal/github/client.go (GitHub API client with ValidateToken)
  - internal/github/client_test.go (6 tests for client)
  - cmd/ghissues/main.go (added validateAuth call on startup)
  - cmd/ghissues/setup.go (added promptToken function)
  - cmd/ghissues/setup_test.go (4 new tests for promptToken)

- **Learnings:**
  - Patterns discovered:
    - Auth Package Pattern: Separate auth logic into internal/auth package
    - Token Source Pattern: Return token along with its source for user feedback
    - Mock External Commands: Create temp bin dir with mock scripts for testing exec.Command
    - HTTP Response Validation: Check status codes and headers for rate limit detection
  - Gotchas encountered:
    - exec.Command mocking requires setting PATH to include temp bin directory
    - HTTP 401 and 403 have different meanings - 401 is invalid token, 403 may be rate limit
    - Import statements must appear at the top of files (not mid-function)

---## ✓ Iteration 2 - US-002: GitHub Authentication
*2026-01-20T08:44:06.871Z (271s)*

**Status:** Completed

**Notes:**
5.031Z (218s)*\n\n**Status:** Completed\n\n**Notes:**\n68→  - Gotchas encountered\\n    69→---\\n    70→```\\n\nIf you discovered a **reusable pattern**, also add it to the `## Codebase Patterns` section at the TOP of progress.md.\n\n## Stop Condition\n**IMPORTANT**: If the work is already complete (implemented in a previous iteration or already exists), verify it meets the acceptance criteria and signal completion immediately.\n\nWhen finished (or if already complete), signal completion with:\n

---
