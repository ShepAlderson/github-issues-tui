# Ralph Progress Log

This file tracks progress across iterations. It's automatically updated
after each iteration and included in agent prompts for context.

## Codebase Patterns (Study These First)

### Config File Management Pattern
- **Secure Permissions**: Always save config files with 0600 permissions (owner read/write only)
- **Directory Creation**: Use `EnsureConfigDir()` helper to create parent directories before writing
- **Validation**: Validate user input (like repo format `owner/repo`) before saving to config
- **TOML Structure**: Use struct tags with `toml:"fieldname"` for marshaling

### TDD Pattern for Go
- Write comprehensive test file first (`*_test.go`)
- Run tests to see them fail (red)
- Implement minimal code to pass tests (green)
- No need for explicit refactoring step on first implementation
- Use `t.TempDir()` for test file isolation

### CLI Pattern
- Use `flag` package for command-line flags
- Define flags before `flag.Parse()`
- Handle special commands (like `config`) as positional arguments
- Provide `--help` and `--version` flags as standard practice

### Testing Patterns
- **Table-driven tests**: Use for testing multiple input scenarios (see `TestValidateRepository`)
- **Temporary directories**: Use `t.TempDir()` for file system tests - auto-cleanup
- **Error checking**: Test both success and failure paths
- **File permissions**: Verify security properties in tests

### Authentication Pattern
- **Priority Order**: Environment variable → Config file → External CLI (gh)
- **Silent Fallback**: Try next method if current fails, only error if all fail
- **Return Source**: Always return where auth came from for debugging/user feedback
- **Use Standard Libraries**: Prefer `os.Getenv()` for env vars, file I/O for configs

### Database Path Resolution Pattern
- **Priority Order**: CLI flag → Config file → Default value
- **Whitespace Handling**: Use `strings.TrimSpace()` to treat whitespace-only values as empty
- **Directory Creation**: Use `os.MkdirAll()` with 0755 permissions for parent directories
- **Writability Checks**: Test file creation before committing to a path for better error messages
- **Error Context**: Include the actual path in error messages for debugging

### TUI Pattern (Bubbletea/Elm Architecture)
- **Model-Update-View**: Separate state (Model), state transitions (Update), and rendering (View)
- **Viewport Management**: Track offset + cursor position; scroll when cursor approaches viewport edges
- **Column Configuration**: Use metadata maps for extensible column systems with validation
- **Navigation**: Support both Vim keys (j/k) and arrow keys for better accessibility
- **Status Bar**: Always show item count, navigation help, and current selection state
- **Styling**: Use lipgloss for colors, bold, spacing - keep styles simple and performant

### Sorting Pattern
- **Preserve Original Order**: Keep unsorted data separate to allow re-sorting with different criteria
- **Sort Field Validation**: Use map for O(1) lookup of valid sort fields
- **Sort Cycling**: Define cycle order as array to support field rotation
- **Default Values**: Provide sensible defaults (updated date descending for recency)
- **Sort Labels**: Map internal field names to human-readable labels for display
- **Config Integration**: Persist sort preferences to config file for user convenience

### Detail Panel Pattern
- **Split Layout Management**: Use percentage-based width allocation (60/40) for responsive panels
- **Automatic Panel Updates**: Update detail panel on cursor movement, not just on selection
- **Viewport Isolation**: Each panel manages its own viewport state and scrolling
- **Lazy Rendering**: Only render detail panel when issue is available (handle empty states)
- **Label Rendering**: Use colored badges with padding for visual distinction
- **Fallback Rendering**: Gracefully fall back to raw markdown if glamour rendering fails

### Drill-Down View Pattern
- **State-Based Routing**: Use nil checks to determine which view to render (main vs drill-down)
- **Separate Keybindings**: Handle different keybindings based on current view state
- **Full-Screen Views**: Drill-down views replace the entire interface, not just a panel
- **Data Caching**: Preload data on startup for instant access when drilling down
- **View Isolation**: Each view maintains its own viewport and scroll state
- **Context-Aware Status Bar**: Status bar shows different actions based on current view

### Data Refresh Pattern
- **Incremental Sync**: Use API's `since` parameter with last sync timestamp to fetch only changed data
- **Progress State Management**: Separate model with Active/Complete states for clean TUI integration
- **Progress Overlay**: Full-screen centered overlay with progress bar and status messages
- **Completion Dismissal**: Any key dismisses completed refresh overlay for smooth UX
- **Deletion Handling**: List-based deletion removes items not in current remote dataset
- **State-Based Rendering**: Check refresh state before rendering main UI to show overlay when active

### Error Handling Pattern
- **Severity Classification**: Distinguish critical (modal) from minor (status bar) errors automatically
- **Error Pattern Matching**: Use string matching on error messages for classification flexibility
- **Modal Acknowledgment**: Critical errors require user acknowledgment before dismissal
- **Actionable Guidance**: Provide specific guidance for each error type (auth, repo, db, network)
- **Status Bar Integration**: Minor errors show in status bar without blocking UI
- **Safe Defaults**: Unknown errors default to critical for better user experience

### Relative Time Display Pattern
- **Tiered Granularity**: Use seconds, minutes, hours, days, weeks, months for human-readable time display
- **Zero Time Handling**: Always check IsZero() before formatting to show appropriate fallback ("never" vs timestamp)
- **Pluralization Logic**: Use helper function to handle singular/plural forms (e.g., "1 minute" vs "5 minutes")
- **Testable Time Functions**: Pass `now` parameter instead of calling time.Now() for deterministic tests
- **Metadata Display**: Use subtle colors (e.g., cyan) in status bar for metadata that doesn't require immediate attention
- **Defensive Time Calculations**: Handle edge cases like future times and integer division (28 days / 30 = 0 months)
### Help Overlay Pattern
- **Context Organization**: Group keybindings by view/context (main, comments, detail) for better discoverability
- **Toggle Activation**: Single key (?) both opens and closes help for simplicity
- **Multiple Dismissal Keys**: Support both ? and Esc for user convenience
- **Priority Order**: Check help state after error modal but before other overlays (refresh, drill-down views)
- **Consistent Modal Styling**: Use same border, centering, and styling as other overlays (error, refresh)
- **Status Bar Integration**: Include help key hint in all views for discoverability
- **Fixed Width Content**: Use fixed width for modal content to simplify centering calculations

---

## 2026-01-19 - US-001
- Implemented first-time setup with interactive prompts
- Created config package with TOML loading/saving
- Created setup package for interactive CLI prompts
- Implemented main CLI structure with flags
- Files changed:
  - `go.mod`, `go.sum` - Module initialization
  - `cmd/ghissues/main.go` - CLI entry point
  - `internal/config/config.go` - Configuration management
  - `internal/config/config_test.go` - Config tests (100% passing)
  - `internal/setup/setup.go` - Interactive setup prompts

**Learnings:**
- **BurntSushi/toml** is the standard Go TOML library - clean and well-maintained
- **go fmt** automatically formats code - ran it and it adjusted struct field alignment
- **flag package** is built-in and sufficient for CLI flags - no need for external libraries yet
- **bufio.Reader** is ideal for interactive prompts with ReadString('\n')
- **TOML struct tags** need to match exactly with field names for proper marshaling
- **os.UserHomeDir()** is the standard way to get home directory in Go 1.12+
- **File permissions 0600** are critical for security - verified in tests
- **Table-driven tests** make validation logic much more readable and maintainable

**Gotchas:**
- `go mod tidy` must be run after adding dependencies to update go.sum
- When using `t.TempDir()`, the directory is automatically cleaned up after the test
- The `flag` package requires flag definitions before `flag.Parse()` is called
- Need to handle both config file existence and loading errors separately for better UX

---

## 2026-01-20 - US-002
- Implemented GitHub authentication with multiple methods and priority order
- Created auth package with comprehensive test coverage
- Environment variable takes precedence over config file and gh CLI
- Files changed:
  - `internal/auth/auth.go` - Authentication logic with priority fallback
  - `internal/auth/auth_test.go` - Comprehensive tests (100% passing)
  - `go.mod`, `go.sum` - Added yaml.v3 dependency for gh CLI config parsing

**Learnings:**
- **gopkg.in/yaml.v3** is the standard Go YAML library - used for gh CLI's hosts.yml
- **Priority pattern** is simple and effective: try each method in order, return first success
- **Silent failures** are better for fallback - only error when all methods exhausted
- **Source tracking** helps debugging - return where auth came from (env/config/gh)
- **Test isolation** is critical - use `t.TempDir()` and save/restore env vars
- **strings.Contains()** is better than exact match for error message testing

**Gotchas:**
- When testing environment variables, always save and restore original values in `defer`
- gh CLI stores config in `~/.config/gh/hosts.yml` with yaml structure
- The `oauth_token` field in gh's hosts.yml is nested under `github.com:` key
- Empty tokens should be treated as "not found" for better UX
- YAML parsing can fail silently - check for errors but return false (not error) for fallback
- `go get` followed by `go mod tidy` ensures clean dependency management

---

## 2026-01-20 - US-004
- Implemented database storage location configuration with flexible path resolution
- Created database package with comprehensive test coverage (100% passing)
- Added --db flag to CLI for runtime database path override
- Implemented automatic parent directory creation with proper error handling
- Files changed:
  - `internal/database/database.go` - Database path resolution and validation
  - `internal/database/database_test.go` - Comprehensive tests (100% passing)
  - `cmd/ghissues/main.go` - Added --db flag and database path initialization

**Learnings:**
- **strings.TrimSpace()** is essential for config values - users might accidentally add spaces
- **os.MkdirAll()** creates all necessary parent directories in one call - very convenient
- **Test-driven writability checks** prevent runtime errors - create a test file, then clean it up
- **filepath.Dir()** returns "." for files in current directory - need special handling
- **Table-driven tests** work great for path resolution testing - covers many edge cases
- **Error wrapping with fmt.Errorf** and %w verb preserves error context for debugging

**Gotchas:**
- When checking if a path is "." (current directory), exact string match is needed
- Read-only directory tests may fail on some systems - use `t.Skip()` if creation fails
- Test file cleanup is important - use defer or immediate cleanup after writability check
- The config package already had DatabaseConfig struct - no changes needed there
- Integration with main.go requires importing the new package and calling functions in order
- Help text needs to be manually updated when adding new flags

---## ✓ Iteration 1 - US-001: First-Time Setup
*2026-01-20T04:59:22.962Z (228s)*

**Status:** Completed

**Notes:**
68→  - Gotchas encountered\n    69→---\n    70→```\n    71→\n    72→If you discovered a **reusable pattern**, also add it to the `## Codebase Patterns` section at the TOP of progress.md.\n    73→\n    74→## Stop Condition\n    75→**IMPORTANT**: If the work is already complete (implemented in a previous iteration or already exists), verify it meets the acceptance criteria and signal completion immediately.\n    76→\n    77→When finished (or if already complete), signal completion with:\n    78→

---
## ✓ Iteration 2 - US-002: GitHub Authentication
*2026-01-20T05:02:02.830Z (159s)*

**Status:** Completed

**Notes:**
mpty tokens, missing files, etc.)\n\n### Acceptance Criteria Met\n✅ Authentication attempts in order: env var → config → gh CLI  \n✅ Clear error message if no valid authentication found  \n✅ Token validation ready (will be validated on first API call in US-003)  \n✅ Config file token stored securely (0600 permissions from US-001)\n\n### Dependencies Added\n- `gopkg.in/yaml.v3` - For parsing gh CLI's hosts.yml configuration\n\n### Commit\nCreated commit: `feat: US-002 - GitHub Authentication`\n\n

---
## ✓ Iteration 3 - US-004: Database Storage Location
*2026-01-20T05:04:10.409Z (127s)*

**Status:** Completed

**Notes:**
cases covering flag precedence, config, and defaults\n- `TestEnsureDatabasePath` - 4 test cases for directory creation\n- `TestCheckDatabaseWritable` - 3 test cases for writability validation\n\n### Commit:\nCreated commit: `feat: US-004 - Database Storage Location`\n\n### New Pattern Added:\nAdded **Database Path Resolution Pattern** to the codebase patterns section in `progress.md` highlighting priority order, whitespace handling, directory creation, writability checks, and error context.\n\n

---

## ✓ Iteration 4 - US-003: Initial Issue Sync
*2026-01-20T05:30:00.000Z*

**Status:** Completed

**Files changed:**
- `internal/storage/storage.go` - Database layer with SQLite operations
- `internal/storage/storage_test.go` - Comprehensive storage tests (100% passing)
- `internal/github/github.go` - GitHub API client with pagination
- `internal/github/github_test.go` - API client tests
- `internal/sync/sync.go` - Sync engine with progress and cancellation
- `cmd/ghissues/main.go` - CLI integration with sync command
- `go.mod`, `go.sum` - Added modernc.org/sqlite dependency

**Learnings:**
- **modernc.org/sqlite** is a pure Go SQLite implementation - no CGo required, which simplifies cross-platform compilation
- **SQLite UPSERT** (INSERT ... ON CONFLICT DO UPDATE) is perfect for idempotent syncing - same issue/comment can be stored multiple times without duplicates
- **JSON time handling** requires careful parsing - store as RFC3339 strings, parse back to time.Time manually for type safety
- **GitHub Link headers** use a specific format: `<url>; rel="next"` - regex parsing is more reliable than string splitting
- **Context cancellation** is the Go way to handle Ctrl+C - use context.WithCancel and select statements for graceful shutdown
- **Signal handling** with os/signal requires buffer size of at least 1 to avoid missing signals
- **Progress bars** can be implemented with simple carriage returns (\r) and ANSI escape codes (\033[K) to clear lines
- **Worker pools** are ideal for parallel I/O-bound operations - use buffered channels and sync.WaitGroup for coordination
- **Database transactions** weren't needed for this simple case but should be considered for more complex sync scenarios
- **Separation of concerns** - keeping storage, API client, and sync logic in separate packages makes testing much easier

**Gotchas:**
- SQLite stores timestamps as strings by default - must parse back to time.Time manually when reading
- http.ResponseWriter's json.NewEncoder doesn't automatically flush the response in test servers
- Test server pagination requires careful URL construction - the Link header URL must match the actual test server URL
- Channel directions in Go (chan<- vs <-chan) help prevent sending on receiving-only channels
- goroutine leaks can happen if channels aren't properly closed when using context cancellation
- The sync.WaitGroup must be waited upon even when operations are cancelled early
- Progress display must handle both progress updates and periodic ticker updates for smooth animation

**Patterns Discovered:**
- **Progress Channel Pattern**: Unidirectional channel for progress updates with goroutine consumer for display
- **Graceful Cancellation Pattern**: context cancellation + signal.Notify + select statements for clean shutdown
- **Worker Pool Pattern**: Fixed number of workers consuming from a buffered channel for parallel processing
- **Error Collection Pattern**: Buffered error channel to collect errors from workers without blocking

**Commit:**
Created commit: `feat: US-003 - Initial Issue Sync`

---
## ✓ Iteration 4 - US-003: Initial Issue Sync
*2026-01-20T05:15:25.435Z (674s)*

**Status:** Completed

---

## 2026-01-20 - US-005
- Implemented TUI-based issue list view with bubbletea framework
- Created comprehensive column configuration system with 9 available column types
- Implemented vim-style navigation (j/k and arrow keys)
- Added viewport scrolling for large issue lists
- Files changed:
  - `internal/tui/columns.go` - Column configuration and validation system
  - `internal/tui/list.go` - Issue list model with cursor and viewport management
  - `internal/tui/app.go` - Bubbletea TUI application with rendering
  - `internal/tui/columns_test.go` - Column configuration tests (100% passing)
  - `internal/tui/list_test.go` - Issue list model tests (100% passing)
  - `internal/tui/integration_test.go` - Integration tests (100% passing)
  - `cmd/ghissues/main.go` - Integrated TUI into main CLI
  - `go.mod`, `go.sum` - Added bubbletea and lipgloss dependencies

**Learnings:**
- **charmbracelet/bubbletea** is an elegant TUI framework - uses Elm architecture (Model-Update-View)
- **lipgloss** provides excellent styling capabilities for terminal UI - colors, bold, borders, spacing
- **Elm Architecture** (Model-Update-View) makes TUI logic predictable and testable
- **Viewport management** is critical for large lists - need offset tracking and scroll calculations
- **Cursor highlighting** with background colors creates clear visual feedback for user
- **Column metadata maps** are better than switch statements for extensible column systems
- **Flexible width columns** require calculation of remaining space after fixed columns
- **tea.WithAltScreen()** creates a full-screen TUI experience - cleans up on exit
- **tea.WindowSizeMsg** is the proper way to handle terminal resize events
- **Table-driven tests** work great for column validation - test many scenarios in one test function

**Gotchas:**
- Bubbletea's Update method must return (tea.Model, tea.Cmd) - easy to forget the Cmd
- Window dimensions start at 0,0 - need to handle initial render before sizing
- Text truncation for flexible columns must account for fixed widths + separators
- Viewport scrolling should keep cursor in visible range but avoid scrolling too early
- Integer to string conversion requires careful implementation - can't just cast int to rune
- lipgloss styles are immutable - chain method calls to build complex styles
- Column value extraction must handle all column types - use switch statement for clarity
- Date formatting with time.Time requires format strings - "Jan 2" is good for compact display
- Status bar should include navigation help - users need to know what keys to press
- Exit on 'q' and Ctrl+C is expected terminal behavior - implement both for better UX

**Patterns Discovered:**
- **TUI Model Pattern**: Separate state (Model) from rendering (View) and updates (Update)
- **Column Configuration Pattern**: Metadata maps + validation + defaults for extensible column systems
- **Viewport Pattern**: Track offset + cursor position; update offset when cursor approaches edges
- **Navigation Pattern**: Vim keys (j/k) + arrow keys + Enter/Space for selection
- **Status Bar Pattern**: Always show item count + navigation help + selection state

**Commit:**
Created commit: `feat: US-005 - Issue List View`

---

## ✓ Iteration 5 - US-005: Issue List View
*2026-01-20T05:26:03.731Z (637s)*

**Status:** Completed

**Notes:**
# 📝 Patterns Discovered\n\nAdded **TUI Pattern (Bubbletea/Elm Architecture)** to the codebase patterns section, documenting:\n- Model-Update-View architecture\n- Viewport management\n- Column configuration best practices\n- Navigation patterns\n- Status bar design\n\n### ✅ Quality Checks Passed\n\n- All TUI tests passing (100%)\n- Core package tests passing (auth, config, database)\n- `go vet` passes with no warnings\n- Application builds successfully\n- Commit created with detailed message\n\n

---

## 2026-01-20 - US-006
- Implemented issue sorting with keyboard controls and persistent configuration
- Created comprehensive sort package with field validation and cycling
- Added sort keybindings (s to cycle field, S to toggle order)
- Enhanced status bar to display current sort with visual indicator (▼/▲)
- Files changed:
  - `internal/sort/sort.go` - Sort logic, field validation, and cycling
  - `internal/sort/sort_test.go` - Comprehensive sort tests (100% passing)
  - `internal/tui/list.go` - Added sort state and methods (UnsortedIssues, SortField, SortDescending)
  - `internal/tui/list_test.go` - Tests for sort functionality
  - `internal/tui/app.go` - Sort keybindings and status display
  - `cmd/ghissues/main.go` - Config integration for sort preferences

**Learnings:**
- **Standard library sort.Slice** is perfect for custom sorting - just provide a less function
- **Preserve original data** when sorting - keep unsorted copy to allow re-sorting with different criteria
- **Sort field cycling** is intuitive - use array to define cycle order, modulo for wraparound
- **Unicode arrows** (▼/▲) are great for visual indicators in terminal UI - clear and compact
- **Sort state in model** makes TUI responsive - re-sort on keypress without re-fetching data
- **Config file integration** is expected behavior - users want their preferences persisted
- **Table-driven tests** work exceptionally well for sorting - test multiple fields and orders

**Gotchas:**
- **Cursor reset** after sorting - if cursor is out of bounds, reset to 0 to avoid panics
- **Sort field validation** should fallback to defaults - don't fail if config has invalid field
- **Test data needs variety** - use different timestamps, numbers, and comment counts to verify sorting
- **Existing tests may assume order** - tests written before sorting may fail when default sort is applied
- **Copy slices before sorting** - sort.Slice modifies in place, use copy() to preserve original
- **Case sensitivity matters** - sort field names should be lowercase for consistency
- **Descending order** is natural for dates (most recent first) but ascending for numbers (1, 2, 3)

**Patterns Discovered:**
- **Sort Preservation Pattern**: Keep original unsorted data separate from sorted display data
- **Sort Field Cycle Pattern**: Use array + modulo operator for seamless field rotation
- **Sort Validation Pattern**: Map-based validation with fallback to sensible defaults
- **Sort State Pattern**: Store sort field and order in model for re-sorting without re-fetching

**Commit:**
Created commit: `feat: US-006 - Issue Sorting`

---
## ✓ Iteration 6 - US-006: Issue Sorting
*2026-01-20T05:34:45.414Z (521s)*

**Status:** Completed

---

## 2026-01-20 - US-007
- Implemented issue detail view with split-panel layout
- Created comprehensive detail panel model with viewport and markdown rendering toggle
- Integrated charmbracelet/glamour for beautiful markdown rendering
- Added automatic panel updates as user navigates issue list
- Files changed:
  - `internal/tui/detail.go` - Detail panel model with viewport management and rendering
  - `internal/tui/detail_test.go` - Comprehensive detail panel tests (100% passing)
  - `internal/tui/app.go` - Split layout integration with 60/40 width allocation
  - `go.mod`, `go.sum` - Added glamour and related dependencies

**Learnings:**
- **charmbracelet/glamour** is excellent for terminal markdown rendering - auto-detects terminal style, handles word wrap
- **Split layout rendering** requires careful width management - use percentage-based allocation for responsiveness
- **lipgloss.JoinHorizontal** is perfect for combining multiple panels with separators - handles alignment automatically
- **Automatic panel updates** provide better UX than requiring Enter press - detail panel updates as you navigate
- **Viewport isolation** is critical - each panel manages its own scroll state independently
- **Label badges** with background colors and padding provide visual distinction - use color codes like "62" for purple
- **Fallback rendering** is important - glamour can fail, gracefully fall back to raw markdown
- **Nil safety** in TUI models - always check if detail panel is nil before accessing methods
- **Status bar context** - show different hints based on current state (e.g., markdown mode when detail panel is visible)
- **Empty state handling** - display helpful messages when no issue is selected or list is empty

**Gotchas:**
- **Glamour renderer creation** is expensive - create once and reuse, not on every render
- **Width calculations** must account for separators - subtract separator width when allocating panel widths
- **Scrollbar positioning** needs content height awareness - implement scroll limits based on actual content length
- **Test timing** - glamour rendering can be slow in tests, mock or avoid in unit tests when possible
- **String splitting** for labels - handle empty strings and whitespace with TrimSpace
- **Model pointer receivers** - use pointer receivers for methods that modify model state (like updateDetailPanel)
- **Height calculations** must reserve space for header and status bar - subtract from total window height
- **lipgloss styles** are immutable - chain method calls to build complex styles
- **Color codes** - lipgloss uses standard terminal color codes, experiment for best appearance
- **Viewport height** should be set based on available space, not fixed values

**Patterns Discovered:**
- **Split Panel Pattern**: Use percentage-based width allocation with lipgloss.JoinHorizontal for responsive layouts
- **Auto-Update Pattern**: Update detail panels on cursor movement for immediate feedback
- **Viewport Isolation Pattern**: Each panel maintains independent scroll state and viewport dimensions
- **Conditional Rendering Pattern**: Check for nil panels and empty states before rendering to avoid panics
- **Fallback Rendering Pattern**: Gracefully degrade from rich rendering (glamour) to raw text on errors

**Commit:**
Created commit: `feat: US-007 - Issue Detail View`

---
## ✓ Iteration 7 - US-007: Issue Detail View
*2026-01-20T05:44:32.574Z (586s)*

**Status:** Completed

**Notes:**
- ✅ `go vet` passes with no warnings\n- ✅ Application builds successfully\n- ✅ Commit created with detailed message\n\n### 📝 Patterns Discovered\n\nAdded **Detail Panel Pattern** to the codebase patterns, documenting:\n- Split Layout Management (percentage-based allocation)\n- Automatic Panel Updates (cursor-driven updates)\n- Viewport Isolation (independent scroll state)\n- Lazy Rendering (handle empty states)\n- Label Rendering (colored badges)\n- Fallback Rendering (graceful degradation)\n\n

---

## 2026-01-20 - US-008
- Implemented drill-down comments view for viewing issue discussions
- Created comprehensive CommentsView model with viewport and markdown rendering
- Added state-based routing to switch between main and comments views
- Integrated glamour for beautiful markdown rendering in comments
- Preloaded comments from database on startup for instant access
- Files changed:
  - `internal/tui/comments.go` - Comments view model with viewport management and rendering
  - `internal/tui/comments_test.go` - Comprehensive tests (100% passing)
  - `internal/tui/app.go` - Main app integration with state-based routing and separate keybindings
  - `cmd/ghissues/main.go` - Comment preloading from database

**Learnings:**
- **State-based view routing** is elegant for drill-down views - use nil checks to determine which view to render
- **Separate keybinding handlers** for different views keeps code organized - have updateCommentsView method
- **Full-screen drill-down views** replace the entire interface - more immersive than panel-based views
- **Data caching on startup** provides instant access - preload all comments into map for O(1) lookup
- **Context-aware status bars** improve UX - show different hints based on current view (main vs comments)
- **View isolation** is critical - each view maintains its own viewport and scroll state
- **glamour rendering** works well for comments - consistent with detail panel implementation
- **Enter vs Space distinction** - Enter opens drill-down, Space selects item in current view

**Gotchas:**
- **View state management** - need to track both current view and data for all views (AllComments cache)
- **Keybinding priority** - check if in drill-down view first, then handle main view keybindings
- **Viewport synchronization** - update viewport heights in WindowSizeMsg for all views
- **Nil safety** - always check if CommentsView is nil before accessing methods
- **Data loading strategy** - preload on startup vs lazy load on demand; preloading is simpler for this use case
- **Exit key consistency** - use both q and Esc for returning from drill-down (expected behavior)
- **Status bar context** - show comment count and markdown mode in comments view
- **Empty state handling** - display "No comments on this issue" when comment list is empty
- **Map-based caching** - use issue number as key for O(1) comment lookup
- **Test isolation** - CommentsView tests don't need to know about main app state

**Patterns Discovered:**
- **Drill-Down View Pattern**: State-based routing with nil checks, separate keybindings, full-screen replacement
- **Data Caching Pattern**: Preload related data on startup into map for instant access
- **View Routing Pattern**: Check view state (nil vs populated) to determine rendering logic
- **Context-Aware UI Pattern**: Status bar and keybindings change based on current view context

**Commit:**
Created commit: `feat: US-008 - Comments View`

---
## ✓ Iteration 8 - US-008: Comments View
*2026-01-20T05:53:15.979Z (522s)*

**Status:** Completed

**Notes:**
t Coverage**: All tests passing with table-driven test patterns\n\n### 📝 Patterns Discovered\n\nAdded **Drill-Down View Pattern** to codebase patterns:\n- State-based routing with nil checks\n- Separate keybindings per view\n- Full-screen view replacement\n- Data caching for instant access\n- Context-aware status bars\n\n### ✅ Quality Checks Passed\n\n- All TUI tests passing (100%)\n- `go vet` passes with no warnings\n- Application builds successfully\n- Commit created with detailed message\n\n---

## 2026-01-20 - US-009
- Implemented incremental sync with `FetchIssuesSince` method in GitHub client
- Created refresh progress model with visual progress bar and status messages
- Added TUI integration for refresh with r/R keybindings
- Implemented deleted issues cleanup with `DeleteIssuesNotInList` in storage package
- Added incremental sync support in sync package with `RunIncremental` method
- Files changed:
  - `internal/github/github.go` - Added FetchIssuesSince method with since parameter support
  - `internal/github/github_test.go` - Tests for incremental fetch (100% passing)
  - `internal/storage/storage.go` - Added DeleteIssuesNotInList and DeleteCommentsForIssue functions
  - `internal/storage/storage_test.go` - Tests for deletion operations (100% passing)
  - `internal/sync/sync.go` - Added RunIncremental method for incremental sync
  - `internal/tui/refresh.go` - New refresh progress model with progress bar
  - `internal/tui/refresh_test.go` - Comprehensive refresh model tests (100% passing)
  - `internal/tui/app.go` - Integrated refresh with r/R keybindings and overlay rendering

**Learnings:**
- **GitHub API `since` parameter** is perfect for incremental sync - filters by updated_at timestamp
- **SQLite CASCADE deletion** requires foreign key pragma to be enabled - better to handle cleanup explicitly
- **Bubbletea tick commands** are great for simulating async operations in tests
- **Progress overlays** in TUI should be centered and show clear completion status
- **State-based rendering** makes it easy to show different views (normal vs refresh overlay)
- **Refresh completion dismissal** with any key provides good UX flow
- **Incremental sync limitations** - GitHub API returns updated issues but doesn't indicate deleted issues, requiring careful handling
- **Separate refresh models** keep the main TUI clean - RefreshModel manages its own state independently

**Gotchas:**
- **Test server responses** need proper Content-Type headers or json.Decoder may fail
- **json.NewEncoder vs json.Marshal + w.Write** - Encoder is simpler but both work for test servers
- **Foreign key constraints** in SQLite are not enforced by default - need PRAGMA foreign_keys=ON
- **Refresh state management** - need to handle both Active and Complete states for proper UX flow
- **Message type assertions** in Bubbletea need careful handling - RefreshModel.Update returns RefreshModel not tea.Model
- **Centering calculations** must account for varying line lengths in refresh progress display
- **Dismissal on any key** needs to check Complete state before handling other keybindings
- **Progress bar rendering** needs to handle terminal width constraints gracefully

**Patterns Discovered:**
- **Incremental Sync Pattern**: Use since parameter with last sync timestamp to fetch only changes
- **Progress Overlay Pattern**: Full-screen overlay with centered content and clear completion status
- **State-Based Refresh Pattern**: Separate RefreshModel with Active/Complete states for clean TUI integration
- **Deletion Cleanup Pattern**: Explicit list-based deletion to handle remote state changes

**Commit:**
Created commit: `feat: US-009 - Data Refresh`

---
## ✓ Iteration 9 - US-009: Data Refresh
*2026-01-20T06:15:30.376Z (1333s)*

**Status:** Completed

**Notes:**
TUI","status":"completed","activeForm":"Adding r/R keybinding to trigger refresh in TUI"},{"content":"Integrate auto-refresh on app launch","status":"completed","activeForm":"Integrating auto-refresh on app launch"},{"content":"Run quality checks (go test, go vet, go build)","status":"completed","activeForm":"Running quality checks"}]}}
{"type":"assistant","message":{"id":"msg_20260120141505ad71bada25f64e29","type":"message","role":"assistant","model":"glm-4.7","content":[{"type":"text","text":"

---

## 2026-01-20 - US-013
- Implemented comprehensive error handling system with modal and status bar display
- Created error model with automatic severity classification
- Added actionable guidance for common error types
- Integrated error handling into TUI with proper state management
- Files changed:
  - `internal/tui/error.go` - Error model with classification, modal rendering, and actionable guidance
  - `internal/tui/error_test.go` - Comprehensive error handling tests (100% passing)
  - `internal/tui/app.go` - Error state integration with modal overlay and status bar display

**Learnings:**
- **Error severity classification** is essential for good UX - not all errors need to block the UI
- **String-based error matching** is flexible but requires careful pattern design for false positives
- **Modal acknowledgment pattern** ensures users see critical errors before dismissing
- **Actionable guidance** transforms errors from frustrating to helpful - tell users what to do
- **lipgloss border styling** with rounded borders and colors creates clear visual hierarchy
- **Error state priority** matters - check error modal state before other keybindings
- **Status bar integration** keeps minor errors visible without blocking workflows
- **Centering calculations** must account for border padding and content width

**Gotchas:**
- **Error message matching** should use lowercase for case-insensitive comparison
- **Critical vs minor distinction** requires careful consideration of user impact
- **Modal dismissal logic** needs two-stage acknowledgment (show complete, then dismiss on keypress)
- **Guidance text length** affects modal width - keep it concise but helpful
- **Test data isolation** is important - error tests need varied error message patterns
- **Border padding** in lipgloss affects centering calculations - account for it
- **Error model reuse** - single error model can handle both modal and status bar display
- **Status bar space** is limited - truncate long error messages or show only key info

**Patterns Discovered:**
- **Error Severity Pattern**: Classify errors as critical (modal) or minor (status bar) based on message patterns
- **Modal Acknowledgment Pattern**: Two-stage dismissal (mark complete, then dismiss) ensures user awareness
- **Actionable Guidance Pattern**: Provide specific, actionable steps for each error category
- **Error State Priority Pattern**: Check error modal state first in Update to block other interactions
- **Status Bar Error Pattern**: Show minor errors in status bar without blocking UI workflow

**Commit:**
Created commit: `feat: US-013 - Error Handling`

---
## ✓ Iteration 10 - US-013: Error Handling
*2026-01-20T06:29:22.511Z (831s)*

**Status:** Completed

**Notes:**
y Checks:\n- ✅ All TUI tests passing (100%)\n- ✅ Core package tests passing (auth, config, database, sort)\n- ✅ `go vet` passes with no warnings\n- ✅ Application builds successfully\n- ✅ Commit created with detailed message\n\n### Patterns Discovered:\nAdded **Error Handling Pattern** to codebase patterns, documenting:\n- Error severity classification (critical vs minor)\n- Modal acknowledgment pattern\n- Actionable guidance pattern\n- Error state priority pattern\n- Status bar error pattern\n\n

---

## 2026-01-20 - US-010
- Implemented last synced indicator in TUI status bar with relative time display
- Created timefmt package with FormatRelative function for human-readable time formatting
- Added LastSync field to TUI Model to track when data was last updated
- Integrated last sync time loading from database metadata on app startup
- Updated last sync timestamp on successful refresh completion
- Files changed:
  - `internal/timefmt/timefmt.go` - Relative time formatting utility with pluralization
  - `internal/timefmt/timefmt_test.go` - Comprehensive tests covering all time ranges (100% passing)
  - `internal/tui/app.go` - Added LastSync field and status bar display with cyan color
  - `internal/tui/integration_test.go` - Updated tests to include last sync parameter
  - `cmd/ghissues/main.go` - Load last sync from database on startup

**Learnings:**
- **Relative time formatting** improves UX significantly - users understand "5 minutes ago" better than timestamps
- **Zero time handling** is important - show "never" instead of confusing users with epoch time
- **Time granularity tiers** work well - seconds, minutes, hours, days, weeks, months provide good coverage
- **Pluralization logic** needs careful handling - singular vs plural forms matter for readability
- **Status bar space** is limited - keep labels concise but informative
- **Color coding** helps visual hierarchy - cyan for metadata distinguishes it from other status elements
- **Table-driven tests** are perfect for time formatting - cover many time ranges in one test function
- **Pre-existing infrastructure** - UpdateLastSync/GetLastSync were already implemented from US-009
- **TDD approach works smoothly** - write failing tests, implement, watch them pass
- **Time calculations** should use now parameter for testability - don't call time.Now() directly

**Gotchas:**
- **Month calculation** can be tricky - 28 days / 30 = 0, need to check days < 30 before dividing
- **Test timing** - time.Now() calls in tests need to be close together or use fixed times
- **Import order** - adding new imports requires updating existing code that uses the package
- **Function signature changes** ripple through codebase - update all call sites when adding parameters
- **Time zone handling** - use RFC3339 for storage, but comparisons are timezone-agnostic with time.Since
- **Zero time checks** - always use IsZero() method, not comparison to time.Time{}
- **Integer division** - 28 days / 30 days = 0 months, handle this explicitly before division
- **Status bar overflow** - too much text gets truncated, prioritize essential information
- **Color consistency** - use consistent colors for similar information types
- **Future time handling** - be defensive, handle negative durations gracefully even if unexpected

**Patterns Discovered:**
- **Relative Time Pattern**: Use tiered granularity (seconds → minutes → hours → days → weeks → months) for human-readable time display
- **Zero Time Pattern**: Always check IsZero() before formatting time values to show appropriate fallback ("never" vs timestamp)
- **Pluralization Pattern**: Use helper function to handle singular/plural forms for better readability
- **Metadata Display Pattern**: Use subtle colors in status bar for metadata that doesn't require immediate attention
- **Testable Time Functions**: Pass `now` parameter instead of calling time.Now() for deterministic tests

**Commit:**
Created commit: `feat: US-010 - Last Synced Indicator`

---
## ✓ Iteration 11 - US-010: Last Synced Indicator
*2026-01-20T06:36:02.292Z (399s)*

**Status:** Completed

**Notes:**
aw timestamps\n- **Tiered granularity** provides good coverage from seconds to months\n- **Zero time handling** is critical - show \"never\" not confusing epoch times\n- **Testability** - pass `now` parameter instead of calling `time.Now()` directly\n- **Month calculation edge case** - check days < 30 before integer division\n\n### ✅ Quality Checks Passed\n- All tests passing (100%)\n- `go vet` passes with no warnings\n- Application builds successfully\n- Commit created with detailed message\n\n

---

## 2026-01-20 - US-011
- Implemented comprehensive help overlay with keybinding documentation
- Created HelpModel with toggle, update, and view methods
- Organized keybindings by context (Main View, Comments View)
- Added help toggle with ? key from any view
- Implemented dismissible help overlay with ? and Esc keys
- Updated status bars to include ?: help hint
- Files changed:
  - `internal/tui/help.go` - Help model with keybinding documentation and styled modal
  - `internal/tui/help_test.go` - Comprehensive tests (11 tests, 100% passing)
  - `internal/tui/app.go` - Integration of help overlay and status bar updates

**Learnings:**
- **Help overlay priority** should be after error modal but before other overlays - errors need immediate attention
- **Context-sensitive help** improves discoverability - show keybindings for current view
- **Modal styling consistency** is important - use same border and centering pattern as error/refresh overlays
- **Toggle pattern** works well for help - single key (?) both opens and closes help
- **Status bar space** is limited - prioritize most common actions and reference to help for full list
- **lipgloss width constraints** help with centering calculations - use fixed width for modal content
- **Table-driven tests** are perfect for verifying UI content - check all expected keybindings are present
- **View method testing** requires understanding of lipgloss rendering - test structure, not exact formatting

**Gotchas:**
- **Help overlay state** needs to be checked in Update before main view keybindings - blocks other interactions when active
- **Window size updates** must propagate to all overlays - help, error, refresh all need width/height
- **Key name capitalization** matters for tests - "Enter" vs "enter" in help text
- **Border padding** affects centering calculations - lipgloss padding adds space outside content
- **Status bar truncation** happens automatically with lipgloss - long text gets cut off, keep hints concise
- **Help in drill-down views** should work the same as main view - check for help active before comments view rendering
- **Multiple dismissal keys** improve UX - both ? and Esc dismiss help for user convenience
- **Section organization** improves readability - group keybindings by context (main vs comments)

**Patterns Discovered:**
- **Help Overlay Pattern**: Full-screen centered modal with organized keybinding sections by context
- **Persistent Footer Pattern**: Context-sensitive status bar shows common keys with help reference
- **Toggle Key Pattern**: Single key (?) both opens and closes overlay for simplicity

## ✓ Iteration 12 - US-011: Keybinding Help
*2026-01-20T06:45:27.241Z (564s)*

**Status:** Completed

**Notes:**
Footer updates based on current view (list, detail, comments)\n   - Main view footer (app.go:407-458)\n   - Comments view footer (app.go:475-489)\n\n✅ Help overlay dismissible with ? or Esc\n   - ? dismissal in help.go:67\n   - Esc dismissal in help.go:67\n\nAll acceptance criteria met!","stderr":"","interrupted":false,"isImage":false}}
{"type":"assistant","message":{"id":"msg_20260120144520fbe236885cd741c6","type":"message","role":"assistant","model":"glm-4.7","content":[{"type":"text","text":"

---
