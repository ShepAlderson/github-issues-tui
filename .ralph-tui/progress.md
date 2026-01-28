# Ralph Progress Log

This file tracks progress across iterations. It's automatically updated
after each iteration and included in agent prompts for context.

## Codebase Patterns (Study These First)

### TDD Pattern
1. Write tests first for each function/feature
2. See tests fail (confirming tests work)
3. Implement minimal code to make tests pass
4. Refactor while keeping tests green

### Bubbletea TUI Testing
- Use `tea.KeyMsg{Type: tea.KeyEnter}` to simulate keypresses
- Test state transitions by checking model fields after Update
- View tests verify UI text content contains expected strings

### Config File Security
- Use `0600` permissions (owner read/write only) for sensitive files
- Implement via `os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)`

### Home Directory Handling
- Use `os.UserHomeDir()` for cross-platform home directory detection
- Fallback gracefully if not available

### Authentication Resolution Pattern
- Define priority order clearly (env var -> config file -> external CLI)
- Each source should return (value, found) tuple pattern
- Provide actionable error messages with all configuration options
- Use `exec.LookPath()` before attempting to execute external commands
- Store token source information for debugging authentication issues

### Database Path Resolution Pattern
- Priority order: CLI flag -> config file -> default path
- Test writability by creating and removing a temp file in the directory
- Use `filepath.Abs()` to normalize relative paths
- Create parent directories with `os.MkdirAll(dir, 0755)` before checking writability
- Provide clear error messages with override options

---

## 2026-01-28 - US-001
- Implemented first-time setup wizard with interactive TUI
- Files changed: cmd/ghissues/main.go, internal/config/*.go
- **Learnings:**
  - Bubbletea input fields require Focus() to capture keyboard input
  - TOML decoding with `toml.DecodeFile` returns (MetaData, error) - handle properly
  - File permissions 0600 is octal, not decimal (use leading 0)
  - `os.UserHomeDir()` is the idiomatic way to get home directory in Go
  - Bubbletea tests: manually trigger state changes since we can't simulate full TUI flow
  - Repository validation: use simple parsing + regex for GitHub username/repo validation
  - Setup flow: state machine pattern works well for multi-step wizards
---

## 2026-01-28 - US-004
- Implemented database storage location configuration with priority-based resolution
- Files changed:
  - internal/database/path.go (new)
  - internal/database/path_test.go (new)
  - cmd/ghissues/main.go (updated with --db flag)
- **Learnings:**
  - `flag` package usage: Define flags with `flag.StringVar()` before `flag.Parse()`
  - Flag package intercepts --help, use custom help subcommand instead
  - Database writability test: Create temp file in directory, then remove it
  - Path resolution pattern: flag > config > default with clear priority
  - `filepath.Abs()` resolves relative paths to absolute paths
  - `os.Chmod()` for read-only testing may not work on all platforms - use `t.Skip()`
---

## 2026-01-28 - US-002
- Implemented GitHub authentication resolver with priority-based token resolution
- Files changed: internal/github/auth.go, internal/github/auth_test.go
- **Learnings:**
  - Authentication priority order: GITHUB_TOKEN env var -> config file -> gh CLI
  - Use `exec.LookPath("gh")` to check if gh CLI is available before executing
  - `os/exec.Command().Output()` returns combined stdout/stderr, use separately if needed
  - Multi-source configuration requires clear error messages when none are available
  - Testing with PATH modification: save/restore PATH to simulate missing commands
  - Writing tests first ensures edge cases (empty strings, missing files) are handled
---

## ✓ Iteration 1 - US-001: First-Time Setup
*2026-01-28T07:53:50.015Z (433s)*

**Status:** Completed

**Notes:**
68→  - Gotchas encountered\n    69→---\n    70→```\n    71→\n    72→If you discovered a **reusable pattern**, also add it to the `## Codebase Patterns` section at the TOP of progress.md.\n    73→\n    74→## Stop Condition\n    75→**IMPORTANT**: If the work is already complete (implemented in a previous iteration or already exists), verify it meets the acceptance criteria and signal completion immediately.\n    76→\n    77→When finished (or if already complete), signal completion with:\n    78→

---
## ✓ Iteration 2 - US-002: GitHub Authentication
*2026-01-28T08:03:31.153Z (580s)*

**Status:** Completed

**Notes:**
n 'ghissues config' to save a token to your config file\n3. Login with 'gh auth login' to use gh CLI authentication\n```\n\n**Token Validation (AC met: ✅)**\n- `ValidateToken()` function created (empty check)\n- Framework for API validation ready for future stories\n\n**Config File Security (AC met: ✅)**\n- Already implemented in US-001 with 0600 permissions\n\n**New Pattern Added to Codebase:**\n- Authentication Resolution Pattern with priority-based resolution and actionable error messages\n\n

---
