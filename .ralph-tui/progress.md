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
