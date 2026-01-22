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

**TUI Pattern with tview**: Use `rivo/tview` for terminal UI:
- Explicit type declarations to avoid type inference issues with tview primitives
- Use `tcell.EventKey` for keyboard input handling
- Chain methods on Flex, Pages, List, TextView for layout composition
- Use SetInputCapture for custom key handling
- Modal dialogs for help screens using Pages
- Track current selection index manually (no GetCurrentIndex method)
- Set Scrollable(true) on TextView for scrollable content

**Markdown Rendering Pattern with glamour**: Use `github.com/charmbracelet/glamour` for terminal markdown rendering:
- Create TermRenderer with `glamour.WithAutoStyle()` for auto-styling
- Fall back to raw text if rendering fails
- Toggle between rendered and raw markdown with state variable
- Update display when markdown mode changes

**Drill-Down View Pattern**: Replace main page content instead of using modal overlays for primary views:
- Use pages.SwitchToPage() to switch between view states
- Set inCommentsView state variable to track current view
- Update status bar based on current view context
- Call updateStatusBar() when switching views

**View State Pattern**: Track view context with boolean flags for conditional UI:
- Declare state variables (inCommentsView, commentsMarkdownRendered) before functions that use them
- Use closures to access and modify state from event handlers
- Update UI immediately after state changes

**Incremental Sync Pattern**: Sync only updated issues using GitHub's `since` parameter:
- Get last sync time from database using MAX(synced_at)
- Use `FetchIssuesSince` with RFC3339 timestamp for incremental fetch
- Track issues present before and after sync to detect deletions
- Always fetch comments for updated issues to get new comments

**TUI Refresh Pattern**: Implement background refresh with progress display in status bar:
- Use goroutine for async refresh with ProgressCallback
- Use app.QueueUpdateDraw() for thread-safe UI updates from goroutine
- Track isRefreshing state to prevent concurrent refreshes
- Update status bar with progress info during refresh
- Refresh issue list after sync completes

**ProgressCallback Pattern**: Report progress during long-running operations:
- Define callback type: `func(current, total int, status string)`
- Call callback periodically with current progress and status message
- Can pass nil for no progress reporting (silent mode)
- Update status bar in UI thread using QueueUpdateDraw()

**Error Handling Pattern**: Use a dedicated `internal/errors` package for categorized error handling:
- `IsMinorError()` identifies network timeouts and rate limits
- `IsCriticalError()` identifies invalid tokens and database corruption
- `GetErrorHint()` returns actionable guidance with retry capability
- Minor errors displayed in status bar, critical errors shown in modal dialog
- Modal dialogs require user acknowledgment via OK button
- Press 'r' on minor error to clear and retry

**Footer Pattern for Context-Sensitive Keybindings**: Display context-sensitive common keys in a persistent footer:
- Add a dedicated footer TextView below the status bar
- Create `GetFooterDisplay(viewState, markdownState)` function that returns context-appropriate text
- Use `SetDynamicColors(true)` on footer TextView for colored key indicators
- Update footer in `updateFooter()` closure when switching views
- Include the footer in the main Flex layout (FlexRow) below the status bar
- Call `updateFooter()` whenever view state changes (displayIssues, returnToIssueList, showComments)

**Keyboard Dismissal Pattern for Modals**: Handle keyboard input for dismissing help/dialog modals:
- Add `SetInputCapture()` on modal views to intercept key events
- Handle `tcell.KeyEscape` and `tcell.KeyRune`('?') for dismissal
- Call `pages.SwitchToPage("main")` and `app.SetFocus(app.GetFocus())` to dismiss
- Return `nil` from the input capture to prevent event propagation

**Color Theme Pattern**: Support multiple color themes using tcell for terminal UI:
- Create `internal/themes` package with Theme struct containing color hex strings
- Define themes as package-level variables (Default, Dracula, Gruvbox, Nord, SolarizedDark, SolarizedLight)
- Use `tcell.NewRGBColor(int32(r), int32(g), int32(b))` to convert hex to tcell.Color
- Theme selected via `display.theme` config field
- `ghissues themes` command allows previewing and changing themes
- Default theme applied when config doesn't specify one

**Multi-Repository Configuration Pattern**: Support multiple repositories in config:
- Add `Repositories []string` field and `DefaultRepository string` to Config struct
- Keep legacy `Repository string` field for backward compatibility
- Helper methods: `HasRepositories()`, `GetRepositories()`, `AddRepository()`, `RemoveRepository()`, `GetDefaultRepository()`, `SetDefaultRepository()`
- Use priority: CLI flag > `DefaultRepository` > legacy `Repository` field
- Each repo has its own database file (owner_repo.db pattern)
- `ghissues repos` command lists configured repositories
- `ghissues --repo owner/repo` selects which repo to view

---

## 2026-01-20 - US-014
- What was implemented:
  - Added `Repositories` slice field and `DefaultRepository` field to Config struct
  - Added helper methods: `HasRepositories()`, `GetRepositories()`, `AddRepository()`, `RemoveRepository()`, `GetDefaultRepository()`, `SetDefaultRepository()`
  - Added `isValidRepoFormat()` validation function for owner/repo format
  - Added `--repo` flag to CLI for selecting which repo to view
  - Added `repos` subcommand to list configured repositories
  - Created helper functions: `runRepoAdd()`, `runRepoRemove()`, `runRepoSetDefault()`
  - Updated `runSync()` to use repoFlag > GetDefaultRepository() > legacy Repository
  - Updated `RefreshSync()` with repo parameter for repo selection priority
  - Updated `RunTUI()` to use the same repo selection logic

- Files changed:
  - `internal/config/config.go` (added multi-repo fields and helper methods)
  - `internal/config/config_test.go` (added 10 new tests for multi-repo functionality)
  - `cmd/ghissues/main.go` (added repoFlag, repos subcommand handler)
  - `cmd/ghissues/repos.go` (new file with repo management functions)
  - `cmd/ghissues/repos_test.go` (new file with 9 tests)
  - `cmd/ghissues/sync.go` (updated runSync and RefreshSync for repo parameter)
  - `cmd/ghissues/tui.go` (updated repo parsing and RefreshSync calls)

- **Learnings:**
  - Patterns discovered:
    - Multi-Repo Config Pattern: Use a slice for repos plus a default field for selection
    - Repo Priority Pattern: CLI flag > config default > legacy single repo field (backward compat)
    - Helper Method Pattern: Add methods on Config struct for repository management
  - Gotchas encountered:
    - Need to maintain backward compatibility with legacy `Repository` field
    - Variable shadowing when renaming parameters (used `repoToUse` vs `repo`)
    - RefreshSync needed new parameter to avoid modifying global repo state

---

## 2026-01-20 - US-013
- What was implemented:
  - Created `internal/errors/errors.go` with error classification functions
  - Added `IsMinorError()` and `IsCriticalError()` for error categorization
  - Added `GetErrorHint()` with actionable guidance for each error type
  - Updated TUI status bar to show minor errors (network timeout, rate limit)
  - Added critical error modal dialog for invalid token, database corruption
  - Modal requires user acknowledgment (OK button)
  - Press 'r' to retry after minor error (clears error and retries)

- Files changed:
  - `internal/errors/errors.go` (new file with error classification)
  - `internal/errors/errors_test.go` (new file with 12 tests)
  - `cmd/ghissues/tui.go` (added error state, status bar error display, error modal)

- **Learnings:**
  - Patterns discovered:
    - Error Handling Pattern: Categorize errors as minor/critical for appropriate UI display
    - Error Hint Pattern: Provide actionable guidance for error resolution
    - Modal Error Pattern: Use tview.Modal with OK button for critical errors
    - Status Bar Error Pattern: Display error message in status bar for minor errors
  - Gotchas encountered:
    - "invalid GitHub token" does not contain "invalid token" as substring (github in between)
    - Must check critical errors first in IsMinorError to avoid false positives
    - Error state variables must be declared before updateStatusBar closure

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
## ✓ Iteration 4 - US-003: Initial Issue Sync
*2026-01-20T08:54:47.570Z (467s)*

**Status:** Completed

**Notes:**
Sync can be cancelled with Ctrl+C gracefully\n\n**Files changed:**\n- `internal/github/client.go` - Added structs and fetching functions\n- `internal/github/issue_test.go`, `internal/github/fetch_test.go` - 21 tests\n- `internal/db/ops.go`, `internal/db/ops_test.go` - Database operations with 14 tests\n- `cmd/ghissues/sync.go`, `cmd/ghissues/sync_test.go` - Sync command with 11 tests\n- `cmd/ghissues/main.go` - Added sync subcommand handler\n- `go.mod` - Added `modernc.org/sqlite` dependency\n\n

---

## 2026-01-20 - US-005
- What was implemented:
  - Added Display struct with Columns field to Config for configurable issue list columns
  - Created IssueList struct and ListIssues, GetIssue functions in db/ops.go
  - Built TUI with issue list view using tview library
  - Vertical split layout: issue list (left) + details placeholder (right)
  - Vim keys (j/k) and arrow keys for navigation
  - Issue count shown in status bar
  - Help modal with keyboard shortcuts
  - Default columns: number, title, author, date, comments

- Files changed:
  - internal/config/config.go (added Display struct, DefaultColumns, apply defaults in Load)
  - internal/config/config_display_test.go (3 tests for Display.Columns)
  - internal/db/ops.go (added IssueList struct, ListIssues, GetIssue functions)
  - internal/db/ops_list_test.go (6 tests for ListIssues/GetIssue)
  - cmd/ghissues/tui.go (new file with TUI, issue list, navigation, status bar)
  - cmd/ghissues/tui_test.go (10 tests for TUI helper functions)
  - cmd/ghissues/main.go (added tui subcommand handler)
  - go.mod (added rivo/tview and gdamore/tcell/v2 dependencies)

- **Learnings:**
  - Patterns discovered:
    - TUI Pattern with tview: Explicit type declarations avoid type inference issues
    - Flex layout with FlexColumn for vertical split (left panel, right panel)
    - Pages for modal dialogs (help screen)
    - SetInputCapture for custom key handling with tcell.EventKey
    - Status bar using TextView at bottom of FlexRow layout
  - Gotchas encountered:
    - tview.NewList() returning *Box instead of *List in type inference - use explicit type declarations
    - tcell.KeyRune, tcell.KeyEscape, tcell.KeyCtrlC for keyboard events
    - Method chaining on tview primitives requires explicit variable declaration
    - ValidateColumns should return false for empty column lists
## ✓ Iteration 5 - US-005: Issue List View
*2026-01-20T09:01:24.447Z (396s)*

**Status:** Completed

**Notes:**
GetIssue\n- `cmd/ghissues/tui.go` - New TUI implementation\n- `cmd/ghissues/tui_test.go` - 10 tests for TUI helpers\n- `cmd/ghissues/main.go` - Added tui subcommand\n- `go.mod` - Added tview and tcell dependencies\n\n**Key patterns discovered:**\n- TUI Pattern with tview: Use explicit type declarations to avoid type inference issues\n- Flex layout with FlexColumn for vertical split panels\n- Pages for modal dialogs (help screens)\n- SetInputCapture with tcell.EventKey for custom key handling\n\n

---

## 2026-01-20 - US-006
- What was implemented:
  - Added SortOption and SortOrder types to config Display struct for sortable issues
  - Available sort options: updated (default), created, number, comments
  - Sort order: desc (default) and asc
  - Added ListIssuesSorted function in db/ops.go with dynamic ORDER BY clause
  - TUI keybindings: 's' to cycle through sort options, 'S' to reverse order
  - Current sort shown in status bar with format "Sort: <Name> <↑↓>"
  - Sort preference persisted to config file automatically on change

- Files changed:
  - internal/config/config.go (added SortOption, SortOrder types, Display struct fields)
  - internal/config/config_display_test.go (7 new tests for sort options)
  - internal/db/ops.go (added ListIssuesSorted function, uses config package)
  - internal/db/ops_list_test.go (6 new tests for sorting)
  - cmd/ghissues/tui.go (added sort helper functions, keybindings, status bar display)
  - cmd/ghissues/tui_test.go (4 new tests for sort helpers)

- **Learnings:**
  - Patterns discovered:
    - Sort Preference Pattern: Store sort settings in config Display struct with TOML serialization
    - SQL ORDER BY Pattern: Use dynamic column name and direction for flexible sorting
    - Status Bar Pattern: Include dynamic sort info in status bar with FormatSortDisplay helper
  - Gotchas encountered:
    - tview.List doesn't have GetCurrentIndex - use SetChangedFunc to track selection
    - Variable scope issues when declaring pages before use in closures
    - fmt.Sprintf for dynamic SQL needs proper escaping/sanitization of column names
## ✓ Iteration 6 - US-006: Issue Sorting
*2026-01-20T09:06:37.674Z (312s)*

**Status:** Completed

**Notes:**
ruct\n- `internal/config/config_display_test.go` - Added 7 tests for sort configuration\n- `internal/db/ops.go` - Added `ListIssuesSorted()` function with dynamic `ORDER BY` clause\n- `internal/db/ops_list_test.go` - Added 6 tests for sorting functionality\n- `cmd/ghissues/tui.go` - Added sort helper functions, keybindings, status bar display\n- `cmd/ghissues/tui_test.go` - Added 4 tests for sort helper functions\n\n**Tests:**\nAll tests pass with `go test ./...` and `go vet ./...` is clean.\n\n

---

---

## 2026-01-20 - US-007
- What was implemented:
  - Added IssueDetail struct with Body, UpdatedAt, HTMLURL, Labels, and Assignees fields
  - Created GetIssueDetail function in db/ops.go that fetches full issue details with labels and assignees
  - Created GetLabels and GetAssignees helper functions to fetch issue metadata
  - Created GetComments function to fetch comments for an issue
  - Created Comment struct for representing comment data
  - Updated TUI detail view to show full issue details: number, title, state badge, author, dates, labels, assignees, comment count, and body
  - Added glamour markdown rendering for issue body with 'm' key toggle
  - Added scrollable detail view using SetScrollable(true)
  - Added comments view modal that opens on Enter key with all comments
  - Added markdown/raw toggle indicator in status bar
  - Added 'm' keybinding to toggle between rendered markdown and raw text
  - Updated help modal with new keybindings for markdown toggle and comments view

- Files changed:
  - internal/db/ops.go (added IssueDetail struct, GetIssueDetail, GetLabels, GetAssignees, GetComments, Comment struct)
  - internal/db/ops_list_test.go (added 8 new tests for issue detail, labels, assignees, and comments)
  - cmd/ghissues/tui.go (updated detail view, added markdown rendering with glamour, comments modal, new keybindings)
  - go.mod, go.sum (added github.com/charmbracelet/glamour dependency)

- **Learnings:**
  - Patterns discovered:
    - Detail View Pattern: Create dedicated detail struct with all display fields
    - Markdown Rendering Pattern: Use glamour with WithAutoStyle() for terminal-aware rendering
    - Comments Modal Pattern: Use Pages with modal dialog for displaying comments
    - Selection Tracking Pattern: Track currentIndex manually since tview.List lacks GetCurrentIndex
    - State Toggle Pattern: Use boolean flag for markdown/raw mode with status bar indicator
  - Gotchas encountered:
    - glamour.WithWidth API changed in v0.10 - use WithAutoStyle() instead
    - tview.List lacks GetCurrentIndex method - need to track selection index manually
    - SQLite foreign key constraints require parent issue to exist before inserting labels/assignees/comments
    - Config field is SortOrder not Order - match exact field names from config struct
## ✓ Iteration 7 - US-007: Issue Detail View
## 2026-01-20 - US-008
- What was implemented:
  - Converted comments modal to drill-down view that replaces main interface
  - Added issue title/number as header in comments view
  - Each comment now shows: author, date, body with markdown rendering via glamour
  - Toggle markdown rendering with 'm' key in comments view (independent from issue body toggle)
  - Comments list is scrollable using SetScrollable(true)
  - 'Esc' or 'q' returns to issue list view from comments
  - Dynamic status bar showing current view context (issue list vs comments view)
  - Updated help modal with dedicated comments view section

- Files changed:
  - cmd/ghissues/tui.go (221 lines): Full implementation of drill-down comments view
  - cmd/ghissues/tui_test.go (53 lines): Added TestFormatComments, TestFormatCommentsEmpty, TestFormatCommentsMarkdown

- **Learnings:**
  - Patterns discovered:
    - Drill-Down View Pattern: Replace pages content instead of overlaying modal for main views
    - View State Pattern: Track inCommentsView boolean to switch status bar content
    - Reusable Package-Level Functions: Move testable functions (formatComments, renderMarkdown) to package level
    - Closures for State Access: Use closures to access RunTUI state from event handlers
  - Gotchas encountered:
    - Variable scope in closures: State variables must be declared before functions that use them
    - Duplicate variable declarations: Cannot redeclare with := in same scope
    - Status bar context: Need to call updateStatusBar() when switching views
    - Package-level vs closure functions: Tests can only access package-level functions
## ✓ Iteration 8 - US-008: Comments View
*2026-01-20T09:18:22.894Z (362s)*

**Status:** Completed

**Notes:**
t\n- Updated help modal with dedicated comments view section\n\n**Files changed:**\n- `cmd/ghissues/tui.go` - 221 lines: Full drill-down implementation\n- `cmd/ghissues/tui_test.go` - 53 lines: 3 new tests for `formatComments`\n- `.ralph-tui/progress.md` - Updated with 2 new patterns and learnings\n\n**New patterns documented:**\n- Drill-Down View Pattern: Replace pages content for primary views instead of modals\n- View State Pattern: Track view context with boolean flags for conditional UI\n\n

---

---

## 2026-01-20 - US-009
- What was implemented:
  - Added auto-refresh on app launch via RunTUIWithRefresh function
  - Added manual refresh with 'r' or 'R' keybinding
  - Progress bar shown in status bar during refresh
  - Incremental sync using GitHub API's `since` parameter
  - Handles deleted issues (removes from local db)
  - Handles new comments on existing issues (always fetches comments for updated issues)
  - Added RefreshSync function with ProgressCallback for progress reporting
  - Added GetLastSyncTime, GetIssueNumbers, DeleteIssue functions to db package

- Files changed:
  - internal/db/ops.go (added GetLastSyncTime, GetIssueNumbers, DeleteIssue, DeleteIssues functions)
  - internal/db/ops_test.go (added 7 tests for new DB functions)
  - internal/github/client.go (added FetchIssuesSince function with since parameter support)
  - internal/github/fetch_test.go (added 2 tests for FetchIssuesSince)
  - cmd/ghissues/sync.go (added RefreshSync function with ProgressCallback type)
  - cmd/ghissues/sync_test.go (added 2 tests for ProgressCallback)
  - cmd/ghissues/tui.go (added RunTUIWithRefresh, isRefreshing state, refreshStatus, 'r' key handler)
  - cmd/ghissues/main.go (updated to use RunTUIWithRefresh for auto-refresh on launch)

- **Learnings:**
  - Patterns discovered:
    - Incremental Sync: Use GitHub's `since` parameter to fetch only updated issues
    - TUI Refresh: Use goroutine + QueueUpdateDraw for async operations in tview
    - ProgressCallback: Define callback type for progress reporting in long-running ops
  - Gotchas encountered:
    - sql.NullString needed for scanning NULL values from MAX() aggregate
    - Variable scope in closures: isRefreshing must be declared before updateStatusBar closure
    - UI updates from goroutine need QueueUpdateDraw for thread safety
    - Prevent concurrent refreshes with isRefreshing flag check
    - Fetch comments for all updated issues to handle new comments
## ✓ Iteration 9 - US-009: Data Refresh
*2026-01-20T09:24:14.066Z (350s)*

**Status:** Completed

**Notes:**
- Added `RunTUIWithRefresh`, refresh state variables, 'r' key handler\n- `cmd/ghissues/main.go` - Updated to use `RunTUIWithRefresh`\n- `.ralph-tui/progress.md` - Added new patterns and US-009 entry\n\n### New Patterns Documented:\n- **Incremental Sync Pattern** - Use GitHub's `since` parameter for efficient syncing\n- **TUI Refresh Pattern** - Goroutine + QueueUpdateDraw for async UI operations\n- **ProgressCallback Pattern** - Callback type for progress reporting in long-running operations\n\n

---
## ✓ Iteration 10 - US-013: Error Handling
*2026-01-20T09:34:16.207Z (601s)*

**Status:** Completed

**Notes:**
tance Criteria Met\n1. **Minor errors shown in status bar** - Network timeouts, rate limits display as `[ERROR] <message>`\n2. **Critical errors shown as modal** - Invalid tokens, database corruption use tview.Modal\n3. **Modal requires acknowledgment** - OK button required to dismiss\n4. **Actionable guidance** - Each error type includes specific action (e.g., \"Update your token\", \"Check your connection\")\n5. **Network error suggestions** - \"Check your internet connection and try again\"\n

---

## 2026-01-20 - US-010
- What was implemented:
  - Added FormatRelativeTime function to format timestamps as relative time (e.g., "5 minutes ago", "3 days ago")
  - Added GetLastSyncedDisplay function returning "Last synced: <relative time>"
  - Updated status bar to show last synced time in issue list view
  - Added lastSynced state variable that gets updated after each displayIssues call
  - Timestamp is already stored via synced_at column, GetLastSyncTime function already exists

- Files changed:
  - `cmd/ghissues/tui.go` (added FormatRelativeTime, GetLastSyncedDisplay functions, updated status bar)
  - `cmd/ghissues/tui_test.go` (added TestFormatRelativeTime, TestGetLastSyncedDisplay tests)

- **Learnings:**
  - Patterns discovered:
    - Relative Time Pattern: Use time.Since() and time.Duration for formatting relative times
    - State Variable Pattern: Declare state variables before closures that need access
  - Gotchas encountered:
    - Future timestamps need to be handled (show as "just now")
    - Test timestamps using Add() need correct sign (negative for past, positive for future)
---
## ✓ Iteration 11 - US-010: Last Synced Indicator
*2026-01-20T09:36:56.579Z (159s)*

**Status:** Completed

**Notes:**
e>\" | Sort: ... | [Markdown] | ...\n\n4. **Added tests** (`cmd/ghissues/tui_test.go`):\n   - `TestFormatRelativeTime` - 9 test cases covering various time ranges\n   - `TestGetLastSyncedDisplay` - 2 test cases for the display function\n\n**Acceptance Criteria Met:**\n- ✅ Status bar shows \"Last synced: <relative time>\" (e.g., \"5 minutes ago\")\n- ✅ Timestamp already stored in database via `synced_at` column\n- ✅ Updates after each successful sync (via `displayIssues()` call after refresh)\n\n

---

## 2026-01-20 - US-011
- What was implemented:
  - Added persistent footer TextView below status bar for context-sensitive keybindings
  - Created `GetFooterDisplay(inCommentsView, commentsMarkdownRendered)` function for testable footer text
  - Footer shows "help | comments | refresh" in issue list view
  - Footer shows "back | markdown/raw" in comments view
  - Added `SetInputCapture()` on help modal to dismiss with ? or Esc keys
  - Updated help modal text to indicate ? dismisses help

- Files changed:
  - `cmd/ghissues/tui.go` (added footer creation, GetFooterDisplay function, showHelp keyboard dismissal)
  - `cmd/ghissues/tui_test.go` (added TestGetFooterDisplay with 3 test cases)

- **Learnings:**
  - Patterns discovered:
    - Footer Pattern: Use separate TextView with SetDynamicColors(true) for context-sensitive keybindings
    - Testable Function Pattern: Extract footer text generation to standalone GetFooterDisplay function
    - Keyboard Dismissal Pattern: Use SetInputCapture on modal to handle ? and Esc for dismissal
  - Gotchas encountered:
    - Modal.SetInputCapture may not capture all keys - need proper focus management
    - Footer TextView needs SetDynamicColors(true) to render [yellow] color tags
    - Must call updateFooter() when switching views (returnToIssueList, showComments)

---
## ✓ Iteration 12 - US-011: Keybinding Help
*2026-01-20T09:40:03.299Z (186s)*

**Status:** Completed

**Notes:**
help modal to handle keyboard dismissal\n\n**Files Changed:**\n- `cmd/ghissues/tui.go` - Added footer TextView, `GetFooterDisplay()` function, keyboard dismissal for help modal\n- `cmd/ghissues/tui_test.go` - Added `TestGetFooterDisplay` with 3 test cases\n\n**New Patterns Documented:**\n- **Footer Pattern** - Use separate TextView with `SetDynamicColors(true)` for context-sensitive keybindings\n- **Keyboard Dismissal Pattern** - Use `SetInputCapture` on modal to handle ? and Esc for dismissal\n

---

## 2026-01-20 - US-012
- What was implemented:
  - Created `internal/themes/themes.go` with 6 built-in themes: default, dracula, gruvbox, nord, solarized-dark, solarized-light
  - Added `Theme` struct with hex color strings for all UI elements (primary/secondary text, backgrounds, borders, highlights)
  - Added `Get()`, `All()`, `IsValid()`, `List()` functions for theme management
  - Updated `internal/config/config.go` to add `Theme` field to Display struct with default "default"
  - Added `ghissues themes` subcommand in `cmd/ghissues/themes.go` for previewing and changing themes
  - Updated `cmd/ghissues/main.go` to handle "themes" subcommand
  - Updated `cmd/ghissues/tui.go` to apply theme colors to issue list and detail views
  - Added `getThemeColor()` helper function to convert hex colors to `tcell.Color` using `tcell.NewRGBColor`

- Files changed:
  - `internal/themes/themes.go` (new file with theme definitions)
  - `internal/themes/themes_test.go` (new file with 8 tests)
  - `internal/config/config.go` (added Theme field to Display struct, DefaultTheme function)
  - `internal/config/config_display_test.go` (added 3 new tests for theme support)
  - `cmd/ghissues/themes.go` (new file with runThemes function)
  - `cmd/ghissues/main.go` (added themes subcommand handler)
  - `cmd/ghissues/tui.go` (added themes import, currentTheme, color application)
  - `cmd/ghissues/tui_test.go` (added TestGetThemeColor)

- **Learnings:**
  - Patterns discovered:
    - Color Theme Pattern: Define themes with hex colors, convert using tcell.NewRGBColor
    - Theme Package Pattern: Centralize theme definitions in dedicated internal package
    - Helper Function Pattern: Extract color conversion to helper function for testability
  - Gotchas encountered:
    - tcell.Color is an int type, cannot directly cast from string - need to parse hex RGB values
    - tcell.NewRGBColor expects int32 parameters, not int - needed explicit int32 cast
    - Default fallback needed for invalid hex colors using tcell.ColorDefault
    - lipgloss.Color and tcell.Color are incompatible types - use string type in theme struct

---
## ✓ Iteration 13 - US-012: Color Themes
*2026-01-20T09:46:16.277Z (372s)*

**Status:** Completed

**Notes:**
handler\n- `cmd/ghissues/tui.go` - Applied theme colors to issue list and detail views\n- `cmd/ghissues/tui_test.go` - Added `TestGetThemeColor`\n\n**Acceptance Criteria Met:**\n- ✅ 6 built-in themes: default, dracula, gruvbox, nord, solarized-dark, solarized-light\n- ✅ Theme selected via config file `display.theme`\n- ✅ Theme preview/change with `ghissues themes` command\n- ✅ Uses tcell for terminal UI styling via `tcell.NewRGBColor()`\n\nAll tests pass and `go vet` / `go build` are clean.\n\n

---
## ✓ Iteration 14 - US-014: Multi-Repository Configuration
*2026-01-20T09:50:51.571Z (274s)*

**Status:** Completed

**Notes:**
ion:\n"}],"model":"hf:MiniMaxAI/MiniMax-M2.1","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":87232,"output_tokens":0,"cache_creation_input_tokens":0,"cache_read_input_tokens":0},"context_management":null},"parent_tool_use_id":null,"session_id":"0613b8e4-1d9c-4dcf-a08d-f0d17bfd1e2c","uuid":"20ccfe06-7648-44ad-b692-57a40e799a53"}
{"type":"assistant","message":{"id":"msg_c7925ae3-dd9b-4c33-b839-1fd777025646","type":"message","role":"assistant","content":[{"type":"text","text":"\n\n

---
