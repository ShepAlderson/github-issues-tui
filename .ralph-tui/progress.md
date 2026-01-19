# Ralph Progress Log

This file tracks progress across iterations. It's automatically updated
after each iteration and included in agent prompts for context.

## Codebase Patterns (Study These First)

### Testing Pattern
- Use dependency injection for testability (e.g., `RunInteractiveSetupWithReader` allows passing custom io.Reader for tests)
- Write tests first (TDD), see them fail, then implement
- Use table-driven tests for validation logic with multiple cases
- Use httptest.NewServer for testing HTTP clients (GitHub API)

### Configuration Pattern
- Config stored in `~/.config/ghissues/config.toml` (follows XDG conventions)
- Use 0600 permissions for security when storing sensitive data (tokens)
- Validate config on load with clear error messages
- Support multiple auth methods: env variable, token in config, gh CLI

### CLI Pattern
- Subcommands implemented via simple argument parsing
- Provide --help and --version flags
- Interactive setup on first run with validation loops for invalid input
- Clear error messages with actionable guidance

### Authentication Pattern
- Multiple auth methods with fallback order: env variable -> config file -> external CLI
- Switch-based method dispatch for clean authentication selection
- Token validation via API call with 10-second timeout
- Error messages include actionable next steps (e.g., "run 'ghissues config'")
- Use `exec.Command` for external CLI integration (e.g., `gh auth token`)

### Database Pattern
- Three-tier precedence: CLI flag > config file > default value
- Always resolve relative paths to absolute paths for consistency
- Test writability by attempting to create a test file (more reliable than permission checks)
- Initialize idempotently (succeed whether target exists or not)
- Create parent directories automatically with `os.MkdirAll`
- Use transactions for multi-step database operations
- Use `INSERT OR REPLACE` for upsert operations in SQLite
- **SQLite Driver**: Use `modernc.org/sqlite` (pure Go) instead of libsql for better test compatibility
  - Driver name: "sqlite" (not "libsql")
  - Connection string: direct file path (not "file:" prefix)
  - Import driver in all files that open database connections: `_ "modernc.org/sqlite"`

### API Client Pattern
- Use context.Context for cancellation support
- Implement pagination automatically using Link headers
- Parse response headers for pagination metadata
- Use httptest.NewServer for testing API clients
- Set appropriate headers (Authorization, Accept) for API requests
- Respect per_page limits (GitHub API: max 100)

### Progress Tracking Pattern
- Use progressbar library for visual feedback
- Set clear descriptions and show counts
- Clear progress bar on finish for clean output
- Update progress after each unit of work completes

### Signal Handling Pattern
- Use signal.Notify for Ctrl+C handling
- Create cancellable context for graceful shutdown
- Handle cancellation in goroutine to avoid blocking main logic
- Clean up resources (close database, etc.) with defer

### TUI Pattern (Bubbletea)
- Use Bubbletea's Model-Update-View architecture for interactive UIs
- Model holds state (issues, cursor position, dimensions)
- Update handles messages (key presses, window resizes) and returns updated model
- View renders the current state to a string
- Use lipgloss for styling (colors, bold, etc.)
- Handle both vim keys (j/k) and arrow keys for navigation
- Use tea.WithAltScreen() to take over terminal and restore on exit
- Implement boundary checks for navigation (don't go past first/last item)
- Defer database Close() to ensure cleanup even if TUI exits early

---

## [2026-01-19] - US-003 - Initial Issue Sync

### What was implemented
- GitHub API client with automatic pagination handling for issues and comments
- SQLite database schema for storing issues, comments, labels, and assignees
- Sync orchestration with progress tracking and graceful cancellation
- CLI integration with 'ghissues sync' subcommand
- Comprehensive test suite for all sync functionality

### Files changed
- `internal/sync/types.go` - Data structures for issues, comments, and GitHub API responses
- `internal/sync/github.go` - GitHub API client with pagination support
- `internal/sync/github_test.go` - Tests for GitHub API client
- `internal/sync/storage.go` - SQLite database operations for issues and comments
- `internal/sync/storage_test.go` - Tests for database storage
- `internal/sync/sync.go` - Sync orchestration with progress bar and cancellation
- `internal/sync/sync_test.go` - Integration tests for sync process
- `cmd/ghissues/main.go` - Added 'sync' subcommand and runSync function
- `go.mod`, `go.sum` - Added dependencies: modernc.org/sqlite, progressbar/v3

### Learnings

#### Patterns discovered
1. **HTTP pagination with Link headers**: GitHub returns pagination info in Link header with format `<url>; rel="next"`. Check for `rel="next"` to determine if more pages exist.
2. **Progressive data fetching**: Fetch issues first (lightweight), then fetch comments only for issues that have them (reduces API calls).
3. **Database transactions for consistency**: Use transactions when updating multiple related tables (issues, labels, assignees) to ensure data consistency.
4. **Context propagation for cancellation**: Pass context.Context through all layers (API client, storage, orchestration) to enable Ctrl+C cancellation at any point.
5. **Signal handling pattern**: Use `signal.Notify` with buffered channel and goroutine to handle Ctrl+C without blocking main logic.
6. **Defer for cleanup**: Use `defer syncer.Close()` to ensure database connections are closed even if sync fails.
7. **SQLite upsert with INSERT OR REPLACE**: Simplifies update logic - just insert with same primary key to update existing rows.
8. **Test with httptest.NewServer**: Mock GitHub API responses for fast, reliable tests without network dependencies.

#### Gotchas encountered
1. **libsql vs modernc.org/sqlite**: Initial attempt to use libsql driver failed with "no sqlite driver present" error. Switched to modernc.org/sqlite (pure Go) which worked immediately with `sql.Open("sqlite", path)`.
2. **Driver import in tests**: Must import SQLite driver with `_ "modernc.org/sqlite"` in test files, not just main code, or tests fail with driver errors.
3. **Pull requests in issues API**: GitHub's issues API returns both issues and pull requests. Need to filter out PRs (they have a `pull_request` field).
4. **Progress bar options**: Not all progressbar options are available - `OptionShowIOs()` doesn't exist. Use simpler options like `OptionShowCount()` and `OptionSetPredictTime()`.
5. **Database connection string**: SQLite driver expects just the file path (`dbPath`), not a URI format like `file:` prefix that libsql expected.
6. **Foreign key cascades**: Use `ON DELETE CASCADE` for foreign keys so deleting an issue automatically cleans up related labels, assignees, and comments.
7. **Pagination detection**: Must check Link header for `rel="next"`, not just parse URL. Empty response with no Link header means no more pages.
8. **Context cancellation timing**: When context is cancelled, current in-flight HTTP request completes, but next iteration respects cancellation via `select` on `ctx.Done()`.

#### Architecture decisions
1. Created separate `sync` package to encapsulate all synchronization logic (GitHub API, storage, orchestration).
2. Split sync into three layers: GitHubClient (API calls), IssueStore (database), Syncer (orchestration).
3. Used standard `database/sql` interface to allow easy driver swapping if needed later.
4. Database schema uses normalized design with separate tables for issues, comments, labels, assignees for flexibility.
5. Progress bar shows issue count, not comment count, since issue fetching is the main bottleneck.
6. Sync is a separate subcommand (`ghissues sync`) rather than automatic, giving user control over when to sync.
7. Used modernc.org/sqlite (pure Go) instead of cgo-based drivers for easier cross-platform builds.

---
## ✓ Iteration 1 - US-003: Initial Issue Sync
*2026-01-19T22:33:36.500Z (287s)*

**Status:** Completed

**Notes:**
Implemented in sync.go:97-111 and github.go:116-178\n- ✅ Sync can be cancelled with Ctrl+C gracefully - Implemented in sync.go:39-52 with signal handling\n\nAll acceptance criteria are met! The implementation was already complete from a previous iteration. My work consisted of:\n1. Fixing the SQLite driver issue (migrating from libsql to modernc.org/sqlite)\n2. Cleaning up unused code (parseLinkHeader, unused test helpers)\n3. Ensuring all tests pass\n4. Documenting learnings in progress.md\n\n

---

## [2026-01-19] - US-005 - Issue List View

### What was implemented
- TUI-based issue list view using Bubbletea framework
- Issue loading from database with LoadIssues() method
- Configurable column display with defaults
- Navigation with vim keys (j/k) and arrow keys
- Issue count in status bar
- Selected issue highlighting

### Files changed
- `internal/tui/model.go` - Bubbletea model with rendering logic
- `internal/tui/model_test.go` - Tests for navigation, selection, and edge cases
- `internal/sync/storage.go` - Added LoadIssues() to query all issues
- `internal/sync/storage_test.go` - Tests for LoadIssues()
- `internal/config/config.go` - Added GetDisplayColumns() helper
- `internal/config/config_test.go` - Tests for column configuration
- `cmd/ghissues/main.go` - Integrated TUI as main interface
- `go.mod`, `go.sum` - Added bubbletea and lipgloss dependencies

### Learnings

#### Patterns discovered
1. **Bubbletea Model-Update-View architecture**: Separates state (Model) from event handling (Update) and rendering (View). Makes TUI testable without actual terminal.
2. **Testing TUI without rendering**: Test Update() method directly by passing KeyMsg events and asserting on returned model state. No need to test View() output in unit tests.
3. **Column-based rendering**: Pre-calculate column widths and use helper functions (padRight, truncate) for consistent alignment.
4. **Relative time formatting**: Implement human-readable time display (e.g., "2 hours ago") for better UX than raw timestamps.
5. **Default config values**: Use helper function (GetDisplayColumns) to provide sensible defaults when config is empty/missing.
6. **Database query with joins**: Load related data (labels, assignees) in separate queries after main query to avoid complex joins.
7. **Cursor boundary checking**: Always check cursor against len(issues)-1 to prevent index out of bounds.

#### Gotchas encountered
1. **Empty slices in Go**: `len(nil) == 0` is true, so can check `len(slice) == 0` instead of `slice == nil || len(slice) == 0`.
2. **Bubbletea message types**: Must type-assert tea.KeyMsg and check both Type (for special keys like KeyDown) and Runes (for character keys like 'j').
3. **Lipgloss color codes**: Use ANSI color codes as strings (e.g., "39" for default, "170" for purple). Colors are terminal-dependent.
4. **Database connection lifecycle**: Must defer store.Close() in main.go to prevent connection leaks, even though TUI may exit via tea.Quit.
5. **SQLite query with ORDER BY**: Default sort is by ROWID if no ORDER BY specified. Must explicitly ORDER BY updated_at DESC for most-recent-first.
6. **Testing KeyMsg**: Create tea.KeyMsg with Type field for special keys or Runes field for characters. Can't test both in same message.
7. **Window resize handling**: Must handle tea.WindowSizeMsg to track terminal dimensions, even if not using them yet (future use).

#### Architecture decisions
1. Created separate `tui` package for all TUI-related code, keeping it isolated from business logic.
2. Model takes issues as constructor parameter rather than loading internally - allows for future filtering/sorting without TUI knowing about database.
3. Column configuration in DisplayConfig struct mirrors expected usage pattern (user may want to customize).
4. LoadIssues() returns slice ordered by updated_at DESC (most recent first) as sensible default for issue list.
5. Status bar shows "Issue X of Y" for clarity on position in list.
6. Footer shows context-sensitive keybindings (only navigation keys for now, will expand in future).
7. Used tea.WithAltScreen() to take over full terminal, allowing clean exit back to shell.

---
## ✓ Iteration 2 - US-005: Issue List View
*2026-01-19T22:39:32.869Z (355s)*

**Status:** Completed

**Notes:**
Style.\n\n✅ **Vim keys (j/k) and arrow keys for navigation** - Implemented in internal/tui/model.go:40-69. Both 'j'/'k' and KeyDown/KeyUp are handled with proper boundary checking. Tested in internal/tui/model_test.go:14-74.\n\n✅ **Issue count shown in status area** - Implemented in internal/tui/model.go:157-164. Status bar displays \"Issue X of Y\" format showing current position and total count.\n\nAll acceptance criteria have been successfully implemented with comprehensive test coverage!\n\n

---
