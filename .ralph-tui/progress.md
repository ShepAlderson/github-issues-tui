# Ralph Progress Log

This file tracks progress across iterations. It's automatically updated
after each iteration and included in agent prompts for context.

## Codebase Patterns (Study These First)

### Project Structure
- `cmd/ghissues/main.go` - Main entry point, minimal - just calls `cmd.Execute()`
- `internal/config/` - Configuration types and TOML file handling
- `internal/setup/` - Interactive setup prompts using charmbracelet/huh
- `internal/cmd/` - Cobra CLI commands (root command and subcommands)

### Configuration Pattern
- Config struct lives in `internal/config/config.go`
- Use `config.DefaultConfigPath()` to get `~/.config/ghissues/config.toml`
- Config saved with 0600 permissions for security (tokens may be stored)
- TOML format for human-readable configuration

### CLI Command Pattern (Cobra)
- Root command created by `NewRootCmd()` function
- Subcommands added via `rootCmd.AddCommand()`
- For testable output, use `cmd.OutOrStdout()` not `fmt.Println()`
- Global state (like configPath, dbPath) exposed via getter/setter for testing
- Use `PersistentFlags()` for flags available to all subcommands (like --db)
- Use `PersistentPreRunE` to set global state from flags before RunE executes

### Database Path Pattern
- `internal/db/path.go` - Database path resolution with precedence: flag -> config -> default
- Default database path is `.ghissues.db` in current working directory
- Config can override via `[database] path = "..."` in TOML
- `--db` flag takes highest precedence
- `EnsureDBPath()` creates parent directories and validates writability
- `IsPathWritable()` uses temp file creation to verify write access

### Testing Pattern
- Tests use `t.TempDir()` for isolated file system tests
- Use `defer SetConfigPath("")` to reset global state after tests
- Interactive prompts (huh) can't be tested directly - use `RunSetupWithValues()` for programmatic setup
- For external dependencies (gh CLI, APIs), use package-level function variables that can be replaced in tests
- Example: `var ghCLITokenFunc = getTokenFromGhCLI` allows mocking in tests

### Authentication Pattern
- `internal/auth/auth.go` - Token retrieval with priority: env var -> config -> gh CLI
- `GetToken(cfg)` returns (token, source, error) - source indicates where token came from
- `ValidateToken(token)` validates against GitHub API with helpful error messages
- Sentinel errors (`ErrNoAuth`, `ErrInvalidToken`) allow callers to check error types with `errors.Is()`

---

## 2026-01-21 - US-001 First-Time Setup
- **What was implemented:**
  - Complete first-time setup flow with interactive prompts
  - Config package for TOML configuration file handling
  - Setup package with charmbracelet/huh for interactive forms
  - CLI commands using cobra (root and config subcommand)
  - Non-interactive setup via flags (--repo, --auth-method, --token)

- **Files changed:**
  - `cmd/ghissues/main.go` - Main entry point
  - `go.mod`, `go.sum` - Module definition and dependencies
  - `internal/config/config.go` - Config types, load/save, validation
  - `internal/config/config_test.go` - Config tests
  - `internal/setup/setup.go` - Interactive and programmatic setup
  - `internal/setup/setup_test.go` - Setup tests
  - `internal/cmd/root.go` - Root and config CLI commands
  - `internal/cmd/root_test.go` - CLI tests

- **Learnings:**
  - **Patterns discovered:**
    - Separating interactive setup (`RunInteractiveSetup`) from programmatic setup (`RunSetupWithValues`) enables testability
    - Use `cmd.OutOrStdout()` in cobra commands for testable output instead of `fmt.Println()`
    - Global package variables (configPath) need getter/setter functions for testing
  - **Gotchas encountered:**
    - `huh` forms require TTY and fail in test environments - provide programmatic alternative
    - When testing cobra subcommands, use `rootCmd.SetArgs([]string{"subcommand", "--flag", "value"})` then `rootCmd.Execute()` - not direct subcommand execution
    - `go mod tidy` may remove indirect dependencies needed by tests - run after adding test imports

---

## 2026-01-21 - US-002 GitHub Authentication
- **What was implemented:**
  - Auth package with `GetToken()` function supporting three authentication methods in priority order:
    1. GITHUB_TOKEN environment variable
    2. Config file token (when auth.method is "token")
    3. GitHub CLI (`gh auth token`)
  - `ValidateToken()` function that validates tokens against GitHub API
  - Clear, helpful error messages for authentication failures
  - TokenSource type to indicate where the token was retrieved from

- **Files changed:**
  - `internal/auth/auth.go` - Token retrieval and validation logic
  - `internal/auth/auth_test.go` - Comprehensive tests with mocked gh CLI

- **Learnings:**
  - **Patterns discovered:**
    - Use package-level function variables (`var ghCLITokenFunc = getTokenFromGhCLI`) to enable mocking external commands in tests
    - Return source information (TokenSource) alongside the token so callers know where auth came from
    - Use sentinel errors (`ErrNoAuth`, `ErrInvalidToken`) with `fmt.Errorf("%w: ...")` to allow `errors.Is()` checks
  - **Gotchas encountered:**
    - When the gh CLI is installed and authenticated on the dev machine, tests that expect "no auth available" will fail - must mock the gh CLI function
    - Use test helper functions (`mockGhCLI`, `mockGhCLIUnavailable`) that return cleanup functions for consistent test isolation

---

## 2026-01-21 - US-004 Database Storage Location
- **What was implemented:**
  - Database path configuration with three-level precedence: `--db` flag > config file > default
  - Default location is `.ghissues.db` in current working directory
  - `DatabaseConfig` struct added to config with TOML support
  - `internal/db/path.go` package for path resolution and validation
  - Parent directory creation when using custom paths
  - Clear error messages when paths are not writable

- **Files changed:**
  - `internal/config/config.go` - Added `DatabaseConfig` struct to `Config`
  - `internal/config/config_test.go` - Tests for database config loading/saving
  - `internal/db/path.go` - New package for database path resolution
  - `internal/db/path_test.go` - Comprehensive tests for path resolution
  - `internal/cmd/root.go` - Added `--db` flag, `SetDBPath()`/`GetDBPath()` functions
  - `internal/cmd/root_test.go` - Tests for `--db` flag behavior and precedence

- **Learnings:**
  - **Patterns discovered:**
    - Use `PersistentPreRunE` to set global state from flags before `RunE` executes
    - Use `PersistentFlags()` for flags that should be available to all subcommands
    - Test writability by attempting to create a temp file, not just checking permissions
  - **Gotchas encountered:**
    - `os.Getuid() == 0` check needed to skip writability tests when running as root
    - When testing path validation, use `/root/...` paths on Unix (generally unwritable)
    - Relative paths like `.ghissues.db` have parent dir `.` which requires special handling

---
