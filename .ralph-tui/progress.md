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

**Database Path Pattern**: Use a dedicated `internal/db` package that:
- Provides `DefaultPath()` returning `.ghissues.db` in current directory
- `GetPath(flag string, cfg *Config)` for priority: flag > config > default
- `EnsureDir(path string)` creates parent directories with `os.MkdirAll`
- `IsWritable(path string)` tests writability and returns descriptive error

**GitHub API Pagination Pattern**: Fetch paginated data by:
- Using `per_page=100` to maximize items per request
- Parsing `Link` header with `rel="next"` for next page URL
- Extracting page number from `page=N` query parameter
- Continuing until `rel="next"` is no longer present

**SQLite Database Pattern**: Use `modernc.org/sqlite` driver with:
- `PRAGMA foreign_keys = ON` to enable foreign key constraints
- `ON CONFLICT DO UPDATE` for UPSERT operations (idempotent inserts)
- Separate tables for related entities (labels, assignees, comments)
- Cascade delete via `ON DELETE CASCADE` foreign key constraint

**Context Cancellation Pattern**: Handle graceful shutdown by:
- Using `context.WithCancel()` for cancellable operations
- Checking `ctx.Done()` in loops to detect cancellation
- Setting up signal handler for SIGINT/SIGTERM
- Returning early with cleanup message on cancellation

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

---

## ✓ Iteration 2 - US-002: GitHub Authentication
*2026-01-20T08:44:06.871Z (271s)*

**Status:** Completed

---

## 2026-01-20 - US-004
- What was implemented:
  - Added Database struct with Path field to Config struct
  - Created internal/db package with GetPath, EnsureDir, IsWritable functions
  - Added --db flag to CLI using standard library flag package
  - Database path priority: --db flag > config file > default (.ghissues.db)
  - Parent directories are automatically created with os.MkdirAll
  - Clear error message if path is not writable using custom PathError type

- Files changed:
  - internal/config/config.go (added Database struct with Path field)
  - internal/db/db.go (new file with database path utilities)
  - internal/db/db_test.go (11 tests for path handling)
  - cmd/ghissues/main.go (added --db flag and db path validation)

- **Learnings:**
  - Patterns discovered:
    - Database Path Pattern: Separate db utilities into internal/db package
    - Priority-based config: flag > config file > default values
    - Path writability testing: create temp file in parent directory
    - Custom error types: PathError implements error interface with Unwrap
  - Gotchas encountered:
    - Return statement with error type must return nil, not empty string
    - `filepath.Dir("test.db")` returns ".", need special handling for current dir
    - `os.Chmod` required to restore permissions after read-only directory test## ✓ Iteration 3 - US-004: Database Storage Location
*2026-01-20T08:46:59.333Z (171s)*

**Status:** Completed

## 2026-01-20 - US-003
- What was implemented:
  - Added Issue, Comment, Label structs to internal/github package
  - Created FetchIssues() with automatic pagination (handles Link header)
  - Created FetchComments() with automatic pagination
  - Added internal/db/ops.go with Open(), UpsertIssue(), UpsertComment(), InsertLabel(), InsertAssignee()
  - Created database schema with foreign key constraints and cascade delete
  - Added `ghissues sync` subcommand to CLI
  - Implemented progress bar showing issues fetched / total
  - Added Ctrl+C handling with context cancellation for graceful shutdown
  - Issues stored with number, title, body, author, created/updated dates, comment count, labels, assignees
  - Comments fetched and stored for each issue

- Files changed:
  - internal/github/client.go (added Issue, Comment, Label structs, FetchIssues, FetchComments)
  - internal/github/issue_test.go (14 tests for issue/comment structures)
  - internal/github/fetch_test.go (7 tests for fetching with pagination)
  - internal/db/ops.go (new file with database operations)
  - internal/db/ops_test.go (14 tests for database operations)
  - cmd/ghissues/sync.go (new file with sync command and progress bar)
  - cmd/ghissues/sync_test.go (11 tests for progress bar and sync progress)
  - cmd/ghissues/main.go (added sync subcommand handler)
  - go.mod, go.sum (added modernc.org/sqlite dependency)

- **Learnings:**
  - Patterns discovered:
    - GitHub API Pagination: Parse Link header for next page URL
    - SQLite foreign keys disabled by default, need `PRAGMA foreign_keys = ON`
    - UPSERT with `ON CONFLICT DO UPDATE` for idempotent database operations
    - Context cancellation allows graceful shutdown of long-running operations
    - Progress bar updates should handle unknown total count
  - Gotchas encountered:
    - SQLite foreign key constraints fail if parent row doesn't exist
    - Link header format requires parsing URL and extracting page parameter
    - `X-Total-Count` header provides total, Link header provides pagination
    - Progress bar needs \r (carriage return) to update in place

---

## ✓ Iteration 4 - US-003: Initial Issue Sync
**Status:** Completed\n"}],"model":"hf:MiniMaxAI/MiniMax-M2.1","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":41020,"output_tokens":0,"cache_creation_input_tokens":0,"cache_read_input_tokens":0},"context_management":null},"parent_tool_use_id":null,"session_id":"bf326f68-c80a-406a-976b-25076c9bf34e","uuid":"8714aa34-993d-4393-b4b1-2db43025b48b"}
{"type":"assistant","message":{"id":"msg_ce244bd4-361d-40fc-949c-20e66dd79b6e","type":"message","role":"assistant","content":[{"type":"text","text":"\n\n

---
