# Ralph Progress Log

This file tracks progress across iterations. It's automatically updated
after each iteration and included in agent prompts for context.

## Codebase Patterns (Study These First)

### TDD Approach
- Write failing tests first, then implement code to make them pass
- Test interactive prompts with buffer-based I/O for deterministic tests
- Use table-driven tests for comprehensive coverage of edge cases

### Package Structure
```
/
├── main.go                 # CLI entry point
├── main_test.go           # Integration tests
└── internal/
    ├── cmd/               # Command implementations
    │   ├── config.go      # Config command
    │   └── config_test.go
    ├── config/            # Configuration logic
    │   ├── config.go      # Load/save config
    │   └── config_test.go
    └── prompt/            # Interactive prompts
        ├── prompt.go      # User input handling
        └── prompt_test.go
```

### Input Handling Pattern
When working with interactive prompts and `bufio.Reader`:
- Create a SINGLE `bufio.NewReader(input)` in the entry function
- Pass the `*bufio.Reader` to all prompt functions
- This ensures the reader state is maintained across multiple reads
- Multiple `bufio.NewReader()` on the same input will cause EOF errors after the first read

---

## [2026-01-20] - US-001 - First-Time Setup

**What was implemented:**
- Interactive configuration setup for first-time users
- Config command (`ghissues config`) to re-run setup
- Configuration file management (`~/.config/ghissues/config.toml`)
- Automatic setup prompting when no config exists
- Support for multiple authentication methods

**Files changed:**
- Created `main.go` - CLI entry point with command routing
- Created `main_test.go` - Integration tests for main flow
- Created `internal/cmd/config.go` - Config command implementation
- Created `internal/cmd/config_test.go` - Config command tests
- Created `internal/config/config.go` - Config file I/O operations
- Created `internal/config/config_test.go` - Config tests
- Created `internal/prompt/prompt.go` - Interactive prompt handling
- Created `internal/prompt/prompt_test.go` - Prompt tests
- Created `internal/prompt/integration_test.go` - Integration tests
- Created `go.mod` - Go module initialization with go-toml dependency
- Updated `.ralph-tui/progress.md` - Documented patterns and learnings

**Learnings:**
- **Pattern Discovered:** When using `bufio.NewReader` with the same `io.Reader`, create it once and reuse it. Creating multiple `bufio.NewReader` instances on the same reader causes state issues and EOF errors.
- **Gotcha Encountered:** Test input for interactive prompts must include newlines (`\n`) to simulate user pressing Enter. Without them, `ReadString('\n')` will hang or return EOF.
- **TDD Benefit:** Writing tests first forced cleaner API design and revealed edge cases early (e.g., invalid repository format, empty tokens).
- **Package Layout:** The `internal/` package structure keeps implementation details private while allowing focused unit tests for each component.

**Testing Notes:**
- All tests use buffer-based I/O (bytes.Buffer, strings.Reader) for deterministic behavior
- Integration tests verify end-to-end flows
- Test coverage includes success paths, error paths, and edge cases
- Verified with: `go test ./...`, `golangci-lint run`, `go vet`, `gofmt`

---
## ✓ Iteration 1 - US-001: First-Time Setup
*2026-01-20T21:07:35.307Z (838s)*

**Status:** Completed

**Notes:**
ion."}],"model":"hf:moonshotai/Kimi-K2-Thinking","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":61157,"output_tokens":0,"cache_creation_input_tokens":0,"cache_read_input_tokens":0},"context_management":null},"parent_tool_use_id":null,"session_id":"6dfd7e38-e6d8-431c-b3a2-2b518c344a20","uuid":"b7718525-087d-4956-b955-80275d682537"}
{"type":"assistant","message":{"id":"msg_c756af15-8241-43d3-818e-89b2c191e4c3","type":"message","role":"assistant","content":[{"type":"text","text":"

---
---
## [2026-01-20] - US-002 - GitHub Authentication

**What was implemented:**
- GitHub token authentication with proper priority order: GITHUB_TOKEN env -> config file -> gh CLI
- Token validation on startup with clear error messages
- New `internal/auth` package for authentication logic
- Integration tests for authentication flow
- Updated main.go to use the authentication system

**Files changed:**
- Created `internal/auth/auth.go` - Authentication logic (GetGitHubToken, ValidateToken)
- Created `internal/auth/auth_test.go` - Authentication tests
- Updated `main.go` - Integrate authentication system with token validation
- Updated `main_test.go` - Added integration tests for authentication flow

**Learnings:**
- **Pattern Discovered:** Dependency injection for testability - passing config struct to GetGitHubToken() instead of having it read from file internally makes testing easier and more reliable
- **Gotcha Encountered:** When creating TOML config files in tests, the format must match exactly what the parser expects. Using `[config]` section when token is at root level causes parsing to fail
- **Design Decision:** Made ValidateToken() a simple non-empty check rather than full API validation to avoid external dependencies during startup. Real API validation would happen on first actual API call
- **Testing Strategy:** Integration tests can't easily mock internal package variables. Testing gh CLI token fallback requires either exposing test hooks or testing at the unit level in the auth package

**Testing Notes:**
- All authentication methods tested independently (env, config, gh CLI)
- Priority order verified with comprehensive table-driven tests
- Integration tests verify end-to-end authentication flow
- Clear error message when no authentication found
- Verified with: `go test ./...`, all tests passing

---
## ✓ Iteration 2 - US-002: GitHub Authentication
*2026-01-20T21:14:43.218Z (427s)*

**Status:** Completed

**Notes:**
68→  - Gotchas encountered\n    69→---\n    70→```\n    71→\n    72→If you discovered a **reusable pattern**, also add it to the `## Codebase Patterns` section at the TOP of progress.md.\n    73→\n    74→## Stop Condition\n    75→**IMPORTANT**: If the work is already complete (implemented in a previous iteration or already exists), verify it meets the acceptance criteria and signal completion immediately.\n    76→\n    77→When finished (or if already complete), signal completion with:\n    78→

---

---

## [2026-01-20] - US-004 - Database Storage Location

**What was implemented:**
- Flexible database storage location with three-tier priority system:
  1. Command-line flag (`--db /path/to/db`) - highest priority
  2. Config file (`database.path` in TOML) - medium priority  
  3. Default (`.ghissues.db` in current directory) - lowest priority
- Automatic parent directory creation with proper error handling
- Path writability validation with clear error messages
- Database location displayed in startup output for transparency

**Files changed:**
- Modified `main.go` - Added flag parsing, database path resolution, and validation
- Modified `main_test.go` - Added comprehensive tests for all priority levels

**Learnings:**
- **Pattern Discovered:** Priority resolution pattern - when multiple sources can provide a value, establish clear precedence rules (flag > config > default) and test each level independently
- **Gotcha Encountered:** macOS path resolution differences - `/var/folders` symlinks to `/private/var/folders`, causing path comparison failures in tests. Solution: avoid strict path equality checks when testing file operations that resolve symlinks
- **Design Decision:** Database file is not created during path validation, only parent directories. This keeps the validation logic focused and avoids side effects before the actual database initialization
- **Testing Strategy:** Table-driven tests work well for testing priority hierarchies. Each test case should verify one priority level in isolation to ensure the hierarchy works correctly

---
## ✓ Iteration 3 - US-004: Database Storage Location
*2026-01-20T21:19:25.245Z (281s)*

**Status:** Completed

---

## [2026-01-20] - US-003 - Initial Issue Sync

**What was implemented:**
- Full sync command (`ghissues sync`) that fetches all open issues from GitHub
- Progress bar showing real-time progress of issues fetched and processed
- Automatic pagination handling for repositories with many issues
- Local libsql database storage at configured path with proper schema
- Complete issue data storage including: number, title, body, author, dates, comment count, labels, and assignees
- Comment fetching and storage for each issue
- Graceful cancellation support with Ctrl+C signal handling
- Integration with existing authentication and database path configuration

**Files changed:**
- Created `internal/db/db.go` - Database operations with libsql driver
- Created `internal/db/db_test.go` - Comprehensive database tests
- Created `internal/github/client.go` - GitHub API client with pagination
- Created `internal/github/client_test.go` - API client tests with mock server
- Created `internal/cmd/sync.go` - Sync command implementation with progress bar
- Created `internal/cmd/sync_test.go` - Sync command tests
- Modified `main.go` - Added sync command routing and integration
- Modified `main_test.go` - Added integration tests for sync command
- Updated dependencies: added libsql driver, sqlite3 driver, and progress bar library

**Learnings:**
- **Pattern Discovered:** Database schema design - store complex data (arrays/objects) as JSON strings in SQLite columns for flexibility. Use foreign keys with CASCADE DELETE for data integrity.
- **Gotcha Encountered:** libsql driver requires URL scheme prefixes (file:, http:, etc.). Regular file paths need conversion. Also need sqlite3 driver for in-memory databases.
- **Design Decision:** Clear existing issues before sync to ensure fresh data. This prevents stale issues from accumulating and ensures the database reflects current GitHub state.
- **Testing Strategy:** Use httptest.NewServer() to mock GitHub API responses. This allows testing sync logic without external dependencies and enables testing error scenarios.
- **Progress UX:** Progress bars should show both progress and count (e.g., "Processing issues 5/42") for better user feedback during long operations.
- **Graceful Degradation:** When comment fetching fails for individual issues, log warnings but continue processing other issues. Don't fail the entire sync for one problematic issue.

**New Patterns Added to Codebase:**
- Database pattern: Store arrays as JSON, use foreign keys with CASCADE DELETE
- API client pattern: Use httptest.NewServer for mocking external APIs in tests
- Progress pattern: Use progressbar library with descriptive messages and error handling

---
