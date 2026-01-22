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
- Use `httptest.NewServer` to mock HTTP APIs in tests
- GitHub client has `SetBaseURL()` method for testing with mock servers

### GitHub API Pattern
- `internal/github/client.go` - GraphQL API client for fetching issues and comments
- Use GraphQL for efficient fetching with automatic pagination
- `FetchIssues()` and `FetchIssueComments()` handle pagination automatically
- Progress callbacks allow real-time progress reporting
- Context cancellation support for graceful shutdown

### Database Pattern
- `internal/db/store.go` - LibSQL/SQLite storage using `github.com/tursodatabase/go-libsql`
- Schema created on `NewStore()` - tables for issues, comments, metadata
- Store labels and assignees as JSON strings in issue table
- Use transactions for batch operations (`SaveIssues`, `SaveComments`)
- `defer func() { _ = tx.Rollback() }()` pattern for safe transaction cleanup

### Sync Pattern
- `internal/sync/sync.go` - Orchestrates fetching and storing
- Progress callback reports phase ("issues" or "comments") and counts
- Fetches all issues first, then comments for issues that have them
- Updates `last_sync` metadata after successful sync

### Authentication Pattern
- `internal/auth/auth.go` - Token retrieval with priority: env var -> config -> gh CLI
- `GetToken(cfg)` returns (token, source, error) - source indicates where token came from
- `ValidateToken(token)` validates against GitHub API with helpful error messages
- Sentinel errors (`ErrNoAuth`, `ErrInvalidToken`) allow callers to check error types with `errors.Is()`

### TUI Pattern
- `internal/tui/model.go` - Bubbletea Model for issue list view
- Model implements `tea.Model` interface: `Init()`, `Update()`, `View()`
- Use `tea.WithAltScreen()` for full-screen TUI that restores terminal on exit
- `SetDisableTUI(true)` package-level variable allows bypassing TUI in tests
- Test TUI by calling `Update()` with `tea.KeyMsg` directly - no TTY needed
- View() can be tested by setting window size and checking output strings

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

## 2026-01-21 - US-003 Initial Issue Sync
- **What was implemented:**
  - GitHub GraphQL API client for fetching issues and comments with automatic pagination
  - LibSQL database store with schema for issues, comments, labels, assignees
  - Sync command (`ghissues sync`) with progress bar and Ctrl+C cancellation
  - Progress callback reporting during fetch operations

- **Files changed:**
  - `internal/github/client.go` - GitHub GraphQL API client
  - `internal/github/client_test.go` - Tests with mock HTTP server
  - `internal/db/store.go` - Database store using libsql
  - `internal/db/store_test.go` - Database operation tests
  - `internal/sync/sync.go` - Sync orchestration logic
  - `internal/sync/sync_test.go` - Sync tests with mocked GitHub
  - `internal/cmd/sync.go` - CLI sync subcommand
  - `internal/cmd/sync_test.go` - CLI sync tests
  - `internal/cmd/root.go` - Added sync subcommand registration
  - `go.mod`, `go.sum` - Added libsql dependency

- **Learnings:**
  - **Patterns discovered:**
    - LibSQL driver requires executing schema statements one at a time (not batched in single Exec)
    - Use `Client.SetBaseURL()` method to allow test injection of mock server URLs
    - Store complex data (labels, assignees) as JSON strings in SQLite for simplicity
    - Use `signal.Notify` with context cancellation for graceful Ctrl+C handling
  - **Gotchas encountered:**
    - LibSQL multi-statement Exec fails silently - split into individual statements
    - In tests with `httptest.NewServer`, need to check for context cancellation in handler
    - Transaction rollback return values should be explicitly ignored with `_ = tx.Rollback()`
    - When gh CLI is installed locally, tests expecting "no auth" may pass due to fallback auth

---

## 2026-01-21 - US-005 Issue List View
- **What was implemented:**
  - TUI package using charmbracelet/bubbletea for the issue list view
  - Configurable columns with defaults: number, title, author, date, comments
  - DisplayConfig struct added to config for storing column configuration
  - Vim keys (j/k) and arrow keys for navigation
  - Issue count shown in status bar
  - Selected issue highlighting
  - Empty state handling when no issues are synced

- **Files changed:**
  - `internal/tui/model.go` - TUI model with bubbletea tea.Model implementation
  - `internal/tui/model_test.go` - Comprehensive tests for TUI navigation and rendering
  - `internal/config/config.go` - Added DisplayConfig struct, DefaultDisplayColumns(), ValidateDisplayColumn()
  - `internal/config/config_test.go` - Tests for display column configuration
  - `internal/cmd/root.go` - Integrated TUI startup, added SetDisableTUI() for testing
  - `internal/cmd/root_test.go` - Updated tests to use SetDisableTUI(true)
  - `go.mod`, `go.sum` - Added bubbletea and lipgloss as direct dependencies

- **Learnings:**
  - **Patterns discovered:**
    - Use `tea.WithAltScreen()` for full-screen TUI that restores terminal on exit
    - Test TUI models by calling Update() with tea.KeyMsg directly - no TTY needed
    - Use `SetDisableTUI(true)` pattern to skip TUI in tests (similar to other mock patterns)
    - Bubbletea Model.View() can be tested by setting window size and checking output strings
  - **Gotchas encountered:**
    - `tea.NewProgram().Run()` requires a TTY - tests fail with "could not open a new TTY"
    - Add package-level `disableTUI` variable with setter for testing to bypass TUI in tests
    - Remember to defer `SetDisableTUI(false)` to reset state after tests

---

## 2026-01-21 - US-006 Issue Sorting
- **What was implemented:**
  - Sorting functionality for the issue list view
  - Default sort: most recently updated first (updated date, descending)
  - Four sort options: updated date, created date, issue number, comment count
  - Keybindings: 's' cycles sort fields, 'S' (shift+s) toggles sort order
  - Status bar shows current sort field and direction with ↑/↓ indicators
  - Sort preference persisted to config file when changed

- **Files changed:**
  - `internal/config/config.go` - Added SortField, SortOrder types, validation functions, and DisplayConfig fields
  - `internal/config/config_test.go` - Tests for sort configuration
  - `internal/tui/model.go` - Added sorting logic, keybindings, status bar display, NewModelWithSort()
  - `internal/tui/model_test.go` - Comprehensive tests for all sort options and keybindings
  - `internal/cmd/root.go` - Load sort prefs from config, save when changed

- **Learnings:**
  - **Patterns discovered:**
    - Use `NewModelWithSort()` constructor variant to allow passing initial sort config from outside
    - Copy slice before sorting with `copy()` to avoid modifying original data
    - Use `sort.Slice()` with a comparison function that checks `sortOrder` to flip direction
    - Track state changes with a `sortChanged` bool to avoid unnecessary config saves
    - Return final model from `p.Run()` to extract state after TUI exits
  - **Gotchas encountered:**
    - When adding default sorting, existing tests that assumed original order will fail - must update to expect sorted order
    - Use `github.Issue.UpdatedAtTime()` / `CreatedAtTime()` helper methods to parse RFC3339 dates for comparison
    - Cursor should reset to 0 when sort changes to avoid confusing UX

---

## 2026-01-21 - US-007 Issue Detail View
- **What was implemented:**
  - Split-panel TUI layout with issue list on left and detail panel on right
  - Detail panel header shows: issue number, title, author, created/updated dates
  - Body rendered with glamour (charmbracelet markdown renderer)
  - Toggle between raw markdown and rendered view with 'm' key
  - Labels and assignees displayed when present
  - Scrollable detail panel with 'h' (up) and 'l' (down) keys
  - Enter key opens comments view (state tracking added)
  - Escape key exits comments view

- **Files changed:**
  - `internal/tui/model.go` - Added detail panel rendering, new state fields, key handlers
  - `internal/tui/model_test.go` - Added comprehensive tests for detail view functionality
  - `go.mod`, `go.sum` - Added glamour dependency

- **Learnings:**
  - **Patterns discovered:**
    - Use `glamour.NewTermRenderer()` with `WithAutoStyle()` and `WithWordWrap()` for terminal-aware markdown rendering
    - Use `lipgloss.Width()` to get visual width of strings with ANSI codes for proper padding
    - Store glamour renderer in Model struct to avoid recreating it on each render
    - Reset detail scroll offset when changing selected issue for better UX
  - **Gotchas encountered:**
    - When refactoring View() to use split panels, old column-based rendering methods become unused - clean them up
    - Glamour renderer width is set at creation time, not per-render - pass width to `WithWordWrap()` at initialization
    - Use `strings.TrimSpace()` on glamour output to remove trailing newlines

---
## ✓ Iteration 7 - US-007: Issue Detail View
*2026-01-21T23:02:27.476Z (302s)*

**Status:** Completed

**Notes:**
ey toggles `rawMarkdown` boolean and `renderBody()` renders accordingly\n- [x] **Labels and assignees displayed if present** - The detail panel conditionally shows labels and assignees when present\n- [x] **Scrollable if content exceeds panel height** - The 'h' and 'l' keys scroll the detail panel, with `detailScrollY` tracking offset\n- [x] **Enter on issue list opens dedicated comments view** - Enter key sets `inCommentsView = true`, Escape returns\n\nAll acceptance criteria are satisfied.\n\n

---

## 2026-01-21 - US-008 Comments View
- **What was implemented:**
  - Full-screen comments drill-down view that replaces the main interface
  - Issue title/number displayed as header in comments view
  - Comments displayed chronologically with author and date for each comment
  - Comment body with markdown rendering (reuses existing glamour renderer)
  - Toggle markdown rendering with 'm' key (shares state with detail panel)
  - Scrollable comment list with 'h' (scroll up) and 'l' (scroll down) keys
  - Esc or 'q' returns to issue list view

- **Files changed:**
  - `internal/tui/model.go` - Added comments field, commentsScrollY, SetComments/GetComments/CommentsScrollOffset methods, renderCommentsView() function, updated View() and Update() to handle comments view
  - `internal/tui/model_test.go` - Added 15 new tests for comments view functionality
  - `go.mod`, `go.sum` - Updated dependencies (go mod tidy)

- **Learnings:**
  - **Patterns discovered:**
    - Drill-down views can be implemented by checking state flag early in View() and returning different render function
    - Share state like `rawMarkdown` between views for consistent UX across detail panel and comments view
    - Use separate scroll offsets (`detailScrollY`, `commentsScrollY`) for independent scrolling contexts
    - When a key should behave differently based on context (like 'q' returning from comments vs quitting app), check state in Update() before processing
  - **Gotchas encountered:**
    - Go time format strings use specific reference date (2006-01-02 15:04:05) - using "2024-01-02" instead causes wrong date parsing
    - Need to reset scroll offset when entering comments view (`commentsScrollY = 0`) and when exiting
    - Disable navigation keys (j/k) when in comments view to prevent cursor changes while viewing comments

---
## ✓ Iteration 8 - US-008: Comments View
*2026-01-21T23:07:06.878Z (278s)*

**Status:** Completed

**Notes:**
odies are rendered with glamour markdown\n- **Markdown toggle** - Press 'm' to toggle between raw and rendered markdown\n- **Scrollable list** - 'h' scrolls up, 'l' scrolls down through comments\n- **Navigation back** - Esc or 'q' returns to the issue list view\n\n**Files changed:**\n- `internal/tui/model.go` - Added comments state, scroll offset, rendering methods\n- `internal/tui/model_test.go` - Added 15 comprehensive tests for comments view\n- `go.mod`, `go.sum` - Updated via go mod tidy\n\n

---

## 2026-01-21 - US-009 Data Refresh
- **What was implemented:**
  - Manual refresh with 'r' or 'R' keybinding in TUI
  - Auto-refresh triggered on app launch via RefreshStartMsg sent after program starts
  - Progress indicator shown in status bar during refresh ("Refreshing..." or "Refreshing issues: X/Y")
  - Sync now removes closed/deleted issues from local database
  - Database methods GetAllIssueNumbers() and DeleteIssues() for tracking and removing closed issues
  - Cursor position maintained after refresh when possible (by remembering selected issue number)
  - Refresh disabled during comments view for better UX

- **Files changed:**
  - `internal/tui/model.go` - Added refresh state (isRefreshing, refreshProgress, refreshError, refreshFunc), message types (RefreshProgressMsg, RefreshDoneMsg, RefreshErrorMsg, RefreshStartMsg), keybindings for 'r'/'R', status bar updates during refresh
  - `internal/tui/model_test.go` - Added 11 tests for refresh functionality
  - `internal/sync/sync.go` - Added logic to detect and remove closed/deleted issues
  - `internal/sync/sync_test.go` - Added 5 tests for incremental sync and deleted issue handling
  - `internal/db/store.go` - Added GetAllIssueNumbers() and DeleteIssues() methods
  - `internal/cmd/root.go` - Integrated refresh with TUI via createRefreshFunc(), auto-refresh on launch

- **Learnings:**
  - **Patterns discovered:**
    - Bubbletea refresh pattern: Model stores refreshFunc (func() tea.Msg), keybinding triggers state change + returns refreshFunc as command
    - Use message types (RefreshDoneMsg, RefreshErrorMsg) to communicate async operation results back to model
    - Auto-refresh on launch: Start TUI first, then use goroutine to send RefreshStartMsg via p.Send()
    - Track existing issue numbers before sync, compare with fetched issues, delete the difference
    - Preserve cursor position across refresh by remembering selected issue number, then finding its new index after sort
  - **Gotchas encountered:**
    - refreshFunc must be set before TUI starts - can't access external dependencies from within tea.Cmd directly
    - Need SetRefreshFunc() method to allow injection of the refresh closure from root command
    - When checking if key was pressed, test for cmd != nil only if refreshFunc is set (tests need to set mock function)
    - p.Send() must be called after p.Run() starts - use goroutine to avoid race

---
## ✓ Iteration 9 - US-009: Data Refresh
*2026-01-21T23:14:29.638Z (442s)*

**Status:** Completed

**Notes:**
ured.\n\n### Files Changed:\n- `internal/tui/model.go` - Added refresh state, message types, and keybindings\n- `internal/tui/model_test.go` - Added 11 tests for refresh functionality\n- `internal/sync/sync.go` - Added closed issue detection and removal logic\n- `internal/sync/sync_test.go` - Added 5 tests for deleted issue handling\n- `internal/db/store.go` - Added GetAllIssueNumbers() and DeleteIssues() methods\n- `internal/cmd/root.go` - Integrated refresh with TUI, auto-refresh on launch\n\n

---

## 2026-01-21 - US-013 Error Handling
- **What was implemented:**
  - Two-tier error display system: minor errors in status bar, critical errors as modal
  - Minor errors (network timeout, rate limit) shown in red text in status bar with retry hint
  - Critical errors (invalid token, database corruption) shown as centered modal dialog
  - Modal requires acknowledgment (Enter, Escape, or 'q') before user can continue
  - Navigation and refresh blocked while error modal is shown
  - Actionable guidance displayed in modal when provided
  - CriticalErrorMsg message type for triggering modal errors from anywhere

- **Files changed:**
  - `internal/tui/model.go` - Added CriticalErrorMsg type, error modal state fields, renderErrorModal(), wrapText(), modal key handling in Update(), error display in status bar
  - `internal/tui/model_test.go` - Added 12 tests for error handling: minor errors in status bar, critical errors as modal, modal acknowledgment, navigation blocking

- **Learnings:**
  - **Patterns discovered:**
    - Use separate message types for minor vs critical errors (RefreshErrorMsg for minor, CriticalErrorMsg for modal)
    - Handle modal keys first in Update() before normal key handling to block actions while modal is shown
    - Use early return pattern `if m.showErrorModal { return m.renderErrorModal() }` at start of View() for modal overlay
    - Text wrapping helper function `wrapText()` useful for fitting content into modal width
    - Modal state requires multiple fields: showErrorModal (bool), errorModalTitle, errorModalMessage, errorModalGuidance
  - **Gotchas encountered:**
    - When blocking keys during modal, still allow Ctrl+C for emergency exit
    - 'q' should dismiss modal, not quit app (unlike normal behavior)
    - Modal centering calculation needs bounds checking for small terminal sizes
    - Error message in status bar should be truncated to fit terminal width

---
## ✓ Iteration 10 - US-013: Error Handling
*2026-01-21T23:19:08.265Z (278s)*

**Status:** Completed

**Notes:**
asErrorModal()`, `GetErrorModalTitle()`, `GetErrorModalMessage()`, `GetErrorModalGuidance()` - State accessors\n   - `renderErrorModal()` - Renders the centered modal dialog\n   - `wrapText()` - Helper for text wrapping in modal\n\n4. **Tests Added:** 12 new tests covering all acceptance criteria\n\n### Files Changed:\n- `internal/tui/model.go` - Core error handling implementation\n- `internal/tui/model_test.go` - Comprehensive test coverage\n- `.ralph-tui/progress.md` - Documented learnings\n\n

---

## 2026-01-21 - US-010 Last Synced Indicator
- **What was implemented:**
  - Status bar now shows "Last synced: <relative time>" (e.g., "5m ago", "2h ago", "3d ago", or "never")
  - Relative time formatting function `formatRelativeTime()` that handles seconds, minutes, hours, and days
  - Last sync time loaded from database on TUI startup
  - Last sync time updated when RefreshDoneMsg is received after successful sync
  - RefreshDoneMsg now includes LastSyncTime field

- **Files changed:**
  - `internal/tui/model.go` - Added lastSyncTime field, SetLastSyncTime/GetLastSyncTime methods, formatRelativeTime function, updated RefreshDoneMsg struct, updated View() status bar
  - `internal/tui/model_test.go` - Added 12 tests for last sync time display and relative time formatting
  - `internal/cmd/root.go` - Load last sync time from database on startup, pass to model, include in RefreshDoneMsg after sync

- **Learnings:**
  - **Patterns discovered:**
    - Relative time formatting is a common UI pattern - keep it simple with buckets (<1m, Nm, Nh, Nd)
    - Zero time value (`time.Time{}`) can represent "never synced" state naturally
    - The database metadata table pattern (key-value store) is already used for last_sync timestamp - no schema changes needed
    - Message types in Bubbletea can be extended with additional fields (RefreshDoneMsg.LastSyncTime) without breaking existing code
  - **Gotchas encountered:**
    - When adding a field to a message type (RefreshDoneMsg), need to update all places that create that message
    - Status bar gets crowded quickly - simplified it to focus on key info (last synced, issue count, sort, refresh, quit)
    - Use `time.Since()` for duration calculation - simpler than manual subtraction

---
## ✓ Iteration 11 - US-010: Last Synced Indicator
*2026-01-21T23:23:09.833Z (241s)*

**Status:** Completed

**Notes:**
ast synced: <time>\" with relative time formatting (e.g., \"5m ago\", \"2h ago\", \"3d ago\", \"never\")\n- [x] **Timestamp stored in database metadata table** - Already existed in `db/store.go` with `GetLastSyncTime()` and `SetLastSyncTime()` methods, using the metadata table\n- [x] **Updates after each successful sync** - `RefreshDoneMsg` now includes `LastSyncTime` field, and the model updates `lastSyncTime` when receiving this message\n\nAll tests pass and the implementation is complete.\n\n

---

## 2026-01-21 - US-011 Keybinding Help
- **What was implemented:**
  - Help overlay toggled with '?' key showing all keybindings organized by context
  - Four sections in help overlay: Navigation, Sorting, Detail Panel/Scrolling, Actions
  - Help overlay dismissible with '?' (toggle off) or 'Esc'
  - Persistent footer shows context-sensitive common keys in both views
  - List view footer: j/k nav, Enter, ?, q
  - Comments view footer: h/l scroll, ?, Esc/q
  - Help overlay blocks all keys except Ctrl+C (quit) and dismiss keys

- **Files changed:**
  - `internal/tui/model.go` - Added showHelpOverlay state, ShowHelpOverlay() method, renderHelpOverlay() function, help key handling in Update(), updated status bar in View() and renderCommentsView() with context-sensitive hints
  - `internal/tui/model_test.go` - Added 11 new tests for help overlay and footer functionality

- **Learnings:**
  - **Patterns discovered:**
    - Help overlay follows same modal pattern as error modal: check state early in Update() to block keys, early return in View() to render overlay
    - Use helper functions inside render methods (e.g., renderLine, renderSection) for consistent formatting of repeated elements
    - Context-sensitive footer: check view state (inCommentsView, etc.) and show relevant keys for that context
    - Modal key blocking pattern: handle modal keys first, return early, then handle normal keys - this prevents accidental actions
  - **Gotchas encountered:**
    - When adding new keybindings ('?'), must add handling both for toggling on (outside modal) and toggling off (inside modal)
    - Helper functions defined inside render methods need proper closure over the strings.Builder
    - Footer gets crowded quickly - balance showing helpful keys vs. keeping it readable
    - Always allow Ctrl+C to quit even when modal is shown (emergency exit)

---
## ✓ Iteration 12 - US-011: Keybinding Help
*2026-01-21T23:27:22.618Z (252s)*

**Status:** Completed

**Notes:**
TestFooterShowsContextSensitiveKeysInCommentsView (0.00s)\n--- PASS: TestFooterShowsHelpHint (0.00s)\n--- PASS: TestHelpOverlayOrganizedByContext (0.00s)\n--- PASS: TestHelpOverlayShowsDismissInstructions (0.00s)\nPASS","stderr":"","interrupted":false,"isImage":false}}
{"type":"assistant","message":{"model":"claude-opus-4-5-20251101","id":"msg_01GTVU5VUTDFRF1FRypDsgCB","type":"message","role":"assistant","content":[{"type":"text","text":"All tests pass. US-011 - Keybinding Help is complete.\n\n

---

## 2026-01-21 - US-012 Color Themes
- **What was implemented:**
  - Six built-in themes: default, dracula, gruvbox, nord, solarized-dark, solarized-light
  - Theme type and validation in config package with TOML persistence via `display.theme`
  - New `internal/themes` package with Theme struct containing color definitions and Styles() method for lipgloss styles
  - `ghissues themes` command to list all themes, preview a theme (`--preview`), or set a theme (`--set`)
  - TUI model integration using themed styles for title, headers, status bar, labels, selections, errors
  - Theme persisted in config.toml and loaded on startup

- **Files changed:**
  - `internal/config/config.go` - Added Theme type, ValidThemes, validation, DisplayName(), AllThemes()
  - `internal/config/config_test.go` - Added 8 theme-related tests
  - `internal/themes/themes.go` - New file with Theme struct, Styles struct, color definitions for all 6 themes
  - `internal/themes/themes_test.go` - New file with 5 test functions covering theme retrieval and styles
  - `internal/cmd/themes.go` - New file with `ghissues themes` command (list/preview/set)
  - `internal/cmd/themes_test.go` - New file with 8 tests for themes command
  - `internal/cmd/root.go` - Added themes subcommand, pass theme to NewModelWithTheme()
  - `internal/tui/model.go` - Added theme field, NewModelWithTheme(), SetTheme()/GetTheme(), getThemeStyles(), updated all render methods to use themed styles
  - `internal/tui/model_test.go` - Added 5 theme-related tests

- **Learnings:**
  - **Patterns discovered:**
    - Centralize theme colors in a dedicated package (`internal/themes`) separate from config types
    - Theme struct holds raw colors; Styles() method generates lipgloss.Style objects on demand
    - Use `getThemeStyles()` helper in model to lazily get styles, avoiding storing styles in model state
    - Preview command pattern: show styled sample output without modifying config, then suggest `--set` command
    - Theme names use kebab-case (`solarized-dark`) for CLI/config, DisplayName() provides human-readable form
  - **Gotchas encountered:**
    - Tests checking output need case-insensitive matching (`strings.ToLower()`) since DisplayName() capitalizes
    - When refactoring hardcoded styles to use theme, must update all render methods consistently
    - Import cycle risk: themes package imports config for Theme type, but config shouldn't import themes
    - lipgloss.Color can use hex codes (`#bd93f9`) or ANSI numbers (`"86"`) - consistency matters for predictability

---
## ✓ Iteration 13 - US-012: Color Themes
*2026-01-21T23:35:49.746Z (506s)*

**Status:** Completed

**Notes:**
8 new theme tests\n- `internal/themes/themes.go` - Theme definitions with colors\n- `internal/themes/themes_test.go` - 5 theme package tests\n- `internal/cmd/themes.go` - New themes command\n- `internal/cmd/themes_test.go` - 8 command tests\n- `internal/cmd/root.go` - Register themes subcommand, pass theme to TUI\n- `internal/tui/model.go` - Theme field, constructors, themed render methods\n- `internal/tui/model_test.go` - 5 TUI theme tests\n- `.ralph-tui/progress.md` - Documented learnings\n\n

---

## 2026-01-21 - US-014 Multi-Repository Configuration
- **What was implemented:**
  - Multi-repository configuration support: `[[repositories]]` TOML array for multiple repos
  - RepositoryConfig struct with Name and DatabasePath fields
  - Automatic migration from old single-repo format (`repository = "..."`) to new multi-repo format
  - `default_repository` field for setting which repo to use by default
  - `--repo owner/repo` flag on root command to select repository
  - `ghissues repos` command to list and manage repositories:
    - Lists all configured repos with `*` indicator for default
    - `--add owner/repo` flag to add new repositories
    - `--db-path /path/to.db` flag for custom database path when adding
    - `--set-default owner/repo` flag to set the default repository
  - Each repository has its own database file (from config or auto-generated from repo name)
  - Repository selection priority: --repo flag > default_repository > single repo (if only one)

- **Files changed:**
  - `internal/config/config.go` - Added RepositoryConfig type, Repositories slice, DefaultRepository field, migration function, GetRepositoryConfig(), GetActiveRepository(), AddRepository(), ListRepositories(), SetDefaultRepository(), GetDatabasePath() methods
  - `internal/config/config_test.go` - Added 17 tests for multi-repo configuration
  - `internal/cmd/repos.go` - New repos command for listing and managing repositories
  - `internal/cmd/repos_test.go` - 10 tests for repos command
  - `internal/cmd/root.go` - Added --repo flag, repoFlag global variable, updated to use GetActiveRepository()

- **Learnings:**
  - **Patterns discovered:**
    - TOML arrays of tables use `[[tablename]]` syntax for multi-entry sections
    - Migration pattern: check for legacy field + empty new field, then transform data in Load()
    - Use `GetActiveRepository(flagValue)` pattern to encapsulate repository selection logic with clear priority
    - Generate default database paths from repo names by replacing `/` with `-` (owner/repo -> .owner-repo.db)
    - When adding new persistent flags, need SetFlagName/GetFlagName functions for testing similar to existing patterns
  - **Gotchas encountered:**
    - When migrating config formats, need to handle both old and new format coexisting during transition
    - Tests that check specific repository output need to reset global flags (defer SetRepoFlag(""))
    - Multiple repos without a default should error with helpful message, not silently pick one
    - The --db flag for root command overrides per-repo database paths, maintaining existing precedence pattern

---
## ✓ Iteration 14 - US-014: Multi-Repository Configuration
*2026-01-21T23:41:21.122Z (330s)*

**Status:** Completed

**Notes:**
ithub.com/shepbook/ghissues/internal/setup\t1.385s\nok  \tgithub.com/shepbook/ghissues/internal/sync\t1.937s\nok  \tgithub.com/shepbook/ghissues/internal/themes\t1.607s\nok  \tgithub.com/shepbook/ghissues/internal/tui\t1.244s","stderr":"","interrupted":false,"isImage":false}}
{"type":"assistant","message":{"model":"claude-opus-4-5-20251101","id":"msg_01XUC9ZrT8tDgibSq3Fzhpzn","type":"message","role":"assistant","content":[{"type":"text","text":"All tests pass. The implementation is complete.\n\n

---
